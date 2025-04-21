/**
 * 获取正在输入的文件/目录名
 */
export function getSearchTerm(value: string): string {
  const lastSlashIndex = value.lastIndexOf('/');

  if (lastSlashIndex >= 0 && lastSlashIndex < value.length - 1) {
    // 返回最后一个斜杠后的部分
    return value.substring(lastSlashIndex + 1);
  }

  if (lastSlashIndex === -1) {
    // 没有斜杠表示我们在根级别输入
    return value;
  }

  // 如果刚输入斜杠则返回空字符串
  return '';
}
