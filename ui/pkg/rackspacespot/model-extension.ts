import { ModelExtensionContext } from '@shell/core/types';

// Provides the cluster list "Provider / Distro" column values for RackspaceSpot clusters.
// - provisionerDisplay  → the muted Distro label (always shown)
// - machineProviderDisplay → the Provider label (shown when machineProvider is truthy,
//   which requires the ui.rancher/provider annotation to be set on the provisioning cluster)
export default (_context: ModelExtensionContext) => ({
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
});
