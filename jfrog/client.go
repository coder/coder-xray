package jfrog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/jfrog/jfrog-client-go/config"
	"github.com/jfrog/jfrog-client-go/http/jfroghttpclient"
	"github.com/jfrog/jfrog-client-go/utils/io/httputils"
	"github.com/jfrog/jfrog-client-go/xray"
	"github.com/jfrog/jfrog-client-go/xray/auth"
	"golang.org/x/xerrors"
)

type Client interface {
	ScanResults(img Image) (ScanResult, error)
	ResultsURL(img Image, packageID string) string
}

type client struct {
	client  *jfroghttpclient.JfrogHttpClient
	baseURL string
	token   string
	user    string
}

func XRayClient(url, user, token string) (Client, error) {
	details := auth.NewXrayDetails()
	details.SetAccessToken(token)
	details.SetUser(user)
	details.SetUrl(url)
	conf, err := config.NewConfigBuilder().SetServiceDetails(details).Build()
	if err != nil {
		return nil, xerrors.Errorf("new config builder: %w", err)
	}
	mgr, err := xray.New(conf)
	if err != nil {
		return nil, xerrors.Errorf("new xray manager: %w", err)
	}
	return &client{
		client:  mgr.Client(),
		baseURL: url,
		user:    user,
		token:   token,
	}, nil
}

type securityResultsPayload struct {
	Data   []ScanResult `json:"data"`
	Offset int          `json:"offset"`
}

type ScanResult struct {
	Version        string         `json:"version"`
	SecurityIssues SecurityIssues `json:"sec_issues"`
	PackageID      string         `json:"package_id"`
}

type SecurityIssues struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Total    int `json:"total"`
}

func (c *client) ScanResults(img Image) (ScanResult, error) {
	path := fmt.Sprintf("%s/xray/api/v1/packages/%s/versions?search=%s", c.baseURL, img.Package, img.Version)
	resp, body, _, err := c.client.SendGet(path, true, &httputils.HttpClientDetails{
		User:        c.user,
		AccessToken: c.token,
	})
	if err != nil {
		return ScanResult{}, xerrors.Errorf("send get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ScanResult{}, xerrors.Errorf("unexpected status code %d body (%s)", resp.StatusCode, body)
	}

	var payload securityResultsPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return ScanResult{}, xerrors.Errorf("unmarshal (%s): %w", body, err)
	}

	if len(payload.Data) == 0 {
		return ScanResult{}, xerrors.Errorf("no results")
	}

	return payload.Data[0], nil
}

type Image struct {
	Repo    string
	Package string
	Version string
}

func (c *client) ResultsURL(img Image, packageID string) string {
	return fmt.Sprintf("%s/ui/scans-list/packages-scans/%s/scan-descendants/%s?package_id=%s&version=%s", c.baseURL, img.Package, img.Version, packageID, img.Version)
}

func ParseImage(image string) (Image, error) {
	tag, err := name.NewTag(image)
	if err != nil {
		return Image{}, xerrors.Errorf("new tag: %w", err)
	}

	repo := root(tag.RepositoryStr())
	pkg, err := filepath.Rel(repo, tag.RepositoryStr())
	if err != nil {
		return Image{}, xerrors.Errorf("rel path between %q and %q", repo, tag.RegistryStr())
	}

	return Image{
		Repo:    repo,
		Package: pkg,
		Version: tag.TagStr(),
	}, nil
}

func root(path string) string {
	dir := filepath.Dir(path)
	if dir == "." {
		return path
	}
	return root(dir)
}
