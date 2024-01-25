package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/coder/xray/jfrog"
)

func root() *cobra.Command {
	var (
		coderURL         string
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
				return xerrors.New("--coder-url is required")
			}

			_, err = url.Parse(artifactoryURL)
			if err != nil {
				return fmt.Errorf("parse coder URL: %w", err)
			}

			if artifactoryUser == "" {
				return xerrors.New("--artifactory-user is required")
			}

			if artifactoryToken == "" {
				return xerrors.New("--artifactory-token is required")
			}

			config, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
			if err != nil {
				return xerrors.Errorf("build kubeconfig: %w", err)
			}

			kclient, err := kubernetes.NewForConfig(config)
			if err != nil {
				return xerrors.Errorf("create kubernetes config: %w", err)
			}

			jClient, err := jfrog.XRayClient(artifactoryURL, artifactoryUser, artifactoryToken)
			if err != nil {
				return xerrors.Errorf("create artifactory client: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&coderURL, "coder-url", "cu", os.Getenv("CODER_URL"), "URL of the Coder instance")
	cmd.Flags().StringVarP(&artifactoryURL, "artifactory-url", "", os.Getenv("ARTIFACTORY_URL"), "URL of the JFrog Artifactory instance")
	cmd.Flags().StringVarP(&artifactoryToken, "artifactory-token", "", os.Getenv("ARTIFACTORY_TOKEN"), "Access Token for JFrog Artifactory instance")
	cmd.Flags().StringVarP(&artifactoryUser, "artifactory-user", "", os.Getenv("ARTIFACTORY_USER"), "User to interface with JFrog Artifactory instance")
	cmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "~/.kube/config", "Path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", os.Getenv("CODER_NAMESPACE"), "Namespace to use when listing pods")
	cmd.Flags().StringVarP(&fieldSelector, "field-selector", "f", "", "Field selector to use when listing pods")
	cmd.Flags().StringVarP(&labelSelector, "label-selector", "l", "", "Label selector to use when listing pods")
	cmd.Flags().StringVarP(&artifactoryToken, "artifactory-token", "", "", "Token to use to fetch scan results for an image")
	return cmd
}
