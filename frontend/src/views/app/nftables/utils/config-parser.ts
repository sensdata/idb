/**
 * nftables配置解析工具函数
 */

import { createLogger } from '@/utils/logger';
import type { PortRule } from '@/api/nftables';

const logger = createLogger('NFTablesConfigParser');

export interface ParsedPortRule {
  port: number;
  protocol?: 'tcp' | 'udp' | 'both';
  action?: 'accept' | 'drop' | 'reject';
}

export interface PortAccessInfo {
  accessible: boolean;
  isLocalOnly: boolean;
  color: string;
  text: string;
}

/**
 * 验证端口号是否有效
 * @param portNum 端口号
 * @returns 是否为有效端口
 */
export function isValidPort(portNum: number): boolean {
  return !Number.isNaN(portNum) && portNum > 0 && portNum <= 65535;
}

/**
 * 解析nftables配置中的开放端口
 * @param configContent nftables配置内容
 * @returns 开放端口的Set集合
 */
export function parseOpenPorts(configContent: string): Set<number> {
  const ports = new Set<number>();

  // 多种正则表达式匹配不同的端口配置格式
  const portPatterns = [
    // 匹配 "tcp dport 22 accept" 或 "tcp dport { 80, 443 } accept"
    /tcp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配 "udp dport 53 accept" 或 "udp dport { 53, 67 } accept"
    /udp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配 "dport 22 accept"
    /dport\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配 "tcp port 22 accept"
    /tcp\s+port\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配 "udp port 53 accept"
    /udp\s+port\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配带有协议的格式 "ip protocol tcp tcp dport 22 accept"
    /ip\s+protocol\s+tcp\s+tcp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
    // 匹配带有协议的格式 "ip protocol udp udp dport 53 accept"
    /ip\s+protocol\s+udp\s+udp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+accept/gi,
  ];

  portPatterns.forEach((regex) => {
    let match;
    // eslint-disable-next-line no-cond-assign
    while ((match = regex.exec(configContent)) !== null) {
      if (match[1]) {
        // 处理端口范围 {80, 443, 9918}
        const portList = match[1].split(/[,\s]+/).filter((p) => p.trim());
        portList.forEach((port) => {
          const portNum = parseInt(port.trim(), 10);
          if (isValidPort(portNum)) {
            ports.add(portNum);
          }
        });
      } else if (match[2]) {
        // 处理单个端口
        const portNum = parseInt(match[2], 10);
        if (isValidPort(portNum)) {
          ports.add(portNum);
        }
      }
    }
  });

  // 额外处理端口范围格式，如 "22-25"
  const portRangeRegex = /dport\s+(\d+)-(\d+)\s+accept/gi;
  let rangeMatch;
  // eslint-disable-next-line no-cond-assign
  while ((rangeMatch = portRangeRegex.exec(configContent)) !== null) {
    const startPort = parseInt(rangeMatch[1], 10);
    const endPort = parseInt(rangeMatch[2], 10);
    if (
      isValidPort(startPort) &&
      isValidPort(endPort) &&
      startPort <= endPort
    ) {
      for (let port = startPort; port <= endPort; port += 1) {
        ports.add(port);
      }
    }
  }

  // 调试输出（开发环境）
  if (ports.size > 0) {
    logger.log(
      '解析到的开放端口:',
      Array.from(ports).sort((a, b) => a - b)
    );
  }

  return ports;
}

/**
 * 解析nftables配置中的完整端口规则信息
 * 从配置内容中解析出端口、协议和动作，不设置任何默认值
 * @param configContent nftables配置内容
 * @returns 解析出的端口规则数组
 */
export function parsePortRules(configContent: string): ParsedPortRule[] {
  const rules: ParsedPortRule[] = [];

  logger.log('开始解析配置内容:', configContent.substring(0, 200) + '...');

  // 定义更详细的正则表达式，捕获协议和动作信息
  // 按优先级排序，先匹配更具体的格式，避免重复匹配
  const rulePatterns = [
    // 匹配带有协议的格式 "ip protocol tcp tcp dport 22 accept"
    {
      regex:
        /ip\s+protocol\s+(tcp)\s+tcp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配带有协议的格式 "ip protocol udp udp dport 53 accept"
    {
      regex:
        /ip\s+protocol\s+(udp)\s+udp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
    // 匹配 "tcp dport 22 accept" 或 "tcp dport { 80, 443 } accept"
    {
      regex: /(tcp)\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配 "udp dport 53 accept" 或 "udp dport { 53, 67 } accept"
    {
      regex: /(udp)\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
    // 匹配 "tcp port 22 accept"
    {
      regex: /(tcp)\s+port\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配 "udp port 53 accept"
    {
      regex: /(udp)\s+port\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
  ];

  rulePatterns.forEach(({ regex, protocol }) => {
    let match;
    // eslint-disable-next-line no-cond-assign
    while ((match = regex.exec(configContent)) !== null) {
      logger.log('匹配到规则:', match[0], '协议:', protocol);
      const action = (match[match.length - 1] || '').toLowerCase() as
        | 'accept'
        | 'drop'
        | 'reject';

      // 找到端口列表（可能在match[2]或match[3]中，取决于正则表达式）
      const portList = match[2] || match[3];
      logger.log('端口列表:', portList, '动作:', action);

      if (portList && (portList.includes(',') || portList.includes(' '))) {
        // 处理端口范围 {80, 443, 9918}
        const ports = portList.split(/[,\s]+/).filter((p) => p.trim());
        ports.forEach((portStr) => {
          const portNum = parseInt(portStr.trim(), 10);
          if (isValidPort(portNum)) {
            rules.push({
              port: portNum,
              protocol,
              action: action || undefined,
            });
          }
        });
      } else if (portList) {
        // 处理单个端口
        const portNum = parseInt(portList, 10);
        if (isValidPort(portNum)) {
          rules.push({
            port: portNum,
            protocol,
            action: action || undefined,
          });
        }
      }
    }
  });

  // 处理端口范围格式，如 "tcp dport 22-25 accept"
  const rangePatterns = [
    {
      regex: /(tcp)\s+dport\s+(\d+)-(\d+)\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    {
      regex: /(udp)\s+dport\s+(\d+)-(\d+)\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
    {
      regex: /dport\s+(\d+)-(\d+)\s+(accept|drop|reject)/gi,
      protocol: undefined,
    },
  ];

  rangePatterns.forEach(({ regex, protocol }) => {
    let rangeMatch;
    // eslint-disable-next-line no-cond-assign
    while ((rangeMatch = regex.exec(configContent)) !== null) {
      const startIdx = protocol ? 2 : 1;
      const startPort = parseInt(rangeMatch[startIdx], 10);
      const endPort = parseInt(rangeMatch[startIdx + 1], 10);
      const action = (rangeMatch[rangeMatch.length - 1] || '').toLowerCase() as
        | 'accept'
        | 'drop'
        | 'reject';

      if (
        isValidPort(startPort) &&
        isValidPort(endPort) &&
        startPort <= endPort
      ) {
        for (let port = startPort; port <= endPort; port += 1) {
          rules.push({
            port,
            protocol,
            action: action || undefined,
          });
        }
      }
    }
  });

  // 调试输出（开发环境）
  if (rules.length > 0) {
    logger.log(
      '解析到的端口规则:',
      rules.map(
        (rule) =>
          `${rule.port}/${rule.protocol || 'unknown'}/${
            rule.action || 'unknown'
          }`
      )
    );
  }

  return rules;
}

/**
 * 获取端口访问状态信息（基于API返回的状态）
 * @param status API返回的访问状态
 * @param t 国际化函数
 * @returns 端口访问信息
 */
export function getPortAccessInfoFromStatus(
  status: string,
  t: (key: string) => string
): PortAccessInfo {
  switch (status) {
    case 'local-only':
      return {
        accessible: true,
        isLocalOnly: true,
        color: 'blue',
        text: t('app.nftables.config.localOnly'),
      };
    case 'fully-accepted':
      return {
        accessible: true,
        isLocalOnly: false,
        color: 'green',
        text: t('app.nftables.config.fullyAccessible'),
      };
    case 'accepted':
      return {
        accessible: true,
        isLocalOnly: false,
        color: 'green',
        text: t('app.nftables.config.accessible'),
      };
    case 'rejected':
      return {
        accessible: false,
        isLocalOnly: false,
        color: 'red',
        text: t('app.nftables.config.rejected'),
      };
    case 'restricted':
      return {
        accessible: false,
        isLocalOnly: false,
        color: 'orange',
        text: t('app.nftables.config.restricted'),
      };
    case 'unknown':
    default:
      return {
        accessible: false,
        isLocalOnly: false,
        color: 'gray',
        text: t('app.nftables.config.unknown'),
      };
  }
}

/**
 * 获取端口访问状态信息（旧版本，保持向后兼容）
 * @param port 端口号
 * @param address 地址
 * @param openPorts 开放端口集合
 * @param t 国际化函数
 * @returns 端口访问信息
 */
export function getPortAccessInfo(
  port: number,
  address: string,
  openPorts: Set<number>,
  t: (key: string) => string
): PortAccessInfo {
  // 如果是localhost或回环地址，只能本地访问
  if (
    address.startsWith('127.0.0.1') ||
    address.startsWith('::1') ||
    address === 'localhost'
  ) {
    return {
      accessible: true,
      isLocalOnly: true,
      color: 'blue',
      text: t('app.nftables.config.localOnly'),
    };
  }

  // 对于外部IP地址，检查防火墙配置
  const isOpen = openPorts.has(port);
  if (isOpen) {
    return {
      accessible: true,
      isLocalOnly: false,
      color: 'green',
      text: t('app.nftables.config.accessible'),
    };
  }

  return {
    accessible: false,
    isLocalOnly: false,
    color: 'red',
    text: t('app.nftables.config.notAccessible'),
  };
}

/**
 * 格式化端口列表为可读字符串
 * @param ports 端口集合
 * @returns 格式化的端口字符串
 */
export function formatPortsList(ports: Set<number>): string {
  const sortedPorts = Array.from(ports).sort((a, b) => a - b);

  if (sortedPorts.length === 0) {
    return '';
  }

  if (sortedPorts.length <= 10) {
    return sortedPorts.join(', ');
  }

  return `${sortedPorts.slice(0, 10).join(', ')} ... (+${
    sortedPorts.length - 10
  } more)`;
}

/**
 * 从nftables配置内容中解析端口规则，保持原始配置的端口分组
 * @param configContent nftables配置内容
 * @returns PortRule数组
 */
export function parsePortRulesFromConfig(configContent: string): PortRule[] {
  const rules: PortRule[] = [];

  logger.log('开始解析配置内容:', configContent.substring(0, 200) + '...');

  // 定义匹配规则的正则表达式模式，保持原始分组
  const rulePatterns = [
    // 匹配带有协议的格式 "ip protocol tcp tcp dport 22 accept"
    {
      regex:
        /ip\s+protocol\s+(tcp)\s+tcp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配带有协议的格式 "ip protocol udp udp dport 53 accept"
    {
      regex:
        /ip\s+protocol\s+(udp)\s+udp\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
    // 匹配 "tcp dport 22 accept" 或 "tcp dport { 80, 443 } accept"
    {
      regex: /(tcp)\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配 "udp dport 53 accept" 或 "udp dport { 53, 67 } accept"
    {
      regex: /(udp)\s+dport\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
    // 匹配 "tcp port 22 accept"
    {
      regex: /(tcp)\s+port\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'tcp' as const,
    },
    // 匹配 "udp port 53 accept"
    {
      regex: /(udp)\s+port\s+(?:\{([^}]+)\}|(\d+))\s+(accept|drop|reject)/gi,
      protocol: 'udp' as const,
    },
  ];

  rulePatterns.forEach(({ regex, protocol }) => {
    let match;
    // eslint-disable-next-line no-cond-assign
    while ((match = regex.exec(configContent)) !== null) {
      logger.log('匹配到规则:', match[0], '协议:', protocol);
      const action = (match[match.length - 1] || '').toLowerCase() as
        | 'accept'
        | 'drop'
        | 'reject';

      // 找到端口列表（match[2]是集合格式，match[3]是单个端口）
      const portGroup = match[2]; // 端口集合 {port1, port2, ...}
      const singlePort = match[3]; // 单个端口

      if (portGroup) {
        // 处理端口集合 {80, 443, 9918}
        const ports = portGroup
          .split(/[,\s]+/)
          .map((p) => parseInt(p.trim(), 10))
          .filter(isValidPort);

        if (ports.length > 0) {
          rules.push({
            port: ports.length === 1 ? ports[0] : ports,
            protocol: protocol || 'tcp',
            action: action || 'accept',
            description: '',
            source: '',
            destination: '',
          });
        }
      } else if (singlePort) {
        // 处理单个端口
        const portNum = parseInt(singlePort, 10);
        if (isValidPort(portNum)) {
          rules.push({
            port: portNum,
            protocol: protocol || 'tcp',
            action: action || 'accept',
            description: '',
            source: '',
            destination: '',
          });
        }
      }
    }
  });

  // 按最小端口号排序
  rules.sort((a, b) => {
    const portA = Array.isArray(a.port) ? Math.min(...a.port) : a.port;
    const portB = Array.isArray(b.port) ? Math.min(...b.port) : b.port;
    return portA - portB;
  });

  logger.log('解析后的端口规则:', rules);

  return rules;
}

/**
 * 根据端口规则生成nftables配置内容
 * @param rules PortRule数组
 * @returns nftables配置字符串
 */
export function generateNftablesConfig(rules: PortRule[]): string {
  const tcpPorts: number[] = [];
  const udpPorts: number[] = [];
  const bothPorts: number[] = [];

  // 按协议分组端口
  rules.forEach((rule) => {
    if (rule.action === 'accept') {
      // 只处理允许的规则
      const ports = Array.isArray(rule.port) ? rule.port : [rule.port];

      switch (rule.protocol) {
        case 'tcp':
          tcpPorts.push(...ports);
          break;
        case 'udp':
          udpPorts.push(...ports);
          break;
        case 'both':
        default:
          bothPorts.push(...ports);
          break;
      }
    }
  });

  // 生成nftables配置
  const config = `# NFTables配置文件
# 声明表，选择协议族（例如：ip 表）
table ip filter {
    
    # 定义 input 链（控制进入本机的流量）
    chain input {
        type filter hook input priority 0; policy drop;

        # 禁止 ping（丢弃 ICMP echo 请求）
        icmp type echo-request drop

        # 允许本地回环接口的流量
        iifname "lo" accept

        # 保留 agent 和 center 的端口
        tcp dport { 9918, 9919 } accept
        
        # 允许SSH端口（请根据实际情况修改）
        tcp dport 22 accept
        
${
  tcpPorts.length > 0
    ? `        # 允许的TCP端口
        tcp dport { ${tcpPorts.sort((a, b) => a - b).join(', ')} } accept
        `
    : ''
}${
    udpPorts.length > 0
      ? `        # 允许的UDP端口
        udp dport { ${udpPorts.sort((a, b) => a - b).join(', ')} } accept
        `
      : ''
  }${
    bothPorts.length > 0
      ? `        # 允许的TCP/UDP端口
        tcp dport { ${bothPorts.sort((a, b) => a - b).join(', ')} } accept
        udp dport { ${bothPorts.sort((a, b) => a - b).join(', ')} } accept
        `
      : ''
  }
        # 允许已建立的连接
        ct state established,related accept
    }

    # 定义 output 链（控制本机发出的流量）
    chain output {
        type filter hook output priority 0; policy accept; 
    }

    # 定义 forward 链（转发流量，适用于路由器等设备）
    chain forward {
        type filter hook forward priority 0; policy drop; 
    }
}`;

  return config;
}
