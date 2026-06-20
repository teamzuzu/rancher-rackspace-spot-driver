package driver

import (
	"context"
	"fmt"
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

// importMockAPI is a minimal spotAPI implementation for import tests.
// GetCloudspace is configured via fields to simulate not-found or an API error.
// CreateSpotNodePool is tracked to verify importCloudspace makes no mutations.
type importMockAPI struct {
	cloudspace  *spotv1.CloudSpace
	notFound    bool
	apiErr      error
	createCalls []string
}

func (m *importMockAPI) Authenticate(_ context.Context) (string, error) { return "", nil }
func (m *importMockAPI) GetCloudspace(_ context.Context, _, _ string) (*spotv1.CloudSpace, error) {
	if m.notFound {
		return nil, fmt.Errorf("cloudspace not found")
	}
	if m.apiErr != nil {
		return nil, m.apiErr
	}
	return m.cloudspace, nil
}
func (m *importMockAPI) CreateCloudspace(_ context.Context, _ spotv1.CloudSpace) error { return nil }
func (m *importMockAPI) DeleteCloudspace(_ context.Context, _, _ string) error         { return nil }
func (m *importMockAPI) GetCloudspaceConfig(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (m *importMockAPI) ListSpotNodePools(_ context.Context, _, _ string) ([]*spotv1.SpotNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) CreateSpotNodePool(_ context.Context, _ string, p spotv1.SpotNodePool) error {
	m.createCalls = append(m.createCalls, p.Name)
	return nil
}
func (m *importMockAPI) UpdateSpotNodePool(_ context.Context, _ string, _ spotv1.SpotNodePool) error {
	return nil
}
func (m *importMockAPI) DeleteSpotNodePool(_ context.Context, _, _ string) error { return nil }
func (m *importMockAPI) GetOnDemandNodePool(_ context.Context, _, _ string) (*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) CreateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *importMockAPI) UpdateOnDemandNodePool(_ context.Context, _ string, _ spotv1.OnDemandNodePool) error {
	return nil
}
func (m *importMockAPI) ListOnDemandNodePools(_ context.Context, _, _ string) ([]*spotv1.OnDemandNodePool, error) {
	return nil, nil
}
func (m *importMockAPI) DeleteOnDemandNodePool(_ context.Context, _, _ string) error { return nil }

func TestImportCloudspace_notFound(t *testing.T) {
	client := &spotClient{api: &importMockAPI{notFound: true}, org: "org"}
	_, err := client.importCloudspace(context.Background(), "org", "missing-cluster", "tok")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error = %q, want it to contain 'not found'", err.Error())
	}
}

func TestImportCloudspace_success(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:              "existing-cluster",
		Region:            "us-east-iad-1",
		KubernetesVersion: "1.33.0",
		CNI:               "calico",
		SpotNodepools: []*spotv1.SpotNodePool{
			{Name: "pool-x", ServerClass: "gp.vs1.medium-iad", Desired: 3, BidPrice: "0.05"},
		},
	}
	mock := &importMockAPI{cloudspace: cs}
	client := &spotClient{api: mock, org: "org"}

	s, err := client.importCloudspace(context.Background(), "org", "existing-cluster", "my-token")
	if err != nil {
		t.Fatalf("importCloudspace() error = %v", err)
	}
	if s.CloudspaceName != "existing-cluster" {
		t.Fatalf("CloudspaceName = %q, want existing-cluster", s.CloudspaceName)
	}
	if s.RefreshToken != "my-token" {
		t.Fatalf("RefreshToken = %q, want my-token", s.RefreshToken)
	}
	if s.SpotPoolName != "pool-x" {
		t.Fatalf("SpotPoolName = %q, want pool-x", s.SpotPoolName)
	}
	if len(mock.createCalls) != 0 {
		t.Fatalf("unexpected CreateSpotNodePool calls: %v", mock.createCalls)
	}
}

func TestImportCloudspace_apiError(t *testing.T) {
	mock := &importMockAPI{apiErr: fmt.Errorf("connection refused")}
	client := &spotClient{api: mock, org: "org"}
	_, err := client.importCloudspace(context.Background(), "org", "my-cluster", "tok")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("error = %q, want it to contain the original error message", err.Error())
	}
}
