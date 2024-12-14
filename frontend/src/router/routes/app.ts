import { RouteRecordRaw } from 'vue-router';
import DeskTopIcon from '@/assets/icons/desktop.svg?raw';
import HomeIcon from '@/assets/icons/home.svg?raw';
import FileIcon from '@/assets/icons/folder.svg?raw';

const appRoutes: RouteRecordRaw[] = [
  {
    path: '/app/sysinfo',
    name: 'appSysinfo',
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
  {
    path: '/app/file',
    name: 'appFile',
    component: () => import('@/views/app/file/main.vue'),
    meta: {
      locale: 'menu.app.file',
      requiresAuth: true,
      icon: FileIcon,
    },
  },
  {
    path: '/app/terminal',
    name: 'appTerminal',
    component: () => import('@/views/app/terminal/main.vue'),
    meta: {
      locale: 'menu.app.terminal',
      requiresAuth: true,
      icon: 'icon-code-square',
    },
  },
  {
    path: '/app/script',
    name: 'appScript',
    component: () => import('@/views/app/script/main.vue'),
    meta: {
      locale: 'menu.app.script',
      requiresAuth: true,
      icon: '<icon-code />',
    },
  },
  {
    path: '/app/test',
    name: 'appTest',
    meta: {
      locale: 'menu.app.test',
      requiresAuth: true,
      card: true,
      icon: DeskTopIcon,
    },
    component: () => import('@/views/app/test/test.vue'),
  },
];

export default appRoutes;
