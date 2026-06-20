package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/sirupsen/logrus"
)

// isNotFound works around a SDK inconsistency: the HTTP 404 path returns a plain
// fmt.Errorf (not *HTTPStatusError), so spotv1.IsNotFound always returns false for it.
// We avoid calling spotv1.IsNotFound (which uses errors.As) because the SDK error type
// has an As() method that causes infinite recursion.
func isNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not found")
}

// spotAPI is the subset of the Rackspace Spot SDK used by spotClient.
// It is defined as an interface to allow substitution in tests.
type spotAPI interface {
	Authenticate(ctx context.Context) (string, error)
	GetCloudspace(ctx context.Context, org, name string) (*spotv1.CloudSpace, error)
	CreateCloudspace(ctx context.Context, cs spotv1.CloudSpace) error
	DeleteCloudspace(ctx context.Context, org, name string) error
	GetCloudspaceConfig(ctx context.Context, namespace, name string) (string, error)
	ListSpotNodePools(ctx context.Context, org, cloudspaceName string) ([]*spotv1.SpotNodePool, error)
	CreateSpotNodePool(ctx context.Context, org string, pool spotv1.SpotNodePool) error
	UpdateSpotNodePool(ctx context.Context, org string, pool spotv1.SpotNodePool) error
	DeleteSpotNodePool(ctx context.Context, org, name string) error
	GetOnDemandNodePool(ctx context.Context, org, name string) (*spotv1.OnDemandNodePool, error)
	CreateOnDemandNodePool(ctx context.Context, org string, pool spotv1.OnDemandNodePool) error
	UpdateOnDemandNodePool(ctx context.Context, org string, pool spotv1.OnDemandNodePool) error
	ListOnDemandNodePools(ctx context.Context, org, cloudspaceName string) ([]*spotv1.OnDemandNodePool, error)
	DeleteOnDemandNodePool(ctx context.Context, org, name string) error
}

type spotClient struct {
	api    spotAPI
	rawSDK *spotv1.RackspaceSpotClient
	org    string
}

func newSpotClient(ctx context.Context, refreshToken, org string) (*spotClient, error) {
	c, err := spotv1.NewSpotClient(&spotv1.Config{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create spot client: %w", err)
	}

	if _, err := c.Authenticate(ctx); err != nil {
		return nil, fmt.Errorf("spot authentication failed: %w", err)
	}

	return &spotClient{api: c, rawSDK: c, org: org}, nil
}

// ensureCloudspace creates a CloudSpace or returns the existing one if already created.
func (c *spotClient) ensureCloudspace(ctx context.Context, s *clusterState) (*spotv1.CloudSpace, error) {
	existing, err := c.api.GetCloudspace(ctx, c.org, s.CloudspaceName)
	if err != nil && !isNotFound(err) {
		return nil, fmt.Errorf("failed to check existing cloudspace: %w", err)
	}
	if existing != nil {
		logrus.Infof("cloudspace %s already exists, reusing", s.CloudspaceName)
		return existing, nil
	}

	cs := spotv1.CloudSpace{
		Name:                 s.CloudspaceName,
		Org:                  c.org,
		Region:               s.Region,
		KubernetesVersion:    s.KubernetesVersion,
		CNI:                  s.CNI,
		GpuEnabled:           s.GPUEnabled,
		PreemptionWebhookURL: s.PreemptionWebhook,
		DeploymentType:       s.DeploymentType,
	}

	logrus.Infof("[rackspacespot] creating cloudspace %s", s.CloudspaceName)
	if err := c.api.CreateCloudspace(ctx, cs); err != nil {
		return nil, fmt.Errorf("failed to create cloudspace: %w", err)
	}

	return c.api.GetCloudspace(ctx, c.org, s.CloudspaceName)
}

// reconcileSpotNodePools creates, updates, and removes spot node pools to match desired state.
//
// Operations are applied in three phases to satisfy the vspotnodepool webhook constraint
// (at most one SpotNodePool with autoscaling enabled per cloudspace at any point in time):
//
//  1. Delete pools no longer in desired state — removes any existing autoscaling pool that is
//     being dropped before a different pool tries to enable autoscaling.
//  2. Disable autoscaling on pools that are losing it — ensures no two pools hold autoscaling
//     simultaneously when one pool is switching off while another switches on.
//  3. Create new pools and update remaining existing pools — safe to enable autoscaling now.
func (c *spotClient) reconcileSpotNodePools(ctx context.Context, s *clusterState) error {
	type poolSpec struct {
		serverClass string
		nodeCount   int
		bidPrice    string
		autoscaling bool
		minNodes    int64
		maxNodes    int64
	}

	desired := map[string]poolSpec{
		s.SpotPoolName: {
			serverClass: s.SpotServerClass,
			nodeCount:   s.SpotNodeCount,
			bidPrice:    s.SpotBidPrice,
			autoscaling: s.SpotAutoscaling,
			minNodes:    s.SpotMinNodes,
			maxNodes:    s.SpotMaxNodes,
		},
	}
	for _, p := range s.AdditionalSpotPools {
		desired[p.Name] = poolSpec{
			serverClass: p.ServerClass,
			nodeCount:   p.NodeCount,
			bidPrice:    p.BidPrice,
			autoscaling: p.Autoscaling,
			minNodes:    p.MinNodes,
			maxNodes:    p.MaxNodes,
		}
	}

	// API limit: only one pool per cloudspace may have autoscaling enabled.
	autoscalingCount := 0
	for _, d := range desired {
		if d.autoscaling {
			autoscalingCount++
		}
	}
	if autoscalingCount > 1 {
		return fmt.Errorf("only one spot node pool may have autoscaling enabled per cloudspace (API limit)")
	}

	existing, err := c.api.ListSpotNodePools(ctx, c.org, s.CloudspaceName)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to list spot node pools: %w", err)
	}
	existingNames := make(map[string]bool, len(existing))
	for _, p := range existing {
		existingNames[p.Name] = true
	}

	buildPool := func(name string, d poolSpec) spotv1.SpotNodePool {
		pool := spotv1.SpotNodePool{
			Name:        name,
			Org:         c.org,
			Cloudspace:  s.CloudspaceName,
			ServerClass: d.serverClass,
			Desired:     d.nodeCount,
			BidPrice:    d.bidPrice,
		}
		pool.Autoscaling.Enabled = d.autoscaling
		pool.Autoscaling.MinNodes = d.minNodes
		pool.Autoscaling.MaxNodes = d.maxNodes
		return pool
	}

	// Phase 1: delete pools no longer desired.
	for _, p := range existing {
		if _, wanted := desired[p.Name]; !wanted {
			if err := c.api.DeleteSpotNodePool(ctx, c.org, p.Name); err != nil && !isNotFound(err) {
				return fmt.Errorf("failed to delete spot node pool %s: %w", p.Name, err)
			}
			logrus.Infof("deleted spot node pool %s (removed from desired state)", p.Name)
		}
	}

	// Phase 2: disable autoscaling on existing pools that are losing it.
	for name, d := range desired {
		if existingNames[name] && !d.autoscaling {
			if err := c.api.UpdateSpotNodePool(ctx, c.org, buildPool(name, d)); err != nil {
				return fmt.Errorf("failed to update spot node pool %s: %w", name, err)
			}
			logrus.Infof("updated spot node pool %s", name)
		}
	}

	// Phase 3: create new pools and update pools keeping or gaining autoscaling.
	for name, d := range desired {
		if existingNames[name] && !d.autoscaling {
			continue // already handled in phase 2
		}
		pool := buildPool(name, d)
		if !existingNames[name] {
			if err := c.api.CreateSpotNodePool(ctx, c.org, pool); err != nil {
				return fmt.Errorf("failed to create spot node pool %s: %w", name, err)
			}
			logrus.Infof("created spot node pool %s", name)
		} else {
			if err := c.api.UpdateSpotNodePool(ctx, c.org, pool); err != nil {
				return fmt.Errorf("failed to update spot node pool %s: %w", name, err)
			}
			logrus.Infof("updated spot node pool %s", name)
		}
	}
	return nil
}

// reconcileOnDemandNodePool creates, updates, or removes the on-demand node pool
// so Rackspace matches the Rancher cluster configuration.
func (c *spotClient) reconcileOnDemandNodePool(ctx context.Context, s *clusterState) error {
	if !s.OnDemandEnabled {
		return c.deleteOnDemandNodePools(ctx, s)
	}

	existing, err := c.api.GetOnDemandNodePool(ctx, c.org, s.OnDemandPoolName)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to check existing on-demand pool: %w", err)
	}

	pool := spotv1.OnDemandNodePool{
		Name:                 s.OnDemandPoolName,
		Org:                  c.org,
		Cloudspace:           s.CloudspaceName,
		ServerClass:          s.OnDemandClass,
		Desired:              s.OnDemandCount,
		OnDemandPricePerHour: s.OnDemandPrice,
	}

	if existing == nil {
		if err := c.api.CreateOnDemandNodePool(ctx, c.org, pool); err != nil {
			return fmt.Errorf("failed to create on-demand node pool: %w", err)
		}
		logrus.Infof("created on-demand node pool %s", s.OnDemandPoolName)
	} else {
		if err := c.api.UpdateOnDemandNodePool(ctx, c.org, pool); err != nil {
			return fmt.Errorf("failed to update on-demand node pool: %w", err)
		}
		logrus.Infof("updated on-demand node pool %s", s.OnDemandPoolName)
	}

	return nil
}

func (c *spotClient) deleteOnDemandNodePools(ctx context.Context, s *clusterState) error {
	pools, err := c.api.ListOnDemandNodePools(ctx, c.org, s.CloudspaceName)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to list on-demand node pools: %w", err)
	}
	for _, p := range pools {
		if err := c.api.DeleteOnDemandNodePool(ctx, c.org, p.Name); err != nil && !isNotFound(err) {
			return fmt.Errorf("failed to delete on-demand node pool %s: %w", p.Name, err)
		}
		logrus.Infof("deleted on-demand node pool %s", p.Name)
	}
	return nil
}

// deleteNodePools removes all node pools for a cloudspace before deletion.
func (c *spotClient) deleteNodePools(ctx context.Context, s *clusterState) error {
	spotPools, err := c.api.ListSpotNodePools(ctx, c.org, s.CloudspaceName)
	if err != nil && !isNotFound(err) {
		return fmt.Errorf("failed to list spot node pools: %w", err)
	}
	for _, p := range spotPools {
		if err := c.api.DeleteSpotNodePool(ctx, c.org, p.Name); err != nil && !isNotFound(err) {
			return fmt.Errorf("failed to delete spot node pool %s: %w", p.Name, err)
		}
		logrus.Infof("deleted spot node pool %s", p.Name)
	}

	return c.deleteOnDemandNodePools(ctx, s)
}

// waitForCloudspace polls until the cloudspace reaches desiredStatus or the context is cancelled.
func (c *spotClient) waitForCloudspace(ctx context.Context, name, desiredStatus string, timeout time.Duration) (*spotv1.CloudSpace, error) {
	deadline := time.Now().Add(timeout)
	for {
		cs, err := c.api.GetCloudspace(ctx, c.org, name)
		if err != nil {
			return nil, fmt.Errorf("failed to poll cloudspace: %w", err)
		}

		logrus.Infof("cloudspace %s: status=%s (waiting for %s)", name, cs.Status, desiredStatus)

		if cs.Status == "Failed" || cs.Status == "Error" {
			return nil, fmt.Errorf("cloudspace %s entered error state: %s", name, cs.Status)
		}
		if cs.Status == desiredStatus {
			return cs, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out after %s waiting for cloudspace %s to reach %s (current: %s)",
				timeout, name, desiredStatus, cs.Status)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval * time.Second):
		}
	}
}

// importCloudspace reads an existing cloudspace from the Spot API and returns
// its config as a clusterState. No infrastructure changes are made.
func (c *spotClient) importCloudspace(ctx context.Context, org, cloudspaceName, token string) (*clusterState, error) {
	cs, err := c.api.GetCloudspace(ctx, org, cloudspaceName)
	if err != nil {
		if isNotFound(err) {
			return nil, fmt.Errorf("cloudspace %q not found: use create mode to provision a new cluster", cloudspaceName)
		}
		return nil, fmt.Errorf("failed to look up cloudspace %q: %w", cloudspaceName, err)
	}
	return stateFromCloudspace(cs, org, token), nil
}

// nodeTotalCount returns the sum of desired nodes across all pools.
func nodeTotalCount(cs *spotv1.CloudSpace) int64 {
	var total int64
	for _, p := range cs.SpotNodepools {
		total += int64(p.Desired)
	}
	for _, p := range cs.OnDemandNodePools {
		total += int64(p.Desired)
	}
	return total
}
