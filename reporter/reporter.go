package reporter

import (
	"context"
	"fmt"

	"golang.org/x/xerrors"

	"github.com/google/uuid"

	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/xray/jfrog"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"cdr.dev/slog"
)

type K8sReporter struct {
	Client        kubernetes.Interface
	LabelSelector string
	FieldSelector string
	Namespace     string
	CoderClient   CoderClient
	Logger        slog.Logger
	JFrogClient   jfrog.Client
	ResultsChan   chan codersdk.JFrogXrayScan

	// Unexported fields are initialized on calls to Init.
	factory informers.SharedInformerFactory
}

func (k *K8sReporter) Init(ctx context.Context) error {
	k.factory = informers.NewSharedInformerFactoryWithOptions(k.Client, 0, informers.WithNamespace(k.Namespace), informers.WithTweakListOptions(func(lo *v1.ListOptions) {
		lo.FieldSelector = k.FieldSelector
		lo.LabelSelector = k.LabelSelector
	}))

	podInformer := k.factory.Core().V1().Pods().Informer()

	_, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				k.Logger.Error(ctx, "unexpected object type", slog.F("type", fmt.Sprintf("%T", obj)))
				return
			}

			log := k.Logger.With(
				slog.F("pod_name", pod.Name),
			)
			for _, container := range pod.Spec.Containers {
				log = log.With(
					slog.F("container_name", container.Name),
					slog.F("container_image", container.Image),
				)

				scan, err := func() (codersdk.JFrogXrayScan, error) {
					var agentToken string
					for _, env := range container.Env {
						if env.Name != "CODER_AGENT_TOKEN" {
							continue
						}
						agentToken = env.Value
						break
					}
					if agentToken == "" {
						return codersdk.JFrogXrayScan{}, nil
					}

					image, err := jfrog.ParseImage(container.Image)
					if err != nil {
						return codersdk.JFrogXrayScan{}, xerrors.Errorf("parse image: %w", err)
					}

					scan, err := k.JFrogClient.ScanResults(image)
					if err != nil {
						return codersdk.JFrogXrayScan{}, xerrors.Errorf("fetch scan results: %w", err)
					}

					manifest, err := k.CoderClient.AgentManifest(ctx, agentToken)
					if err != nil {
						return codersdk.JFrogXrayScan{}, xerrors.Errorf("agent manifest: %w", err)
					}

					log = log.With(
						slog.F("workspace_id", manifest.WorkspaceID),
						slog.F("agent_id", manifest.AgentID),
						slog.F("workspace_name", manifest.WorkspaceName),
					)

					req := codersdk.JFrogXrayScan{
						WorkspaceID: manifest.WorkspaceID,
						AgentID:     manifest.AgentID,
						Critical:    scan.SecurityIssues.Critical,
						High:        scan.SecurityIssues.High,
						Medium:      scan.SecurityIssues.Medium,
					}
					err = k.CoderClient.PostJFrogXrayScan(ctx, req)
					if err != nil {
						return codersdk.JFrogXrayScan{}, xerrors.Errorf("post xray scan: %w", err)
					}

					return req, nil
				}()
				if err != nil {
					log.Error(ctx, "scan agent", slog.Error(err))
					break
				}
				if scan.AgentID != uuid.Nil {
					log.Info(ctx, "uploaded agent results!", slog.F("pod_name", pod.Name), slog.F("namespace", pod.Namespace))
					if k.ResultsChan != nil {
						// This should only be populated during tests
						// so it's ok to assume an unbuffered channel is
						// going to block until read.
						k.ResultsChan <- scan
					}
				}
			}
		},
	})
	if err != nil {
		return fmt.Errorf("register pod handler: %w", err)
	}
	return nil
}

func (k *K8sReporter) Start(stop chan struct{}) {
	k.factory.Start(stop)
}
