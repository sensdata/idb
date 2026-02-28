export const KEY_ALGORITHM_OPTIONS = [
  { value: 'RSA 2048', label: 'RSA 2048' },
  { value: 'RSA 3072', label: 'RSA 3072' },
  { value: 'RSA 4096', label: 'RSA 4096' },
  { value: 'EC 256', label: 'EC 256' },
  { value: 'EC 384', label: 'EC 384' },
] as const;

export const DEFAULT_KEY_ALGORITHM = 'RSA 2048';
