import { App } from 'vue';
import Breadcrumb from './breadcrumb/index.vue';
import IdbTable from './idb-table/index.vue';
import IdbTableOperation from './idb-table-operation/index.vue';
import FixedFooterBar from './fixed-footer-bar/index.vue';
import DockerInstallGuide from './docker-install-guide/index.vue';

export default {
  install(Vue: App) {
    Vue.component('IdbTable', IdbTable);
    Vue.component('IdbTableOperation', IdbTableOperation);
    Vue.component('FixedFooterBar', FixedFooterBar);
    Vue.component('Breadcrumb', Breadcrumb);
    Vue.component('DockerInstallGuide', DockerInstallGuide);
  },
};
