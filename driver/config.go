package driver

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/rancher/kontainer-engine/types"
)

const (
	flagRefreshToken      = "rackspace-spot-refresh-token"
	flagOrganization      = "rackspace-spot-organization"
	flagRegion            = "rackspace-spot-region"
	flagK8sVersion        = "kubernetes-version"
	flagCNI               = "cni"
	flagGPUEnabled        = "gpu-enabled"
	flagPreemptionWebhook = "preemption-webhook-url"
	flagDeploymentType    = "deployment-type"

	flagSpotPoolName    = "spot-node-pool-name"
	flagSpotServerClass = "spot-server-class"
	flagSpotNodeCount   = "spot-node-count"
	flagSpotBidPrice    = "spot-bid-price"

	flagSpotAutoscaling = "spot-autoscaling-enabled"
	flagSpotMinNodes    = "spot-autoscaling-min-nodes"
	flagSpotMaxNodes    = "spot-autoscaling-max-nodes"

	flagOnDemandEnabled     = "on-demand-enabled"
	flagOnDemandPoolName    = "on-demand-node-pool-name"
	flagOnDemandClass       = "on-demand-server-class"
	flagOnDemandCount       = "on-demand-node-count"
	flagOnDemandPrice       = "on-demand-price-per-hour"
	flagAdditionalSpotPools = "additional-spot-pools"

	defaultRegion     = "colo-lax-1"
	defaultK8sVersion = "1.32.9"
	defaultCNI        = "calico"
	defaultSpotClass  = "rxtx.4xlarge-mi300x"
	defaultSpotCount  = int64(3)
	defaultSpotBid    = "0.50"

	metaStateKey = "state"

	clusterReadyTimeout = 30 // minutes
	pollInterval        = 15 // seconds
)

// SpotPoolConfig describes a single spot node pool.
type SpotPoolConfig struct {
	Name        string `json:"name"`
	ServerClass string `json:"serverClass"`
	NodeCount   int    `json:"nodeCount"`
	BidPrice    string `json:"bidPrice"`
	Autoscaling bool   `json:"autoscaling"`
	MinNodes    int64  `json:"minNodes,omitempty"`
	MaxNodes    int64  `json:"maxNodes,omitempty"`
}

type clusterState struct {
	// Auth
	RefreshToken string `json:"refreshToken,omitempty"`
	Organization string `json:"organization"`

	// Cluster identity
	CloudspaceName string `json:"cloudspaceName"`
	Region         string `json:"region"`

	// Cluster config
	KubernetesVersion string `json:"kubernetesVersion"`
	CNI               string `json:"cni"`
	GPUEnabled        bool   `json:"gpuEnabled"`
	PreemptionWebhook string `json:"preemptionWebhookURL,omitempty"`
	DeploymentType    string `json:"deploymentType,omitempty"`

	// Spot node pool
	SpotPoolName    string `json:"spotPoolName"`
	SpotServerClass string `json:"spotServerClass"`
	SpotNodeCount   int    `json:"spotNodeCount"`
	SpotBidPrice    string `json:"spotBidPrice"`

	// Spot autoscaling
	SpotAutoscaling bool  `json:"spotAutoscaling"`
	SpotMinNodes    int64 `json:"spotMinNodes,omitempty"`
	SpotMaxNodes    int64 `json:"spotMaxNodes,omitempty"`

	// Additional spot node pools (beyond the primary)
	AdditionalSpotPools []SpotPoolConfig `json:"additionalSpotPools,omitempty"`

	// Optional on-demand node pool
	OnDemandEnabled  bool   `json:"onDemandEnabled"`
	OnDemandPoolName string `json:"onDemandPoolName,omitempty"`
	OnDemandClass    string `json:"onDemandServerClass,omitempty"`
	OnDemandCount    int    `json:"onDemandNodeCount,omitempty"`
	OnDemandPrice    string `json:"onDemandPricePerHour,omitempty"`
}

func stateFromOptions(opts *types.DriverOptions) (*clusterState, error) {
	s := &clusterState{
		RefreshToken:      getStringOption(opts, flagRefreshToken, "rackspaceSpotRefreshToken"),
		Organization:      getStringOption(opts, flagOrganization, "rackspaceSpotOrganization"),
		Region:            getStringOption(opts, flagRegion, "rackspaceSpotRegion"),
		KubernetesVersion: getStringOption(opts, flagK8sVersion, "kubernetesVersion"),
		CNI:               getStringOption(opts, flagCNI, "cni"),
		GPUEnabled:        getBoolOption(opts, flagGPUEnabled, "gpuEnabled"),
		PreemptionWebhook: getStringOption(opts, flagPreemptionWebhook, "preemptionWebhookUrl"),
		DeploymentType:    getStringOption(opts, flagDeploymentType, "deploymentType"),
		SpotPoolName:      getStringOption(opts, flagSpotPoolName, "spotNodePoolName"),
		SpotServerClass:   getStringOption(opts, flagSpotServerClass, "spotServerClass"),
		SpotBidPrice:      getStringOption(opts, flagSpotBidPrice, "spotBidPrice"),
		SpotAutoscaling:   getBoolOption(opts, flagSpotAutoscaling, "spotAutoscalingEnabled"),
		SpotMinNodes:      getIntOption(opts, flagSpotMinNodes, "spotAutoscalingMinNodes"),
		SpotMaxNodes:      getIntOption(opts, flagSpotMaxNodes, "spotAutoscalingMaxNodes"),
		OnDemandEnabled:   getBoolOption(opts, flagOnDemandEnabled, "onDemandEnabled"),
		OnDemandPoolName:  getStringOption(opts, flagOnDemandPoolName, "onDemandNodePoolName"),
		OnDemandClass:     getStringOption(opts, flagOnDemandClass, "onDemandServerClass"),
		OnDemandPrice:     getStringOption(opts, flagOnDemandPrice, "onDemandPricePerHour"),
	}

	if n, ok := lookupIntOption(opts, flagSpotNodeCount, "spotNodeCount"); ok {
		s.SpotNodeCount = int(n)
	}
	if n, ok := lookupIntOption(opts, flagOnDemandCount, "onDemandNodeCount"); ok {
		s.OnDemandCount = int(n)
	}

	if raw := getStringOption(opts, flagAdditionalSpotPools, "additionalSpotPools"); raw != "" {
		if err := json.Unmarshal([]byte(raw), &s.AdditionalSpotPools); err != nil {
			return nil, fmt.Errorf("failed to parse additional spot pools: %w", err)
		}
	}

	applyDefaults(s)

	if err := validate(s); err != nil {
		return nil, err
	}

	return s, nil
}

// k8sVersionMap normalises short-form versions (e.g. "1.33") to the full
// semver the API requires.
var k8sVersionMap = map[string]string{
	"1.29": "1.29.6",
	"1.30": "1.30.10",
	"1.31": "1.31.1",
	"1.32": "1.32.9",
	"1.33": "1.33.0",
	"1.28": "1.32.9", // old default → current stable
}

func applyDefaults(s *clusterState) {
	if s.Region == "" {
		s.Region = defaultRegion
	}
	if s.KubernetesVersion == "" {
		s.KubernetesVersion = defaultK8sVersion
	}
	// Normalise "1.X" → "1.X.Y" if the API requires full semver.
	if !strings.Contains(s.KubernetesVersion, ".") || len(strings.Split(s.KubernetesVersion, ".")) == 2 {
		if full, ok := k8sVersionMap[s.KubernetesVersion]; ok {
			s.KubernetesVersion = full
		}
	}
	if s.CNI == "" {
		s.CNI = defaultCNI
	}
	if s.SpotPoolName == "" || uuid.Validate(s.SpotPoolName) != nil {
		s.SpotPoolName = uuid.New().String()
	}
	if s.SpotServerClass == "" {
		s.SpotServerClass = defaultSpotClass
	}
	if s.SpotNodeCount == 0 {
		s.SpotNodeCount = int(defaultSpotCount)
	}
	if s.SpotBidPrice == "" {
		s.SpotBidPrice = defaultSpotBid
	}
	if s.OnDemandEnabled && (s.OnDemandPoolName == "" || uuid.Validate(s.OnDemandPoolName) != nil) {
		s.OnDemandPoolName = uuid.New().String()
	}

	for i := range s.AdditionalSpotPools {
		p := &s.AdditionalSpotPools[i]
		if p.Name == "" || uuid.Validate(p.Name) != nil {
			p.Name = uuid.New().String()
		}
		if p.BidPrice == "" {
			p.BidPrice = defaultSpotBid
		}
		if p.NodeCount == 0 {
			p.NodeCount = int(defaultSpotCount)
		}
	}
}

func validate(s *clusterState) error {
	if s.RefreshToken == "" {
		return fmt.Errorf("%s is required", flagRefreshToken)
	}
	if s.Organization == "" {
		return fmt.Errorf("%s is required", flagOrganization)
	}
	return nil
}

func stateFromClusterInfo(info *types.ClusterInfo) (*clusterState, error) {
	s := &clusterState{}
	if info == nil {
		return s, nil
	}
	if info.Metadata == nil || info.Metadata[metaStateKey] == "" {
		s.RefreshToken = info.Password
		return s, nil
	}
	if err := json.Unmarshal([]byte(info.Metadata[metaStateKey]), s); err != nil {
		return nil, fmt.Errorf("failed to deserialize cluster state: %w", err)
	}
	if s.RefreshToken == "" {
		s.RefreshToken = info.Password
	}
	return s, nil
}

func (s *clusterState) save(info *types.ClusterInfo) error {
	persisted := *s
	persisted.RefreshToken = ""
	data, err := json.Marshal(&persisted)
	if err != nil {
		return fmt.Errorf("failed to serialize cluster state: %w", err)
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	info.Metadata[metaStateKey] = string(data)
	info.Password = s.RefreshToken
	return nil
}

// mergeState overwrites the mutable fields from opts into an existing state,
// preserving identity fields (CloudspaceName, Organization) that cannot change.
func mergeState(existing *clusterState, opts *types.DriverOptions) {
	if v := getStringOption(opts, flagRefreshToken, "rackspaceSpotRefreshToken"); v != "" {
		existing.RefreshToken = v
	}
	if v := getStringOption(opts, flagK8sVersion, "kubernetesVersion"); v != "" {
		existing.KubernetesVersion = v
	}
	if v := getStringOption(opts, flagSpotBidPrice, "spotBidPrice"); v != "" {
		existing.SpotBidPrice = v
	}
	if n, ok := lookupIntOption(opts, flagSpotNodeCount, "spotNodeCount"); ok {
		existing.SpotNodeCount = int(n)
	}
	if v, ok := lookupBoolOption(opts, flagSpotAutoscaling, "spotAutoscalingEnabled"); ok {
		existing.SpotAutoscaling = v
	}
	if n, ok := lookupIntOption(opts, flagSpotMinNodes, "spotAutoscalingMinNodes"); ok {
		existing.SpotMinNodes = n
	}
	if n, ok := lookupIntOption(opts, flagSpotMaxNodes, "spotAutoscalingMaxNodes"); ok {
		existing.SpotMaxNodes = n
	}
	if v, ok := lookupBoolOption(opts, flagOnDemandEnabled, "onDemandEnabled"); ok {
		existing.OnDemandEnabled = v
	}
	if n, ok := lookupIntOption(opts, flagOnDemandCount, "onDemandNodeCount"); ok {
		existing.OnDemandCount = int(n)
	}

	if raw := getStringOption(opts, flagAdditionalSpotPools, "additionalSpotPools"); raw != "" {
		var pools []SpotPoolConfig
		if err := json.Unmarshal([]byte(raw), &pools); err == nil {
			existing.AdditionalSpotPools = pools
		}
	}
}

func getStringOption(opts *types.DriverOptions, keys ...string) string {
	value, _ := lookupStringOption(opts, keys...)
	return value
}

func getBoolOption(opts *types.DriverOptions, keys ...string) bool {
	value, _ := lookupBoolOption(opts, keys...)
	return value
}

func getIntOption(opts *types.DriverOptions, keys ...string) int64 {
	value, _ := lookupIntOption(opts, keys...)
	return value
}

func lookupStringOption(opts *types.DriverOptions, keys ...string) (string, bool) {
	if opts == nil {
		return "", false
	}
	for _, key := range keys {
		if value, ok := opts.StringOptions[key]; ok {
			return value, true
		}
	}
	return "", false
}

func lookupBoolOption(opts *types.DriverOptions, keys ...string) (bool, bool) {
	if opts == nil {
		return false, false
	}
	for _, key := range keys {
		if value, ok := opts.BoolOptions[key]; ok {
			return value, true
		}
	}
	return false, false
}

func lookupIntOption(opts *types.DriverOptions, keys ...string) (int64, bool) {
	if opts == nil {
		return 0, false
	}
	for _, key := range keys {
		if value, ok := opts.IntOptions[key]; ok {
			return value, true
		}
	}
	return 0, false
}
