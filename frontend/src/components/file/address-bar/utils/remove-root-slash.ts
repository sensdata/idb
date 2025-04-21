export function removeRootSlash(path: string): string {
  if (path.startsWith('/')) {
    return path.substring(1);
  }
  return path;
}
