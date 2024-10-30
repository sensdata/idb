import { RouteRecordRaw } from 'vue-router';
import DeskTopIcon from '@/assets/icons/desktop.svg?raw';
import { DEFAULT_LAYOUT } from './base';

const manageRoutes: RouteRecordRaw[] = [
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
