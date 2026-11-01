package driver

import (
	_ "embed"
	"encoding/json"
	"strings"
)

//go:embed catalog.json
var catalogJSON []byte

type optionCatalog struct {
	Defaults struct {
		KubernetesVersion string `json:"kubernetesVersion"`
		CNI               string `json:"cni"`
		ServerClass       string `json:"serverClass"`
	} `json:"defaults"`
	KubernetesVersions []string `json:"kubernetesVersions"`
}

func loadOptionCatalog() optionCatalog {
	var catalog optionCatalog
	if err := json.Unmarshal(catalogJSON, &catalog); err != nil {
		panic("invalid embedded Rackspace catalogue: " + err.Error())
	}
	if catalog.Defaults.KubernetesVersion == "" || catalog.Defaults.CNI == "" || catalog.Defaults.ServerClass == "" || len(catalog.KubernetesVersions) == 0 {
		panic("incomplete embedded Rackspace catalogue")
	}
	return catalog
}

var providerCatalog = loadOptionCatalog()
var defaultK8sVersion = providerCatalog.Defaults.KubernetesVersion
var defaultCNI = providerCatalog.Defaults.CNI
var defaultSpotClass = providerCatalog.Defaults.ServerClass

// Catalogue versions are sorted newest first. Resolve short versions using
// Rackspace's latest supported patch for that minor, rather than a manual map.
var k8sVersionMap = func() map[string]string {
	versions := map[string]string{}
	for _, version := range providerCatalog.KubernetesVersions {
		parts := strings.Split(version, ".")
		if len(parts) == 3 {
			minor := strings.Join(parts[:2], ".")
			if _, exists := versions[minor]; !exists {
				versions[minor] = version
			}
		}
	}
	// Preserve the migration from the original 1.28 default for old configurations.
	if stable, ok := versions["1.32"]; ok {
		versions["1.28"] = stable
	}
	return versions
}()
