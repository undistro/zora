package kubeconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigFromSecretName return a rest.Config from a kubeconfig secret name
func ConfigFromSecretName(ctx context.Context, client ctrlclient.Client, name types.NamespacedName) (*rest.Config, error) {
	existing := &corev1.Secret{}
	if err := client.Get(ctx, name, existing); err != nil {
		return nil, fmt.Errorf("failed to get secret %v: %w", name, err)
	}
	return ConfigFromSecret(existing)
}

// ConfigFromSecret return a rest.Config from a kubeconfig secret
func ConfigFromSecret(secret *corev1.Secret) (*rest.Config, error) {
	b, ok := secret.Data[SecretField]
	if !ok {
		return nil, fmt.Errorf("missing key %q in secret data", SecretField)
	}
	return clientcmd.RESTConfigFromKubeConfig(b)
}
