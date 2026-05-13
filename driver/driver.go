package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rancher/kontainer-engine/types"
	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/sirupsen/logrus"
)

const driverName = "rackspacespot"

// Driver implements the kontainer-engine types.Driver interface for Rackspace Spot.
type Driver struct {
	driverCapabilities types.Capabilities
}

// NewDriver returns a new Rackspace Spot cluster driver.
func NewDriver() types.Driver {
	return &Driver{
		driverCapabilities: types.Capabilities{
			Capabilities: map[int64]bool{
				types.GetVersionCapability:      true,
				types.SetVersionCapability:      true,
				types.GetClusterSizeCapability:  true,
				types.SetClusterSizeCapability:  true,
			},
		},
	}
}

// GetDriverCreateOptions returns the flags shown in the Rancher UI when creating a cluster.
func (d *Driver) GetDriverCreateOptions(ctx context.Context) (*types.DriverFlags, error) {
	flags := &types.DriverFlags{
		Options: map[string]*types.Flag{
			flagRefreshToken: {
				Type:     types.StringType,
				Usage:    "Rackspace Spot refresh token",
				Password: true,
			},
			flagOrganization: {
				Type:  types.StringType,
				Usage: "Rackspace Spot organization name",
			},
			flagRegion: {
				Type:    types.StringType,
				Usage:   "Rackspace Spot region",
				Default: &types.Default{DefaultString: defaultRegion},
			},
			flagK8sVersion: {
				Type:    types.StringType,
				Usage:   "Kubernetes version",
				Default: &types.Default{DefaultString: defaultK8sVersion},
			},
			flagCNI: {
				Type:    types.StringType,
				Usage:   "CNI plugin (calico, flannel)",
				Default: &types.Default{DefaultString: defaultCNI},
			},
			flagGPUEnabled: {
				Type:    types.BoolType,
				Usage:   "Enable GPU support on the CloudSpace",
				Default: &types.Default{DefaultBool: false},
			},
			flagPreemptionWebhook: {
				Type:  types.StringType,
				Usage: "Webhook URL called before spot nodes are preempted",
			},
			flagDeploymentType: {
				Type:  types.StringType,
				Usage: "Deployment type (e.g. spot, on-demand)",
			},
			flagSpotPoolName: {
				Type:  types.StringType,
				Usage: "Name for the spot node pool (leave blank to auto-generate a UUID)",
			},
			flagSpotServerClass: {
				Type:    types.StringType,
				Usage:   "Server class for spot nodes",
				Default: &types.Default{DefaultString: defaultSpotClass},
			},
			flagSpotNodeCount: {
				Type:    types.IntType,
				Usage:   "Desired number of spot nodes",
				Default: &types.Default{DefaultInt: defaultSpotCount},
			},
			flagSpotBidPrice: {
				Type:    types.StringType,
				Usage:   "Maximum bid price per node-hour for spot nodes",
				Default: &types.Default{DefaultString: defaultSpotBid},
			},
			flagSpotAutoscaling: {
				Type:    types.BoolType,
				Usage:   "Enable autoscaling for the spot node pool",
				Default: &types.Default{DefaultBool: false},
			},
			flagSpotMinNodes: {
				Type:    types.IntType,
				Usage:   "Minimum nodes when autoscaling is enabled",
				Default: &types.Default{DefaultInt: 1},
			},
			flagSpotMaxNodes: {
				Type:    types.IntType,
				Usage:   "Maximum nodes when autoscaling is enabled",
				Default: &types.Default{DefaultInt: 10},
			},
			flagOnDemandEnabled: {
				Type:    types.BoolType,
				Usage:   "Create an on-demand node pool alongside the spot pool",
				Default: &types.Default{DefaultBool: false},
			},
			flagOnDemandPoolName: {
				Type:  types.StringType,
				Usage: "Name for the on-demand node pool (leave blank to auto-generate a UUID)",
			},
			flagOnDemandClass: {
				Type:  types.StringType,
				Usage: "Server class for on-demand nodes",
			},
			flagOnDemandCount: {
				Type:    types.IntType,
				Usage:   "Desired number of on-demand nodes",
				Default: &types.Default{DefaultInt: 1},
			},
			flagOnDemandPrice: {
				Type:  types.StringType,
				Usage: "Maximum price per node-hour for on-demand nodes",
			},
		},
	}
	return flags, nil
}

// GetDriverUpdateOptions returns the flags available when updating a cluster.
func (d *Driver) GetDriverUpdateOptions(ctx context.Context) (*types.DriverFlags, error) {
	flags := &types.DriverFlags{
		Options: map[string]*types.Flag{
			flagK8sVersion: {
				Type:  types.StringType,
				Usage: "Kubernetes version to upgrade to",
			},
			flagSpotNodeCount: {
				Type:  types.IntType,
				Usage: "Desired number of spot nodes",
			},
			flagSpotBidPrice: {
				Type:  types.StringType,
				Usage: "Maximum bid price for spot nodes",
			},
			flagSpotAutoscaling: {
				Type:  types.BoolType,
				Usage: "Enable or disable spot pool autoscaling",
			},
			flagSpotMinNodes: {
				Type:  types.IntType,
				Usage: "Autoscaling minimum node count",
			},
			flagSpotMaxNodes: {
				Type:  types.IntType,
				Usage: "Autoscaling maximum node count",
			},
			flagOnDemandEnabled: {
				Type:  types.BoolType,
				Usage: "Enable or disable the on-demand node pool",
			},
			flagOnDemandCount: {
				Type:  types.IntType,
				Usage: "Desired number of on-demand nodes",
			},
		},
	}
	return flags, nil
}

// Create provisions a new CloudSpace and its node pools.
func (d *Driver) Create(ctx context.Context, opts *types.DriverOptions, clusterInfo *types.ClusterInfo) (retInfo *types.ClusterInfo, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("[%s] Create() panic: %v", driverName, r)
			logrus.Errorf(retErr.Error())
		}
	}()

	logrus.Infof("[%s] Create() started", driverName)

	s, err := stateFromOptions(opts)
	if err != nil {
		return nil, err
	}

	// Derive a valid CloudSpace name from the cluster name supplied by Rancher.
	rawName := opts.StringOptions["name"]
	logrus.Infof("[%s] raw cluster name from opts: %q", driverName, rawName)
	cloudspaceName, err := sanitizeResourceName(rawName)
	if err != nil {
		return nil, fmt.Errorf("invalid cluster name: %w", err)
	}
	s.CloudspaceName = cloudspaceName
	logrus.Infof("[%s] sanitized cloudspace name: %q", driverName, cloudspaceName)

	info := clusterInfo
	if info == nil {
		info = &types.ClusterInfo{}
	}

	if err := s.save(info); err != nil {
		return info, err
	}

	logrus.Infof("[%s] authenticating with Rackspace Spot API (org: %s)", driverName, s.Organization)
	client, err := newSpotClient(ctx, s.RefreshToken, s.Organization)
	if err != nil {
		return info, err
	}
	logrus.Infof("[%s] authenticated OK", driverName)

	logrus.Infof("[%s] creating cloudspace %s in org %s (region: %s, k8s: %s)",
		driverName, s.CloudspaceName, s.Organization, s.Region, s.KubernetesVersion)

	if _, err := client.ensureCloudspace(ctx, s); err != nil {
		logrus.Errorf("[%s] ensureCloudspace failed: %v", driverName, err)
		return info, err
	}
	logrus.Infof("[%s] ensureCloudspace OK", driverName)

	if err := client.ensureSpotNodePool(ctx, s); err != nil {
		logrus.Errorf("[%s] ensureSpotNodePool failed: %v", driverName, err)
		return info, err
	}
	logrus.Infof("[%s] ensureSpotNodePool OK", driverName)

	if err := client.ensureOnDemandNodePool(ctx, s); err != nil {
		logrus.Errorf("[%s] ensureOnDemandNodePool failed: %v", driverName, err)
		return info, err
	}
	logrus.Infof("[%s] ensureOnDemandNodePool OK", driverName)

	if err := s.save(info); err != nil {
		return info, err
	}

	logrus.Infof("[%s] Create() completed successfully", driverName)
	return info, nil
}

// Update modifies the node pools and/or Kubernetes version of an existing cluster.
func (d *Driver) Update(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions) (*types.ClusterInfo, error) {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return clusterInfo, err
	}

	mergeState(s, opts)

	client, err := newSpotClient(ctx, s.RefreshToken, s.Organization)
	if err != nil {
		return clusterInfo, err
	}

	logrus.Infof("[%s] updating cloudspace %s", driverName, s.CloudspaceName)

	if err := client.ensureSpotNodePool(ctx, s); err != nil {
		return clusterInfo, err
	}
	if err := client.ensureOnDemandNodePool(ctx, s); err != nil {
		return clusterInfo, err
	}

	if err := s.save(clusterInfo); err != nil {
		return clusterInfo, err
	}

	return clusterInfo, nil
}

// PostCheck waits for the CloudSpace to be ready, fetches the kubeconfig, and
// creates the Rancher service account.
func (d *Driver) PostCheck(ctx context.Context, clusterInfo *types.ClusterInfo) (*types.ClusterInfo, error) {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return clusterInfo, err
	}

	client, err := newSpotClient(ctx, s.RefreshToken, s.Organization)
	if err != nil {
		return clusterInfo, err
	}

	logrus.Infof("[%s] waiting for cloudspace %s to become ready", driverName, s.CloudspaceName)

	cs, err := client.waitForCloudspace(ctx, s.CloudspaceName, "Running",
		time.Duration(clusterReadyTimeout)*time.Minute)
	if err != nil {
		return clusterInfo, err
	}

	clusterInfo.Endpoint = cs.APIServerEndpoint
	clusterInfo.NodeCount = nodeTotalCount(cs)

	logrus.Infof("[%s] fetching kubeconfig for cloudspace %s", driverName, s.CloudspaceName)

	kubeconfig, err := client.api.GetCloudspaceConfig(ctx, s.Organization, s.CloudspaceName)
	if err != nil {
		return clusterInfo, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if err := populateClusterInfoFromKubeconfig(clusterInfo, kubeconfig); err != nil {
		return clusterInfo, err
	}

	logrus.Infof("[%s] ensuring rancher service account in cloudspace %s", driverName, s.CloudspaceName)

	if err := ensureRancherServiceAccount(ctx, kubeconfig, clusterInfo); err != nil {
		return clusterInfo, err
	}

	return clusterInfo, nil
}

// Remove deletes all node pools and then the CloudSpace itself.
func (d *Driver) Remove(ctx context.Context, clusterInfo *types.ClusterInfo) error {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return err
	}

	if s.CloudspaceName == "" {
		logrus.Warnf("[%s] no cloudspace name in state, nothing to remove", driverName)
		return nil
	}

	client, err := newSpotClient(ctx, s.RefreshToken, s.Organization)
	if err != nil {
		return err
	}

	logrus.Infof("[%s] removing node pools for cloudspace %s", driverName, s.CloudspaceName)

	if err := client.deleteNodePools(ctx, s); err != nil {
		return err
	}

	logrus.Infof("[%s] deleting cloudspace %s", driverName, s.CloudspaceName)

	if err := client.api.DeleteCloudspace(ctx, s.Organization, s.CloudspaceName); err != nil {
		if isNotFound(err) {
			logrus.Infof("[%s] cloudspace %s already gone", driverName, s.CloudspaceName)
			return nil
		}
		return fmt.Errorf("failed to delete cloudspace: %w", err)
	}

	return nil
}

// GetVersion returns the current Kubernetes version of the cluster.
func (d *Driver) GetVersion(ctx context.Context, clusterInfo *types.ClusterInfo) (*types.KubernetesVersion, error) {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return nil, err
	}
	return &types.KubernetesVersion{Version: s.KubernetesVersion}, nil
}

// SetVersion requests a Kubernetes version upgrade.
func (d *Driver) SetVersion(ctx context.Context, clusterInfo *types.ClusterInfo, version *types.KubernetesVersion) error {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return err
	}

	// Rackspace Spot manages control-plane upgrades via CloudSpace patches; currently
	// the public SDK does not expose a direct version-update call. Persist the desired
	// version so it is applied on the next Update() reconciliation.
	s.KubernetesVersion = version.Version
	return s.save(clusterInfo)
}

// GetClusterSize returns the total node count across all pools.
func (d *Driver) GetClusterSize(ctx context.Context, clusterInfo *types.ClusterInfo) (*types.NodeCount, error) {
	return &types.NodeCount{Count: clusterInfo.NodeCount}, nil
}

// SetClusterSize scales the primary spot node pool to the requested count.
func (d *Driver) SetClusterSize(ctx context.Context, clusterInfo *types.ClusterInfo, count *types.NodeCount) error {
	s, err := stateFromClusterInfo(clusterInfo)
	if err != nil {
		return err
	}

	client, err := newSpotClient(ctx, s.RefreshToken, s.Organization)
	if err != nil {
		return err
	}

	s.SpotNodeCount = int(count.Count)

	pool := spotv1.SpotNodePool{
		Name:        s.SpotPoolName,
		Org:         s.Organization,
		Cloudspace:  s.CloudspaceName,
		ServerClass: s.SpotServerClass,
		Desired:     s.SpotNodeCount,
		BidPrice:    s.SpotBidPrice,
	}

	if err := client.api.UpdateSpotNodePool(ctx, s.Organization, pool); err != nil {
		return fmt.Errorf("failed to scale spot node pool: %w", err)
	}

	clusterInfo.NodeCount = count.Count
	return s.save(clusterInfo)
}

// GetCapabilities declares which optional driver features are supported.
func (d *Driver) GetCapabilities(ctx context.Context) (*types.Capabilities, error) {
	return &d.driverCapabilities, nil
}

// GetK8SCapabilities returns Kubernetes-level capabilities for this provider.
func (d *Driver) GetK8SCapabilities(ctx context.Context, opts *types.DriverOptions) (*types.K8SCapabilities, error) {
	return &types.K8SCapabilities{
		L4LoadBalancer: &types.LoadBalancerCapabilities{
			Enabled:              true,
			Provider:             "Rackspace",
			ProtocolsSupported:   []string{"TCP", "UDP"},
			HealthCheckSupported: true,
		},
		NodePoolScalingSupported: true,
	}, nil
}

// RemoveLegacyServiceAccount is a no-op; legacy service accounts are not created.
func (d *Driver) RemoveLegacyServiceAccount(ctx context.Context, clusterInfo *types.ClusterInfo) error {
	return nil
}

// ETCDSave is not supported — control plane ETCD is fully managed by Rackspace Spot.
func (d *Driver) ETCDSave(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) error {
	return fmt.Errorf("ETCD backup is managed by Rackspace Spot and cannot be triggered via the driver")
}

// ETCDRestore is not supported — control plane ETCD is fully managed by Rackspace Spot.
func (d *Driver) ETCDRestore(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) (*types.ClusterInfo, error) {
	return nil, fmt.Errorf("ETCD restore is managed by Rackspace Spot and cannot be triggered via the driver")
}

// ETCDRemoveSnapshot is not supported — control plane ETCD is fully managed by Rackspace Spot.
func (d *Driver) ETCDRemoveSnapshot(ctx context.Context, clusterInfo *types.ClusterInfo, opts *types.DriverOptions, snapshotName string) error {
	return fmt.Errorf("ETCD snapshot management is not supported by the Rackspace Spot driver")
}

// sanitizeResourceName converts a Rancher cluster name into a name that satisfies
// Rackspace Spot's resource naming rules (lowercase, alphanumeric + hyphens, ≤ 63 chars).
func sanitizeResourceName(name string) (string, error) {
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, name)
	name = strings.Trim(name, "-")
	if len(name) > 63 {
		name = name[:63]
	}
	if err := spotv1.ValidateResourceName(name); err != nil {
		return "", err
	}
	return name, nil
}
