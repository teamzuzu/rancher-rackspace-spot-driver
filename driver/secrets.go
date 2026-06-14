package driver

import (
	"context"

	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	credentialsSecretName = "rackspace-spot-credentials"
	credentialsNamespace  = "cattle-system"
	credentialsOrgKey     = "org"
	credentialsTokenKey   = "refreshToken"
)

type credentialDefaults struct {
	Org          string
	RefreshToken string
}

// defaultCredentialLoader is called by stateFromOptions and GetDriverCreateOptions
// to read credential defaults. Replaced in tests to avoid real k8s access.
var defaultCredentialLoader func(ctx context.Context) credentialDefaults = loadCredentialDefaults

func loadCredentialDefaults(ctx context.Context) credentialDefaults {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		logrus.Debugf("[%s] not running in-cluster, skipping credential secret lookup: %v", driverName, err)
		return credentialDefaults{}
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logrus.Warnf("[%s] failed to create k8s client for credential lookup: %v", driverName, err)
		return credentialDefaults{}
	}
	return loadCredentialDefaultsFromClient(ctx, client)
}

func loadCredentialDefaultsFromClient(ctx context.Context, client kubernetes.Interface) credentialDefaults {
	secret, err := client.CoreV1().Secrets(credentialsNamespace).Get(ctx, credentialsSecretName, metav1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			logrus.Warnf("[%s] failed to read credential secret %s/%s: %v",
				driverName, credentialsNamespace, credentialsSecretName, err)
		}
		return credentialDefaults{}
	}
	return credentialDefaults{
		Org:          string(secret.Data[credentialsOrgKey]),
		RefreshToken: string(secret.Data[credentialsTokenKey]),
	}
}
