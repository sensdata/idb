import { RouteRecordRaw } from 'vue-router';
import DeskTopIcon from '@/assets/icons/desktop.svg?raw';

const manageRoutes: RouteRecordRaw[] = [
  {
    path: '/manage/host',
    name: 'manageHost',
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
