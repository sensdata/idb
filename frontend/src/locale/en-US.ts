import localeMessageBox from '@/components/message-box/locale/en-US';
import localeLogin from '@/views/common/login/locale/en-US';
import { LocaleModules } from './types';

const locales: LocaleModules = import.meta.glob('./en-US/*.ts', {
  eager: true,
});
const viewLocales: LocaleModules = import.meta.glob('../views/**/en-US.ts', {
  eager: true,
});
const componentLocales: LocaleModules = import.meta.glob(
  '../components/**/en-US.ts',
  {
    eager: true,
  }
);
const routerLocales: LocaleModules = import.meta.glob('../router/**/en-US.ts', {
  eager: true,
});

export default {
  'menu.dashboard': 'Dashboard',
  'menu.profile': 'Profile',
  'locale.switchLocale': 'Switch to English',
  ...localeMessageBox,
  ...localeLogin,
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
