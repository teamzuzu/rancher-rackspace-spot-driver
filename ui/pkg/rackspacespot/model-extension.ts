import { ModelExtensionContext } from '@shell/core/types';

// Provides the cluster list "Provider / Distro" column values for RackspaceSpot clusters.
// - provisionerDisplay  → the muted Distro label (always shown)
// - machineProviderDisplay → the Provider label (shown when machineProvider is truthy,
//   which requires the ui.rancher/provider annotation to be set on the provisioning cluster)
//
// Must be a regular function (not an arrow function) — Rancher's extension manager
// calls it with `new`, and arrow functions cannot be used as constructors.
export default function RackspaceSpotModelExtension(_context: ModelExtensionContext) {
  return {
    useFor(cluster: any): boolean {
      return cluster.provisioner === 'rackspacespot'
        || cluster.mgmt?.spec?.genericEngineConfig?.driverName === 'rackspacespot';
    },

    provisionerDisplay(_cluster: any): string {
      return 'Rackspace Spot';
    },

    machineProviderDisplay(_cluster: any): string {
      return 'Rackspace Spot';
    },
  };
}
