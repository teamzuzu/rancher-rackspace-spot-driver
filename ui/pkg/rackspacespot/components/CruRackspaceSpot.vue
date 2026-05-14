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
          v-model:value="config.rackspaceSpotRefreshToken"
          label="Refresh Token"
          type="password"
          placeholder="Rackspace Spot refresh token"
          :required="true"
          :mode="mode"
        />
      </div>
      <div class="col span-6">
        <LabeledInput
          v-model:value="config.rackspaceSpotOrganization"
          label="Organization"
          placeholder="your-org-name"
          :required="true"
          :mode="mode"
        />
      </div>
    </div>

    <!-- ── Cluster ─────────────────────────────────────────── -->
    <h3 class="mt-20">Cluster</h3>
    <div class="row">
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.rackspaceSpotRegion"
          label="Region"
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <LabeledSelect
          v-model:value="config.kubernetesVersion"
          label="Kubernetes Version"
          :options="k8sVersionOptions"
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <LabeledSelect
          v-model:value="config.cni"
          label="CNI"
          :options="cniOptions"
          :mode="mode"
        />
      </div>
    </div>
    <div class="row mt-10">
      <div class="col span-4">
        <Checkbox
          v-model:value="config.gpuEnabled"
          label="Enable GPU support"
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.preemptionWebhookUrl"
          label="Preemption Webhook URL"
          placeholder="https://..."
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.deploymentType"
          label="Deployment Type"
          placeholder="optional"
          :mode="mode"
        />
      </div>
    </div>

    <!-- ── Server Classes & Pricing ──────────────────────── -->
    <div class="section-header mt-20">
      <h3>Spot Node Pools</h3>
      <button
        class="btn btn-sm btn-default"
        type="button"
        :disabled="!config.rackspaceSpotRefreshToken || priceLoading"
        @click="loadServerClasses"
      >
        {{ priceLoading ? 'Loading…' : (serverClasses.length ? 'Refresh Prices' : 'Load Server Classes & Prices') }}
      </button>
    </div>
    <div v-if="priceError" class="price-error mt-5">
      {{ priceError }}
    </div>
    <div v-if="serverClasses.length" class="price-table-wrap mt-10">
      <table class="price-table">
        <thead>
          <tr>
            <th>Server Class</th>
            <th>Region</th>
            <th>CPU</th>
            <th>Memory</th>
            <th>Market $/hr</th>
            <th>Min Bid $/hr</th>
            <th>On-Demand $/hr</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="sc in filteredServerClasses" :key="sc.name">
            <td><code>{{ sc.name }}</code></td>
            <td>{{ sc.region }}</td>
            <td>{{ sc.cpu }}</td>
            <td>{{ sc.memory }}</td>
            <td class="price">{{ sc.marketPrice }}</td>
            <td class="price">{{ sc.minBidPrice }}</td>
            <td class="price">{{ sc.onDemandPrice }}</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- ── Primary spot pool ─────────────────────────────── -->
    <div class="pool-card mt-15">
      <div class="pool-card-header">
        <span class="pool-label">Pool 1 (primary)</span>
      </div>
      <div class="row mt-10">
        <div class="col span-4">
          <LabeledInput
            v-model:value="config.spotServerClass"
            label="Server Class"
            :placeholder="serverClassPlaceholder"
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
            placeholder="0.50"
            :mode="mode"
          />
        </div>
      </div>
      <div class="row mt-10">
        <div class="col span-4">
          <Checkbox
            v-model:value="config.spotAutoscalingEnabled"
            label="Enable Autoscaling"
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
          <LabeledInput
            v-model:value="pool.serverClass"
            label="Server Class"
            :placeholder="serverClassPlaceholder"
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
            placeholder="0.50"
            :mode="mode"
          />
        </div>
      </div>
      <div class="row mt-10">
        <div class="col span-4">
          <Checkbox
            v-model:value="pool.autoscaling"
            label="Enable Autoscaling"
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
          <LabeledInput
            v-model:value="config.onDemandServerClass"
            label="Server Class"
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
  </CruResource>
</template>

<script>
import { defineComponent } from 'vue';
import CruResource       from '@shell/components/CruResource';
import Banner            from '@components/Banner/Banner';
import LabeledInput      from '@components/Form/LabeledInput/LabeledInput';
import LabeledSelect     from '@shell/components/form/LabeledSelect';
import Checkbox          from '@components/Form/Checkbox/Checkbox';

const DRIVER   = 'rackspacespot';
const AUTH_URL = 'https://login.spot.rackspace.com';
const API_URL  = 'https://spot.rackspace.com';
const CLIENT_ID = 'mwG3lUMV8KyeMqHe4fJ5Bb3nM1vBvRNa';

const DEFAULTS = {
  driverName:              DRIVER,
  rackspaceSpotRegion:     'us-east-iad-1',
  kubernetesVersion:       '1.33.0',
  cni:                     'calico',
  gpuEnabled:              false,
  spotServerClass:         'ch.vs1.medium-iad',
  spotNodeCount:           1,
  spotBidPrice:            '0.50',
  spotAutoscalingEnabled:  false,
  spotAutoscalingMinNodes: 1,
  spotAutoscalingMaxNodes: 10,
  onDemandEnabled:         false,
  onDemandNodeCount:       1,
};

function emptyPool() {
  return {
    serverClass:  '',
    nodeCount:    1,
    bidPrice:     '0.50',
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
      serverClasses:        [],
      priceLoading:         false,
      priceError:           null,
      cniOptions:           [
        { label: 'calico', value: 'calico' },
        { label: 'cilium', value: 'cilium' },
        { label: 'byocni', value: 'byocni' },
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
    filteredServerClasses() {
      const region = this.config.rackspaceSpotRegion;
      if (!region) return this.serverClasses;
      return this.serverClasses.filter(sc => sc.region === region);
    },
    serverClassPlaceholder() {
      if (this.serverClasses.length) {
        const first = this.filteredServerClasses[0];
        return first ? first.name : 'e.g. ch.vs1.medium-iad';
      }
      return 'e.g. ch.vs1.medium-iad';
    },
  },

  methods: {
    addSpotPool() {
      this.additionalSpotPools.push(emptyPool());
    },

    removeSpotPool(idx) {
      this.additionalSpotPools.splice(idx, 1);
    },

    async loadServerClasses() {
      if (!this.config.rackspaceSpotRefreshToken) return;
      this.priceLoading = true;
      this.priceError   = null;
      try {
        const authResp = await fetch(`${AUTH_URL}/oauth/token`, {
          method:  'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body:    new URLSearchParams({
            grant_type:    'refresh_token',
            client_id:     CLIENT_ID,
            refresh_token: this.config.rackspaceSpotRefreshToken,
          }),
        });
        if (!authResp.ok) throw new Error(`Auth failed: ${authResp.status}`);
        const authData = await authResp.json();
        const token = authData.id_token;
        if (!token) throw new Error('No id_token in auth response');

        const scResp = await fetch(`${API_URL}/apis/ngpc.rxt.io/v1/serverclasses`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!scResp.ok) throw new Error(`Server class fetch failed: ${scResp.status}`);
        const scData = await scResp.json();

        this.serverClasses = (scData.items || [])
          .filter(i => i.spec?.availability === 'available')
          .map(i => ({
            name:          i.metadata?.name || '',
            region:        i.spec?.region || '',
            cpu:           i.spec?.resources?.cpu || '',
            memory:        i.spec?.resources?.memory || '',
            marketPrice:   i.status?.spotPricing?.marketPricePerHour
              ? `$${i.status.spotPricing.marketPricePerHour}`
              : 'N/A',
            minBidPrice:   i.spec?.minBidPricePerHour
              ? `$${i.spec.minBidPricePerHour}`
              : 'N/A',
            onDemandPrice: i.spec?.onDemandPricing?.cost
              ? `$${i.spec.onDemandPricing.cost}`
              : 'N/A',
          }))
          .sort((a, b) => a.name.localeCompare(b.name));
      } catch (e) {
        this.priceError = `Could not load server classes: ${e.message}`;
      } finally {
        this.priceLoading = false;
      }
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
          existing.rackspacespotEngineConfig = cfg;
          await existing.save();
        } else {
          const cluster = await this.$store.dispatch('rancher/create', {
            type:                      'cluster',
            name:                      this.clusterName,
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
.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.section-header h3 {
  margin: 0;
}
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
.price-table-wrap {
  overflow-x: auto;
}
.price-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85em;
}
.price-table th,
.price-table td {
  padding: 4px 10px;
  border: 1px solid var(--border);
  text-align: left;
  white-space: nowrap;
}
.price-table th {
  background: var(--accent-btn);
  font-weight: 600;
}
.price-table .price {
  font-family: monospace;
}
.price-error {
  color: var(--error);
  font-size: 0.85em;
}
.mt-5  { margin-top: 5px; }
.mt-10 { margin-top: 10px; }
.mt-15 { margin-top: 15px; }
.mt-20 { margin-top: 20px; }
.mb-20 { margin-bottom: 20px; }
</style>
