/**
 * 在选项中查找最长的共同前缀
 */
export function findCommonPrefix(
  options: Array<{ value: string; isDir?: boolean }>
): string {
  if (!options.length) return '';
  if (options.length === 1) return options[0].value;

  const firstItem = options[0].value;
  let prefixLength = firstItem.length;

  for (let i = 1; i < options.length; i++) {
    const currentItem = options[i].value;
    let j = 0;
    while (
      j < prefixLength &&
      j < currentItem.length &&
      firstItem[j] === currentItem[j]
    ) {
      j++;
    }
    prefixLength = j;
  }

  return firstItem.substring(0, prefixLength);
}
