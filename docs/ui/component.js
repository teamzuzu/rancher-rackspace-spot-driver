/* global Ember */
/* Rancher Cluster Driver — Rackspace Spot
 * AMD module: ui/components/cluster-driver/driver-rackspacespot/component
 *
 * Template pre-compiled with ember-source@3.24.7 (Rancher's Ember version).
 * Ember.HTMLBars.template() is available in both dev and production builds;
 * Ember.HTMLBars.compile() is NOT available in production, so we avoid it.
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

  // Pre-compiled with: ember-source@3.24.7
  // ember-template-compiler.precompile(template, { moduleName: '...' })
  var COMPILED_TEMPLATE = {
    "id": "IQR3xs+i",
    "block": "{\"symbols\":[\"opt\",\"err\"],\"statements\":[[10,\"div\"],[14,0,\"driver-rackspacespot\"],[12],[2,\"\\n\"],[6,[37,6],[[35,3]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"banner bg-error mb-10\"],[12],[2,\"\\n    \"],[6,[37,5],[[30,[36,4],[[30,[36,4],[[35,3]],null]],null]],null,[[\"default\"],[{\"statements\":[[10,\"p\"],[12],[1,[32,2]],[13]],\"parameters\":[2]}]]],[2,\"\\n  \"],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"h3\"],[12],[2,\"Authentication\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-6\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label required\"],[12],[2,\"Refresh Token\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"type\",\"value\",\"placeholder\",\"class\"],[\"password\",[35,1,[\"rackspaceSpotRefreshToken\"]],\"Rackspace Spot refresh token\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-6\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label required\"],[12],[2,\"Organization\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"placeholder\",\"class\"],[[35,1,[\"rackspaceSpotOrganization\"]],\"your-org-name\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"Cluster\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Region\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"rackspaceSpotRegion\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Kubernetes Version\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"kubernetesVersion\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"CNI\"],[13],[2,\"\\n      \"],[10,\"select\"],[14,0,\"form-control\"],[15,\"onchange\",[30,[36,7],[[32,0],\"setCNI\"],[[\"value\"],[\"target.value\"]]]],[12],[2,\"\\n\"],[6,[37,5],[[30,[36,4],[[30,[36,4],[[35,8]],null]],null]],null,[[\"default\"],[{\"statements\":[[2,\"          \"],[10,\"option\"],[15,2,[32,1,[\"value\"]]],[15,\"selected\",[32,1,[\"selected\"]]],[12],[1,[32,1,[\"label\"]]],[13],[2,\"\\n\"]],\"parameters\":[1]}]]],[2,\"      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"GPU Enabled\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[1,[30,[36,2],null,[[\"type\",\"checked\"],[\"checkbox\",[35,1,[\"gpuEnabled\"]]]]]],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Enable GPU support\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Preemption Webhook URL\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"placeholder\",\"class\"],[[35,1,[\"preemptionWebhookUrl\"]],\"https://...\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Deployment Type\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"placeholder\",\"class\"],[[35,1,[\"deploymentType\"]],\"optional\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"Spot Node Pool\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Pool Name\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"spotNodePoolName\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Server Class\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"spotServerClass\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Node Count\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"type\",\"value\",\"min\",\"class\"],[\"number\",[35,1,[\"spotNodeCount\"]],\"0\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Bid Price (USD/hr)\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"placeholder\",\"class\"],[[35,1,[\"spotBidPrice\"]],\"0.50\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Enable Autoscaling\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[1,[30,[36,2],null,[[\"type\",\"checked\"],[\"checkbox\",[35,1,[\"spotAutoscalingEnabled\"]]]]]],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Autoscale spot nodes\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[6,[37,6],[[35,1,[\"spotAutoscalingEnabled\"]]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Min Nodes\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"type\",\"value\",\"min\",\"class\"],[\"number\",[35,1,[\"spotAutoscalingMinNodes\"]],\"0\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Max Nodes\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"type\",\"value\",\"min\",\"class\"],[\"number\",[35,1,[\"spotAutoscalingMaxNodes\"]],\"1\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"h3\"],[14,5,\"margin-top:20px\"],[12],[2,\"On-Demand Node Pool\"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-12\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Enable On-Demand Pool\"],[13],[2,\"\\n      \"],[10,\"div\"],[12],[2,\"\\n        \"],[1,[30,[36,2],null,[[\"type\",\"checked\"],[\"checkbox\",[35,1,[\"onDemandEnabled\"]]]]]],[2,\"\\n        \"],[10,\"span\"],[14,5,\"margin-left:6px\"],[12],[2,\"Add stable capacity alongside spot nodes\"],[13],[2,\"\\n      \"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[6,[37,6],[[35,1,[\"onDemandEnabled\"]]],null,[[\"default\"],[{\"statements\":[[2,\"  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Pool Name\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"onDemandNodePoolName\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Server Class\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"class\"],[[35,1,[\"onDemandServerClass\"]],\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Node Count\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"type\",\"value\",\"min\",\"class\"],[\"number\",[35,1,[\"onDemandNodeCount\"]],\"0\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:10px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-4\"],[12],[2,\"\\n      \"],[10,\"label\"],[14,0,\"acc-label\"],[12],[2,\"Price Per Hour (USD)\"],[13],[2,\"\\n      \"],[1,[30,[36,2],null,[[\"value\",\"placeholder\",\"class\"],[[35,1,[\"onDemandPricePerHour\"]],\"0.00\",\"form-control\"]]]],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"]],\"parameters\":[]}]]],[2,\"\\n  \"],[10,\"div\"],[14,0,\"row\"],[14,5,\"margin-top:30px\"],[12],[2,\"\\n    \"],[10,\"div\"],[14,0,\"col span-12\"],[12],[2,\"\\n      \"],[11,\"button\"],[24,0,\"btn bg-primary\"],[16,\"disabled\",[34,9]],[4,[38,7],[[32,0],\"save\"],null],[12],[2,\"\\n        \"],[6,[37,6],[[35,9]],null,[[\"default\",\"else\"],[{\"statements\":[[2,\"Saving\\u2026\"]],\"parameters\":[]},{\"statements\":[[1,[34,0]]],\"parameters\":[]}]]],[2,\"\\n      \"],[13],[2,\"\\n       \\n      \"],[11,\"button\"],[24,0,\"btn bg-transparent\"],[4,[38,7],[[32,0],\"cancel\"],null],[12],[2,\"Cancel\"],[13],[2,\"\\n    \"],[13],[2,\"\\n  \"],[13],[2,\"\\n\"],[13],[2,\"\\n\"]],\"hasEval\":false,\"upvars\":[\"saveButtonLabel\",\"config\",\"input\",\"errors\",\"-track-array\",\"each\",\"if\",\"action\",\"cniOptions\",\"saving\"]}",
    "moduleName": "ui/templates/components/cluster-driver/driver-rackspacespot/component"
  };

  var Component    = (_component    && _component.default)    || _component;
  var ClusterDriver = (_clusterDriver && _clusterDriver.default) || _clusterDriver;
  var layout       = Ember.HTMLBars.template(COMPILED_TEMPLATE);

  exports.default = Component.extend(ClusterDriver, {
    layout:     layout,
    driverName: DRIVER,

    config: Ember.computed('cluster.genericEngineConfig', function() {
      return this.get('cluster.genericEngineConfig');
    }),

    cniOptions: Ember.computed('config.cni', function() {
      var current = this.get('config.cni') || 'calico';
      return [
        { value: 'calico',  label: 'calico',  selected: current === 'calico'  },
        { value: 'flannel', label: 'flannel', selected: current === 'flannel' },
      ];
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
    },
  });
});
