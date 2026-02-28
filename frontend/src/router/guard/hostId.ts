import type { Router } from 'vue-router';
import { Message } from '@arco-design/web-vue';
import i18n from '@/locale';
import { isLogin } from '@/helper/auth';
import { useHostStore } from '@/store';
import { SELECT_HOST } from '../constants';

function parseHostId(value: unknown): number | undefined {
  const raw = Array.isArray(value) ? value[0] : value;
  if (typeof raw === 'number' && Number.isInteger(raw) && raw > 0) {
    return raw;
  }
  if (typeof raw === 'string' && /^\d+$/.test(raw)) {
    const num = Number(raw);
    if (num > 0) return num;
  }
  return undefined;
}

export default function setupHostIdGuard(router: Router) {
  router.beforeEach(async (to, _from, next) => {
    const hostStore = useHostStore();
    const {
      global: { t },
    } = i18n;

    if (isLogin() && !hostStore.isReady) {
      try {
        await hostStore.init();
      } catch (error) {
        Message.error(t('router.guard.hostId.message'));
        next(SELECT_HOST);
        return;
      }
    }

    const requiresHostId =
      to.path.startsWith('/app') || to.meta?.requiresHostId;
    if (!requiresHostId) {
      next();
      return;
    }

    const queryHostId = parseHostId(to.query.id);
    const resolvedHostId =
      queryHostId || hostStore.currentId || hostStore.defaultId;

    if (!resolvedHostId) {
      Message.error(t('router.guard.hostId.message'));
      next(SELECT_HOST);
      return;
    }

    if (hostStore.currentId !== resolvedHostId) {
      hostStore.setCurrentId(resolvedHostId);
    }

    if (queryHostId !== resolvedHostId) {
      next({
        ...to,
        query: {
          ...to.query,
          id: resolvedHostId,
        },
        replace: true,
      });
      return;
    }

    next();
  });
}
