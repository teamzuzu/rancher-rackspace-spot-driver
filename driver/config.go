package driver

import (
	"encoding/json"
	"fmt"

	"github.com/rancher/kontainer-engine/drivers/options"
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

	flagOnDemandEnabled  = "on-demand-enabled"
	flagOnDemandPoolName = "on-demand-node-pool-name"
	flagOnDemandClass    = "on-demand-server-class"
	flagOnDemandCount    = "on-demand-node-count"
	flagOnDemandPrice    = "on-demand-price-per-hour"

	defaultRegion      = "colo-lax-1"
	defaultK8sVersion  = "1.32.9"
	defaultCNI         = "calico"
	defaultSpotPool    = "spot-pool"
	defaultSpotClass   = "rxtx.4xlarge-mi300x"
	defaultSpotCount   = int64(3)
	defaultSpotBid     = "0.50"
	defaultOnDemandPool = "on-demand-pool"

	metaStateKey = "state"

	clusterReadyTimeout = 30 // minutes
	pollInterval        = 15 // seconds
)

type clusterState struct {
	// Auth
	RefreshToken string `json:"refreshToken"`
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

	// Optional on-demand node pool
	OnDemandEnabled  bool   `json:"onDemandEnabled"`
	OnDemandPoolName string `json:"onDemandPoolName,omitempty"`
	OnDemandClass    string `json:"onDemandServerClass,omitempty"`
	OnDemandCount    int    `json:"onDemandNodeCount,omitempty"`
	OnDemandPrice    string `json:"onDemandPricePerHour,omitempty"`
}

func stateFromOptions(opts *types.DriverOptions) (*clusterState, error) {
	// Rancher stores genericEngineConfig keys in camelCase; we accept both forms.
	str := func(kebab, camel string) string {
		return options.GetValueFromDriverOptions(opts, types.StringType, kebab, camel).(string)
	}
	bl := func(kebab, camel string) bool {
		return options.GetValueFromDriverOptions(opts, types.BoolType, kebab, camel).(bool)
	}
	num := func(kebab, camel string) int64 {
		return options.GetValueFromDriverOptions(opts, types.IntType, kebab, camel).(int64)
	}

	s := &clusterState{
		RefreshToken:      str(flagRefreshToken, "rackspaceSpotRefreshToken"),
		Organization:      str(flagOrganization, "rackspaceSpotOrganization"),
		Region:            str(flagRegion, "rackspaceSpotRegion"),
		KubernetesVersion: str(flagK8sVersion, "kubernetesVersion"),
		CNI:               str(flagCNI, "cni"),
		GPUEnabled:        bl(flagGPUEnabled, "gpuEnabled"),
		PreemptionWebhook: str(flagPreemptionWebhook, "preemptionWebhookUrl"),
		DeploymentType:    str(flagDeploymentType, "deploymentType"),
		SpotPoolName:      str(flagSpotPoolName, "spotNodePoolName"),
		SpotServerClass:   str(flagSpotServerClass, "spotServerClass"),
		SpotBidPrice:      str(flagSpotBidPrice, "spotBidPrice"),
		SpotAutoscaling:   bl(flagSpotAutoscaling, "spotAutoscalingEnabled"),
		SpotMinNodes:      num(flagSpotMinNodes, "spotAutoscalingMinNodes"),
		SpotMaxNodes:      num(flagSpotMaxNodes, "spotAutoscalingMaxNodes"),
		OnDemandEnabled:   bl(flagOnDemandEnabled, "onDemandEnabled"),
		OnDemandPoolName:  str(flagOnDemandPoolName, "onDemandNodePoolName"),
		OnDemandClass:     str(flagOnDemandClass, "onDemandServerClass"),
		OnDemandPrice:     str(flagOnDemandPrice, "onDemandPricePerHour"),
	}

	if n := num(flagSpotNodeCount, "spotNodeCount"); n > 0 {
		s.SpotNodeCount = int(n)
	}
	if n := num(flagOnDemandCount, "onDemandNodeCount"); n > 0 {
		s.OnDemandCount = int(n)
	}

	applyDefaults(s)

	if err := validate(s); err != nil {
		return nil, err
	}

	return s, nil
}

func applyDefaults(s *clusterState) {
	if s.Region == "" {
		s.Region = defaultRegion
	}
	if s.KubernetesVersion == "" {
		s.KubernetesVersion = defaultK8sVersion
	}
	if s.CNI == "" {
		s.CNI = defaultCNI
	}
	if s.SpotPoolName == "" {
		s.SpotPoolName = defaultSpotPool
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
	if s.OnDemandEnabled && s.OnDemandPoolName == "" {
		s.OnDemandPoolName = defaultOnDemandPool
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
	if info.Metadata == nil || info.Metadata[metaStateKey] == "" {
		return s, nil
	}
	if err := json.Unmarshal([]byte(info.Metadata[metaStateKey]), s); err != nil {
		return nil, fmt.Errorf("failed to deserialize cluster state: %w", err)
	}
	return s, nil
}

func (s *clusterState) save(info *types.ClusterInfo) error {
	data, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("failed to serialize cluster state: %w", err)
	}
	if info.Metadata == nil {
		info.Metadata = map[string]string{}
	}
	info.Metadata[metaStateKey] = string(data)
	return nil
}

// mergeState overwrites the mutable fields from opts into an existing state,
// preserving identity fields (CloudspaceName, Organization) that cannot change.
func mergeState(existing *clusterState, opts *types.DriverOptions) {
	str := func(kebab, camel string) string {
		return options.GetValueFromDriverOptions(opts, types.StringType, kebab, camel).(string)
	}
	bl := func(kebab, camel string) bool {
		return options.GetValueFromDriverOptions(opts, types.BoolType, kebab, camel).(bool)
	}
	num := func(kebab, camel string) int64 {
		return options.GetValueFromDriverOptions(opts, types.IntType, kebab, camel).(int64)
	}

	if v := str(flagK8sVersion, "kubernetesVersion"); v != "" {
		existing.KubernetesVersion = v
	}
	if v := str(flagSpotBidPrice, "spotBidPrice"); v != "" {
		existing.SpotBidPrice = v
	}
	if n := num(flagSpotNodeCount, "spotNodeCount"); n > 0 {
		existing.SpotNodeCount = int(n)
	}
	if v := bl(flagSpotAutoscaling, "spotAutoscalingEnabled"); v {
		existing.SpotAutoscaling = v
	}
	if n := num(flagSpotMinNodes, "spotAutoscalingMinNodes"); n > 0 {
		existing.SpotMinNodes = n
	}
	if n := num(flagSpotMaxNodes, "spotAutoscalingMaxNodes"); n > 0 {
		existing.SpotMaxNodes = n
	}
	if v := bl(flagOnDemandEnabled, "onDemandEnabled"); v {
		existing.OnDemandEnabled = v
	}
	if n := num(flagOnDemandCount, "onDemandNodeCount"); n > 0 {
		existing.OnDemandCount = int(n)
	}
}
