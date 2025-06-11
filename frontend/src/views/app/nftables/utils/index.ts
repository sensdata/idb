/**
 * nftables工具函数入口文件
 */

// 配置解析相关工具
export {
  parseOpenPorts,
  parsePortRules,
  isValidPort,
  getPortAccessInfo,
  formatPortsList,
  type PortAccessInfo,
  type ParsedPortRule,
} from './config-parser';

// 防火墙状态管理工具
export {
  getInstallStatusText,
  getActiveSystemColor,
  getActiveSystemText,
  isNftablesActive,
  isIptablesActive,
  shouldShowSwitchButton,
  getSwitchButtonText,
  getFirewallStatusInfo,
  type FirewallStatusInfo,
} from './firewall-status';
