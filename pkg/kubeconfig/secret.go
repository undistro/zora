package kubeconfig

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const SecretField = "value"

// ApplyKubeconfigSecret create or update secret with kubeconfig data
func ApplyKubeconfigSecret(ctx context.Context, client ctrlclient.Client, config *api.Config, name types.NamespacedName, owner ctrlclient.Object) (*corev1.Secret, error) {
	kubeconfig, err := clientcmd.Write(*config)
	if err != nil {
		return nil, err
	}
	existing := &corev1.Secret{}
	if err := client.Get(ctx, name, existing); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get secret %v: %w", name, err)
	}
	data := map[string][]byte{SecretField: kubeconfig}

	// update if already exists
	if existing.Name != "" {
		existing.Data = data
		if err := client.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update secret %v: %w", name, err)
		}
		return existing, nil
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.Name,
			Namespace: name.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}
	if err := controllerutil.SetOwnerReference(owner, secret, client.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to set owner reference in secret %v: %w", name, err)
	}
	if err := client.Create(ctx, secret); err != nil {
		return nil, fmt.Errorf("failed to create secret %v: %w", name, err)
	}

	return secret, nil
}
