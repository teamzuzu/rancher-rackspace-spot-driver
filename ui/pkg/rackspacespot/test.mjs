// Run with: node ui/pkg/rackspacespot/test.mjs
// Tests the pure-JS logic from CruRackspaceSpot.vue without a browser or framework.

import assert from 'node:assert/strict';

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

console.log(`\n${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);
