import { h } from 'vue';
import CruRackspaceSpot from './components/CruRackspaceSpot.vue';
import { registerClusterEditor } from './cluster-editor.mjs';
import { IPlugin } from '@shell/core/types';
import RackspaceSpotProvisioner from './provisioner';

export default function(plugin: IPlugin): void {
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  plugin.metadata = require('./package.json');

  plugin.register('provisioner', RackspaceSpotProvisioner.ID, RackspaceSpotProvisioner);

  let editorRegistered = false;
  plugin.addProduct((extension, store) => {
    if (!editorRegistered) {
      registerClusterEditor(extension, store, CruRackspaceSpot, h);
      editorRegistered = true;
    }
  });

  plugin.addL10n('en-us', () => ({
    cluster: {
      provider: {
        rackspacespot: 'Rackspace Spot',
      },
    },
  }));
}
