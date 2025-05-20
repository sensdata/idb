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
      {
        path: 'network',
        name: 'sysinfoNetwork',
        component: () => import('@/views/app/sysinfo/network.vue'),
        meta: {
          locale: 'menu.app.sysinfo.network',
          requiresAuth: true,
        },
      },
      {
        path: 'system',
        name: 'sysinfoSystem',
        component: () => import('@/views/app/sysinfo/system.vue'),
        meta: {
          locale: 'menu.app.sysinfo.system',
          requiresAuth: true,
        },
      },
      {
        path: 'config',
        name: 'sysinfoConfig',
        component: () => import('@/views/app/sysinfo/config.vue'),
        meta: {
          locale: 'menu.app.sysinfo.config',
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
  {
    path: '/app/ssh',
    name: 'ssh',
    meta: {
      locale: 'menu.app.ssh',
      requiresAuth: true,
      icon: 'icon-command',
    },
    redirect: '/app/ssh/config',
    children: [
      {
        path: 'config',
        name: 'sshConfig',
        component: () => import('@/views/app/ssh/pages/config/main.vue'),
        meta: {
          locale: 'menu.app.ssh.config',
          requiresAuth: true,
        },
      },
      {
        path: 'key-pairs',
        name: 'sshKeyPairs',
        component: () => import('@/views/app/ssh/pages/key-pairs/main.vue'),
        meta: {
          locale: 'menu.app.ssh.keyPairs',
          requiresAuth: true,
        },
      },
      {
        path: 'public-keys',
        name: 'sshPublicKeys',
        component: () => import('@/views/app/ssh/pages/public-keys/main.vue'),
        meta: {
          locale: 'menu.app.ssh.publicKeys',
          requiresAuth: true,
        },
      },
    ],
  },
];

export default appRoutes;
