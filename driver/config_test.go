package driver

import (
	"strings"
	"testing"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
	"github.com/rancher/kontainer-engine/types"
)

func TestSaveRedactsRefreshTokenFromMetadata(t *testing.T) {
	state := &clusterState{
		RefreshToken:      "refresh-token",
		Organization:      "org",
		CloudspaceName:    "cluster",
		Region:            "test-region",
		KubernetesVersion: defaultK8sVersion,
		CNI:               defaultCNI,
		SpotPoolName:      "pool",
		SpotServerClass:   defaultSpotClass,
		SpotNodeCount:     int(defaultSpotCount),
		SpotBidPrice:      defaultSpotBid,
	}
	info := &types.ClusterInfo{}

	if err := state.save(info); err != nil {
		t.Fatalf("save() error = %v", err)
	}

	if info.Password != "refresh-token" {
		t.Fatalf("Password = %q, want refresh token", info.Password)
	}
	if strings.Contains(info.Metadata[metaStateKey], "refresh-token") {
		t.Fatalf("metadata contains refresh token: %s", info.Metadata[metaStateKey])
	}
}

func TestStateFromClusterInfoSupportsLegacyMetadataToken(t *testing.T) {
	info := &types.ClusterInfo{
		Metadata: map[string]string{
			metaStateKey: `{"refreshToken":"legacy-token","organization":"org"}`,
		},
	}

	state, err := stateFromClusterInfo(info)
	if err != nil {
		t.Fatalf("stateFromClusterInfo() error = %v", err)
	}

	if state.RefreshToken != "legacy-token" {
		t.Fatalf("RefreshToken = %q, want legacy-token", state.RefreshToken)
	}
}

func TestStateFromClusterInfoReadsTokenFromPassword(t *testing.T) {
	info := &types.ClusterInfo{
		Password: "stored-token",
		Metadata: map[string]string{
			metaStateKey: `{"organization":"org"}`,
		},
	}

	state, err := stateFromClusterInfo(info)
	if err != nil {
		t.Fatalf("stateFromClusterInfo() error = %v", err)
	}

	if state.RefreshToken != "stored-token" {
		t.Fatalf("RefreshToken = %q, want stored-token", state.RefreshToken)
	}
}

func TestMergeStateAppliesFalseAndZeroValues(t *testing.T) {
	state := &clusterState{
		RefreshToken:        "old-token",
		SpotAutoscaling:     true,
		SpotNodeCount:       3,
		SpotMinNodes:        1,
		SpotMaxNodes:        10,
		OnDemandEnabled:     true,
		OnDemandCount:       2,
		AdditionalSpotPools: []SpotPoolConfig{{Name: "old"}},
	}
	opts := &types.DriverOptions{
		BoolOptions: map[string]bool{
			"spotAutoscalingEnabled": false,
			"onDemandEnabled":        false,
		},
		IntOptions: map[string]int64{
			"spotNodeCount":           0,
			"spotAutoscalingMinNodes": 0,
			"spotAutoscalingMaxNodes": 0,
			"onDemandNodeCount":       0,
		},
		StringOptions: map[string]string{
			"rackspaceSpotRefreshToken": "new-token",
			"additionalSpotPools":       "[]",
		},
	}

	mergeState(state, opts)

	if state.RefreshToken != "new-token" {
		t.Fatalf("RefreshToken = %q, want new-token", state.RefreshToken)
	}
	if state.SpotAutoscaling {
		t.Fatal("SpotAutoscaling = true, want false")
	}
	if state.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = true, want false")
	}
	if state.SpotNodeCount != 0 || state.SpotMinNodes != 0 || state.SpotMaxNodes != 0 || state.OnDemandCount != 0 {
		t.Fatalf("zero values not applied: %+v", state)
	}
	if len(state.AdditionalSpotPools) != 0 {
		t.Fatalf("AdditionalSpotPools len = %d, want 0", len(state.AdditionalSpotPools))
	}
}

// TestSaveDoesNotClearPasswordWhenTokenEmpty verifies that save() does not
// overwrite info.Password with an empty string when RefreshToken is unset.
// This prevents token loss in SetClusterSize/SetVersion when Rancher does not
// round-trip ClusterInfo.Password back to the driver.
func TestSaveDoesNotClearPasswordWhenTokenEmpty(t *testing.T) {
	state := &clusterState{
		RefreshToken: "", // token not loaded (e.g. empty state from SetClusterSize)
		Organization: "org",
	}
	info := &types.ClusterInfo{
		Password: "pre-existing-token",
		Metadata: map[string]string{},
	}

	if err := state.save(info); err != nil {
		t.Fatalf("save() error = %v", err)
	}

	if info.Password != "pre-existing-token" {
		t.Fatalf("save() overwrote Password with empty token: got %q", info.Password)
	}
}

// TestMergeStatePreservesTokenWhenAbsentFromOpts verifies that mergeState does
// not clear an existing RefreshToken when the token key is absent from opts.
func TestMergeStatePreservesTokenWhenAbsentFromOpts(t *testing.T) {
	state := &clusterState{RefreshToken: "existing-token"}
	opts := &types.DriverOptions{
		StringOptions: map[string]string{}, // token not present
		BoolOptions:   map[string]bool{},
		IntOptions:    map[string]int64{},
	}

	mergeState(state, opts)

	if state.RefreshToken != "existing-token" {
		t.Fatalf("mergeState clobbered RefreshToken: got %q", state.RefreshToken)
	}
}

func TestValidateRejectsMultipleAutoscalingPools(t *testing.T) {
	state := &clusterState{
		RefreshToken:    "token",
		Organization:   "org",
		Region:          "us-east-1",
		SpotAutoscaling: true,
		AdditionalSpotPools: []SpotPoolConfig{
			{Autoscaling: true},
		},
	}
	if err := validate(state); err == nil {
		t.Fatal("validate() should return error when multiple pools have autoscaling enabled")
	}
}

func TestStateFromOptionsAppliesDefaults(t *testing.T) {
	opts := &types.DriverOptions{
		StringOptions: map[string]string{
			"rackspaceSpotRefreshToken": "token",
			"rackspaceSpotOrganization": "org",
			"rackspaceSpotRegion":       "us-east-1",
		},
		BoolOptions: map[string]bool{},
		IntOptions:  map[string]int64{},
	}

	state, err := stateFromOptions(opts)
	if err != nil {
		t.Fatalf("stateFromOptions() error = %v", err)
	}

	if state.Region != "us-east-1" {
		t.Fatalf("Region = %q, want us-east-1", state.Region)
	}
	if state.KubernetesVersion != defaultK8sVersion {
		t.Fatalf("KubernetesVersion = %q, want %q", state.KubernetesVersion, defaultK8sVersion)
	}
	if state.SpotServerClass != defaultSpotClass {
		t.Fatalf("SpotServerClass = %q, want %q", state.SpotServerClass, defaultSpotClass)
	}
	if state.SpotNodeCount != int(defaultSpotCount) {
		t.Fatalf("SpotNodeCount = %d, want %d", state.SpotNodeCount, defaultSpotCount)
	}
}

func TestStateFromCloudspace_fullConfig(t *testing.T) {
	p0 := &spotv1.SpotNodePool{
		Name:        "pool-a",
		ServerClass: "gp.vs1.medium-iad",
		Desired:     3,
		BidPrice:    "0.05",
	}
	p0.Autoscaling.Enabled  = true
	p0.Autoscaling.MinNodes = 2
	p0.Autoscaling.MaxNodes = 8

	p1 := &spotv1.SpotNodePool{
		Name:        "pool-b",
		ServerClass: "gp.vs1.large-iad",
		Desired:     2,
		BidPrice:    "0.08",
	}

	od := &spotv1.OnDemandNodePool{
		Name:                 "od-pool",
		ServerClass:          "gp.vs1.large-iad",
		Desired:              1,
		OnDemandPricePerHour: "0.50",
	}

	cs := &spotv1.CloudSpace{
		Name:                 "my-cluster",
		Region:               "us-east-iad-1",
		KubernetesVersion:    "1.33.0",
		CNI:                  "calico",
		GpuEnabled:           true,
		PreemptionWebhookURL: "https://example.com/webhook",
		DeploymentType:       "spot",
		SpotNodepools:        []*spotv1.SpotNodePool{p0, p1},
		OnDemandNodePools:    []*spotv1.OnDemandNodePool{od},
	}

	s := stateFromCloudspace(cs, "my-org", "tok")

	if s.Organization != "my-org" {
		t.Fatalf("Organization = %q, want my-org", s.Organization)
	}
	if s.RefreshToken != "tok" {
		t.Fatalf("RefreshToken = %q, want tok", s.RefreshToken)
	}
	if s.CloudspaceName != "my-cluster" {
		t.Fatalf("CloudspaceName = %q, want my-cluster", s.CloudspaceName)
	}
	if s.Region != "us-east-iad-1" {
		t.Fatalf("Region = %q, want us-east-iad-1", s.Region)
	}
	if s.KubernetesVersion != "1.33.0" {
		t.Fatalf("KubernetesVersion = %q, want 1.33.0", s.KubernetesVersion)
	}
	if s.CNI != "calico" {
		t.Fatalf("CNI = %q, want calico", s.CNI)
	}
	if !s.GPUEnabled {
		t.Fatal("GPUEnabled = false, want true")
	}
	if s.PreemptionWebhook != "https://example.com/webhook" {
		t.Fatalf("PreemptionWebhook = %q", s.PreemptionWebhook)
	}
	if s.DeploymentType != "spot" {
		t.Fatalf("DeploymentType = %q, want spot", s.DeploymentType)
	}
	if s.SpotPoolName != "pool-a" {
		t.Fatalf("SpotPoolName = %q, want pool-a", s.SpotPoolName)
	}
	if s.SpotServerClass != "gp.vs1.medium-iad" {
		t.Fatalf("SpotServerClass = %q", s.SpotServerClass)
	}
	if s.SpotNodeCount != 3 {
		t.Fatalf("SpotNodeCount = %d, want 3", s.SpotNodeCount)
	}
	if s.SpotBidPrice != "0.05" {
		t.Fatalf("SpotBidPrice = %q, want 0.05", s.SpotBidPrice)
	}
	if !s.SpotAutoscaling {
		t.Fatal("SpotAutoscaling = false, want true")
	}
	if s.SpotMinNodes != 2 {
		t.Fatalf("SpotMinNodes = %d, want 2", s.SpotMinNodes)
	}
	if s.SpotMaxNodes != 8 {
		t.Fatalf("SpotMaxNodes = %d, want 8", s.SpotMaxNodes)
	}
	if len(s.AdditionalSpotPools) != 1 {
		t.Fatalf("len(AdditionalSpotPools) = %d, want 1", len(s.AdditionalSpotPools))
	}
	p := s.AdditionalSpotPools[0]
	if p.Name != "pool-b" || p.ServerClass != "gp.vs1.large-iad" || p.NodeCount != 2 || p.BidPrice != "0.08" {
		t.Fatalf("AdditionalSpotPools[0] = %+v", p)
	}
	if !s.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = false, want true")
	}
	if s.OnDemandPoolName != "od-pool" {
		t.Fatalf("OnDemandPoolName = %q, want od-pool", s.OnDemandPoolName)
	}
	if s.OnDemandClass != "gp.vs1.large-iad" {
		t.Fatalf("OnDemandClass = %q", s.OnDemandClass)
	}
	if s.OnDemandCount != 1 {
		t.Fatalf("OnDemandCount = %d, want 1", s.OnDemandCount)
	}
	if s.OnDemandPrice != "0.50" {
		t.Fatalf("OnDemandPrice = %q, want 0.50", s.OnDemandPrice)
	}
}

func TestStateFromCloudspace_noOnDemand(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:          "cs",
		Region:        "us-east-iad-1",
		SpotNodepools: []*spotv1.SpotNodePool{{Name: "pool-a", Desired: 3}},
	}
	s := stateFromCloudspace(cs, "org", "tok")
	if s.OnDemandEnabled {
		t.Fatal("OnDemandEnabled = true, want false")
	}
	if s.OnDemandPoolName != "" || s.OnDemandClass != "" || s.OnDemandCount != 0 {
		t.Fatalf("unexpected on-demand fields: poolName=%q class=%q count=%d",
			s.OnDemandPoolName, s.OnDemandClass, s.OnDemandCount)
	}
}

func TestStateFromCloudspace_singleSpotPool(t *testing.T) {
	cs := &spotv1.CloudSpace{
		Name:          "cs",
		Region:        "us-east-iad-1",
		SpotNodepools: []*spotv1.SpotNodePool{{Name: "pool-a", Desired: 2}},
	}
	s := stateFromCloudspace(cs, "org", "tok")
	if s.SpotPoolName != "pool-a" {
		t.Fatalf("SpotPoolName = %q, want pool-a", s.SpotPoolName)
	}
	if len(s.AdditionalSpotPools) != 0 {
		t.Fatalf("len(AdditionalSpotPools) = %d, want 0", len(s.AdditionalSpotPools))
	}
}
