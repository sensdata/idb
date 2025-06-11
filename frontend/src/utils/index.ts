type TargetContext = '_self' | '_parent' | '_blank' | '_top';

export function isArray(obj: any) {
  return Array.isArray
    ? Array.isArray(obj)
    : Object.prototype.toString.call(obj) === '[object Array]';
}

export function serializeQueryParams(params: Record<string, any>) {
  const res: string[] = [];
  for (const name in params) {
    if (
      Object.prototype.hasOwnProperty.call(params, name) &&
      params[name] != null
    ) {
      let value = params[name];
      if (isArray(value)) {
        value = value.join(',');
      }
      res.push(
        encodeURIComponent(name) + '=' + encodeURIComponent(String(value))
      );
    }
  }
  return res.join('&');
}

export const openWindow = (
  url: string,
  opts?: { target?: TargetContext; [key: string]: any }
) => {
  const { target = '_blank', ...others } = opts || {};
  window.open(
    url,
    target,
    Object.entries(others)
      .reduce((preValue: string[], curValue) => {
        const [key, value] = curValue;
        return [...preValue, `${key}=${value}`];
      }, [])
      .join(',')
  );
};

export const regexUrl = new RegExp(
  '^(?!mailto:)(?:(?:http|https|ftp)://)(?:\\S+(?::\\S*)?@)?(?:(?:(?:[1-9]\\d?|1\\d\\d|2[01]\\d|22[0-3])(?:\\.(?:1?\\d{1,2}|2[0-4]\\d|25[0-5])){2}(?:\\.(?:[0-9]\\d?|1\\d\\d|2[0-4]\\d|25[0-4]))|(?:(?:[a-z\\u00a1-\\uffff0-9]+-?)*[a-z\\u00a1-\\uffff0-9]+)(?:\\.(?:[a-z\\u00a1-\\uffff0-9]+-?)*[a-z\\u00a1-\\uffff0-9]+)*(?:\\.(?:[a-z\\u00a1-\\uffff]{2,})))|localhost)(?::\\d{2,5})?(?:(/|\\?|#)[^\\s]*)?$',
  'i'
);

// IP 地址验证函数
export function validateIPFormat(ip: string): boolean {
  if (!ip) return false;

  // IPv4 地址正则
  const ipv4Regex =
    /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

  // IPv4 CIDR 正则
  const ipv4CidrRegex =
    /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\/(?:[0-9]|[1-2][0-9]|3[0-2])$/;

  // IPv4 范围正则
  const ipv4RangeRegex =
    /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)-(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;

  // IPv6 地址正则（简化版）
  const ipv6Regex = /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;

  // IPv6 CIDR 正则
  const ipv6CidrRegex =
    /^([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\/(?:[0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$/;

  // 检查各种格式
  if (
    ipv4Regex.test(ip) ||
    ipv4CidrRegex.test(ip) ||
    ipv6Regex.test(ip) ||
    ipv6CidrRegex.test(ip)
  ) {
    return true;
  }

  // 验证IP范围格式
  if (ipv4RangeRegex.test(ip)) {
    const [startIP, endIP] = ip.split('-');
    // 简单验证起始IP应该小于结束IP（可选，根据需求）
    return ipv4Regex.test(startIP) && ipv4Regex.test(endIP);
  }

  return false;
}

// 获取 IP 地址类型
export function getIPType(ip: string): 'single' | 'cidr' | 'range' {
  if (!ip) return 'single';

  if (ip.includes('/')) {
    return 'cidr';
  }

  if (ip.includes('-')) {
    return 'range';
  }

  return 'single';
}

export default null;
