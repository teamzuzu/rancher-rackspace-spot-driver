import { IPlugin } from '@shell/core/types';
import RackspaceSpotProvisioner from './provisioner';

export default function(plugin: IPlugin): void {
  // eslint-disable-next-line @typescript-eslint/no-var-requires
  plugin.metadata = require('./package.json');

  plugin.register('provisioner', RackspaceSpotProvisioner.ID, RackspaceSpotProvisioner);
}
