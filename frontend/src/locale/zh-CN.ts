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
  'menu.dashboard': '仪表盘',
  'menu.profile': '个人中心',
  'menu.faq': '常见问题',
  'menu.manage.host': '节点管理',
  'menu.manage.host.list': '节点列表',
  'menu.app.sysinfo': '系统信息',
  'menu.app.sysinfo.overview': '状态概览',
  'navbar.docs': '文档中心',
  'navbar.action.locale': '切换为中文',
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
