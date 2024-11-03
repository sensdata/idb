import type { Router } from 'vue-router';
import NProgress from 'nprogress'; // progress bar
import usetCurrentHost from '@/hooks/current-host';
import { Message } from '@arco-design/web-vue';
import i18n from '@/locale';
import { SELECT_HOST } from '../constants';

export default function setupHostIdGuard(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const { currentHostId, switchHost } = usetCurrentHost();
    const {
      global: { t },
    } = i18n;
    if (to.path.startsWith('/app') || to.meta?.requiresHostId) {
      if (!currentHostId.value && !to.query.id) {
        Message.error(t('router.guard.hostId.message'));
        next(SELECT_HOST);
      } else {
        if (to.query.id && +to.query.id !== currentHostId.value) {
          switchHost(+to.query.id);
        }
        if (!to.query.id && currentHostId.value) {
          next({
            ...to,
            query: {
              ...to.query,
              id: currentHostId.value,
            },
          });
        } else {
          next();
        }
      }
    } else {
      next();
    }
    NProgress.done();
  });
}
