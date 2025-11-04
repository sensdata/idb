import localeMessageBox from '@/components/message-box/locale/zh-CN';
import { LocaleModules } from './types';

const locales: LocaleModules = import.meta.glob('./zh-CN/*.ts', {
  eager: true,
});

const viewLocales: LocaleModules = import.meta.glob('../views/**/zh-CN.ts', {
  eager: true,
});
const componentLocales: LocaleModules = import.meta.glob(
  '../components/**/zh-CN.ts',
  {
    eager: true,
  }
);
const routerLocales: LocaleModules = import.meta.glob('../router/**/zh-CN.ts', {
  eager: true,
});

export default {
  'menu.app.certificate': '证书管理',
  'menu.app.rsync': '文件同步',
  'locale.switchLocale': '切换为中文',
  ...localeMessageBox,
  ...Object.values({
    ...componentLocales,
    ...viewLocales,
    ...routerLocales,
    ...locales,
  }).reduce((result, locale) => {
    return {
      ...result,
      ...locale.default,
    };
  }, {}),
};
