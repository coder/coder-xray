package reporter

import (
	"context"
	"fmt"

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
	JFrogClient   *jfrog.Client

	// Unexported fields are initialized on calls to Init.
	podInformer cache.SharedIndexInformer
}

type WorkspaceAgent struct {
	Image string
	Token string
}

func (k *K8sReporter) Init(ctx context.Context) error {
	podFactory := informers.NewSharedInformerFactoryWithOptions(k.Client, 0, informers.WithNamespace(k.Namespace), informers.WithTweakListOptions(func(lo *v1.ListOptions) {
		lo.FieldSelector = k.FieldSelector
		lo.LabelSelector = k.LabelSelector
	}))

	k.podInformer = podFactory.Core().V1().Pods().Informer()

	_, err := k.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				k.Logger.Error(ctx, "unexpected object type", slog.F("type", fmt.Sprintf("%T", obj)))
				return
			}

			log := k.Logger.With(
				slog.F("pod_name", pod.Name),
			)
			var isWorkspace bool
			for _, container := range pod.Spec.Containers {
				var agentToken string
				for _, env := range container.Env {
					if env.Name != "CODER_AGENT_TOKEN" {
						continue
					}
					isWorkspace = true
					agentToken = env.Value
					break
				}
				if agentToken == "" {
					continue
				}

				log = log.With(
					slog.F("container_name", container.Name),
					slog.F("container_image", container.Image),
				)

				image, err := jfrog.ParseImage(container.Image)
				if err != nil {
					log.Error(ctx, "parse image", slog.Error(err))
					return
				}

				scan, err := k.JFrogClient.ScanResults(image)
				if err != nil {
					log.Error(ctx, "fetch scan results", slog.Error(err))
					return
				}

				manifest, err := k.CoderClient.AgentManifest(ctx, agentToken)
				if err != nil {
					log.Error(ctx, "Get agent manifest", slog.Error(err))
					return
				}

				log = log.With(
					slog.F("workspace_id", manifest.WorkspaceID),
					slog.F("agent_id", manifest.AgentID),
					slog.F("workspace_name", manifest.WorkspaceName),
				)

				err = k.CoderClient.PostJFrogXrayScan(ctx, codersdk.JFrogXrayScan{
					WorkspaceID: manifest.WorkspaceID,
					AgentID:     manifest.AgentID,
					Critical:    scan.SecurityIssues.Critical,
					High:        scan.SecurityIssues.High,
				})
				if err != nil {
					log.Error(ctx, "post xray results", slog.Error(err))
					return
				}
			}
			if isWorkspace {
				log.Info(ctx, "uploaded workspace results!", slog.F("pod_name", pod.Name), slog.F("namespace", pod.Namespace))
			}
		},
	})
	if err != nil {
		return fmt.Errorf("register pod handler: %w", err)
	}
	return nil
}

func (k *K8sReporter) Start(stop chan struct{}) {
	k.podInformer.Run(stop)
}
