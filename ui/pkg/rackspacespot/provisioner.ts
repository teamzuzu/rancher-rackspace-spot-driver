import { IClusterProvisioner, ClusterProvisionerContext } from '@shell/core/types';
import CruRackspaceSpot from './components/CruRackspaceSpot.vue';

const DRIVER = 'rackspacespot';

// Inline data URI so the icon resolves correctly when the extension bundle
// runs in Rancher's context (a relative path from require() would resolve
// against Rancher's host, not the extension endpoint).
const ICON = `data:image/svg+xml,${encodeURIComponent('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="50" fill="#012A4E"/><text x="50" y="66" font-family="Arial,Helvetica,sans-serif" font-weight="bold" font-size="38" fill="#FFFFFF" text-anchor="middle">RS</text></svg>')}`;

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

  get icon(): any { return ICON; }

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
