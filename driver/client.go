package driver

import (
	"context"
	"fmt"
	"time"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/sirupsen/logrus"
)

type spotClient struct {
	api *spotv1.RackspaceSpotClient
	org string
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

	return &spotClient{api: c, org: org}, nil
}

// ensureCloudspace creates a CloudSpace or returns the existing one if already created.
func (c *spotClient) ensureCloudspace(ctx context.Context, s *clusterState) (*spotv1.CloudSpace, error) {
	existing, err := c.api.GetCloudspace(ctx, c.org, s.CloudspaceName)
	if err != nil && !spotv1.IsNotFound(err) {
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

	if err := c.api.CreateCloudspace(ctx, cs); err != nil {
		return nil, fmt.Errorf("failed to create cloudspace: %w", err)
	}

	return c.api.GetCloudspace(ctx, c.org, s.CloudspaceName)
}

// ensureSpotNodePool creates or updates the spot node pool.
func (c *spotClient) ensureSpotNodePool(ctx context.Context, s *clusterState) error {
	existing, err := c.api.GetSpotNodePool(ctx, c.org, s.SpotPoolName)
	if err != nil && !spotv1.IsNotFound(err) {
		return fmt.Errorf("failed to check existing spot pool: %w", err)
	}

	pool := spotv1.SpotNodePool{
		Name:        s.SpotPoolName,
		Org:         c.org,
		Cloudspace:  s.CloudspaceName,
		ServerClass: s.SpotServerClass,
		Desired:     s.SpotNodeCount,
		BidPrice:    s.SpotBidPrice,
	}
	pool.Autoscaling.Enabled = s.SpotAutoscaling
	pool.Autoscaling.MinNodes = s.SpotMinNodes
	pool.Autoscaling.MaxNodes = s.SpotMaxNodes

	if existing == nil {
		if err := c.api.CreateSpotNodePool(ctx, c.org, pool); err != nil {
			return fmt.Errorf("failed to create spot node pool: %w", err)
		}
		logrus.Infof("created spot node pool %s", s.SpotPoolName)
	} else {
		if err := c.api.UpdateSpotNodePool(ctx, c.org, pool); err != nil {
			return fmt.Errorf("failed to update spot node pool: %w", err)
		}
		logrus.Infof("updated spot node pool %s", s.SpotPoolName)
	}

	return nil
}

// ensureOnDemandNodePool creates or updates the on-demand node pool when enabled.
func (c *spotClient) ensureOnDemandNodePool(ctx context.Context, s *clusterState) error {
	if !s.OnDemandEnabled {
		return nil
	}

	existing, err := c.api.GetOnDemandNodePool(ctx, c.org, s.OnDemandPoolName)
	if err != nil && !spotv1.IsNotFound(err) {
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

// deleteNodePools removes all node pools for a cloudspace before deletion.
func (c *spotClient) deleteNodePools(ctx context.Context, s *clusterState) error {
	spotPools, err := c.api.ListSpotNodePools(ctx, c.org, s.CloudspaceName)
	if err != nil && !spotv1.IsNotFound(err) {
		return fmt.Errorf("failed to list spot node pools: %w", err)
	}
	for _, p := range spotPools {
		if err := c.api.DeleteSpotNodePool(ctx, c.org, p.Name); err != nil && !spotv1.IsNotFound(err) {
			return fmt.Errorf("failed to delete spot node pool %s: %w", p.Name, err)
		}
		logrus.Infof("deleted spot node pool %s", p.Name)
	}

	odmPools, err := c.api.ListOnDemandNodePools(ctx, c.org, s.CloudspaceName)
	if err != nil && !spotv1.IsNotFound(err) {
		return fmt.Errorf("failed to list on-demand node pools: %w", err)
	}
	for _, p := range odmPools {
		if err := c.api.DeleteOnDemandNodePool(ctx, c.org, p.Name); err != nil && !spotv1.IsNotFound(err) {
			return fmt.Errorf("failed to delete on-demand node pool %s: %w", p.Name, err)
		}
		logrus.Infof("deleted on-demand node pool %s", p.Name)
	}

	return nil
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
