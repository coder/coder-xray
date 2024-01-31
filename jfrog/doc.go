// package jfrog contains an abstraction for interfacing with an JFrog
// artifactory instance.
package jfrog

//go:generate mockgen -destination ./mock.go -package jfrog github.com/coder/xray/jfrog Client
