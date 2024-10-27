import DeskTopIcon from '@/assets/icons/desktop.svg?raw';
import { DEFAULT_LAYOUT } from '../base';
import { AppRouteRecordRaw } from '../types';

const ManageHost: AppRouteRecordRaw = {
  path: '/manage/host',
  name: 'manageHost',
  component: DEFAULT_LAYOUT,
  meta: {
    locale: 'menu.manage.host',
    requiresAuth: true,
    icon: DeskTopIcon,
    order: 0,
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
};

export default ManageHost;
