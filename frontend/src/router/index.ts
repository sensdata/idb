import { createRouter, createWebHistory } from 'vue-router';
import NProgress from 'nprogress'; // progress bar
import 'nprogress/nprogress.css';

import { manageRoutes, appRoutes } from './routes';
import { REDIRECT_MAIN, NOT_FOUND_ROUTE, DEFAULT_LAYOUT } from './routes/base';
import createRouteGuard from './guard';
import { DEFAULT_ROUTE_NAME } from './constants';

NProgress.configure({ showSpinner: false }); // NProgress Configuration

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: {
        name: DEFAULT_ROUTE_NAME,
      },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/common/login/index.vue'),
      meta: {
        requiresAuth: false,
      },
    },
    REDIRECT_MAIN,
    {
      path: '/:pathMatch(.*)*',
      name: 'all',
      component: DEFAULT_LAYOUT,
      children: [...manageRoutes, ...appRoutes, NOT_FOUND_ROUTE],
    },
  ],
  scrollBehavior() {
    return { top: 0 };
  },
});

createRouteGuard(router);

export default router;
