package main

import (
	"os"

	"github.com/spf13/cobra"
)

func root() *cobra.Command {
	var (
		coderURL         string
		fieldSelector    string
		kubeConfig       string
		namespace        string
		labelSelector    string
		artifactoryToken string
	)
	cmd := &cobra.Command{}
	cmd.Flags().StringVarP(&coderURL, "coder-url", "u", os.Getenv("CODER_URL"), "URL of the Coder instance")
	cmd.Flags().StringVarP(&kubeConfig, "kubeconfig", "k", "~/.kube/config", "Path to the kubeconfig file")
	cmd.Flags().StringVarP(&namespace, "namespace", "n", os.Getenv("CODER_NAMESPACE"), "Namespace to use when listing pods")
	cmd.Flags().StringVarP(&fieldSelector, "field-selector", "f", "", "Field selector to use when listing pods")
	cmd.Flags().StringVarP(&labelSelector, "label-selector", "l", "", "Label selector to use when listing pods")
	cmd.Flags().StringVarP(&artifactoryToken, "artifactory-token", "", "", "Token to use to fetch scan results for an image")
	return cmd
}
