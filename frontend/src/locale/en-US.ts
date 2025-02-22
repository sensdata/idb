import localeMessageBox from '@/components/message-box/locale/en-US';
import localeLogin from '@/views/common/login/locale/en-US';
import { LocaleModules } from './types';

const locales: LocaleModules = import.meta.glob('./en-US/*.ts', {
  eager: true,
});

export default {
  'menu.dashboard': 'Dashboard',
  'menu.profile': 'Profile',
  'locale.switchLocale': 'Switch to English',
  ...localeMessageBox,
  ...localeLogin,
  ...Object.values(locales).reduce((result, locale) => {
    return {
      ...result,
      ...locale.default,
    };
  }, {}),
};
