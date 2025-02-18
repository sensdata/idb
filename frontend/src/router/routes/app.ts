import { RouteRecordRaw } from 'vue-router';
import HomeIcon from '@/assets/icons/home.svg?raw';
import FileIcon from '@/assets/icons/folder.svg?raw';

const appRoutes: RouteRecordRaw[] = [
  {
    path: '/app/sysinfo',
    name: 'sysinfo',
    meta: {
      locale: 'menu.app.sysinfo',
      requiresAuth: true,
      icon: HomeIcon,
    },
    redirect: '/app/sysinfo/overview',
    children: [
      {
        path: 'overview',
        name: 'sysinfoOverview',
        component: () => import('@/views/app/sysinfo/overview.vue'),
        meta: {
          locale: 'menu.app.sysinfo.overview',
          requiresAuth: true,
        },
      },
    ],
  },
  {
    path: '/app/crontab',
    name: 'crontab',
    component: () => import('@/views/app/crontab/main.vue'),
    meta: {
      locale: 'menu.app.crontab',
      requiresAuth: true,
      icon: 'icon-clock-circle',
    },
  },
  {
    path: '/app/file',
    name: 'file',
    component: () => import('@/views/app/file/main.vue'),
    meta: {
      locale: 'menu.app.file',
      requiresAuth: true,
      icon: FileIcon,
    },
  },
  {
    path: '/app/terminal',
    name: 'terminal',
    component: () => import('@/views/app/terminal/main.vue'),
    meta: {
      locale: 'menu.app.terminal',
      requiresAuth: true,
      icon: 'icon-code-square',
      command: 'openTerminal',
    },
  },
  {
    path: '/app/script',
    name: 'script',
    component: () => import('@/views/app/script/main.vue'),
    meta: {
      locale: 'menu.app.script',
      requiresAuth: true,
      icon: 'icon-code',
    },
  },
  {
    path: '/app/process',
    name: 'process',
    component: () => import('@/views/app/process/main.vue'),
    meta: {
      locale: 'menu.app.process',
      requiresAuth: true,
      icon: 'icon-branch',
    },
  },
];

export default appRoutes;
