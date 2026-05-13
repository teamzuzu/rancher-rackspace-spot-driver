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
    <div
      v-if="config.spotAutoscalingEnabled"
      class="row mt-10"
    >
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
    return {
      clusterName: this.value?.displayName || this.value?.name || '',
      config:      { ...DEFAULTS, ...(this.value?.rackspacespotEngineConfig || {}) },
      errors:      [],
      cniOptions:  [
        { label: 'calico', value: 'calico' },
        { label: 'flannel', value: 'flannel' },
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
  },

  methods: {
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
        const cfg = { ...this.config };

        if (this.value?.id) {
          const existing = await this.$store.dispatch('rancher/find', {
            type: 'cluster',
            id:   this.value.id,
            opt:  { force: true },
          });
          existing.rackspacespotEngineConfig = cfg;
          await existing.save();
        } else {
          const cluster = await this.$store.dispatch('rancher/create', {
            type:                       'cluster',
            name:                       this.clusterName,
            rackspacespotEngineConfig:  cfg,
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
.driver-rackspacespot {
  padding: 10px 0;
}
.mt-10  { margin-top: 10px; }
.mt-20  { margin-top: 20px; }
.mb-20  { margin-bottom: 20px; }
</style>
