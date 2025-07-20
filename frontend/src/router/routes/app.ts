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
    path: '/app/store',
    name: 'store',
    component: () => import('@/views/app/store/main.vue'),
    meta: {
      locale: 'menu.app.store',
      requiresAuth: true,
      icon: 'icon-apps',
    },
  },
  {
    path: '/app/docker',
    name: 'docker',
    meta: {
      locale: 'menu.app.docker',
      requiresAuth: true,
      icon: 'icon-common',
    },
    children: [
      {
        path: 'compose',
        name: 'compose',
        component: () => import('@/views/app/docker/compose/list.vue'),
        meta: {
          locale: 'menu.app.docker.compose',
          requiresAuth: true,
        },
      },
      {
        path: 'container/:composeId?',
        name: 'container',
        component: () => import('@/views/app/docker/container/list.vue'),
        meta: {
          locale: 'menu.app.docker.container',
          requiresAuth: true,
        },
      },
      {
        path: 'image',
        name: 'image',
        component: () => import('@/views/app/docker/image/list.vue'),
        meta: {
          locale: 'menu.app.docker.image',
          requiresAuth: true,
        },
      },
      {
        path: 'network',
        name: 'network',
        component: () => import('@/views/app/docker/network/list.vue'),
        meta: {
          locale: 'menu.app.docker.network',
          requiresAuth: true,
        },
      },
      {
        path: 'volume',
        name: 'volume',
        component: () => import('@/views/app/docker/volume/list.vue'),
        meta: {
          locale: 'menu.app.docker.volume',
          requiresAuth: true,
        },
      },
      {
        path: 'setting',
        name: 'setting',
        component: () => import('@/views/app/docker/setting/main.vue'),
        meta: {
          locale: 'menu.app.docker.setting',
          requiresAuth: true,
        },
      },
    ],
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
    path: '/app/logrotate',
    name: 'logrotate',
    component: () => import('@/views/app/logrotate/main.vue'),
    meta: {
      locale: 'menu.app.logrotate',
      requiresAuth: true,
      icon: 'icon-refresh',
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
    path: '/app/nftables',
    name: 'nftables',
    meta: {
      locale: 'menu.app.nftables',
      requiresAuth: true,
      icon: 'icon-thunderbolt',
    },
    redirect: '/app/nftables/config',
    children: [
      {
        path: 'config',
        name: 'nftables-config',
        component: () =>
          import('@/views/app/nftables/pages/process-config/main.vue'),
        meta: {
          locale: 'menu.app.nftables.config',
          requiresAuth: true,
        },
      },
      {
        path: 'ports',
        name: 'nftables-ports',
        component: () => import('@/views/app/nftables/pages/ports/main.vue'),
        meta: {
          locale: 'menu.app.nftables.ports',
          requiresAuth: true,
        },
      },
      {
        path: 'ip-blacklist',
        name: 'nftables-ip-blacklist',
        component: () =>
          import('@/views/app/nftables/pages/ip-blacklist/main.vue'),
        meta: {
          locale: 'menu.app.nftables.ipBlacklist',
          requiresAuth: true,
        },
      },
      {
        path: 'ping',
        name: 'nftables-ping',
        component: () => import('@/views/app/nftables/pages/ping/main.vue'),
        meta: {
          locale: 'menu.app.nftables.ping',
          requiresAuth: true,
        },
      },
      {
        path: 'file',
        name: 'nftables-file',
        component: () => import('@/views/app/nftables/pages/file/main.vue'),
        meta: {
          locale: 'menu.app.nftables.file',
          requiresAuth: true,
        },
      },
    ],
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
  {
    path: '/app/certificate',
    name: 'certificate',
    component: () => import('@/views/app/certificate/main.vue'),
    meta: {
      locale: 'menu.app.certificate',
      requiresAuth: true,
      icon: 'icon-safe',
    },
  },
  {
    path: '/app/service',
    name: 'service',
    component: () => import('@/views/app/service/main.vue'),
    meta: {
      locale: 'menu.app.service',
      requiresAuth: true,
      icon: 'icon-settings',
    },
  },
];

export default appRoutes;
