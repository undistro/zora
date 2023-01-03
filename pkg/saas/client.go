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

	"github.com/undistro/zora/pkg/payloads/v1alpha1"
)

const (
	workspacePathF = "zora/api/%s/workspaces/%s"
	clusterPathF   = "namespaces/%s/clusters/%s"
)

var allowedStatus = []int{
	http.StatusOK,
	http.StatusCreated,
	http.StatusAccepted,
	http.StatusNoContent,
}

type Client interface {
	PutCluster(ctx context.Context, cluster v1alpha1.Cluster) error
	DeleteCluster(ctx context.Context, namespace, name string) error
	PutClusterScan(ctx context.Context, namespace, name string, pluginStatus map[string]*v1alpha1.PluginStatus) error
	DeleteClusterScan(ctx context.Context, namespace, name string) error
}

type client struct {
	client      *http.Client
	baseURL     *url.URL
	workspaceID string
	version     string
}

func NewClient(baseURL, version, workspaceID string, httpclient *http.Client) (Client, error) {
	u, err := validateURL(baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf(workspacePathF, version, workspaceID))
	return &client{
		version:     version,
		baseURL:     u,
		workspaceID: workspaceID,
		client:      httpclient,
	}, nil
}

func (r *client) PutCluster(ctx context.Context, cluster v1alpha1.Cluster) error {
	u := r.clusterURL(cluster.Namespace, cluster.Name)
	b, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) DeleteCluster(ctx context.Context, namespace, name string) error {
	u := r.clusterURL(namespace, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) PutClusterScan(ctx context.Context, namespace, name string, pluginStatus map[string]*v1alpha1.PluginStatus) error {
	u := r.clusterURL(namespace, name, "scan")
	b, err := json.Marshal(pluginStatus)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) DeleteClusterScan(ctx context.Context, namespace, name string) error {
	u := r.clusterURL(namespace, name, "scan")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	res, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return validateStatus(res)
}

func (r *client) clusterURL(namespace, name string, extra ...string) string {
	p := path.Join(r.baseURL.Path, fmt.Sprintf(clusterPathF, namespace, name))
	if len(extra) > 0 {
		tmp := []string{p}
		p = path.Join(append(tmp, extra...)...)
	}
	u := *r.baseURL
	u.Path = p
	return u.String()
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
