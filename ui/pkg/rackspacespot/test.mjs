// Run with: node ui/pkg/rackspacespot/test.mjs
// Tests the pure-JS logic from CruRackspaceSpot.vue without a browser or framework.

import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { serverClassOptions, regionOptions, versionOptions, cniOptions, defaultServerClass } from './catalog.mjs';
const catalog = JSON.parse(readFileSync(new URL('../../../driver/catalog.json', import.meta.url)));
import { CLUSTER_RESOURCE, isRackspaceSpotCluster, registerClusterEditor, createClusterProduct } from './cluster-editor.mjs';
import { normalizeAutopilotPools, validateAutopilotPools } from './autopilot.mjs';

let passed = 0;
let failed = 0;

function test(name, fn) {
  try {
    fn();
    console.log(`  ✓ ${name}`);
    passed++;
  } catch (e) {
    console.error(`  ✗ ${name}`);
    console.error(`    ${e.message}`);
    failed++;
  }
}

console.log('\nRackspace catalogue:');
test('server classes use exact region membership, including Gen 2', () => {
  for (const region of catalog.regions) {
    const options = serverClassOptions(catalog, region.name);
    const expected = catalog.serverClasses.filter(serverClass => serverClass.region === region.name);
    assert.deepEqual(options.map(option => option.value), expected.map(serverClass => serverClass.name));
    assert.ok(options.length, `no active classes for ${region.name}`);
  }
  const regional = { serverClasses: [{ name: 'gen1', region: 'us-central-dfw-1' }, { name: 'gen2', region: 'us-central-dfw-2' }] };
  assert.deepEqual(serverClassOptions(regional, 'us-central-dfw-2').map(option => option.value), ['gen2']);
});
test('unknown regions never synthesize server class names', () => {
  assert.deepEqual(serverClassOptions(catalog, 'missing-region'), []);
});
test('new cluster defaults are present in the catalogue', () => {
  assert.ok(versionOptions(catalog).some(option => option.value === catalog.defaults.kubernetesVersion));
  assert.ok(regionOptions(catalog).some(option => option.value === catalog.defaults.region));
  assert.ok(cniOptions(catalog).some(option => option.value === catalog.defaults.cni));
  for (const region of catalog.regions) {
    assert.ok(serverClassOptions(catalog, region.name).some(option => option.value === defaultServerClass(catalog, region.name)));
  }
});
test('editing preserves retired versions, regions, CNIs and server classes', () => {
  assert.ok(versionOptions(catalog, '1.20.0').some(option => option.value === '1.20.0'));
  assert.ok(regionOptions(catalog, 'retired-region').some(option => option.value === 'retired-region'));
  assert.ok(cniOptions(catalog, 'existing-cni').some(option => option.value === 'existing-cni'));
  const options = serverClassOptions(catalog, catalog.defaults.region, ['retired-class', 'retired-class']);
  assert.equal(options.filter(option => option.value === 'retired-class').length, 1);
});
test('catalogue options are sorted by numeric version and reflect new upstream values', () => {
  const updated = { ...catalog, kubernetesVersions: ['1.100.0', '1.35.9'], cnis: ['new-cni'] };
  assert.deepEqual(versionOptions(updated).map(option => option.value), updated.kubernetesVersions);
  assert.deepEqual(cniOptions(updated), [{ label: 'new-cni', value: 'new-cni' }]);
});

// Mirror of the save() validation logic in CruRackspaceSpot.vue.
// Returns an error message string, or null if valid.
function validateConfig(clusterName, config, autoscalingPoolCount) {
  if (!clusterName) return 'Cluster Name is required';
  if (!config.rackspaceSpotRefreshToken) return 'Refresh Token is required';
  if (!config.rackspaceSpotOrganization) return 'Organization is required';
  if (config.importExistingCluster && !config.importCloudspaceName) {
    return 'Cloudspace Name is required when importing an existing cluster';
  }
  if (!config.importExistingCluster) {
    if (config.onDemandEnabled && !config.onDemandServerClass) {
      return 'On-Demand Server Class is required when the on-demand pool is enabled';
    }
    if (autoscalingPoolCount > 1) {
      return 'Only one spot node pool may have autoscaling enabled per cloudspace (API limit).';
    }
  }
  return null;
}

console.log('\nimport mode validation:');
test('import mode requires cloudspace name', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    importExistingCluster:     true,
    importCloudspaceName:      '',
  }, 0);
  assert.ok(err !== null && err.includes('Cloudspace Name'), `expected Cloudspace Name error, got: ${err}`);
});
test('import mode with cloudspace name skips pool checks', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    onDemandEnabled:           true,
    onDemandServerClass:       '',
    importExistingCluster:     true,
    importCloudspaceName:      'us',
  }, 0);
  assert.equal(err, null);
});
test('normal mode requires on-demand server class when enabled', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    onDemandEnabled:           true,
    onDemandServerClass:       '',
    importExistingCluster:     false,
  }, 0);
  assert.ok(err !== null && err.includes('On-Demand'), `expected On-Demand error, got: ${err}`);
});
test('import mode skips autoscaling pool count check', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    importExistingCluster:     true,
    importCloudspaceName:      'us',
  }, 2);
  assert.equal(err, null);
});
test('normal mode rejects multiple autoscaling pools', () => {
  const err = validateConfig('my-cluster', {
    rackspaceSpotRefreshToken: 'tok',
    rackspaceSpotOrganization: 'org',
    importExistingCluster:     false,
  }, 2);
  assert.ok(err !== null && err.includes('autoscaling'), `expected autoscaling error, got: ${err}`);
});

test('Autopilot normalizes form strings and optional bounds', () => {
  const pools = normalizeAutopilotPools([{ name: 'pool-id', vcpus: '16', memoryPerVCPU: 4, budgetPerHour: '0.50', minVCPUsPerNode: '', maxVCPUsPerNode: '8' }]);
  assert.deepEqual(pools, [{ name: 'pool-id', vcpus: 16, memoryPerVCPU: '4', budgetPerHour: '0.50', minVCPUsPerNode: 0, maxVCPUsPerNode: 8 }]);
  assert.equal(validateAutopilotPools(pools), '');
});
test('Autopilot validates capacity, memory, budget and bounds', () => {
  const valid = { vcpus: 16, memoryPerVCPU: 'any', budgetPerHour: '0.50', minVCPUsPerNode: 0, maxVCPUsPerNode: 0 };
  assert.equal(validateAutopilotPools([valid]), '');
  assert.equal(validateAutopilotPools([]), '');
  for (const change of [{ vcpus: 0 }, { vcpus: 1.5 }, { memoryPerVCPU: '3' }, { budgetPerHour: 'NaN' }, { budgetPerHour: '0' }, { minVCPUsPerNode: -1 }, { minVCPUsPerNode: 8, maxVCPUsPerNode: 2 }]) {
    assert.ok(validateAutopilotPools([{ ...valid, ...change }]));
  }
});

test('cluster editor recognizes existing Spot clusters, including older clusters without annotations', () => {
  assert.ok(isRackspaceSpotCluster({ id: 'fleet-default/c-spot', mgmt: { spec: { genericEngineConfig: { driverName: 'rackspacespot' } } } }));
  assert.ok(isRackspaceSpotCluster({ id: 'fleet-default/c-spot', metadata: { annotations: { 'ui.rancher/provider': 'rackspacespot' } } }));
  assert.ok(isRackspaceSpotCluster({ id: 'fleet-default/c-spot', provisioner: 'RackspaceSpot' }));
  assert.equal(isRackspaceSpotCluster({ id: 'fleet-default/c-other', provisioner: 'eks' }), false);
  assert.equal(isRackspaceSpotCluster({ rackspacespotEngineConfig: {} }), false);
});
test('cluster editor captures the host fallback before registering its adapter', () => {
  const original = { name: 'HostClusterEditor' };
  const spot = { name: 'SpotEditor' };
  const calls = [];
  let loadEditor;
  registerClusterEditor({ register(kind, resource, loader) {
    calls.push('register');
    assert.equal(kind, 'edit');
    assert.equal(resource, CLUSTER_RESOURCE);
    loadEditor = loader;
  } }, { getters: { 'type-map/importEdit'(resource) {
    calls.push('capture');
    assert.equal(resource, CLUSTER_RESOURCE);
    return original;
  } } }, spot, (component, props, slots) => ({ component, props, slots }));
  assert.deepEqual(calls, ['capture', 'register']);
  // The loader resolves to the component; rendering is checked below after loading.
  assert.equal(typeof loadEditor, 'function');
});
const originalClusterEditor = { name: 'OriginalClusterEditor' };
const spotClusterEditor = { name: 'SpotClusterEditor' };
let clusterEditorLoader;
registerClusterEditor({ register(_kind, _resource, loader) { clusterEditorLoader = loader; } }, {
  getters: { 'type-map/importEdit': () => originalClusterEditor },
}, spotClusterEditor, (component, props, slots) => ({ component, props, slots }));
const clusterEditor = await clusterEditorLoader();
test('Spot view and edit bypass the legacy wrapper and forward values, attrs and events', () => {
  for (const mode of ['view', 'edit']) {
    const value = { id: 'fleet-default/c-spot', provisioner: 'rackspacespot' };
    const events = [];
    const context = { value, mode, $attrs: { 'done-route': 'clusters' }, $slots: {}, $emit: (...args) => events.push(args) };
    const rendered = clusterEditor.render.call(context);
    assert.equal(rendered.component, spotClusterEditor);
    assert.equal(rendered.props.value, value);
    assert.equal(rendered.props.mode, mode);
    assert.equal(rendered.props['done-route'], 'clusters');
    rendered.props['onUpdate:value'](value);
    assert.deepEqual(events, [['update:value', value]]);
  }
});
test('other providers and the create wizard retain the original Rancher editor', () => {
  for (const value of [{ id: 'fleet-default/c-other', provisioner: 'gke' }, {}]) {
    const rendered = clusterEditor.render.call({ value, mode: 'create', $attrs: {}, $slots: {}, $emit() {} });
    assert.equal(rendered.component, originalClusterEditor);
  }
});

test('Rancher product lifecycle initializes the editor once per store', () => {
  const calls = [];
  const product = createClusterProduct(spotClusterEditor, (component, props) => ({ component, props }));
  const plugin = { register(kind, resource) { calls.push(['register', kind, resource]); } };
  const store = { getters: { 'type-map/importEdit': () => { calls.push(['capture']); return originalClusterEditor; } } };
  // Mirror extension-manager.loadProducts, including its init export contract.
  const implementation = product;
  assert.equal(typeof implementation.init, 'function');
  implementation.init(plugin, store);
  implementation.init(plugin, store);
  assert.deepEqual(calls, [['capture'], ['register', 'edit', CLUSTER_RESOURCE]]);
});

test('product registration updates the live manager after plugin types were applied', () => {
  const types = {};
  const dynamic = {};
  const plugin = { register(kind, resource, loader) { (types[kind] ||= {})[resource] = loader; } };
  const manager = { register(kind, resource, loader) { (dynamic[kind] ||= {})[resource] = loader; } };
  const store = { state: { $extension: manager }, getters: { 'type-map/importEdit': () => originalClusterEditor } };
  // Rancher copies types before invoking products, so plugin.register alone
  // would update types but leave the runtime registry without an editor.
  Object.assign(dynamic, types);
  createClusterProduct(spotClusterEditor, () => {}).init(plugin, store);
  assert.equal(typeof dynamic.edit[CLUSTER_RESOURCE], 'function');
  assert.equal(dynamic.edit[CLUSTER_RESOURCE], types.edit[CLUSTER_RESOURCE]);
});

console.log(`\n${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);
