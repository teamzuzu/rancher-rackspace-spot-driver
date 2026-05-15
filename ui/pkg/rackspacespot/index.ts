import { IPlugin } from '@shell/core/types';
import RackspaceSpotProvisioner from './provisioner';

export default function(plugin: IPlugin): void {
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  plugin.metadata = require('./package.json');

  plugin.register('provisioner', RackspaceSpotProvisioner.ID, RackspaceSpotProvisioner);

  plugin.addL10n('en-us', () => ({
    cluster: {
      providerGroup: {
        'create-kontainer': 'Provision a new cluster in Rackspace Spot',
      },
      provider: {
        rackspacespot: 'Rackspace Spot',
      },
    },
  }));
}
