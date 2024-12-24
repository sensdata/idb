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
      card: true,
    },
    component: () => import('@/views/mange/host/list.vue'),
  },
];

export default manageRoutes;
