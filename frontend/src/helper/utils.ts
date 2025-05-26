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
  return '#' + (0x1000000 + Math.random() * 0xffffff).toString(16).substr(1, 6);
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
  const defaultColor = '#cccccc';
  if (!char || typeof char !== 'string') {
    return defaultColor;
  }

  const code = char.charCodeAt(0);
  const hue = (code * 137) % 360;

  const saturation = 60;
  const lightness = 50;

  return hslToHex(hue, saturation, lightness);
}
