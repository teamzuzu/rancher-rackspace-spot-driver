export const CLUSTER_RESOURCE = 'provisioning.cattle.io.cluster';

export function isRackspaceSpotCluster(value) {
  if (!value?.id) return false;

  const providers = [
    value.metadata?.annotations?.['ui.rancher/provider'],
    value.annotations?.['ui.rancher/provider'],
    value.mgmt?.metadata?.annotations?.['ui.rancher/provider'],
    value.mgmt?.annotations?.['ui.rancher/provider'],
    value.mgmt?.spec?.genericEngineConfig?.driverName,
    value.spec?.genericEngineConfig?.driverName,
    value.provisioner,
  ];

  return providers.some(provider => provider?.toLowerCase() === 'rackspacespot');
}

// Rancher 2.15's default cluster editor rejects external KontainerDrivers
// before resolving the extension's provisioner component. Capture the host's
// editor before registering this adapter, and delegate all other clusters to it.
// Capturing through the host store avoids bundling an older Rancher editor.
export function registerClusterEditor(plugin, store, spotEditor, h) {
  const defaultEditor = store.getters['type-map/importEdit'](CLUSTER_RESOURCE);
  const editor = {
    name: 'RackspaceSpotClusterEditor',
    inheritAttrs: false,
    props: {
      value: { type: Object, default: () => ({}) },
      mode: { type: String, default: 'create' },
    },
    emits: ['update:value'],
    render() {
      return h(isRackspaceSpotCluster(this.value) ? spotEditor : defaultEditor, {
        ...this.$attrs,
        value: this.value,
        mode: this.mode,
        'onUpdate:value': value => this.$emit('update:value', value),
      }, this.$slots);
    },
  };

  const loader = () => Promise.resolve(editor);
  plugin.register('edit', CLUSTER_RESOURCE, loader);
  // Product initialization runs after applyPlugin has copied plugin.types.
  // Publish the late registration to the live manager as well; plugin.types
  // still records it so unloading this extension removes the adapter.
  store.state?.$extension?.register('edit', CLUSTER_RESOURCE, loader);
}

// addProduct consumes a module with an init export, rather than calling a
// callback directly. Repeated product initialization must not capture our own
// adapter as the fallback and recurse indefinitely.
export function createClusterProduct(spotEditor, h) {
  const registeredStores = new WeakSet();
  return {
    init(plugin, store) {
      if (registeredStores.has(store)) return;
      registerClusterEditor(plugin, store, spotEditor, h);
      registeredStores.add(store);
    },
  };
}
