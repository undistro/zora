// Copyright 2022 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package saas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/undistro/zora/api/zora/v1alpha2"
	"github.com/undistro/zora/pkg/authentication"
	"github.com/undistro/zora/pkg/filemonitor"
)

const (
	clusterPathF  = "zora/api/%s/workspaces/%s/namespaces/%s/clusters/%s"
	versionHeader = "x-zora-version"
)

var allowedStatus = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNoContent,
}

type Client interface {
	PutCluster(ctx context.Context, cluster Cluster) error
	DeleteCluster(ctx context.Context, namespace, name string) error
	PutClusterScan(ctx context.Context, namespace, name string, pluginStatus map[string]*PluginStatus) error
	DeleteClusterScan(ctx context.Context, namespace, name string) error
	PutVulnerabilityReport(ctx context.Context, namespace, name string, vulnReport *v1alpha2.VulnerabilityReport) error
	PutClusterStatus(ctx context.Context, namespace, name string, pluginStatus map[string]*PluginStatus) error
}

type client struct {
	client       *http.Client
	baseURL      *url.URL
	workspaceID  string
	version      string
	tokenMonitor *filemonitor.FileMonitor
}

func NewClient(baseURL, version, workspaceID string, httpclient *http.Client, tokenPath string, done <-chan struct{}) (Client, error) {
	u, err := validateURL(baseURL)
	if err != nil {
		return nil, err
	}
	tokenMonitor := filemonitor.NewFileMonitor(tokenPath, func(content []byte) (any, error) {
		return authentication.ParseTokenData(content)
	})
	go tokenMonitor.MonitorFile(done)

	return &client{
		version:      version,
		baseURL:      u,
		workspaceID:  workspaceID,
		client:       httpclient,
		tokenMonitor: tokenMonitor,
	}, nil
}

func (r *client) PutCluster(ctx context.Context, cluster Cluster) error {
	u := r.clusterURL("v1alpha1", cluster.Namespace, cluster.Name)
	b, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set("content-type", "application/json")
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) DeleteCluster(ctx context.Context, namespace, name string) error {
	u := r.clusterURL("v1alpha1", namespace, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) PutClusterScan(ctx context.Context, namespace, name string, pluginStatus map[string]*PluginStatus) error {
	u := r.clusterURL("v1alpha1", namespace, name, "scan")
	b, err := json.Marshal(pluginStatus)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set("content-type", "application/json")
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) PutVulnerabilityReport(ctx context.Context, namespace, name string, vulnReport *v1alpha2.VulnerabilityReport) error {
	if vulnReport == nil {
		return nil
	}
	u := r.clusterURL("v1alpha2", namespace, name, "vulnerabilityreports")
	b, err := json.Marshal(vulnReport)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set("content-type", "application/json")
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) DeleteClusterScan(ctx context.Context, namespace, name string) error {
	u := r.clusterURL("v1alpha1", namespace, name, "scan")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) PutClusterStatus(ctx context.Context, namespace, name string, pluginStatus map[string]*PluginStatus) error {
	u := r.clusterURL("v1alpha1", namespace, name, "status")
	b, err := json.Marshal(pluginStatus)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	r.addAuthorizationHeader(req)
	req.Header.Set("content-type", "application/json")
	req.Header.Set(versionHeader, r.version)
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) clusterURL(version, namespace, name string, extra ...string) string {
	p := path.Join(r.baseURL.Path, fmt.Sprintf(clusterPathF, version, r.workspaceID, namespace, name))
	if len(extra) > 0 {
		tmp := []string{p}
		p = path.Join(append(tmp, extra...)...)
	}
	u := *r.baseURL
	u.Path = p
	return u.String()
}

func (r *client) addAuthorizationHeader(req *http.Request) {
	tokenContent := r.tokenMonitor.GetContent()
	if tokenContent != nil {
		if tokenData, ok := tokenContent.(*authentication.TokenData); ok {
			req.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenData.TokenType, tokenData.AccessToken))
		}
	}
}

func validateURL(u string) (*url.URL, error) {
	uri, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, err
	}
	if uri.Scheme != "http" && uri.Scheme != "https" {
		return nil, fmt.Errorf("invalid URL scheme")
	}
	if uri.Host == "" {
		return nil, fmt.Errorf("invalid URL host")
	}
	return uri, nil
}

func validateStatus(res *http.Response) error {
	for _, s := range allowedStatus {
		if res.StatusCode == s {
			return nil
		}
	}
	if res.StatusCode == http.StatusUnprocessableEntity {
		serr := &saasError{}
		if err := json.NewDecoder(res.Body).Decode(serr); err != nil {
			return fmt.Errorf("failed to decode SaaS error in response body: %v", err)
		}
		return serr
	}
	return fmt.Errorf("invalid HTTP status: %d", res.StatusCode)
}

type saasError struct {
	Err    string `json:"error,omitempty"`
	Detail string `json:"detail,omitempty"`
}

func (r saasError) Error() string {
	return fmt.Sprintf("%s: %s", r.Err, r.Detail)
}
