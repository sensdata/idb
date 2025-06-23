import {
  IconCode,
  IconStorage,
  IconDesktop,
  IconWifi,
  IconLock,
  IconFile,
  IconCloud,
  IconBug,
  IconTool,
  IconSafe,
  IconUser,
  IconSettings,
  IconDashboard,
  IconComputer,
  IconThunderbolt,
  IconMobile,
  IconCamera,
  IconPrinter,
  IconCommand,
  IconMenu,
} from '@arco-design/web-vue/es/icon';
import type { Component } from 'vue';
import { useI18n } from 'vue-i18n';

// 导入具体的软件图标
import NginxIcon from '@/assets/icons/nginx.svg';
import PhpIcon from '@/assets/icons/php.svg';
import ServiceIcon from '@/assets/icons/service.svg';
import StorageIcon from '@/assets/icons/storage.svg';
import ToolIcon from '@/assets/icons/tool.svg';
import SafeIcon from '@/assets/icons/safe.svg';
import DashboardIcon from '@/assets/icons/dashboard.svg';
import SettingsIcon from '@/assets/icons/settings.svg';
import UserIcon from '@/assets/icons/user.svg';
// 导入新增的业务图标
import DockerIcon from '@/assets/icons/docker.svg';
import MySQLIcon from '@/assets/icons/mysql.svg';
import PostgreSQLIcon from '@/assets/icons/postgresql.svg';
import RedisIcon from '@/assets/icons/redis.svg';
import MongoDBIcon from '@/assets/icons/mongodb.svg';
import ApacheIcon from '@/assets/icons/apache.svg';
import KubernetesIcon from '@/assets/icons/kubernetes.svg';
import NodejsIcon from '@/assets/icons/nodejs.svg';
import JavaIcon from '@/assets/icons/java.svg';

// 定义图标映射类型
type IconRecord = Record<string, any>;

// 具体软件图标映射（优先级最高）
const specificIconMap: IconRecord = {
  'nginx': NginxIcon,
  'php': PhpIcon,
  'php-fpm': PhpIcon,
  'service': ServiceIcon,
  'systemd': ServiceIcon,
  'systemctl': ServiceIcon,
  'storage': StorageIcon,
  'tool': ToolIcon,
  'safe': SafeIcon,
  'dashboard': DashboardIcon,
  'settings': SettingsIcon,
  'user': UserIcon,
  // 新增业务图标映射
  'docker': DockerIcon,
  'dockerd': DockerIcon,
  'docker-proxy': DockerIcon,
  'docker-containerd': DockerIcon,
  'containerd': DockerIcon,
  'containerd-shim': DockerIcon,
  'mysql': MySQLIcon,
  'mysqld': MySQLIcon,
  'postgresql': PostgreSQLIcon,
  'postgres': PostgreSQLIcon,
  'redis': RedisIcon,
  'mongodb': MongoDBIcon,
  'apache': ApacheIcon,
  'httpd': ApacheIcon,
  'kubernetes': KubernetesIcon,
  'kubectl': KubernetesIcon,
  'k8s': KubernetesIcon,
  'node': NodejsIcon,
  'nodejs': NodejsIcon,
  'npm': NodejsIcon,
  'java': JavaIcon,
  'jar': JavaIcon,
  'tomcat': JavaIcon,
};

// 通用图标映射（回退方案）
const genericIconMap: Record<string, Component> = {
  // 数据库相关
  mysql: IconStorage,
  mysqld: IconStorage,
  postgresql: IconStorage,
  postgres: IconStorage,
  redis: IconStorage,
  mongodb: IconStorage,
  mariadb: IconStorage,

  // Web服务器
  apache: IconMenu,
  httpd: IconMenu,
  tomcat: IconMenu,
  node: IconCode,
  nodejs: IconCode,

  // 容器化
  docker: IconComputer,
  dockerd: IconComputer,
  containerd: IconComputer,
  podman: IconComputer,
  kubernetes: IconCloud,
  kubectl: IconCloud,

  // 系统服务
  systemd: IconSettings,
  systemctl: IconSettings,
  init: IconSettings,
  kernel: IconThunderbolt,
  kthreadd: IconThunderbolt,

  // 网络服务
  ssh: IconLock,
  sshd: IconLock,
  ftp: IconFile,
  ftpd: IconFile,
  telnet: IconWifi,
  dhcp: IconWifi,
  dhcpd: IconWifi,
  dns: IconWifi,
  named: IconWifi,
  bind: IconWifi,

  // 安全相关
  firewall: IconSafe,
  iptables: IconSafe,
  fail2ban: IconSafe,
  selinux: IconSafe,

  // 监控工具
  top: IconDashboard,
  htop: IconDashboard,
  ps: IconDashboard,
  netstat: IconWifi,
  tcpdump: IconBug,
  wireshark: IconBug,

  // 文件系统
  mount: IconFile,
  umount: IconFile,
  nfs: IconFile,
  samba: IconFile,
  smb: IconFile,

  // 用户管理
  login: IconUser,
  su: IconUser,
  sudo: IconUser,
  passwd: IconUser,

  // 开发工具
  git: IconTool,
  vim: IconTool,
  nano: IconTool,
  emacs: IconTool,

  // 桌面环境
  gnome: IconDesktop,
  kde: IconDesktop,
  xfce: IconDesktop,
  x11: IconDesktop,
  wayland: IconDesktop,

  // 多媒体
  pulse: IconCamera,
  pulseaudio: IconCamera,
  alsa: IconCamera,

  // 打印服务
  cups: IconPrinter,
  cupsd: IconPrinter,

  // 移动设备
  android: IconMobile,
  adb: IconMobile,

  // Shell
  bash: IconCommand,
  sh: IconCommand,
  zsh: IconCommand,
  fish: IconCommand,
  csh: IconCommand,
  tcsh: IconCommand,
};

// 进程类型描述的翻译键映射
const specificDescriptionKeys: Record<string, string> = {
  'nginx': 'app.nftables.process.nginx',
  'mysql': 'app.nftables.process.mysql',
  'mysqld': 'app.nftables.process.mysqld',
  'postgresql': 'app.nftables.process.postgresql',
  'postgres': 'app.nftables.process.postgres',
  'redis': 'app.nftables.process.redis',
  'mongodb': 'app.nftables.process.mongodb',
  'apache': 'app.nftables.process.apache',
  'httpd': 'app.nftables.process.httpd',
  'docker': 'app.nftables.process.docker',
  'dockerd': 'app.nftables.process.dockerd',
  'docker-proxy': 'app.nftables.process.dockerProxy',
  'docker-containerd': 'app.nftables.process.dockerContainerd',
  'containerd': 'app.nftables.process.containerd',
  'containerd-shim': 'app.nftables.process.containerdShim',
  'kubernetes': 'app.nftables.process.kubernetes',
  'kubectl': 'app.nftables.process.kubectl',
  'ssh': 'app.nftables.process.ssh',
  'sshd': 'app.nftables.process.sshd',
  'systemd': 'app.nftables.process.systemd',
  'nodejs': 'app.nftables.process.nodejs',
  'node': 'app.nftables.process.node',
  'java': 'app.nftables.process.java',
  'tomcat': 'app.nftables.process.tomcat',
  'php': 'app.nftables.process.php',
  'php-fpm': 'app.nftables.process.phpFpm',
  'podman': 'app.nftables.process.podman',
  'mariadb': 'app.nftables.process.mariadb',
  'bind': 'app.nftables.process.bind',
  'named': 'app.nftables.process.named',
  'fail2ban': 'app.nftables.process.fail2ban',
  'iptables': 'app.nftables.process.iptables',
  'firewall': 'app.nftables.process.firewall',
  'selinux': 'app.nftables.process.selinux',
};

// 图标缓存，避免重复计算
const iconCache = new Map<string, any>();
const iconClassCache = new Map<string, string>();
const descriptionCache = new Map<string, string>();

/**
 * 提供进程图标相关功能的组合式函数
 */
export function useProcessIcons() {
  const { t } = useI18n();
  /**
   * 判断是否为SVG图标组件
   */
  const isSvgIcon = (icon: any): boolean => {
    return Object.values(specificIconMap).includes(icon);
  };

  /**
   * 根据进程名称获取对应图标
   */
  const getProcessIcon = (processName: string): any => {
    if (!processName) return IconCode;

    const name = processName.toLowerCase();

    // 检查缓存
    if (iconCache.has(name)) {
      return iconCache.get(name);
    }

    let icon;

    // 1. 优先检查具体软件图标
    if (specificIconMap[name]) {
      icon = specificIconMap[name];
    } else {
      // 2. 模糊匹配具体软件图标
      for (const [key, value] of Object.entries(specificIconMap)) {
        if (name.includes(key) || key.includes(name)) {
          icon = value;
          break;
        }
      }

      if (!icon) {
        // 3. 根据进程名特征优先匹配特定SVG图标
        if (name.includes('docker')) {
          icon = DockerIcon;
        } else if (name.includes('php')) {
          icon = PhpIcon;
        } else if (
          name.includes('service') ||
          name.includes('daemon') ||
          name.endsWith('d')
        ) {
          icon = ServiceIcon;
        } else if (
          name.includes('db') ||
          name.includes('data') ||
          name.includes('storage')
        ) {
          icon = StorageIcon;
        } else if (name.includes('log') || name.includes('monitor')) {
          icon = DashboardIcon;
        } else if (
          name.includes('security') ||
          name.includes('auth') ||
          name.includes('safe')
        ) {
          icon = SafeIcon;
        } else if (
          name.includes('tool') ||
          name.includes('git') ||
          name.includes('vim')
        ) {
          icon = ToolIcon;
        } else if (name.includes('user') || name.includes('admin')) {
          icon = UserIcon;
        } else if (name.includes('setting') || name.includes('config')) {
          icon = SettingsIcon;
        } else if (genericIconMap[name]) {
          // 4. 使用通用图标
          icon = genericIconMap[name];
        } else {
          // 5. 模糊匹配通用图标
          for (const [key, value] of Object.entries(genericIconMap)) {
            if (name.includes(key) || key.includes(name)) {
              icon = value;
              break;
            }
          }

          if (!icon) {
            // 6. 根据进程名特征进行模式匹配
            if (name.includes('java') || name.includes('jar')) {
              icon = IconMenu; // 代表服务器应用
            } else if (name.includes('python') || name.includes('py')) {
              icon = IconCode;
            } else if (name.includes('ruby') || name.includes('rb')) {
              icon = IconCode;
            } else if (name.includes('go') || name.includes('golang')) {
              icon = IconCode;
            } else if (name.includes('rust') || name.includes('cargo')) {
              icon = IconCode;
            } else if (name.includes('server') || name.includes('srv')) {
              icon = IconMenu;
            } else if (name.includes('network') || name.includes('net')) {
              icon = IconWifi;
            } else if (name.includes('file') || name.includes('fs')) {
              icon = IconFile;
            } else if (name.includes('cloud') || name.includes('cluster')) {
              icon = IconCloud;
            } else {
              // 默认图标
              icon = IconCode;
            }
          }
        }
      }
    }

    // 存入缓存
    iconCache.set(name, icon);
    return icon;
  };

  /**
   * 根据进程类型获取图标样式类
   */
  const getProcessIconClass = (processName: string): string => {
    if (!processName) return 'process-icon-default';

    const name = processName.toLowerCase();

    // 检查缓存
    if (iconClassCache.has(name)) {
      return iconClassCache.get(name)!;
    }

    let iconClass = 'process-icon-default';

    // 数据库类
    if (
      [
        'mysql',
        'mysqld',
        'postgresql',
        'postgres',
        'redis',
        'mongodb',
        'mariadb',
      ].some((db) => name.includes(db)) ||
      name.includes('db') ||
      name.includes('data')
    ) {
      iconClass = 'process-icon-database';
    }
    // Web服务器类
    else if (
      ['nginx', 'apache', 'httpd', 'tomcat'].some((server) =>
        name.includes(server)
      ) ||
      name.includes('java') ||
      name.includes('jar') ||
      name.includes('server')
    ) {
      iconClass = 'process-icon-server';
    }
    // 容器化类
    else if (
      ['docker', 'kubernetes', 'kubectl', 'containerd', 'podman'].some(
        (container) => name.includes(container)
      ) ||
      name.includes('cloud') ||
      name.includes('cluster')
    ) {
      iconClass = 'process-icon-container';
    }
    // 网络类
    else if (
      [
        'ssh',
        'sshd',
        'ftp',
        'telnet',
        'dhcp',
        'dns',
        'named',
        'bind',
        'netstat',
      ].some((net) => name.includes(net)) ||
      name.includes('network') ||
      name.includes('net')
    ) {
      iconClass = 'process-icon-network';
    }
    // 安全类
    else if (
      ['firewall', 'iptables', 'fail2ban', 'selinux'].some((sec) =>
        name.includes(sec)
      ) ||
      name.includes('security') ||
      name.includes('auth')
    ) {
      iconClass = 'process-icon-security';
    }
    // 开发工具类
    else if (
      ['git', 'vim', 'nano', 'emacs', 'python', 'node', 'nodejs'].some((dev) =>
        name.includes(dev)
      ) ||
      name.includes('code')
    ) {
      iconClass = 'process-icon-development';
    }
    // 系统服务类
    else if (
      ['systemd', 'systemctl', 'init', 'kernel'].some((sys) =>
        name.includes(sys)
      ) ||
      name.includes('service') ||
      name.includes('daemon') ||
      name.endsWith('d')
    ) {
      iconClass = 'process-icon-system';
    }

    // 存入缓存
    iconClassCache.set(name, iconClass);
    return iconClass;
  };

  /**
   * 获取进程类型描述
   */
  const getProcessTypeDescription = (processName: string): string => {
    if (!processName) return t('app.nftables.process.unknown');

    const name = processName.toLowerCase();

    // 检查缓存
    if (descriptionCache.has(name)) {
      return descriptionCache.get(name)!;
    }

    let description = '';

    // 直接匹配
    if (specificDescriptionKeys[name]) {
      description = t(specificDescriptionKeys[name]);
    } else {
      // 模糊匹配
      let matched = false;
      for (const [key, translationKey] of Object.entries(
        specificDescriptionKeys
      )) {
        if (name.includes(key)) {
          description = t(translationKey);
          matched = true;
          break;
        }
      }

      // 分类描述
      if (!matched) {
        const iconClass = getProcessIconClass(processName);
        switch (iconClass) {
          case 'process-icon-database':
            description = t('app.nftables.process.category.database');
            break;
          case 'process-icon-server':
            description = t('app.nftables.process.category.webServer');
            break;
          case 'process-icon-container':
            description = t('app.nftables.process.category.container');
            break;
          case 'process-icon-network':
            description = t('app.nftables.process.category.network');
            break;
          case 'process-icon-security':
            description = t('app.nftables.process.category.security');
            break;
          case 'process-icon-development':
            description = t('app.nftables.process.category.development');
            break;
          case 'process-icon-system':
            description = t('app.nftables.process.category.system');
            break;
          default:
            description = t('app.nftables.process.category.application');
        }
      }
    }

    // 存入缓存
    descriptionCache.set(name, description);
    return description;
  };

  return {
    getProcessIcon,
    isSvgIcon,
    getProcessIconClass,
    getProcessTypeDescription,
  };
}
