package driver

import "testing"

func TestCatalogDefaultsAndVersionNormalization(t *testing.T) {
	if defaultK8sVersion != providerCatalog.KubernetesVersions[0] {
		t.Fatal("driver default does not follow catalogue")
	}
	for _, version := range providerCatalog.KubernetesVersions {
		state := &clusterState{KubernetesVersion: version}
		applyDefaults(state)
		if state.KubernetesVersion != version {
			t.Fatalf("existing version changed: %s to %s", version, state.KubernetesVersion)
		}
	}
	state := &clusterState{}
	applyDefaults(state)
	if state.KubernetesVersion != defaultK8sVersion || state.CNI != defaultCNI || state.SpotServerClass != defaultSpotClass {
		t.Fatal("defaults do not match catalogue")
	}
	for minor, full := range k8sVersionMap {
		state := &clusterState{KubernetesVersion: minor}
		applyDefaults(state)
		if state.KubernetesVersion != full {
			t.Fatalf("%s resolved to %s, want %s", minor, state.KubernetesVersion, full)
		}
	}
}
