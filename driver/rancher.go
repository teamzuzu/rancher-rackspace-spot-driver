package driver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	k8stypes "k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

var rancherClusterGVR = schema.GroupVersionResource{
	Group:    "management.cattle.io",
	Version:  "v3",
	Resource: "clusters",
}

// syncGenericEngineConfig patches spec.genericEngineConfig on the Rancher cluster
// object so the edit form shows the actual pool state rather than form defaults.
// Called non-fatally after a successful import PostCheck.
// Uses in-cluster config (the driver subprocess runs inside the Rancher pod).
func syncGenericEngineConfig(ctx context.Context, s *clusterState) {
	if s.RancherClusterID == "" {
		logrus.Warnf("[%s] syncGenericEngineConfig: no RancherClusterID in state, skipping", driverName)
		return
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		logrus.Warnf("[%s] syncGenericEngineConfig: in-cluster config unavailable (non-fatal): %v", driverName, err)
		return
	}

	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		logrus.Warnf("[%s] syncGenericEngineConfig: dynamic client failed (non-fatal): %v", driverName, err)
		return
	}

	additionalPools := "[]"
	if len(s.AdditionalSpotPools) > 0 {
		if data, merr := json.Marshal(s.AdditionalSpotPools); merr == nil {
			additionalPools = string(data)
		}
	}

	engineConfig := map[string]interface{}{
		"spotNodePoolName":        s.SpotPoolName,
		"spotServerClass":         s.SpotServerClass,
		"spotNodeCount":           int64(s.SpotNodeCount),
		"spotBidPrice":            s.SpotBidPrice,
		"spotAutoscalingEnabled":  s.SpotAutoscaling,
		"spotAutoscalingMinNodes": s.SpotMinNodes,
		"spotAutoscalingMaxNodes": s.SpotMaxNodes,
		"onDemandEnabled":         s.OnDemandEnabled,
		"onDemandNodePoolName":    s.OnDemandPoolName,
		"onDemandServerClass":     s.OnDemandClass,
		"onDemandNodeCount":       int64(s.OnDemandCount),
		"onDemandPricePerHour":    s.OnDemandPrice,
		"additionalSpotPools":     additionalPools,
		"rackspaceSpotRegion":     s.Region,
		"kubernetesVersion":       s.KubernetesVersion,
		"cni":                     s.CNI,
		"gpuEnabled":              s.GPUEnabled,
		"deploymentType":          s.DeploymentType,
		"preemptionWebhookUrl":    s.PreemptionWebhook,
	}

	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"genericEngineConfig": engineConfig,
		},
	}

	patchData, err := json.Marshal(patch)
	if err != nil {
		logrus.Warnf("[%s] syncGenericEngineConfig: marshal failed (non-fatal): %v", driverName, err)
		return
	}

	if _, err := dc.Resource(rancherClusterGVR).Patch(ctx, s.RancherClusterID,
		k8stypes.MergePatchType, patchData, metav1.PatchOptions{}); err != nil {
		logrus.Warnf("[%s] syncGenericEngineConfig: patch failed (non-fatal): %v", driverName, err)
		return
	}

	logrus.Infof("[%s] synced genericEngineConfig for cluster %s (spotNodeCount=%d, spotServerClass=%s)",
		driverName, s.RancherClusterID, s.SpotNodeCount, s.SpotServerClass)
}

// buildRancherManagementClient creates a dynamic k8s client for the management cluster
// using in-cluster config. Returns an error instead of logging so callers can decide.
func buildRancherManagementClient() (dynamic.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}
	return dynamic.NewForConfig(cfg)
}
