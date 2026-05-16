import { IClusterProvisioner, ClusterProvisionerContext } from '@shell/core/types';
import CruRackspaceSpot from './components/CruRackspaceSpot.vue';

const DRIVER = 'rackspacespot';

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

  get icon(): any { return require('./assets/rackspace-spot.svg'); }

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
}
