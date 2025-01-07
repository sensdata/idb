import { createApp, markRaw } from 'vue';
import ArcoVue from '@arco-design/web-vue';
import ArcoVueIcon from '@arco-design/web-vue/es/icon';
import globalComponents from '@/components';
import router from './router';
import pinia from './store';
import i18n from './locale';
import directive from './directive';
import App from './App.vue';
// Styles are imported via arco-plugin. See config/plugin/arcoStyleImport.ts in the directory for details
// 样式通过 arco-plugin 插件导入。详见目录文件 config/plugin/arcoStyleImport.ts
// https://arco.design/docs/designlab/use-theme-package
import '@/assets/style/util.less';
import '@/assets/style/twilwind.css';
import '@/assets/style/global.less';
import '@/assets/style/overlay.less';

const app = createApp(App);

app.use(ArcoVue, {});
app.use(ArcoVueIcon);

pinia.use(({ store }) => {
  store.router = markRaw(router);
});
app.use(pinia);
app.use(router);
app.use(i18n);
app.use(globalComponents);
app.use(directive);

app.mount('#app');
