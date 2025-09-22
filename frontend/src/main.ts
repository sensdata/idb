import { createApp, markRaw } from 'vue';
import ArcoVue, { Message } from '@arco-design/web-vue';
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
// Force all error messages to be manual-close with an X (type-safe override)
const originalMessageError = Message.error.bind(Message);
Message.error = (arg) => {
  if (typeof arg === 'string') {
    return originalMessageError({ content: arg, closable: true, duration: 0 });
  }
  // Enforce manual-close for all error messages globally
  return originalMessageError({ ...arg, closable: true, duration: 0 });
};
// Make success messages stay longer (3s) by default
const originalMessageSuccess = Message.success.bind(Message);
Message.success = (arg) => {
  if (typeof arg === 'string') {
    return originalMessageSuccess({ content: arg, duration: 3000 });
  }
  const hasDuration = Object.prototype.hasOwnProperty.call(arg, 'duration');
  return originalMessageSuccess(hasDuration ? arg : { ...arg, duration: 3000 });
};
// Make warning messages stay 3s by default
const originalMessageWarning = Message.warning.bind(Message);
Message.warning = (arg) => {
  if (typeof arg === 'string') {
    return originalMessageWarning({ content: arg, duration: 3000 });
  }
  const hasDuration = Object.prototype.hasOwnProperty.call(arg, 'duration');
  return originalMessageWarning(hasDuration ? arg : { ...arg, duration: 3000 });
};
// Make info messages stay 3s by default
const originalMessageInfo = Message.info.bind(Message);
Message.info = (arg) => {
  if (typeof arg === 'string') {
    return originalMessageInfo({ content: arg, duration: 3000 });
  }
  const hasDuration = Object.prototype.hasOwnProperty.call(arg, 'duration');
  return originalMessageInfo(hasDuration ? arg : { ...arg, duration: 3000 });
};

app.use(i18n);
app.use(globalComponents);
app.use(directive);

app.mount('#app');
