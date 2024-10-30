import { RouteRecordRaw } from 'vue-router';
import HomeIcon from '@/assets/icons/home.svg?raw';
import { DEFAULT_LAYOUT } from './base';

const appRoutes: RouteRecordRaw[] = [
  {
    path: '/app/sysinfo',
    name: 'appSysinfo',
    component: DEFAULT_LAYOUT,
    meta: {
      locale: 'menu.app.sysinfo',
      requiresAuth: true,
      icon: HomeIcon,
    },
    redirect: '/app/sysinfo/overview',
    children: [
      {
        path: 'overview',
        name: 'appSysinfoOverview',
        component: () => import('@/views/app/sysinfo/overview.vue'),
        meta: {
          locale: 'menu.app.sysinfo.overview',
          requiresAuth: true,
          card: true,
        },
      },
    ],
  },
];

export default appRoutes;
