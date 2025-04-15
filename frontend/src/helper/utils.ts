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
