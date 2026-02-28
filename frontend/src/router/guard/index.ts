import type { Router } from 'vue-router';
import NProgress from 'nprogress';
import { setRouteEmitter } from '@/utils/route-listener';
import setupUserLoginInfoGuard from './userLoginInfo';
import setupPermissionGuard from './permission';
import setupHostIdGuard from './hostId';

function setupPageGuard(router: Router) {
  router.beforeEach(async (to) => {
    NProgress.start();
    // emit route change
    setRouteEmitter(to);
  });
  router.afterEach(() => {
    NProgress.done();
  });
  router.onError(() => {
    NProgress.done();
  });
}

export default function createRouteGuard(router: Router) {
  setupPageGuard(router);
  setupUserLoginInfoGuard(router);
  setupPermissionGuard(router);
  setupHostIdGuard(router);
}
