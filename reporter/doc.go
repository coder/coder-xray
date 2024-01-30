// package reporter contains an abstraction for scanning Coder Kubernetes workspaces using JFrog JFrog
// and uploading results to a Coder deployment.
package reporter

//go:generate mockgen -destination ./codermock.go -package reporter github.com/coder/xray/reporter CoderClient
