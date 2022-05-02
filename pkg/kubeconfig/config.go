package kubeconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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
		return nil, fmt.Errorf("missing key '.data.%s' in secret %q", SecretField, secret.Name)
	}

	config, err := clientcmd.Load(b)
	if err != nil {
		if err = Check(config); err != nil {
			return nil, fmt.Errorf("invalid kubeconfig: %w", err)
		}
	}
	return clientcmd.RESTConfigFromKubeConfig(b)
}

// Check return an error if config is invalid (contains unsupported entries)
func Check(config *clientcmdapi.Config) error {
	// checking current context
	name := config.CurrentContext
	ctx, ok := config.Contexts[name]
	if !ok || ctx == nil {
		return fmt.Errorf("can't find referenced current context %q", name)
	}

	// checking referenced cluster
	cluster, ok := config.Clusters[ctx.Cluster]
	if !ok || cluster == nil {
		return fmt.Errorf("can't find referenced cluster %q from current context %q", ctx.Cluster, name)
	}
	if cluster.CertificateAuthority != "" {
		return fmt.Errorf("CA files are not supported (try 'certificate-authority-data' instead)")
	}

	// checking referenced user
	user, ok := config.AuthInfos[ctx.AuthInfo]
	if !ok || user == nil {
		return fmt.Errorf("can't find referenced user %q from current context %q", ctx.AuthInfo, name)
	}
	if user.ClientCertificate != "" {
		return fmt.Errorf("client certificate files are not supported (try 'client-certificate-data' instead)")
	}
	if user.ClientKey != "" {
		return fmt.Errorf("client key files are not supported (try 'client-key-data' instead)")
	}
	if user.TokenFile != "" {
		return fmt.Errorf("token files are not supported (try 'token' instead)")
	}
	if user.Impersonate != "" || len(user.ImpersonateGroups) > 0 {
		return fmt.Errorf("impersonation is not supported")
	}
	if user.AuthProvider != nil && len(user.AuthProvider.Config) > 0 {
		return fmt.Errorf("auth provider configurations are not supported")
	}
	if user.Exec != nil {
		return fmt.Errorf("exec entrypoint is not supported")
	}

	return nil
}
