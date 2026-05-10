<template>
  <div class="driver-rackspacespot">
    <Banner
      v-if="validationErrors.length"
      color="error"
      :label="validationErrors.join(' | ')"
    />

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
        <LabeledInput
          v-model:value="config.kubernetesVersion"
          label="Kubernetes Version"
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

    <!-- ── Spot Node Pool ──────────────────────────────────── -->
    <h3 class="mt-20">Spot Node Pool</h3>
    <div class="row">
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.spotNodePoolName"
          label="Pool Name"
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.spotServerClass"
          label="Server Class"
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
    </div>
    <div class="row mt-10">
      <div class="col span-4">
        <LabeledInput
          v-model:value="config.spotBidPrice"
          label="Bid Price (USD/hr)"
          placeholder="0.50"
          :mode="mode"
        />
      </div>
      <div class="col span-4">
        <Checkbox
          v-model:value="config.spotAutoscalingEnabled"
          label="Enable Autoscaling"
          :mode="mode"
        />
      </div>
    </div>
    <div v-if="config.spotAutoscalingEnabled" class="row mt-10">
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
            v-model:value="config.onDemandNodePoolName"
            label="Pool Name"
            :mode="mode"
          />
        </div>
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
      </div>
      <div class="row mt-10">
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
  </div>
</template>

<script>
import { defineComponent } from 'vue';
import Banner        from '@components/Banner/Banner';
import LabeledInput  from '@components/Form/LabeledInput/LabeledInput';
import LabeledSelect from '@shell/components/form/LabeledSelect';
import Checkbox      from '@components/Form/Checkbox/Checkbox';

const DRIVER   = 'rackspacespot';
const DEFAULTS = {
  driverName:              DRIVER,
  rackspaceSpotRegion:     'colo-lax-1',
  kubernetesVersion:       '1.28',
  cni:                     'calico',
  gpuEnabled:              false,
  spotNodePoolName:        'spot-pool',
  spotServerClass:         'rxtx.4xlarge-mi300x',
  spotNodeCount:           3,
  spotBidPrice:            '0.50',
  spotAutoscalingEnabled:  false,
  spotAutoscalingMinNodes: 1,
  spotAutoscalingMaxNodes: 10,
  onDemandEnabled:         false,
  onDemandNodePoolName:    'on-demand-pool',
  onDemandNodeCount:       1,
};

export default defineComponent({
  name: 'CruRackspaceSpot',

  components: {
    Banner,
    LabeledInput,
    LabeledSelect,
    Checkbox,
  },

  props: {
    // provisioning.cattle.io.cluster or management.cattle.io.cluster value
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
    return {
      validationErrors: [],
      cniOptions:       [
        { label: 'calico', value: 'calico' },
        { label: 'flannel', value: 'flannel' },
      ],
    };
  },

  computed: {
    // Reactive proxy into value.genericEngineConfig with defaults merged in.
    config() {
      if (!this.value.genericEngineConfig) {
        // Initialise on first access so Vue can track it reactively.
        this.$set
          ? this.$set(this.value, 'genericEngineConfig', { ...DEFAULTS })
          : (this.value.genericEngineConfig = { ...DEFAULTS });
      }

      // Backfill any missing keys without overwriting user data.
      const cfg = this.value.genericEngineConfig;
      Object.keys(DEFAULTS).forEach((k) => {
        if (cfg[k] === undefined || cfg[k] === null) {
          const set = this.$set || ((obj, key, val) => { obj[key] = val; });
          set(cfg, k, DEFAULTS[k]);
        }
      });

      return cfg;
    },
  },

  methods: {
    // Called by the parent provisioner / CruResource before the cluster is saved.
    validate() {
      this.validationErrors = [];
      if (!this.config.rackspaceSpotRefreshToken) {
        this.validationErrors.push('Refresh Token is required');
      }
      if (!this.config.rackspaceSpotOrganization) {
        this.validationErrors.push('Organization is required');
      }
      return this.validationErrors.length === 0;
    },
  },
});
</script>

<style scoped>
.driver-rackspacespot {
  padding: 10px 0;
}
.mt-10 { margin-top: 10px; }
.mt-20 { margin-top: 20px; }
</style>
