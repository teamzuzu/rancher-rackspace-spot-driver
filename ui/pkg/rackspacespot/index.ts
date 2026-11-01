import { IPlugin } from '@shell/core/types';
import RackspaceSpotProvisioner from './provisioner';

export default function(plugin: IPlugin): void {
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  plugin.metadata = require('./package.json');

  plugin.register('provisioner', RackspaceSpotProvisioner.ID, RackspaceSpotProvisioner);

  // Rancher's product loader invokes the module's init export.
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  plugin.addProduct(require('./cluster-product'));

  plugin.addL10n('en-us', () => ({
    cluster: {
      provider: {
        rackspacespot: 'Rackspace Spot',
      },
    },
  }));
}
