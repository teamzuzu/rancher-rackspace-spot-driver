/* global Ember */
/* Rancher Cluster Driver — Rackspace Spot
 * AMD module ID: ui/components/cluster-driver/driver-rackspacespot/component
 * Loaded by Rancher as the Custom UI URL for the rackspacespot cluster driver.
 */
define('ui/components/cluster-driver/driver-rackspacespot/component', [
  'exports',
  '@ember/component',
  'shared/mixins/cluster-driver',
], function(exports, _component, _clusterDriver) {
  'use strict';

  var DRIVER = 'rackspacespot';

  var DEFAULTS = {
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

  var TEMPLATE = '\
<div class="driver-rackspacespot">\
  {{#if errors}}\
  <div class="banner bg-error mb-10">\
    {{#each errors as |err|}}<p>{{err}}</p>{{/each}}\
  </div>\
  {{/if}}\
\
  <h3>Authentication</h3>\
  <div class="row">\
    <div class="col span-6">\
      <label class="acc-label required">Refresh Token</label>\
      {{input type="password" value=config.rackspaceSpotRefreshToken placeholder="Rackspace Spot refresh token" class="form-control"}}\
    </div>\
    <div class="col span-6">\
      <label class="acc-label required">Organization</label>\
      {{input value=config.rackspaceSpotOrganization placeholder="your-org-name" class="form-control"}}\
    </div>\
  </div>\
\
  <h3 style="margin-top:20px">Cluster</h3>\
  <div class="row">\
    <div class="col span-4">\
      <label class="acc-label">Region</label>\
      {{input value=config.rackspaceSpotRegion class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Kubernetes Version</label>\
      {{input value=config.kubernetesVersion class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">CNI</label>\
      <select class="form-control" onchange={{action (mut config.cni) value="target.value"}}>\
        <option value="calico" selected={{eq config.cni "calico"}}>calico</option>\
        <option value="flannel" selected={{eq config.cni "flannel"}}>flannel</option>\
      </select>\
    </div>\
  </div>\
  <div class="row" style="margin-top:10px">\
    <div class="col span-4">\
      <label class="acc-label">GPU Enabled</label>\
      <div>\
        {{input type="checkbox" checked=config.gpuEnabled}}\
        <span style="margin-left:6px">Enable GPU support</span>\
      </div>\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Preemption Webhook URL</label>\
      {{input value=config.preemptionWebhookUrl placeholder="https://..." class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Deployment Type</label>\
      {{input value=config.deploymentType placeholder="optional" class="form-control"}}\
    </div>\
  </div>\
\
  <h3 style="margin-top:20px">Spot Node Pool</h3>\
  <div class="row">\
    <div class="col span-4">\
      <label class="acc-label">Pool Name</label>\
      {{input value=config.spotNodePoolName class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Server Class</label>\
      {{input value=config.spotServerClass class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Node Count</label>\
      {{input type="number" value=config.spotNodeCount min="0" class="form-control"}}\
    </div>\
  </div>\
  <div class="row" style="margin-top:10px">\
    <div class="col span-4">\
      <label class="acc-label">Bid Price (USD/hr)</label>\
      {{input value=config.spotBidPrice placeholder="0.50" class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Enable Autoscaling</label>\
      <div>\
        {{input type="checkbox" checked=config.spotAutoscalingEnabled}}\
        <span style="margin-left:6px">Autoscale spot nodes</span>\
      </div>\
    </div>\
  </div>\
  {{#if config.spotAutoscalingEnabled}}\
  <div class="row" style="margin-top:10px">\
    <div class="col span-4">\
      <label class="acc-label">Min Nodes</label>\
      {{input type="number" value=config.spotAutoscalingMinNodes min="0" class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Max Nodes</label>\
      {{input type="number" value=config.spotAutoscalingMaxNodes min="1" class="form-control"}}\
    </div>\
  </div>\
  {{/if}}\
\
  <h3 style="margin-top:20px">On-Demand Node Pool</h3>\
  <div class="row">\
    <div class="col span-12">\
      <label class="acc-label">Enable On-Demand Pool</label>\
      <div>\
        {{input type="checkbox" checked=config.onDemandEnabled}}\
        <span style="margin-left:6px">Add an on-demand pool for stable baseline capacity</span>\
      </div>\
    </div>\
  </div>\
  {{#if config.onDemandEnabled}}\
  <div class="row" style="margin-top:10px">\
    <div class="col span-4">\
      <label class="acc-label">Pool Name</label>\
      {{input value=config.onDemandNodePoolName class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Server Class</label>\
      {{input value=config.onDemandServerClass class="form-control"}}\
    </div>\
    <div class="col span-4">\
      <label class="acc-label">Node Count</label>\
      {{input type="number" value=config.onDemandNodeCount min="0" class="form-control"}}\
    </div>\
  </div>\
  <div class="row" style="margin-top:10px">\
    <div class="col span-4">\
      <label class="acc-label">Price Per Hour (USD)</label>\
      {{input value=config.onDemandPricePerHour placeholder="0.00" class="form-control"}}\
    </div>\
  </div>\
  {{/if}}\
\
  <div class="row" style="margin-top:30px">\
    <div class="col span-12">\
      <button class="btn bg-primary" onclick={{action "save"}} disabled={{saving}}>\
        {{#if saving}}Saving…{{else}}{{saveButtonLabel}}{{/if}}\
      </button>\
      &nbsp;\
      <button class="btn bg-transparent" onclick={{action "cancel"}}>Cancel</button>\
    </div>\
  </div>\
</div>';

  var compiled;
  try {
    compiled = Ember.HTMLBars.compile(TEMPLATE);
  } catch (e) {
    // eslint-disable-next-line no-console
    console.error('rackspacespot: failed to compile template', e);
  }

  exports.default = _component.default.extend(_clusterDriver.default, {
    layout:     compiled,
    driverName: DRIVER,

    config: Ember.computed('cluster.genericEngineConfig', function() {
      return this.get('cluster.genericEngineConfig');
    }),

    init: function() {
      this._super.apply(this, arguments);

      var config = this.get('cluster.genericEngineConfig');
      if (!config) {
        config = Ember.Object.create(Object.assign({}, DEFAULTS));
        this.set('cluster.genericEngineConfig', config);
        return;
      }

      // Fill in any missing defaults without overwriting existing values.
      Object.keys(DEFAULTS).forEach(function(k) {
        var v = config.get(k);
        if (v === undefined || v === null || v === '') {
          config.set(k, DEFAULTS[k]);
        }
      });
    },

    validate: function() {
      this._super.apply(this, arguments);

      var errors  = this.get('errors') || [];
      var config  = this.get('config');

      if (!config || !config.get('rackspaceSpotRefreshToken')) {
        errors.push('Refresh Token is required');
      }
      if (!config || !config.get('rackspaceSpotOrganization')) {
        errors.push('Organization is required');
      }

      this.set('errors', errors);
      return errors.length === 0;
    },

    actions: {
      save: function(cb) {
        this.send('driverSave', cb);
      },
      cancel: function() {
        this.send('driverCancel');
      },
    },
  });
});
