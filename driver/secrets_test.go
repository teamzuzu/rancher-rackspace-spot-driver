package driver

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestLoadCredentialDefaultsFromClientBothKeys(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsSecretName,
			Namespace: credentialsNamespace,
		},
		Data: map[string][]byte{
			credentialsOrgKey:   []byte("my-org"),
			credentialsTokenKey: []byte("my-token"),
		},
	})

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "my-org" {
		t.Errorf("Org = %q, want my-org", got.Org)
	}
	if got.RefreshToken != "my-token" {
		t.Errorf("RefreshToken = %q, want my-token", got.RefreshToken)
	}
}

func TestLoadCredentialDefaultsFromClientMissingOrgKey(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      credentialsSecretName,
			Namespace: credentialsNamespace,
		},
		Data: map[string][]byte{
			credentialsTokenKey: []byte("my-token"),
			// org key absent
		},
	})

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "" {
		t.Errorf("Org = %q, want empty", got.Org)
	}
	if got.RefreshToken != "my-token" {
		t.Errorf("RefreshToken = %q, want my-token", got.RefreshToken)
	}
}

func TestLoadCredentialDefaultsFromClientSecretNotFound(t *testing.T) {
	client := fake.NewSimpleClientset() // empty — no secret

	got := loadCredentialDefaultsFromClient(context.Background(), client)

	if got.Org != "" || got.RefreshToken != "" {
		t.Errorf("got non-empty defaults when secret missing: %+v", got)
	}
}
