import { RouteRecordRaw } from 'vue-router';

const manageRoutes: RouteRecordRaw[] = [
  {
    path: '/manage/host',
    name: 'host',
    meta: {
      locale: 'menu.manage.host',
      requiresAuth: true,
      icon: 'icon-desktop',
    },
    component: () => import('@/views/manage/host/list.vue'),
  },
  {
    path: '/manage/settings',
    name: 'settings',
    meta: {
      locale: 'menu.manage.settings',
      requiresAuth: true,
      icon: 'icon-settings',
    },
    component: () => import('@/views/manage/settings/index.vue'),
  },
];

export default manageRoutes;
