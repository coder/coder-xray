package reporter

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coder/coder/v2/codersdk"
	"github.com/coder/coder/v2/codersdk/agentsdk"
	"github.com/coder/xray/jfrog"

	"cdr.dev/slog"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type K8sReporter struct {
	Client        kubernetes.Interface
	LabelSelector string
	FieldSelector string
	Namespace     string
	CoderURL      *url.URL
	Logger        slog.Logger
	CoderToken    string
	JFrogClient   *jfrog.Client

	ctx         context.Context
	podInformer cache.SharedIndexInformer
}

type WorkspaceAgent struct {
	Image string
	Token string
}

func (k *K8sReporter) Init(ctx context.Context) error {
	k.ctx = ctx

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

				image, err := jfrog.ParseImage(container.Image)
				if err != nil {
					k.Logger.Error(ctx, "parse image",
						slog.F("pod_name", pod.Name),
						slog.F("container_name", container.Name),
						slog.F("container_image", container.Image),
						slog.Error(err),
					)
					return
				}

				scan, err := k.JFrogClient.ScanResults(image)
				if err != nil {
					k.Logger.Error(ctx, "fetch scan results", slog.Error(err))
					return
				}

				agentClient := agentsdk.New(k.CoderURL)
				agentClient.SetSessionToken(agentToken)
				manifest, err := agentClient.Manifest(ctx)
				if err != nil {
					k.Logger.Error(ctx, "Get agent manifest", slog.Error(err))
					return
				}

				cclient := codersdk.New(k.CoderURL)
				cclient.SetSessionToken(k.CoderToken)
				err = cclient.PostJFrogXrayScan(ctx, codersdk.JFrogXrayScan{
					WorkspaceID: manifest.WorkspaceID,
					AgentID:     manifest.AgentID,
					Critical:    scan.SecurityIssues.Critical,
					High:        scan.SecurityIssues.High,
				})
				if err != nil {
					k.Logger.Error(ctx, "post xray results", slog.Error(err))
					return
				}
			}
			if isWorkspace {
				k.Logger.Info(ctx, "uploaded workspace results!", slog.F("name", pod.Name), slog.F("namespace", pod.Namespace))
			}
		},
	})
	if err != nil {
		return fmt.Errorf("register pod handler: %w", err)
	}
	return nil
}
