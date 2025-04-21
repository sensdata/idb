/**
 * 获取当前路径的目录部分
 */
import { addRootSlash } from './add-root-slash';

export function getSearchPath(value: string): string {
  const lastSlashIndex = value.lastIndexOf('/');

  if (lastSlashIndex >= 0) {
    // 如果路径中有斜杠，则最后一个斜杠前的所有内容都是目录
    return addRootSlash(value.substring(0, lastSlashIndex + 1));
  }

  // 如果我们在根级别且输入中没有斜杠，则在根目录中搜索
  return '/';
}
