function normalizeAbsolutePath(path: string): string {
  const segments = path.split('/').filter(Boolean);
  const stack: string[] = [];

  segments.forEach((segment) => {
    if (segment === '.' || segment === '') return;
    if (segment === '..') {
      if (stack.length > 0) stack.pop();
      return;
    }
    stack.push(segment);
  });

  return `/${stack.join('/')}`.replace(/\/+/g, '/');
}

export function resolveLinuxPathInput(
  rawInput: string,
  currentPath: string,
  homePath = '/'
): string {
  const input = rawInput.trim();
  const normalizedCurrent = normalizeAbsolutePath(currentPath || '/');

  if (!input) return normalizedCurrent;

  const withHome =
    input === '~' || input.startsWith('~/')
      ? `${homePath}${input.slice(1)}`
      : input;

  const absolute = withHome.startsWith('/')
    ? withHome
    : `${normalizedCurrent}/${withHome}`;

  return normalizeAbsolutePath(absolute);
}

export function resolveSearchContext(
  rawInput: string,
  currentPath: string,
  homePath = '/'
): {
  searchPath: string;
  searchTerm: string;
} {
  const input = rawInput.trim();
  if (!input) {
    return {
      searchPath: resolveLinuxPathInput('', currentPath, homePath),
      searchTerm: '',
    };
  }

  if (input.endsWith('/')) {
    return {
      searchPath: resolveLinuxPathInput(input, currentPath, homePath),
      searchTerm: '',
    };
  }

  const lastSlashIndex = input.lastIndexOf('/');
  if (lastSlashIndex >= 0) {
    const baseInput = input.slice(0, lastSlashIndex + 1);
    const searchTerm = input.slice(lastSlashIndex + 1);
    return {
      searchPath: resolveLinuxPathInput(baseInput, currentPath, homePath),
      searchTerm,
    };
  }

  return {
    searchPath: resolveLinuxPathInput('', currentPath, homePath),
    searchTerm: input,
  };
}
