<template>
  <CruResource
    :mode="mode"
    :resource="value"
    :errors="errors"
    :validation-passed="validationPassed"
    @finish="save"
    @cancel="cancel"
  >
    <Banner
      v-if="errors.length"
      color="error"
      :label="errors.join(' | ')"
    />

    <!-- ── Cluster name ───────────────────────────────────── -->
    <div class="row mb-20">
      <div class="col span-6">
        <LabeledInput
          v-model:value="clusterName"
          label="Cluster Name"
          :required="true"
          :mode="mode"
        />
      </div>
    </div>

    <!-- ── Authentication ─────────────────────────────────── -->
    <h3>Authentication</h3>
    <div class="row">
      <div class="col span-6">
        <LabeledInput
          v-model:value="config.rackspaceSpotOrganization"
          label="Organization"
          placeholder="your-org-name"
          :required="true"
          :mode="mode"
        />
      </div>
      <div class="col span-6">
        <LabeledInput
          v-model:value="config.rackspaceSpotRefreshToken"
          label="Refresh Token"
          type="password"
          placeholder="Rackspace Spot refresh token"
          :required="true"
          :mode="mode"
        />
      </div>
    </div>

    <!-- ── Import mode ───────────────────────────────────────── -->
    <div class="row mt-20">
      <div class="col span-12">
        <!-- mode forced to 'view' in edit mode: import is a one-time create-time action -->
        <Checkbox
          v-model:value="config.importExistingCluster"
          label="Import existing cluster"
          :mode="mode === 'edit' ? 'view' : mode"
        />
        <p v-if="config.importExistingCluster" class="import-note mt-5">
          Node pool configuration will be read from the existing cloudspace after import.
        </p>
      </div>
    </div>

    <!-- ── Cluster ─────────────────────────────────────────── -->
    <h3 class="mt-20">Cluster</h3>
    <div class="row">
      <div class="col span-4">
        <LabeledSelect
          v-model:value="config.rackspaceSpotRegion"
          label="Region"
          :options="regionOptions"
          :required="true"
          :mode="mode === 'edit' ? 'view' : mode"
        />
      </div>
      <div class="col span-4">
        <LabeledSelect
          v-model:value="config.kubernetesVersion"
          label="Kubernetes Version"
          :options="k8sVersionOptions"
          :mode="mode === 'edit' ? 'view' : mode"
        />
      </div>
      <div class="col span-4">
        <LabeledSelect
          v-model:value="config.cni"
          label="CNI"
          :options="cniOptions"
          :mode="mode === 'edit' ? 'view' : mode"
        />
      </div>
    </div>
    <div class="row mt-10">
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.preemptionWebhookUrl"
          label="Preemption Webhook URL"
          placeholder="https://..."
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <Checkbox
          v-model:value="config.gpuEnabled"
          label="Enable GPU support"
          :mode="mode === 'edit' ? 'view' : mode"
        />
      </div>
    </div>
    <!-- ── Node pools (hidden in import mode) ───────────── -->
    <template v-if="!config.importExistingCluster">
      <!-- ── Spot Node Pools ───────────────────────────────── -->
      <h3 class="mt-20">Spot Node Pools</h3>
      <a href="https://tombojer.github.io/spot-cost-analyzer/" target="_blank" rel="noopener" class="spot-cost-link">Estimate costs with the Spot Cost Analyzer ↗</a>

      <!-- ── Primary spot pool ─────────────────────────────── -->
      <div class="pool-card mt-15">
        <div class="pool-card-header">
          <span class="pool-label">Pool 1 (primary)</span>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="config.spotServerClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.spotNodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.spotBidPrice"
              label="Bid Price (USD/hr)"
              placeholder="0.01"
              :mode="mode"
            />
          </div>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <Checkbox
              v-model:value="config.spotAutoscalingEnabled"
              :label="primaryAutoscalingDisabled ? 'Enable Autoscaling (another pool has autoscaling)' : 'Enable Autoscaling'"
              :disabled="primaryAutoscalingDisabled"
              :mode="mode"
            />
          </div>
          <template v-if="config.spotAutoscalingEnabled">
            <div class="col span-4">
              <LabeledInput
                v-model:value="config.spotAutoscalingMinNodes"
                label="Min Nodes"
                type="number"
                :min="0"
                :mode="mode"
              />
            </div>
            <div class="col span-4">
              <LabeledInput
                v-model:value="config.spotAutoscalingMaxNodes"
                label="Max Nodes"
                type="number"
                :min="1"
                :mode="mode"
              />
            </div>
          </template>
        </div>
      </div>

      <!-- ── Additional spot pools ─────────────────────────── -->
      <div
        v-for="(pool, idx) in additionalSpotPools"
        :key="idx"
        class="pool-card mt-10"
      >
        <div class="pool-card-header">
          <span class="pool-label">Pool {{ idx + 2 }}</span>
          <button
            class="btn btn-sm btn-danger"
            type="button"
            @click="removeSpotPool(idx)"
          >
            Remove
          </button>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="pool.serverClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="pool.nodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="pool.bidPrice"
              label="Bid Price (USD/hr)"
              placeholder="0.01"
              :mode="mode"
            />
          </div>
        </div>
        <div class="row mt-10">
          <div class="col span-4">
            <Checkbox
              v-model:value="pool.autoscaling"
              :label="autoscalingPoolCount >= 1 && !pool.autoscaling ? 'Enable Autoscaling (another pool has autoscaling)' : 'Enable Autoscaling'"
              :disabled="autoscalingPoolCount >= 1 && !pool.autoscaling"
              :mode="mode"
            />
          </div>
          <template v-if="pool.autoscaling">
            <div class="col span-4">
              <LabeledInput
                v-model:value="pool.minNodes"
                label="Min Nodes"
                type="number"
                :min="0"
                :mode="mode"
              />
            </div>
            <div class="col span-4">
              <LabeledInput
                v-model:value="pool.maxNodes"
                label="Max Nodes"
                type="number"
                :min="1"
                :mode="mode"
              />
            </div>
          </template>
        </div>
      </div>

      <div class="mt-10">
        <button
          class="btn btn-sm btn-primary"
          type="button"
          @click="addSpotPool"
        >
          + Add Spot Pool
        </button>
      </div>

      <!-- ── On-Demand Node Pool ─────────────────────────────── -->
      <h3 class="mt-20">On-Demand Node Pool</h3>
      <div class="row">
        <div class="col span-12">
          <Checkbox
            v-model:value="config.onDemandEnabled"
            label="Add stable capacity alongside spot nodes"
            :mode="mode"
          />
        </div>
      </div>
      <template v-if="config.onDemandEnabled">
        <div class="row mt-10">
          <div class="col span-4">
            <LabeledSelect
              v-model:value="config.onDemandServerClass"
              label="Server Class"
              :options="serverClassOptions"
              :taggable="true"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.onDemandNodeCount"
              label="Node Count"
              type="number"
              :min="0"
              :mode="mode"
            />
          </div>
          <div class="col span-4">
            <LabeledInput
              v-model:value="config.onDemandPricePerHour"
              label="Price Per Hour (USD)"
              placeholder="0.00"
              :mode="mode"
            />
          </div>
        </div>
      </template>
    </template>
  </CruResource>
</template>

<script>
import { defineComponent } from 'vue';
import CruResource       from '@shell/components/CruResource';
import Banner            from '@components/Banner/Banner';
import LabeledInput      from '@components/Form/LabeledInput/LabeledInput';
import LabeledSelect     from '@shell/components/form/LabeledSelect';
import Checkbox          from '@components/Form/Checkbox/Checkbox';

const DRIVER = 'rackspacespot';

const DEFAULTS = {
  driverName:              DRIVER,
  rackspaceSpotRegion:     'us-east-iad-1',
  kubernetesVersion:       '1.33.0',
  cni:                     'calico',
  gpuEnabled:              false,
  importExistingCluster:   false,
  spotServerClass:         'gp.vs1.medium-iad',
  spotNodeCount:           3,
  spotBidPrice:            '0.01',
  spotAutoscalingEnabled:  false,
  spotAutoscalingMinNodes: 1,
  spotAutoscalingMaxNodes: 10,
  onDemandEnabled:         false,
  onDemandServerClass:     'gp.vs1.medium-iad',
  onDemandNodeCount:       1,
};

function emptyPool() {
  return {
    serverClass:  '',
    nodeCount:    1,
    bidPrice:     '0.01',
    autoscaling:  false,
    minNodes:     1,
    maxNodes:     10,
  };
}

export default defineComponent({
  name: 'CruRackspaceSpot',

  components: {
    CruResource,
    Banner,
    LabeledInput,
    LabeledSelect,
    Checkbox,
  },

  props: {
    value: {
      type:    Object,
      default: () => ({}),
    },
    mode: {
      type:    String,
      default: 'create',
    },
  },

  emits: ['update:value'],

  data() {
    // In edit mode value is a Steve provisioning.cattle.io.cluster; config lives on mgmt.
    // In create mode it may be a plain object or newly constructed Steve resource.
    const raw = this.value?.rackspacespotEngineConfig
      || this.value?.mgmt?.spec?.genericEngineConfig
      || {};
    let additionalSpotPools = [];
    if (raw.additionalSpotPools) {
      try {
        additionalSpotPools = JSON.parse(raw.additionalSpotPools);
      } catch (e) {}
    }
    return {
      clusterName: this.value?.displayName
        || this.value?.spec?.displayName
        || this.value?.name
        || this.value?.metadata?.name
        || '',
      config:               { ...DEFAULTS, ...raw },
      additionalSpotPools,
      errors:               [],
      availableRegions:     [],
      knownRegions: [
        { name: 'aus-syd-1',        description: 'Sydney, Australia' },
        { name: 'hkg-hkg-1',        description: 'Hong Kong, Hong Kong' },
        { name: 'uk-lon-1',         description: 'United Kingdom, London' },
        { name: 'us-central-dfw-1', description: 'US Central, Dallas Fort Worth, TX' },
        { name: 'us-central-dfw-2', description: 'US Central, Dallas Fort Worth, TX' },
        { name: 'us-central-ord-1', description: 'US Central, Chicago, IL' },
        { name: 'us-east-iad-1',    description: 'US East, Ashburn, VA' },
        { name: 'us-east-iad-2',    description: 'US East, Ashburn, VA' },
        { name: 'us-west-sjc-1',    description: 'US West, San Jose, CA' },
      ],
      knownServerClasses: [
        { name: 'ch.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '16', memory: '30GB' },
        { name: 'ch.vs1.large-iad',   region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '4',  memory: '7.5GB' },
        { name: 'ch.vs1.medium-iad',  region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '2',  memory: '3.75GB' },
        { name: 'ch.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'Compute Heavy',   cpu: '8',  memory: '15GB' },
        { name: 'gp.bm2.small-iad',   region: 'us-east-iad-1', category: 'Bare Metal',       cpu: '12', memory: '32GB' },
        { name: 'gp.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'General Purpose', cpu: '16', memory: '60GB' },
        { name: 'gp.vs1.large-iad',   region: 'us-east-iad-1', category: 'General Purpose', cpu: '4',  memory: '15GB' },
        { name: 'gp.vs1.medium-iad',  region: 'us-east-iad-1', category: 'General Purpose', cpu: '2',  memory: '3.75GB' },
        { name: 'gp.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'General Purpose', cpu: '8',  memory: '30GB' },
        { name: 'io.bm2-iad',         region: 'us-east-iad-1', category: 'Bare Metal',       cpu: '40', memory: '128GB' },
        { name: 'mh.vs1.2xlarge-iad', region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '16', memory: '120GB' },
        { name: 'mh.vs1.large-iad',   region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '4',  memory: '30GB' },
        { name: 'mh.vs1.medium-iad',  region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '2',  memory: '15GB' },
        { name: 'mh.vs1.xlarge-iad',  region: 'us-east-iad-1', category: 'Memory Heavy',    cpu: '8',  memory: '60GB' },
      ],
      cniOptions:           [
        { label: 'calico', value: 'calico' },
        { label: 'cilium', value: 'cilium' },
      ],
      k8sVersionOptions: [
        { label: '1.33.0', value: '1.33.0' },
        { label: '1.32.9', value: '1.32.9' },
        { label: '1.31.1', value: '1.31.1' },
        { label: '1.30.10', value: '1.30.10' },
        { label: '1.29.6', value: '1.29.6' },
      ],
    };
  },

  computed: {
    validationPassed() {
      return !!(this.clusterName && this.config.rackspaceSpotRefreshToken && this.config.rackspaceSpotOrganization);
    },
    autoscalingPoolCount() {
      return (this.config.spotAutoscalingEnabled ? 1 : 0)
        + this.additionalSpotPools.filter(p => p.autoscaling).length;
    },
    primaryAutoscalingDisabled() {
      return this.autoscalingPoolCount >= 1 && !this.config.spotAutoscalingEnabled;
    },
    regionOptions() {
      const source = this.availableRegions.length ? this.availableRegions : this.knownRegions;
      return source.map(r => ({
        label: r.description ? `${r.name} — ${r.description}` : r.name,
        value: r.name,
      }));
    },
    serverClassOptions() {
      const region = this.config.rackspaceSpotRegion;
      // Derive location code from region string: 'us-east-iad-1' → 'iad', 'aus-syd-1' → 'syd'
      const parts = region ? region.split('-') : [];
      const locCode = parts.length >= 2 ? parts[parts.length - 2] : 'iad';
      return this.knownServerClasses.map(sc => {
        const name = sc.name.replace(/-iad$/, `-${locCode}`);
        return {
          label: `${name} — ${sc.category}, ${sc.cpu} CPU, ${sc.memory}`,
          value: name,
        };
      });
    },
  },

  async created() {
    if (this.mode !== 'create') return;
    try {
      const secret = await this.$store.dispatch('management/find', {
        type: 'secret',
        id:   'cattle-system/rackspace-spot-credentials',
      });
      if (!this.config.rackspaceSpotOrganization && secret?.data?.org) {
        this.config.rackspaceSpotOrganization = atob(secret.data.org);
      }
      if (!this.config.rackspaceSpotRefreshToken && secret?.data?.refreshToken) {
        this.config.rackspaceSpotRefreshToken = atob(secret.data.refreshToken);
      }
    } catch (_) {
      // Secret absent or no permission — silent fallback, fields stay blank
    }
  },

  watch: {
    'config.rackspaceSpotRegion'(newRegion, oldRegion) {
      if (this.mode === 'edit' || newRegion === oldRegion) return;
      this.config.spotServerClass     = '';
      this.config.onDemandServerClass = '';
      this.additionalSpotPools.forEach(p => { p.serverClass = ''; });
    },
  },

  methods: {
    addSpotPool() {
      this.additionalSpotPools.push(emptyPool());
    },

    removeSpotPool(idx) {
      this.additionalSpotPools.splice(idx, 1);
    },

    async save(btnCb) {
      this.errors = [];

      if (!this.clusterName) {
        this.errors = ['Cluster Name is required'];
        if (btnCb) btnCb(false);
        return;
      }
      if (!this.config.rackspaceSpotRefreshToken) {
        this.errors = ['Refresh Token is required'];
        if (btnCb) btnCb(false);
        return;
      }
      if (!this.config.rackspaceSpotOrganization) {
        this.errors = ['Organization is required'];
        if (btnCb) btnCb(false);
        return;
      }
      if (!this.config.importExistingCluster) {
        if (this.config.onDemandEnabled && !this.config.onDemandServerClass) {
          this.errors = ['On-Demand Server Class is required when the on-demand pool is enabled'];
          if (btnCb) btnCb(false);
          return;
        }
        if (this.autoscalingPoolCount > 1) {
          this.errors = ['Only one spot node pool may have autoscaling enabled per cloudspace (API limit).'];
          if (btnCb) btnCb(false);
          return;
        }
      }

      try {
        const cfg = {
          ...this.config,
          additionalSpotPools: JSON.stringify(this.additionalSpotPools),
        };

        if (this.value?.id) {
          // mgmtClusterId is the Norman v3 cluster ID (e.g. c-w467n).
          // this.value.id is the Steve provisioning cluster name which is NOT a valid Norman ID.
          const mgmtId = this.value?.mgmtClusterId
            || this.value?.mgmt?.id
            || this.value?.id;
          const existing = await this.$store.dispatch('rancher/find', {
            type: 'cluster',
            id:   mgmtId,
            opt:  { force: true },
          });
          existing.annotations = { ...(existing.annotations || {}), 'ui.rancher/provider': 'rackspacespot' };
          existing.rackspacespotEngineConfig = cfg;
          await existing.save();
        } else {
          const cluster = await this.$store.dispatch('rancher/create', {
            type:                      'cluster',
            name:                      this.clusterName,
            annotations:               { 'ui.rancher/provider': 'rackspacespot' },
            rackspacespotEngineConfig: cfg,
          });
          await cluster.save();
        }

        if (btnCb) btnCb(true);
        this.$router.push({ name: 'c-cluster-manager' });
      } catch (e) {
        this.errors = [e?.message || 'Failed to save cluster'];
        if (btnCb) btnCb(false);
      }
    },

    cancel() {
      this.$router.back();
    },
  },
});
</script>

<style scoped>
.pool-card {
  border: 1px solid var(--border);
  border-radius: 4px;
  padding: 12px 16px;
}
.pool-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.pool-label {
  font-weight: 600;
  font-size: 0.9em;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--default-text);
}
.mt-5  { margin-top: 5px; }
.mt-10 { margin-top: 10px; }
.mt-15 { margin-top: 15px; }
.mt-20 { margin-top: 20px; }
.mb-20 { margin-bottom: 20px; }
.spot-cost-link { font-size: 0.85em; color: var(--primary); text-decoration: none; }
.spot-cost-link:hover { text-decoration: underline; }
.import-note {
  font-size: 0.875rem;
  color: var(--muted);
}
</style>
