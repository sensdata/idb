/**
 * IP地址验证和处理工具函数
 * @description 使用成熟的库提供IP地址格式验证、类型识别等功能
 */

import { isIP, isIPv4, isIPv6 } from 'is-ip';
import isCidr from 'is-cidr';

/**
 * IP地址类型
 */
export type IPType = 'single' | 'cidr' | 'range';

/**
 * 验证单个IPv4地址格式
 * @param ip - IPv4地址字符串
 * @returns 是否为有效的IPv4地址
 */
export const validateSingleIP = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') {
    return false;
  }
  return isIPv4(ip.trim());
};

/**
 * 验证IPv6地址格式
 * @param ip - IPv6地址字符串
 * @returns 是否为有效的IPv6地址
 */
export const validateSingleIPv6 = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') {
    return false;
  }
  return isIPv6(ip.trim());
};

/**
 * 验证任意IP地址格式（IPv4或IPv6）
 * @param ip - IP地址字符串
 * @returns 是否为有效的IP地址
 */
export const validateAnyIP = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') {
    return false;
  }
  return isIP(ip.trim());
};

/**
 * 验证CIDR网段格式
 * @param cidr - CIDR格式字符串（如: 192.168.1.0/24）
 * @returns 是否为有效的CIDR格式
 */
export const validateCIDR = (cidr: string): boolean => {
  if (!cidr || typeof cidr !== 'string') {
    return false;
  }
  // is-cidr returns a number (version) or false, so we convert to boolean
  return Boolean(isCidr(cidr.trim()));
};

/**
 * 验证IP范围格式
 * @param range - IP范围字符串（如: 192.168.1.1-192.168.1.100）
 * @returns 是否为有效的IP范围格式
 */
export const validateIPRange = (range: string): boolean => {
  if (!range || typeof range !== 'string') {
    return false;
  }

  const trimmedRange = range.trim();

  // 检查是否包含连字符
  if (!trimmedRange.includes('-')) {
    return false;
  }

  // 分割起始和结束IP
  const parts = trimmedRange.split('-');
  if (parts.length !== 2) {
    return false;
  }

  const [startIP, endIP] = parts.map((ip) => ip.trim());

  // 验证两个IP都是有效的
  if (!validateAnyIP(startIP) || !validateAnyIP(endIP)) {
    return false;
  }

  // 确保两个IP是同一类型（都是IPv4或都是IPv6）
  const startIsIPv4 = isIPv4(startIP);
  const endIsIPv4 = isIPv4(endIP);

  if (startIsIPv4 !== endIsIPv4) {
    return false;
  }

  // 对于IPv4，检查起始IP是否小于等于结束IP
  if (startIsIPv4) {
    const ipToNumber = (ip: string): number => {
      return (
        ip
          .split('.')
          .reduce((acc, octet) => (acc << 8) + parseInt(octet, 10), 0) >>> 0
      );
    };
    return ipToNumber(startIP) <= ipToNumber(endIP);
  }

  // 对于IPv6，暂时只进行基本验证
  return true;
};

/**
 * 验证IP格式（支持单个IP、CIDR网段、IP范围）
 * @param ip - 待验证的IP字符串
 * @returns 是否为有效的IP格式
 */
export const validateIPFormat = (ip: string): boolean => {
  if (!ip || typeof ip !== 'string') {
    return false;
  }

  const trimmedIP = ip.trim();

  // 检查单个IP
  if (validateAnyIP(trimmedIP)) {
    return true;
  }

  // 检查CIDR格式
  if (trimmedIP.includes('/') && validateCIDR(trimmedIP)) {
    return true;
  }

  // 检查IP范围格式
  if (trimmedIP.includes('-') && validateIPRange(trimmedIP)) {
    return true;
  }

  return false;
};

/**
 * 获取IP类型
 * @param ip - IP字符串
 * @returns IP类型
 */
export const getIPType = (ip: string): IPType => {
  if (!ip || typeof ip !== 'string') {
    return 'single';
  }

  const trimmedIP = ip.trim();

  if (trimmedIP.includes('/')) {
    return 'cidr';
  }

  if (trimmedIP.includes('-')) {
    return 'range';
  }

  return 'single';
};

/**
 * 获取IP版本
 * @param ip - IP字符串
 * @returns IP版本 (4, 6, 或 undefined)
 */
export const getIPVersion = (ip: string): 4 | 6 | undefined => {
  if (!ip || typeof ip !== 'string') {
    return undefined;
  }

  const trimmedIP = ip.trim();

  if (isIPv4(trimmedIP)) {
    return 4;
  }

  if (isIPv6(trimmedIP)) {
    return 6;
  }

  return undefined;
};

/**
 * 格式化IP显示
 * @param ip - IP字符串
 * @param type - IP类型（可选，如果不提供会自动检测）
 * @returns 格式化后的IP显示字符串
 */
export const formatIPDisplay = (ip: string, type?: IPType): string => {
  if (!ip) return '';

  const trimmedIP = ip.trim();
  const actualType = type || getIPType(trimmedIP);
  const version = getIPVersion(trimmedIP.split('/')[0].split('-')[0]);
  const versionText = version ? `IPv${version}` : '';

  switch (actualType) {
    case 'cidr':
      return `${trimmedIP} (${versionText}网段)`;
    case 'range':
      return `${trimmedIP} (${versionText}范围)`;
    case 'single':
    default:
      return `${trimmedIP} (${versionText})`;
  }
};

/**
 * 解析CIDR网段信息
 * @param cidr - CIDR格式字符串
 * @returns 解析后的网段信息
 */
export const parseCIDR = (
  cidr: string
): {
  network: string;
  prefixLength: number;
  hostCount: number;
  version: 4 | 6;
} | null => {
  if (!validateCIDR(cidr)) {
    return null;
  }

  const [network, prefixLengthStr] = cidr.split('/');
  const prefixLength = parseInt(prefixLengthStr, 10);
  const version = getIPVersion(network);

  if (!version) {
    return null;
  }

  let hostCount = 0;
  if (version === 4) {
    const hostBits = 32 - prefixLength;
    hostCount = hostBits === 0 ? 1 : 2 ** hostBits - 2; // 减去网络地址和广播地址
  } else {
    const hostBits = 128 - prefixLength;
    // 对于IPv6，主机数量可能非常大，这里简化处理
    hostCount = hostBits === 0 ? 1 : 2 ** Math.min(hostBits, 32);
  }

  return {
    network,
    prefixLength,
    hostCount: Math.max(0, hostCount),
    version,
  };
};

/**
 * 检查IP是否在私有网段内
 * @param ip - 单个IP地址
 * @returns 是否为私有IP
 */
export const isPrivateIP = (ip: string): boolean => {
  if (!validateSingleIP(ip)) {
    return false;
  }

  const octets = ip.split('.').map(Number);
  const [a, b] = octets;

  // 10.0.0.0/8
  if (a === 10) {
    return true;
  }

  // 172.16.0.0/12
  if (a === 172 && b >= 16 && b <= 31) {
    return true;
  }

  // 192.168.0.0/16
  if (a === 192 && b === 168) {
    return true;
  }

  // 127.0.0.0/8 (loopback)
  if (a === 127) {
    return true;
  }

  return false;
};

/**
 * 检查IPv6是否在私有网段内
 * @param ip - IPv6地址
 * @returns 是否为私有IPv6
 */
export const isPrivateIPv6 = (ip: string): boolean => {
  if (!validateSingleIPv6(ip)) {
    return false;
  }

  const lowerIP = ip.toLowerCase();

  // Unique Local Addresses (fc00::/7)
  if (lowerIP.startsWith('fc') || lowerIP.startsWith('fd')) {
    return true;
  }

  // Link-Local (fe80::/10)
  if (
    lowerIP.startsWith('fe8') ||
    lowerIP.startsWith('fe9') ||
    lowerIP.startsWith('fea') ||
    lowerIP.startsWith('feb')
  ) {
    return true;
  }

  // Loopback (::1)
  if (lowerIP === '::1') {
    return true;
  }

  return false;
};

/**
 * 统一的私有IP检查函数
 * @param ip - IP地址（IPv4或IPv6）
 * @returns 是否为私有IP
 */
export const isPrivateAddress = (ip: string): boolean => {
  if (isIPv4(ip)) {
    return isPrivateIP(ip);
  }

  if (isIPv6(ip)) {
    return isPrivateIPv6(ip);
  }

  return false;
};
