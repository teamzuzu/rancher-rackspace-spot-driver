import { h } from 'vue';
import CruRackspaceSpot from './components/CruRackspaceSpot.vue';
import { createClusterProduct } from './cluster-editor.mjs';

export const { init } = createClusterProduct(CruRackspaceSpot, h);
