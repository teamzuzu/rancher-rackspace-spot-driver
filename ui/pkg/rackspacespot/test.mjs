// Run with: node ui/pkg/rackspacespot/test.mjs
// Tests the pure-JS logic from CruRackspaceSpot.vue without a browser or framework.

import assert from 'node:assert/strict';
import { CLUSTER_RESOURCE, isRackspaceSpotCluster, registerClusterEditor, createClusterProduct } from './cluster-editor.mjs';
import { normalizeAutopilotPools, validateAutopilotPools } from './autopilot.mjs';

const knownServerClasses = [
  { name: 'ch.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '16', memory: '30GB'   },
  { name: 'ch.vs1.large-iad',   region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '4',  memory: '7.5GB'  },
  { name: 'ch.vs1.medium-iad',  region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '2',  memory: '3.75GB' },
  { name: 'ch.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '8',  memory: '15GB'   },
  { name: 'gp.bm2.small-iad',   region: 'us-east-iad-1', category: 'Bare Metal',       cpu: '12', memory: '32GB'   },
  { name: 'gp.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'General Purpose', cpu: '16', memory: '60GB'   },
  { name: 'gp.vs1.large-iad',   region: 'us-east-iad-1', category: 'General Purpose', cpu: '4',  memory: '15GB'   },
  { name: 'gp.vs1.medium-iad',  region: 'us-east-iad-1', category: 'General Purpose', cpu: '2',  memory: '3.75GB' },
  { name: 'gp.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'General Purpose', cpu: '8',  memory: '30GB'   },
  { name: 'io.bm2-iad',         region: 'us-east-iad-1', category: 'Bare Metal',       cpu: '40', memory: '128GB'  },
  { name: 'mh.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '16', memory: '120GB'  },
  { name: 'mh.vs1.large-iad',   region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '4',  memory: '30GB'   },
  { name: 'mh.vs1.medium-iad',  region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '2',  memory: '15GB'   },
  { name: 'mh.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '8',  memory: '60GB'   },
];

// Mirror of the serverClassOptions computed property in CruRackspaceSpot.vue
function serverClassOptions(region) {
  const parts = region ? region.split('-') : [];
  const locCode = parts.length >= 2 ? parts[parts.length - 2] : 'iad';
  return knownServerClasses.map(sc => {
    const name = sc.name.replace(/-iad$/, `-${locCode}`);
    return { label: `${name} — ${sc.category}, ${sc.cpu} CPU, ${sc.memory}`, value: name };
  });
}

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

console.log('\nLocation code extraction:');
const locCases = [
  ['us-east-iad-1',    'iad'],
  ['us-east-iad-2',    'iad'],
  ['us-central-dfw-1', 'dfw'],
  ['us-central-dfw-2', 'dfw'],
  ['us-central-ord-1', 'ord'],
  ['us-west-sjc-1',    'sjc'],
  ['aus-syd-1',        'syd'],
  ['hkg-hkg-1',        'hkg'],
  ['uk-lon-1',         'lon'],
];
for (const [region, expected] of locCases) {
  test(`${region} → ${expected}`, () => {
    const parts = region.split('-');
    const got = parts[parts.length - 2];
    assert.equal(got, expected);
  });
}

console.log('\nserverClassOptions:');
test('us-east-iad-1 keeps -iad suffix', () => {
  const opts = serverClassOptions('us-east-iad-1');
  assert.equal(opts.length, knownServerClasses.length);
  assert.ok(opts.every(o => o.value.endsWith('-iad')), 'all values end with -iad');
});
test('us-east-iad-2 also uses -iad suffix', () => {
  const opts = serverClassOptions('us-east-iad-2');
  assert.ok(opts.every(o => o.value.endsWith('-iad')));
});
test('aus-syd-1 produces -syd names', () => {
  const opts = serverClassOptions('aus-syd-1');
  assert.ok(opts.every(o => o.value.endsWith('-syd')));
  assert.ok(opts.some(o => o.value === 'gp.vs1.medium-syd'));
});
test('uk-lon-1 produces -lon names', () => {
  const opts = serverClassOptions('uk-lon-1');
  assert.ok(opts.every(o => o.value.endsWith('-lon')));
});
test('us-central-dfw-1 produces -dfw names', () => {
  const opts = serverClassOptions('us-central-dfw-1');
  assert.ok(opts.every(o => o.value.endsWith('-dfw')));
});
test('hkg-hkg-1 produces -hkg names', () => {
  const opts = serverClassOptions('hkg-hkg-1');
  assert.ok(opts.every(o => o.value.endsWith('-hkg')));
});
test('io.bm2-iad → io.bm2-syd (no spurious -iad left)', () => {
  const opts = serverClassOptions('aus-syd-1');
  const bm2 = opts.find(o => o.value.startsWith('io.bm2'));
  assert.equal(bm2.value, 'io.bm2-syd');
});
test('no region defaults to -iad', () => {
  const opts = serverClassOptions('');
  assert.ok(opts.every(o => o.value.endsWith('-iad')));
});
test('label contains the derived value', () => {
  serverClassOptions('aus-syd-1').forEach(o => assert.ok(o.label.includes(o.value)));
});
test('same count for every region', () => {
  ['us-east-iad-1','aus-syd-1','uk-lon-1','us-central-dfw-1','hkg-hkg-1'].forEach(r => {
    assert.equal(serverClassOptions(r).length, knownServerClasses.length);
  });
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

test('Rancher product lifecycle initializes the editor once per store', async () => {
  const calls = [];
  const product = createClusterProduct(spotClusterEditor, (component, props) => ({ component, props }));
  const plugin = { register(kind, resource) { calls.push(['register', kind, resource]); } };
  const store = { getters: { 'type-map/importEdit': () => { calls.push(['capture']); return originalClusterEditor; } } };
  // Mirror extension-manager.loadProducts, including its init export contract.
  const implementation = await product;
  assert.equal(typeof implementation.init, 'function');
  implementation.init(plugin, store);
  implementation.init(plugin, store);
  assert.deepEqual(calls, [['capture'], ['register', 'edit', CLUSTER_RESOURCE]]);
});

console.log(`\n${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);
