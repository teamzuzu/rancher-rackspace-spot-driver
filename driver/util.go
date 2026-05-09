package driver

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/rancher/kontainer-engine/types"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	rancherServiceAccount   = "rancher"
	rancherNamespace        = "kube-system"
	rancherClusterRoleBinding = "rancher"
	tokenSecretName         = "rancher-token"
	tokenWaitTimeout        = 30 * time.Second
)

// populateClusterInfoFromKubeconfig extracts the endpoint and CA certificate
// from a raw kubeconfig string and stores them in ClusterInfo.
func populateClusterInfoFromKubeconfig(info *types.ClusterInfo, kubeconfig string) error {
	cfg, err := clientcmd.Load([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	ctxName := cfg.CurrentContext
	if ctxName == "" {
		return fmt.Errorf("kubeconfig has no current-context")
	}
	kubeCtx, ok := cfg.Contexts[ctxName]
	if !ok {
		return fmt.Errorf("context %q not found in kubeconfig", ctxName)
	}

	cluster, ok := cfg.Clusters[kubeCtx.Cluster]
	if !ok {
		return fmt.Errorf("cluster %q not found in kubeconfig", kubeCtx.Cluster)
	}

	info.Endpoint = cluster.Server

	if len(cluster.CertificateAuthorityData) > 0 {
		info.RootCaCertificate = base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
	}

	authInfo, ok := cfg.AuthInfos[kubeCtx.AuthInfo]
	if ok {
		if len(authInfo.ClientCertificateData) > 0 {
			info.ClientCertificate = base64.StdEncoding.EncodeToString(authInfo.ClientCertificateData)
		}
		if len(authInfo.ClientKeyData) > 0 {
			info.ClientKey = base64.StdEncoding.EncodeToString(authInfo.ClientKeyData)
		}
		if authInfo.Token != "" {
			info.ServiceAccountToken = authInfo.Token
		}
	}

	return nil
}

// ensureRancherServiceAccount creates a service account with cluster-admin rights and
// returns its token. Safe to call multiple times.
func ensureRancherServiceAccount(ctx context.Context, kubeconfig string, info *types.ClusterInfo) error {
	restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return fmt.Errorf("failed to build REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	if err := createServiceAccount(ctx, clientset); err != nil {
		return err
	}
	if err := createClusterRoleBinding(ctx, clientset); err != nil {
		return err
	}

	token, err := getOrCreateToken(ctx, clientset)
	if err != nil {
		return err
	}

	info.ServiceAccountToken = token
	return nil
}

func createServiceAccount(ctx context.Context, cs kubernetes.Interface) error {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rancherServiceAccount,
			Namespace: rancherNamespace,
		},
	}
	_, err := cs.CoreV1().ServiceAccounts(rancherNamespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create service account: %w", err)
	}
	return nil
}

func createClusterRoleBinding(ctx context.Context, cs kubernetes.Interface) error {
	crb := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: rancherClusterRoleBinding,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      rancherServiceAccount,
				Namespace: rancherNamespace,
			},
		},
	}
	_, err := cs.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create cluster role binding: %w", err)
	}
	return nil
}

// getOrCreateToken handles both old-style auto-created tokens (k8s <1.24) and the
// explicit Secret approach required in k8s 1.24+.
func getOrCreateToken(ctx context.Context, cs kubernetes.Interface) (string, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tokenSecretName,
			Namespace: rancherNamespace,
			Annotations: map[string]string{
				corev1.ServiceAccountNameKey: rancherServiceAccount,
			},
		},
		Type: corev1.SecretTypeServiceAccountToken,
	}

	_, err := cs.CoreV1().Secrets(rancherNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return "", fmt.Errorf("failed to create token secret: %w", err)
	}

	// The token is populated asynchronously by the token controller.
	deadline := time.Now().Add(tokenWaitTimeout)
	for {
		s, err := cs.CoreV1().Secrets(rancherNamespace).Get(ctx, tokenSecretName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get token secret: %w", err)
		}
		if token := string(s.Data["token"]); token != "" {
			return token, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("timed out waiting for service account token")
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}
