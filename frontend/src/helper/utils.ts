export function compareVersion(version1: string, version2: string) {
  if (!version1 || !version2) {
    return 0;
  }

  version1 = version1.replace(/^v/, '');
  version2 = version2.replace(/^v/, '');

  const v1 = version1.split('.');
  const v2 = version2.split('.');
  const len = Math.max(v1.length, v2.length);

  for (let i = 0; i < len; i += 1) {
    const num1 = parseInt(v1[i] || '0', 10);
    const num2 = parseInt(v2[i] || '0', 10);

    if (num1 > num2) {
      return 1;
    }

    if (num1 < num2) {
      return -1;
    }
  }

  return 0;
}

export function getRandomColor() {
  // 从品牌色系统中随机选择一个颜色，而不是生成随机十六进制
  const brandColors = [
    'var(--idblue-6)',
    'var(--idbgreen-6)',
    'var(--idbcyan-6)',
    'var(--idbautumn-6)',
    'var(--idbred-6)',
    'var(--idbdusk-6)',
    'var(--idbturquoise-6)',
  ];

  const randomIndex = Math.floor(Math.random() * brandColors.length);
  return brandColors[randomIndex];
}

export function hslToHex(h: number, s: number, l: number) {
  s /= 100;
  l /= 100;

  const a = s * Math.min(l, 1 - l);
  const f = (n: number) => {
    const k = (n + h / 30) % 12;
    const color = l - a * Math.max(Math.min(k - 3, 9 - k, 1), -1);
    return Math.round(255 * color)
      .toString(16)
      .padStart(2, '0');
  };

  return `#${f(0)}${f(8)}${f(4)}`;
}

export function getHexColorByChar(char: string) {
  // 使用CSS变量作为默认颜色，优先使用系统色，回退到品牌定义的灰色
  const defaultColor =
    typeof window !== 'undefined'
      ? getComputedStyle(document.documentElement)
          .getPropertyValue('--color-text-4')
          .trim() ||
        getComputedStyle(document.documentElement)
          .getPropertyValue('--idb-fallback-gray')
          .trim() ||
        '#cccccc'
      : '#cccccc';

  if (!char || typeof char !== 'string') {
    return defaultColor;
  }

  const code = char.charCodeAt(0);
  const hue = (code * 137) % 360;

  const saturation = 60;
  const lightness = 50;

  return hslToHex(hue, saturation, lightness);
}

export function checkIpV6(value: string) {
  if (value === '' || value == null) {
    return true;
  }
  const ipv6Regex =
    /^(([0-9a-fA-F]{1,4}:){7}([0-9a-fA-F]{1,4}|:)|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9])?[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9])?[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9])?[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9])?[0-9]))$/;
  return ipv6Regex.test(value);
}
