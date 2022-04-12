package cloud

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretKeySelectorValueFunc func(credentialsRef *corev1.SecretReference, key string) (string, error)

func DefaultSecretKeySelectorValueFunc(ctx context.Context, reader client.Reader, defaultNamespace string) SecretKeySelectorValueFunc {
	return func(credentialsRef *corev1.SecretReference, key string) (string, error) {
		if credentialsRef == nil {
			return "", errors.New("credentialsRef is nil")
		}
		if credentialsRef.Name == "" {
			return "", errors.New("credentialsRef.Name is empty")
		}
		if credentialsRef.Namespace == "" {
			credentialsRef.Namespace = defaultNamespace
		}
		if key == "" {
			return "", errors.New("key is empty")
		}
		secret := &corev1.Secret{}
		k := client.ObjectKey{Namespace: credentialsRef.Namespace, Name: credentialsRef.Name}
		if err := reader.Get(ctx, k, secret); err != nil {
			return "", fmt.Errorf("failed to get secret %q: %w", k.String(), err)
		}

		if _, ok := secret.Data[key]; !ok {
			return "", fmt.Errorf("secret %q has no key %q", k.String(), key)
		}

		return string(secret.Data[key]), nil
	}
}
