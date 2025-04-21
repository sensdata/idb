export function addRootSlash(path: string): string {
  if (!path.startsWith('/')) {
    return '/' + path;
  }
  return path;
}
