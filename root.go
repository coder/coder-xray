package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"cdr.dev/slog"
	"cdr.dev/slog/sloggers/sloghuman"

	"github.com/coder/coder-xray/jfrog"
	"github.com/coder/coder-xray/reporter"
)

func root() *cobra.Command {
	var (
		coderURL         string
		coderToken       string
		artifactoryURL   string
		artifactoryUser  string
		artifactoryToken string
		fieldSelector    string
		kubeConfig       string
		namespace        string
		labelSelector    string
	)
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan Coder Kubernetes workspace images for vulnerabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			if coderURL == "" {
				return xerrors.New("--coder-url is required")
			}

			coderParsed, err := url.Parse(coderURL)
			if err != nil {
				return fmt.Errorf("parse coder URL: %w", err)
			}

			if artifactoryURL == "" {
				return xerrors.New("--artifactory-url is required")
			}

			_, err = url.Parse(artifactoryURL)
			if err != nil {
				return fmt.Errorf("parse artifactory URL: %w", err)
			}

			if artifactoryUser == "" {
				return xerrors.New("--artifactory-user is required")
			}

			if artifactoryToken == "" {
				return xerrors.New("--artifactory-token is required")
			}

			config, err := restclient.InClusterConfig()
			if xerrors.Is(err, restclient.ErrNotInCluster) {
				config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
			}
			if err != nil {
				return xerrors.Errorf("build kubeconfig: %w", err)
			}

			kclient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return xerrors.Errorf("create kubernetes config: %w", err)
			}

			jclient, err := jfrog.XRayClient(artifactoryURL, artifactoryUser, artifactoryToken)
			if err != nil {
				return xerrors.Errorf("create artifactory client: %w", err)
			}

			coderClient := reporter.NewClient(coderParsed, coderToken)
			logger := slog.Make(sloghuman.Sink(cmd.ErrOrStderr()))
			kr := reporter.K8sReporter{
				Client:      kclient,
				JFrogClient: jclient,
				CoderClient: coderClient,
				Namespace:   namespace,
				Logger:      logger,
			}

			err = kr.Init(cmd.Context())
			if err != nil {
				return xerrors.Errorf("initialize reporter: %w", err)
			}

			stopCh := make(chan struct{})
			defer close(stopCh)
			kr.Start(stopCh)
			<-cmd.Context().Done()

			logger.Info(cmd.Context(), "exiting")

			return nil
		},
	}
	cmd.Flags().StringVarP(&coderURL, "coder-url", "", os.Getenv("CODER_URL"), "URL of the Coder instance")
	cmd.Flags().StringVarP(&coderToken, "coder-token", "", os.Getenv("CODER_TOKEN"), "Access Token for the Coder instance. Requires Template Admin privileges.")
	cmd.Flags().StringVarP(&artifactoryURL, "artifactory-url", "", os.Getenv("CODER_ARTIFACTORY_URL"), "URL of the JFrog Artifactory instance")
	cmd.Flags().StringVarP(&artifactoryToken, "artifactory-token", "", os.Getenv("CODER_ARTIFACTORY_TOKEN"), "Access Token for JFrog Artifactory instance")
	cmd.Flags().StringVarP(&artifactoryUser, "artifactory-user", "", os.Getenv("CODER_ARTIFACTORY_USER"), "User to interface with JFrog Artifactory instance")
	cmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "/home/coder/.kube/config", "Path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", os.Getenv("CODER_NAMESPACE"), "Namespace to use when listing pods")
	cmd.Flags().StringVarP(&fieldSelector, "field-selector", "f", "", "Field selector to use when listing pods")
	cmd.Flags().StringVarP(&labelSelector, "label-selector", "l", "", "Label selector to use when listing pods")
	return cmd
}
