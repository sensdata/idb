import { RouteRecordRaw } from 'vue-router';
import DeskTopIcon from '@/assets/icons/desktop.svg?raw';
import { DEFAULT_LAYOUT } from './base';

const manageRoutes: RouteRecordRaw[] = [
  {
    path: '/manage/test',
    name: 'manageTest',
    component: () => import('@/views/mange/test/test.vue'),
    meta: {
      locale: 'menu.manage.test',
      requiresAuth: true,
      card: true,
      icon: DeskTopIcon,
    },
  },
  {
    path: '/manage/host',
    name: 'manageHost',
    component: DEFAULT_LAYOUT,
    meta: {
      locale: 'menu.manage.host',
      requiresAuth: true,
      icon: DeskTopIcon,
    },
    redirect: '/manage/host/list',
    children: [
      {
        path: 'list',
        name: 'manageHostList',
        component: () => import('@/views/mange/host/list.vue'),
        meta: {
          locale: 'menu.manage.host.list',
          requiresAuth: true,
          card: true,
        },
      },
    ],
  },
];

export default manageRoutes;
