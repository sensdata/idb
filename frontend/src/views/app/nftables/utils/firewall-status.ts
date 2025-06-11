/**
 * 防火墙状态管理工具函数
 */

import type { NftablesStatus } from '@/api/nftables';

export interface FirewallStatusInfo {
  installStatusText: string;
  activeSystemColor: string;
  activeSystemText: string;
  canSwitch: boolean;
  switchButtonText: string;
  isNftablesActive: boolean;
  isIptablesActive: boolean;
}

/**
 * 获取安装状态文本
 * @param status 安装状态
 * @param t 国际化函数
 * @returns 状态文本
 */
export function getInstallStatusText(
  status: string | undefined,
  t: (key: string) => string
): string {
  if (!status) return t('app.nftables.status.unknown');
  return status === 'installed'
    ? t('app.nftables.status.installed')
    : t('app.nftables.status.notInstalled');
}

/**
 * 获取激活系统颜色
 * @param active 激活状态
 * @returns 颜色值
 */
export function getActiveSystemColor(active: string | undefined): string {
  if (!active) return 'gray';

  if (active.includes('nftables')) return 'green';
  if (active.includes('iptables')) return 'blue';
  if (active.includes('no firewall')) return 'red';
  if (active.includes('uncertain')) return 'orange';

  return 'gray';
}

/**
 * 获取激活系统文本
 * @param active 激活状态
 * @param t 国际化函数
 * @returns 状态文本
 */
export function getActiveSystemText(
  active: string | undefined,
  t: (key: string) => string
): string {
  if (!active) return t('app.nftables.status.unknown');

  if (active.includes('nftables is active'))
    return t('app.nftables.status.nftablesActive');
  if (active.includes('iptables (legacy) is active'))
    return t('app.nftables.status.iptablesLegacyActive');
  if (active.includes('iptables-nft'))
    return t('app.nftables.status.iptablesNftActive');
  if (active.includes('no firewall'))
    return t('app.nftables.status.noFirewall');
  if (active.includes('uncertain')) return t('app.nftables.status.uncertain');

  return active;
}

/**
 * 检查是否为nftables激活状态
 * @param active 激活状态
 * @returns 是否激活
 */
export function isNftablesActive(active: string | undefined): boolean {
  return active?.includes('nftables is active') ?? false;
}

/**
 * 检查是否为iptables激活状态
 * @param active 激活状态
 * @returns 是否激活
 */
export function isIptablesActive(active: string | undefined): boolean {
  return (
    (active?.includes('iptables') && !active?.includes('nftables')) ?? false
  );
}

/**
 * 判断是否应该显示切换按钮
 * @param firewallStatus 防火墙状态
 * @returns 是否显示切换按钮
 */
export function shouldShowSwitchButton(
  firewallStatus: NftablesStatus | null
): boolean {
  if (!firewallStatus) return false;

  const { status, active } = firewallStatus;

  // 当nftables已安装且激活时，可以切换到iptables
  if (status === 'installed' && isNftablesActive(active)) {
    return true;
  }

  // 当iptables激活时，可以切换到nftables
  if (isIptablesActive(active)) {
    return true;
  }

  return false;
}

/**
 * 获取切换按钮文本
 * @param firewallStatus 防火墙状态
 * @param t 国际化函数
 * @returns 按钮文本
 */
export function getSwitchButtonText(
  firewallStatus: NftablesStatus | null,
  t: (key: string) => string
): string {
  if (!firewallStatus) return '';

  const { active } = firewallStatus;

  // 当前是nftables，显示切换到iptables
  if (isNftablesActive(active)) {
    return t('app.nftables.button.switchToIptables');
  }

  // 当前是iptables，显示切换到nftables
  if (isIptablesActive(active)) {
    return t('app.nftables.button.switchToNftables');
  }

  return '';
}

/**
 * 获取防火墙状态的完整信息
 * @param firewallStatus 防火墙状态
 * @param t 国际化函数
 * @returns 状态信息对象
 */
export function getFirewallStatusInfo(
  firewallStatus: NftablesStatus | null,
  t: (key: string) => string
): FirewallStatusInfo {
  const active = firewallStatus?.active;
  const status = firewallStatus?.status;

  return {
    installStatusText: getInstallStatusText(status, t),
    activeSystemColor: getActiveSystemColor(active),
    activeSystemText: getActiveSystemText(active, t),
    canSwitch: shouldShowSwitchButton(firewallStatus),
    switchButtonText: getSwitchButtonText(firewallStatus, t),
    isNftablesActive: isNftablesActive(active),
    isIptablesActive: isIptablesActive(active),
  };
}
