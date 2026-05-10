import { IClusterProvisioner, ClusterProvisionerContext } from '@shell/core/types';
import CruRackspaceSpot from './components/CruRackspaceSpot.vue';

const DRIVER        = 'rackspacespot';
const NORMAN_CLUSTER = 'cluster';

export default class RackspaceSpotProvisioner implements IClusterProvisioner {
  static ID = DRIVER;

  private context: ClusterProvisionerContext;

  constructor(context: ClusterProvisionerContext) {
    this.context = context;
  }

  get id(): string { return DRIVER; }

  get group(): string { return 'kontainer'; }

  get label(): string { return 'Rackspace Spot'; }

  get description(): string { return 'Kubernetes clusters on Rackspace Spot instances'; }

  get component(): any { return CruRackspaceSpot; }

  get hidden(): boolean { return false; }

  get detailTabs() {
    return {
      machines:     false,
      logs:         false,
      registration: false,
      snapshots:    false,
      related:      true,
      events:       false,
      conditions:   false,
    };
  }

  async provision(cluster: any, _pools: any[]): Promise<string[]> {
    const cfg: Record<string, any> = cluster?.genericEngineConfig || {};
    const errors: string[] = [];

    if (!cfg.rackspaceSpotRefreshToken) {
      errors.push('Refresh Token is required');
    }
    if (!cfg.rackspaceSpotOrganization) {
      errors.push('Organization is required');
    }
    if (errors.length) return errors;

    const { dispatch } = this.context;

    try {
      if (cluster.id) {
        const existing = await dispatch('rancher/find', {
          type: NORMAN_CLUSTER,
          id:   cluster.id,
          opt:  { force: true },
        });
        existing.genericEngineConfig = { ...cfg };
        await existing.save();
      } else {
        const normanCluster = await dispatch('rancher/create', {
          type:                NORMAN_CLUSTER,
          name:                cluster.metadata?.name || cfg.clusterName,
          genericEngineConfig: { ...cfg },
        });
        await normanCluster.save();
      }
    } catch (e: any) {
      return [e?.message || 'Failed to save cluster'];
    }

    return [];
  }
}
