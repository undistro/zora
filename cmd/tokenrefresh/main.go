// Copyright 2024 Undistro Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/undistro/zora/pkg/authentication"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	secretName       string
	secretNamespace  string
	tokenName        string
	domain           string
	clientID         string
	minRefreshTime   time.Duration
	refreshThreshold time.Duration
	kubeconfig       string
)

const (
	annotationStatus = "zora.undistro.io/status"
)

type SecretStatus struct {
	LastRefreshTime      time.Time `json:"lastRefreshTime"`
	NextScheduledRefresh time.Time `json:"nextScheduledRefresh"`
	TokenExpiry          time.Time `json:"tokenExpiry"`
}

type Controller struct {
	clientset *kubernetes.Clientset
	client    *http.Client
	refreshCh chan struct{}
	mutex     sync.Mutex
}

func init() {
	flag.StringVar(&secretName, "secret-name", "oauth-tokens", "Name of the secret containing OAuth tokens")
	flag.StringVar(&secretNamespace, "namespace", "default", "Namespace of the secret")
	flag.StringVar(&tokenName, "token-name", "token", "Name of token within the secret's data")
	flag.StringVar(&domain, "domain", "", "Domain name associated with the token")
	flag.StringVar(&clientID, "client-id", "", "Client ID associated with the token")
	flag.DurationVar(&minRefreshTime, "min-refresh-time", 5*time.Minute, "Minimum time between refresh attempts")
	flag.DurationVar(&refreshThreshold, "refresh-threshold", 5*time.Minute, "Time before token expiry to attempt refresh")
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file. If not set, in-cluster config will be used")
}

func main() {
	flag.Parse()

	// Set up Kubernetes client
	var config *rest.Config
	var err error
	if kubeconfig == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	controller := &Controller{
		client:    &http.Client{Transport: &http.Transport{Proxy: http.ProxyFromEnvironment}},
		clientset: clientset,
		refreshCh: make(chan struct{}, 1),
	}

	// Create a watch on the secret
	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"secrets",
		secretNamespace,
		fields.OneTermEqualSelector("metadata.name", secretName),
	)

	_, informer := cache.NewInformer(
		watchlist,
		&v1.Secret{},
		0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { controller.handleSecretChange("added", obj) },
			UpdateFunc: func(_, newObj interface{}) { controller.handleSecretChange("updated", newObj) },
		},
	)

	// Start the informer
	stop := make(chan struct{})
	defer close(stop)
	go informer.Run(stop)

	// Start the refresh loop
	controller.refreshLoop(context.Background())
}

func (c *Controller) handleSecretChange(operation string, obj interface{}) {
	secret, ok := obj.(*v1.Secret)
	if !ok {
		slog.Error(fmt.Sprintf("Error: Unexpected object type %s", reflect.TypeOf(secret)))
		return
	}

	slog.Info(fmt.Sprintf("Secret Change: %s %s", secret.Name, operation))

	// Trigger an asynchronous refresh
	select {
	case c.refreshCh <- struct{}{}:
		slog.Info("Secret Change:Triggered asynchronous refresh")
	default:
		slog.Info("Secret Change:Refresh already queued")
	}
}

func (c *Controller) refreshLoop(ctx context.Context) {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		c.mutex.Lock()
		_, expiryTime, _, err := c.getTokenData(ctx)
		c.mutex.Unlock()
		if err != nil {
			slog.Error(fmt.Sprintf("Error getting token data: %v", err))
			time.Sleep(minRefreshTime)
			continue
		}

		timeUntilExpiry := time.Until(expiryTime)
		refreshTime := timeUntilExpiry - refreshThreshold

		if refreshTime < minRefreshTime {
			refreshTime = minRefreshTime
		}

		slog.Info(fmt.Sprintf("Refresh Loop: next refresh scheduled in %v", refreshTime))

		select {
		case <-time.After(refreshTime):
			slog.Info("Refresh Loop: refreshing token due to scheduled refresh")
		case <-c.refreshCh:
			slog.Info("Refresh Loop: refreshing token due to secret change")
		case sig := <-signalCh:
			slog.Info(fmt.Sprintf("Refresh Loop: received signal %s, terminating application", sig))
			return
		}

		c.refreshTokenIfNeeded(ctx)
	}
}

func (c *Controller) getTokenData(ctx context.Context) (*authentication.TokenData, time.Time, *SecretStatus, error) {
	secret, err := c.clientset.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, time.Time{}, nil, err
	}

	tokenData, err := authentication.ParseTokenData(secret.Data[tokenName])
	if err != nil {
		return nil, time.Time{}, nil, err
	}

	expiryTime, err := authentication.GetJWTExpiry(tokenData)
	if err != nil {
		return nil, time.Time{}, nil, err
	}

	secretStatus := getSecretStatus(secret)
	return tokenData, expiryTime, secretStatus, nil
}

func (c *Controller) refreshTokenIfNeeded(ctx context.Context) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tokenData, expiryTime, secretStatus, err := c.getTokenData(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("Error getting token data: %v", err))
		return
	}

	if time.Until(expiryTime) > refreshThreshold {
		slog.Info("Refresh Loop: token is still valid, no need to refresh")
		if secretStatus == nil || !secretStatus.TokenExpiry.Equal(expiryTime) {
			// Update secret status
			err = c.updateSecretStatus(ctx, time.Now(), expiryTime)
			if err != nil {
				slog.Error(fmt.Sprintf("Error updating secret status: %v", err))
			}
		}
	} else {
		tokenData, err = c.refreshToken(domain, clientID, tokenData.RefreshToken)
		if err != nil {
			slog.Error(fmt.Sprintf("Error refreshing token: %v", err))
			return
		}

		err = c.updateSecret(ctx, tokenData)
		if err != nil {
			slog.Error(fmt.Sprintf("Error updating secret: %v", err))
			return
		}

		slog.Info("Refresh Loop: token refreshed successfully")
	}
}

func (c *Controller) refreshToken(domain, clientID, refreshToken string) (*authentication.TokenData, error) {
	url := fmt.Sprintf("https://%s/oauth/token", domain)
	data := fmt.Sprintf("grant_type=refresh_token&client_id=%s&refresh_token=%s", clientID, refreshToken)

	resp, err := c.client.Post(url, "application/x-www-form-urlencoded", bytes.NewBufferString(data))
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tokenData authentication.TokenData
	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return &tokenData, nil
}

func (c *Controller) updateSecret(ctx context.Context, tokenData *authentication.TokenData) error {
	secret, err := c.clientset.CoreV1().Secrets(secretNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	expiryTime, _ := authentication.GetJWTExpiry(tokenData)
	status := newSecretStatus(expiryTime, time.Now())
	err = setSecretStatus(secret, status)
	if err != nil {
		return err
	}

	tokenBytes, err := json.Marshal(tokenData)
	if err != nil {
		return err
	}

	secret.Data[tokenName] = tokenBytes
	_, err = c.clientset.CoreV1().Secrets(secretNamespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (c *Controller) updateSecretStatus(ctx context.Context, lastRefreshTime, tokenExpiry time.Time) error {
	status := newSecretStatus(tokenExpiry, lastRefreshTime)
	secret := v1.Secret{}
	err := setSecretStatus(&secret, status)
	if err != nil {
		return err
	}

	patch, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	_, err = c.clientset.CoreV1().Secrets(secretNamespace).Patch(ctx, secretName, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	return err
}

func newSecretStatus(tokenExpiry time.Time, lastRefreshTime time.Time) *SecretStatus {
	nextScheduledRefresh := tokenExpiry.Add(-refreshThreshold)

	status := SecretStatus{
		LastRefreshTime:      lastRefreshTime,
		NextScheduledRefresh: nextScheduledRefresh,
		TokenExpiry:          tokenExpiry,
	}
	return &status
}

func getSecretStatus(secret *v1.Secret) *SecretStatus {
	if secret != nil && secret.Annotations != nil {
		statusVal := secret.Annotations[annotationStatus]
		secretStatus := SecretStatus{}
		err := json.Unmarshal([]byte(statusVal), &secretStatus)

		if err == nil {
			return &secretStatus
		}
	}
	return nil
}

func setSecretStatus(secret *v1.Secret, status *SecretStatus) error {
	if secret == nil {
		return errors.New("missing secret")
	}
	if status == nil {
		return errors.New("missing status")
	}
	if secret.Annotations == nil {
		secret.Annotations = map[string]string{}
	}
	statusVal, err := json.Marshal(status)
	if err != nil {
		return err
	}
	secret.Annotations[annotationStatus] = string(statusVal)
	return nil
}
