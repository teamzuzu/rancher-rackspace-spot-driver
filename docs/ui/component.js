/* global Ember */
/* Rancher Cluster Driver — Rackspace Spot
 * AMD module: ui/components/cluster-driver/driver-rackspacespot/component
 *
 * Renders entirely via didInsertElement + innerHTML to avoid all Glimmer VM
 * helper resolution (the component owner is not fully wired when loaded as
 * an external AMD module, causing null manager errors for every helper).
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

  // Empty pre-compiled template — no opcodes, no helper lookups.
  var EMPTY_LAYOUT = {"id":"iiOJ48wv","block":"{\"symbols\":[],\"statements\":[],\"hasEval\":false,\"upvars\":[]}","moduleName":"empty"};

  var Component     = (_component    && _component.default)    || _component;
  var ClusterDriver = (_clusterDriver && _clusterDriver.default) || _clusterDriver;

  var layout = null;
  try { layout = Ember.HTMLBars.template(EMPTY_LAYOUT); } catch(e) {}

  function esc(v) {
    return String(v === null || v === undefined ? '' : v)
      .replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  function buildHTML(config) {
    function v(k) { return esc(config ? config.get(k) : DEFAULTS[k]); }
    function chk(k) { return (config ? config.get(k) : DEFAULTS[k]) ? ' checked' : ''; }
    function sel(k, val) { return (config ? (config.get(k) || 'calico') : 'calico') === val ? ' selected' : ''; }

    var autoVis  = (config ? config.get('spotAutoscalingEnabled') : false) ? '' : 'display:none';
    var demandVis = (config ? config.get('onDemandEnabled') : false) ? '' : 'display:none';

    return '<div class="driver-rackspacespot">'
      + '<div id="rsp-errors" class="banner bg-error mb-10" style="display:none"><p id="rsp-error-text"></p></div>'

      + '<h3>Authentication</h3>'
      + '<div class="row">'
      +   '<div class="col span-6"><label class="acc-label required">Refresh Token</label>'
      +     '<input id="rsp-token" type="password" class="form-control" placeholder="Rackspace Spot refresh token" value="' + v('rackspaceSpotRefreshToken') + '"></div>'
      +   '<div class="col span-6"><label class="acc-label required">Organization</label>'
      +     '<input id="rsp-org" type="text" class="form-control" placeholder="your-org-name" value="' + v('rackspaceSpotOrganization') + '"></div>'
      + '</div>'

      + '<h3 style="margin-top:20px">Cluster</h3>'
      + '<div class="row">'
      +   '<div class="col span-4"><label class="acc-label">Region</label>'
      +     '<input id="rsp-region" type="text" class="form-control" value="' + v('rackspaceSpotRegion') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Kubernetes Version</label>'
      +     '<input id="rsp-k8s" type="text" class="form-control" value="' + v('kubernetesVersion') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">CNI</label>'
      +     '<select id="rsp-cni" class="form-control">'
      +       '<option value="calico"' + sel('cni','calico') + '>calico</option>'
      +       '<option value="flannel"' + sel('cni','flannel') + '>flannel</option>'
      +     '</select></div>'
      + '</div>'
      + '<div class="row" style="margin-top:10px">'
      +   '<div class="col span-4"><label class="acc-label">GPU Enabled</label><div>'
      +     '<input id="rsp-gpu" type="checkbox"' + chk('gpuEnabled') + '>'
      +     '<span style="margin-left:6px">Enable GPU support</span></div></div>'
      +   '<div class="col span-4"><label class="acc-label">Preemption Webhook URL</label>'
      +     '<input id="rsp-webhook" type="text" class="form-control" placeholder="https://..." value="' + v('preemptionWebhookUrl') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Deployment Type</label>'
      +     '<input id="rsp-deploy" type="text" class="form-control" placeholder="optional" value="' + v('deploymentType') + '"></div>'
      + '</div>'

      + '<h3 style="margin-top:20px">Spot Node Pool</h3>'
      + '<div class="row">'
      +   '<div class="col span-4"><label class="acc-label">Pool Name</label>'
      +     '<input id="rsp-spname" type="text" class="form-control" value="' + v('spotNodePoolName') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Server Class</label>'
      +     '<input id="rsp-spclass" type="text" class="form-control" value="' + v('spotServerClass') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Node Count</label>'
      +     '<input id="rsp-spcount" type="number" class="form-control" min="0" value="' + v('spotNodeCount') + '"></div>'
      + '</div>'
      + '<div class="row" style="margin-top:10px">'
      +   '<div class="col span-4"><label class="acc-label">Bid Price (USD/hr)</label>'
      +     '<input id="rsp-bid" type="text" class="form-control" placeholder="0.50" value="' + v('spotBidPrice') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Enable Autoscaling</label><div>'
      +     '<input id="rsp-autoscale" type="checkbox"' + chk('spotAutoscalingEnabled') + '>'
      +     '<span style="margin-left:6px">Autoscale spot nodes</span></div></div>'
      + '</div>'
      + '<div id="rsp-autoscale-section" class="row" style="margin-top:10px;' + autoVis + '">'
      +   '<div class="col span-4"><label class="acc-label">Min Nodes</label>'
      +     '<input id="rsp-amin" type="number" class="form-control" min="0" value="' + v('spotAutoscalingMinNodes') + '"></div>'
      +   '<div class="col span-4"><label class="acc-label">Max Nodes</label>'
      +     '<input id="rsp-amax" type="number" class="form-control" min="1" value="' + v('spotAutoscalingMaxNodes') + '"></div>'
      + '</div>'

      + '<h3 style="margin-top:20px">On-Demand Node Pool</h3>'
      + '<div class="row">'
      +   '<div class="col span-12"><label class="acc-label">Enable On-Demand Pool</label><div>'
      +     '<input id="rsp-ondemand" type="checkbox"' + chk('onDemandEnabled') + '>'
      +     '<span style="margin-left:6px">Add stable capacity alongside spot nodes</span></div></div>'
      + '</div>'
      + '<div id="rsp-demand-section" style="' + demandVis + '">'
      +   '<div class="row" style="margin-top:10px">'
      +     '<div class="col span-4"><label class="acc-label">Pool Name</label>'
      +       '<input id="rsp-odname" type="text" class="form-control" value="' + v('onDemandNodePoolName') + '"></div>'
      +     '<div class="col span-4"><label class="acc-label">Server Class</label>'
      +       '<input id="rsp-odclass" type="text" class="form-control" value="' + v('onDemandServerClass') + '"></div>'
      +     '<div class="col span-4"><label class="acc-label">Node Count</label>'
      +       '<input id="rsp-odcount" type="number" class="form-control" min="0" value="' + v('onDemandNodeCount') + '"></div>'
      +   '</div>'
      +   '<div class="row" style="margin-top:10px">'
      +     '<div class="col span-4"><label class="acc-label">Price Per Hour (USD)</label>'
      +       '<input id="rsp-odprice" type="text" class="form-control" placeholder="0.00" value="' + v('onDemandPricePerHour') + '"></div>'
      +   '</div>'
      + '</div>'

      + '<div class="row" style="margin-top:30px">'
      +   '<div class="col span-12">'
      +     '<button id="rsp-save" class="btn bg-primary">Create</button>'
      +     '&nbsp;'
      +     '<button id="rsp-cancel" class="btn bg-transparent">Cancel</button>'
      +   '</div>'
      + '</div>'
      + '</div>';
  }

  exports.default = Component.extend(ClusterDriver, {
    layout:     layout,
    driverName: DRIVER,

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

    didInsertElement: function() {
      this._super.apply(this, arguments);
      var config = this.get('config') || this.get('cluster.genericEngineConfig');
      var el = this.element;
      el.innerHTML = buildHTML(config);
      this._wire(el, config);
    },

    _wire: function(el, config) {
      var self = this;

      function txt(id, field) {
        var inp = el.querySelector('#' + id);
        if (inp) inp.addEventListener('input', function(e) { config.set(field, e.target.value); });
      }
      function num(id, field) {
        var inp = el.querySelector('#' + id);
        if (inp) inp.addEventListener('input', function(e) { config.set(field, parseInt(e.target.value, 10) || 0); });
      }
      function chk(id, field, onChange) {
        var inp = el.querySelector('#' + id);
        if (inp) inp.addEventListener('change', function(e) {
          config.set(field, e.target.checked);
          if (onChange) onChange(e.target.checked);
        });
      }

      txt('rsp-token',   'rackspaceSpotRefreshToken');
      txt('rsp-org',     'rackspaceSpotOrganization');
      txt('rsp-region',  'rackspaceSpotRegion');
      txt('rsp-k8s',     'kubernetesVersion');
      txt('rsp-webhook', 'preemptionWebhookUrl');
      txt('rsp-deploy',  'deploymentType');
      txt('rsp-spname',  'spotNodePoolName');
      txt('rsp-spclass', 'spotServerClass');
      txt('rsp-bid',     'spotBidPrice');
      txt('rsp-odname',  'onDemandNodePoolName');
      txt('rsp-odclass', 'onDemandServerClass');
      txt('rsp-odprice', 'onDemandPricePerHour');

      num('rsp-spcount', 'spotNodeCount');
      num('rsp-amin',    'spotAutoscalingMinNodes');
      num('rsp-amax',    'spotAutoscalingMaxNodes');
      num('rsp-odcount', 'onDemandNodeCount');

      var cni = el.querySelector('#rsp-cni');
      if (cni) cni.addEventListener('change', function(e) { config.set('cni', e.target.value); });

      chk('rsp-gpu', 'gpuEnabled');
      chk('rsp-autoscale', 'spotAutoscalingEnabled', function(v) {
        var s = el.querySelector('#rsp-autoscale-section');
        if (s) s.style.display = v ? '' : 'none';
      });
      chk('rsp-ondemand', 'onDemandEnabled', function(v) {
        var s = el.querySelector('#rsp-demand-section');
        if (s) s.style.display = v ? '' : 'none';
      });

      var saveBtn = el.querySelector('#rsp-save');
      var cancelBtn = el.querySelector('#rsp-cancel');
      if (saveBtn) saveBtn.addEventListener('click', function() { self.send('save'); });
      if (cancelBtn) cancelBtn.addEventListener('click', function() { self.send('cancel'); });
    },

    config: Ember.computed('cluster.genericEngineConfig', function() {
      return this.get('cluster.genericEngineConfig');
    }),

    showErrors: function(errors) {
      var el = this.element;
      if (!el) return;
      var box = el.querySelector('#rsp-errors');
      var txt = el.querySelector('#rsp-error-text');
      if (!box || !txt) return;
      if (errors && errors.length) {
        txt.textContent = errors.join(' ');
        box.style.display = '';
      } else {
        box.style.display = 'none';
      }
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
      this.showErrors(errors);
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
