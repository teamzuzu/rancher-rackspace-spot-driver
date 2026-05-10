/* global Ember */
/* Rancher Cluster Driver — Rackspace Spot
 * AMD module: ui/components/cluster-driver/driver-rackspacespot/component
 *
 * Template pre-compiled with ember-source@3.24.7.
 * No {{input}} (not registered in Rancher's minimal Ember iframe context).
 * No {{each}} / -track-array. Uses plain <input> elements + explicit actions.
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

  // Pre-compiled with ember-source@3.24.7.
  // Upvars: saveButtonLabel, config, action, errorMessage, if,
  //         isCalicoSelected, isFlannelSelected, saving
  var COMPILED_TEMPLATE = {"id":"BQWYwZCT","block":"{\"symbols\":[],\"statements\":[[10,\"div\"],[14,0,\"driver-rackspacespot\"],[12],[2,\"\\n\"],[6,[37,4],[[35,3]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"banner bg-error mb-10\"],[12],[10,\"p\"],[12],[1,[34,3]],[13],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"h3\"],[12],[2,\"Authentication\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-6\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label required\"],[12],[2,\"Refresh Token\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"rackspaceSpotRefreshToken\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"rackspaceSpotRefreshToken\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"Rackspace Spot refresh token\"],[14,0,\"form-control\"],[14,4,\"password\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-6\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label required\"],[12],[2,\"Organization\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"rackspaceSpotOrganization\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"rackspaceSpotOrganization\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"your-org-name\"],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"Cluster\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Region\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"rackspaceSpotRegion\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"rackspaceSpotRegion\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Kubernetes Version\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"kubernetesVersion\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"kubernetesVersion\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"CNI\"],[13],[2,\"\\n      \"],[10,\"select\"],[14,0,\"form-control\"],[15,\"onchange\",[30,[36,2],[[32,0],\"setCNI\"],[[\"value\"],[\"target.value\"]]]],[12],[2,\"\\n        \"],[10,\"option\"],[14,2,\"calico\"],[15,\"selected\",[34,5]],[12],[2,\"calico\"],[13],[2,\"\\n        \"],[10,\"option\"],[14,2,\"flannel\"],[15,\"selected\",[34,6]],[12],[2,\"flannel\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"GPU Enabled\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[10,\"input\"],[15,\"checked\",[34,1,[\"gpuEnabled\"]]],[15,\"onchange\",[30,[36,2],[[32,0],\"setField\",\"gpuEnabled\"],[[\"value\"],[\"target.checked\"]]]],[14,4,\"checkbox\"],[12],[13],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Enable GPU support\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Preemption Webhook URL\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"preemptionWebhookUrl\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"preemptionWebhookUrl\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"https://...\"],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Deployment Type\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"deploymentType\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"deploymentType\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"optional\"],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"Spot Node Pool\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Pool Name\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotNodePoolName\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"spotNodePoolName\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Server Class\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotServerClass\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"spotServerClass\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Node Count\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotNodeCount\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setNumber\",\"spotNodeCount\"],[[\"value\"],[\"target.value\"]]]],[14,\"min\",\"0\"],[14,0,\"form-control\"],[14,4,\"number\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Bid Price (USD/hr)\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotBidPrice\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"spotBidPrice\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"0.50\"],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Enable Autoscaling\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[10,\"input\"],[15,\"checked\",[34,1,[\"spotAutoscalingEnabled\"]]],[15,\"onchange\",[30,[36,2],[[32,0],\"setField\",\"spotAutoscalingEnabled\"],[[\"value\"],[\"target.checked\"]]]],[14,4,\"checkbox\"],[12],[13],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Autoscale spot nodes\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[6,[37,4],[[35,1,[\"spotAutoscalingEnabled\"]]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Min Nodes\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotAutoscalingMinNodes\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setNumber\",\"spotAutoscalingMinNodes\"],[[\"value\"],[\"target.value\"]]]],[14,\"min\",\"0\"],[14,0,\"form-control\"],[14,4,\"number\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Max Nodes\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"spotAutoscalingMaxNodes\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setNumber\",\"spotAutoscalingMaxNodes\"],[[\"value\"],[\"target.value\"]]]],[14,\"min\",\"1\"],[14,0,\"form-control\"],[14,4,\"number\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"On-Demand Node Pool\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-12\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Enable On-Demand Pool\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[10,\"input\"],[15,\"checked\",[34,1,[\"onDemandEnabled\"]]],[15,\"onchange\",[30,[36,2],[[32,0],\"setField\",\"onDemandEnabled\"],[[\"value\"],[\"target.checked\"]]]],[14,4,\"checkbox\"],[12],[13],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Add stable capacity alongside spot nodes\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[6,[37,4],[[35,1,[\"onDemandEnabled\"]]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Pool Name\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"onDemandNodePoolName\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"onDemandNodePoolName\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Server Class\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"onDemandServerClass\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"onDemandServerClass\"],[[\"value\"],[\"target.value\"]]]],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Node Count\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"onDemandNodeCount\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setNumber\",\"onDemandNodeCount\"],[[\"value\"],[\"target.value\"]]]],[14,\"min\",\"0\"],[14,0,\"form-control\"],[14,4,\"number\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Price Per Hour (USD)\"],[13],[2,\"\\n      \"],[10,\"input\"],[15,2,[34,1,[\"onDemandPricePerHour\"]]],[15,\"oninput\",[30,[36,2],[[32,0],\"setField\",\"onDemandPricePerHour\"],[[\"value\"],[\"target.value\"]]]],[14,\"placeholder\",\"0.00\"],[14,0,\"form-control\"],[14,4,\"text\"],[12],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:30px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-12\"],[12],[2,\"\\n      \"],[10,\"button\"],[14,0,\"btn bg-primary\"],[15,\"onclick\",[30,[36,2],[[32,0],\"save\"],null]],[15,\"disabled\",[34,7]],[12],[2,\"\\n        \"],[6,[37,4],[[35,7]],null,[[\"default\",\"else\"],[{\"statements\":[[2,\"Saving\\u2026\"]],\"parameters\":[]},{\"statements\":[[1,[34,0]]],\"parameters\":[]}]]],[2,\"\\n      \"],[13],[2,\"\\n       \\n      \"],[10,\"button\"],[14,0,\"btn bg-transparent\"],[15,\"onclick\",[30,[36,2],[[32,0],\"cancel\"],null]],[12],[2,\"Cancel\"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[13],[2,\"\\n\"]],\"hasEval\":false,\"upvars\":[\"saveButtonLabel\",\"config\",\"action\",\"errorMessage\",\"if\",\"isCalicoSelected\",\"isFlannelSelected\",\"saving\"]}","moduleName":"ui/templates/components/cluster-driver/driver-rackspacespot/component"};

  var Component     = (_component    && _component.default)    || _component;
  var ClusterDriver = (_clusterDriver && _clusterDriver.default) || _clusterDriver;
  var layout        = Ember.HTMLBars.template(COMPILED_TEMPLATE);

  exports.default = Component.extend(ClusterDriver, {
    layout:     layout,
    driverName: DRIVER,

    config: Ember.computed('cluster.genericEngineConfig', function() {
      return this.get('cluster.genericEngineConfig');
    }),

    errorMessage: Ember.computed('errors.[]', function() {
      var errors = this.get('errors') || [];
      return errors.length ? errors.join(' ') : null;
    }),

    isCalicoSelected: Ember.computed('config.cni', function() {
      return (this.get('config.cni') || 'calico') === 'calico';
    }),

    isFlannelSelected: Ember.computed('config.cni', function() {
      return this.get('config.cni') === 'flannel';
    }),

    init: function() {
      this._super.apply(this, arguments);

      var config = this.get('cluster.genericEngineConfig');
      if (!config) {
        config = Ember.Object.create(Object.assign({}, DEFAULTS));
        this.set('cluster.genericEngineConfig', config);
        return;
      }

      Object.keys(DEFAULTS).forEach(function(k) {
        var v = config.get(k);
        if (v === undefined || v === null || v === '') {
          config.set(k, DEFAULTS[k]);
        }
      });
    },

    validate: function() {
      this._super.apply(this, arguments);

      var errors = this.get('errors') || [];
      var config = this.get('config');

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
      setCNI: function(val) {
        var config = this.get('config');
        if (config) { config.set('cni', val); }
      },
      setField: function(field, value) {
        var config = this.get('config');
        if (config) { config.set(field, value); }
      },
      setNumber: function(field, value) {
        var config = this.get('config');
        if (config) { config.set(field, parseInt(value, 10) || 0); }
      },
    },
  });
});
