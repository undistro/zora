package eks

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/eks"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"

	"github.com/getupio-undistro/snitch/api/v1alpha1"
	"github.com/getupio-undistro/snitch/pkg/provider/cloud"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd/api"
)

func GetCredentials(credentialsRef *corev1.SecretReference, secretKeySelector cloud.SecretKeySelectorValueFunc) (accessKeyID, secretAccessKey string, err error) {
	accessKeyID, err = secretKeySelector(credentialsRef, "accessKeyId")
	if err != nil {
		return "", "", err
	}
	secretAccessKey, err = secretKeySelector(credentialsRef, "secretAccessKey")
	if err != nil {
		return "", "", err
	}
	return accessKeyID, secretAccessKey, nil
}

func GetConfig(secretKeySelector cloud.SecretKeySelectorValueFunc, spec *v1alpha1.ClusterEKSSpec) (*api.Config, error) {
	accessKeyID, secretAccessKey, err := GetCredentials(&spec.CredentialsRef, secretKeySelector)
	if err != nil {
		return nil, err
	}
	sess, err := getAWSSession(accessKeyID, secretAccessKey, spec.Region, "")
	if err != nil {
		return nil, err
	}
	client := eks.New(sess)
	cluster, err := client.DescribeCluster(&eks.DescribeClusterInput{Name: &spec.Name})
	if err != nil {
		return nil, fmt.Errorf("failed to describe EKS cluster %s: %w", spec.Name, err)
	}

	config := api.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters:   map[string]*api.Cluster{},
		AuthInfos:  map[string]*api.AuthInfo{},
		Contexts:   map[string]*api.Context{},
	}
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token with aws-iam-authenticator: %w", err)
	}
	t, err := gen.GetWithOptions(&token.GetTokenOptions{ClusterID: *cluster.Cluster.Name, Session: sess})
	if err != nil {
		return nil, fmt.Errorf("failed to generate token with aws-iam-authenticator: %w", err)
	}
	name := fmt.Sprintf("eks_%s_%s", spec.Region, spec.Name)
	cert, err := base64.StdEncoding.DecodeString(*cluster.Cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode EKS cluster CA data: %w", err)
	}

	config.Clusters[name] = &api.Cluster{CertificateAuthorityData: cert, Server: *cluster.Cluster.Endpoint}
	config.CurrentContext = name

	// Just reuse the context name as an auth name.
	config.Contexts[name] = &api.Context{Cluster: name, AuthInfo: name}

	// AWS specific configation; use cloud platform scope.
	config.AuthInfos[name] = &api.AuthInfo{Token: t.Token}

	return &config, nil
}

func ClusterIsReady(secretKeySelector cloud.SecretKeySelectorValueFunc, spec *v1alpha1.ClusterEKSSpec) (bool, error) {
	accessKeyID, secretAccessKey, err := GetCredentials(&spec.CredentialsRef, secretKeySelector)
	if err != nil {
		return false, err
	}
	client, err := getClient(accessKeyID, secretAccessKey, spec.Region, "")
	if err != nil {
		return false, err
	}
	cluster, err := client.DescribeCluster(&eks.DescribeClusterInput{Name: &spec.Name})
	if err != nil {
		return false, err
	}

	switch *cluster.Cluster.Status {
	case "ACTIVE", "UPDATING":
		return true, nil
	}
	return false, nil
}

func getClient(accessKeyID, secretAccessKey, region, endpoint string) (*eks.EKS, error) {
	s, err := getAWSSession(accessKeyID, secretAccessKey, region, endpoint)
	if err != nil {
		return nil, err
	}
	return eks.New(s), nil
}

func getAWSSession(accessKeyID, secretAccessKey, region, endpoint string) (*session.Session, error) {
	config := aws.NewConfig().
		WithRegion(region).
		WithCredentials(credentials.NewStaticCredentials(accessKeyID, secretAccessKey, "")).
		WithMaxRetries(3)
	if endpoint != "" {
		config = config.WithEndpoint(endpoint)
	}
	awsSession, err := session.NewSession(config)
	if err != nil {
		return awsSession, fmt.Errorf("failed to create an AWS session: %w", err)
	}

	return awsSession, nil
}
