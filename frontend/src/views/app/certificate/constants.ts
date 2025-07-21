// Certificate related constants

// Key algorithm options supported by the backend
export const KEY_ALGORITHM_OPTIONS = [
  { value: 'RSA 2048', label: 'RSA 2048' },
  { value: 'RSA 3072', label: 'RSA 3072' },
  { value: 'RSA 4096', label: 'RSA 4096' },
  { value: 'EC 256', label: 'EC 256' },
  { value: 'EC 384', label: 'EC 384' },
] as const;

// Default key algorithm
export const DEFAULT_KEY_ALGORITHM = 'RSA 2048';

// Key algorithm values as union type
export type KeyAlgorithmValue = (typeof KEY_ALGORITHM_OPTIONS)[number]['value'];

// Validation function for key algorithm
export function isValidKeyAlgorithm(value: string): value is KeyAlgorithmValue {
  return KEY_ALGORITHM_OPTIONS.some((option) => option.value === value);
}

// Certificate status options
export const CERTIFICATE_STATUS = {
  VALID: 'valid',
  EXPIRED: 'expired',
  EXPIRING_SOON: 'expiring_soon',
  INVALID: 'invalid',
} as const;

// Expire unit options
export const EXPIRE_UNIT_OPTIONS = [
  { value: 'day', label: 'Days' },
  { value: 'year', label: 'Years' },
] as const;

// Import type options
export const IMPORT_TYPE = {
  FILE_UPLOAD: 0,
  TEXT_INPUT: 1,
  LOCAL_PATH: 2,
} as const;
