// Certificate entity types for frontend

// Enums
export enum CertificateStatus {
  VALID = 'valid',
  EXPIRED = 'expired',
  EXPIRING_SOON = 'expiring_soon',
  INVALID = 'invalid',
}

export enum KeyAlgorithm {
  RSA = 'rsa',
  ECDSA = 'ecdsa',
  ED25519 = 'ed25519',
}

export enum ImportType {
  FILE_UPLOAD = 0,
  TEXT_INPUT = 1,
  LOCAL_PATH = 2,
}

export interface CertificateSimpleEntity {
  domain: string;
  alt_names: string[];
  not_before: Date;
  not_after: Date;
  issuer_organization: string;
  status: CertificateStatus;
  source: string;
}

export interface CertificateEntity {
  alias: string;
  requester: string;
  certificates: CertificateSimpleEntity[];
}

export interface CertificateDetailEntity {
  // 证书信息
  domain: string;
  alt_names: string[];
  not_before: Date;
  not_after: Date;
  country: string;
  organization: string;
  key_algorithm: string;
  key_size: number;
  is_ca: boolean;

  // 签发机构信息
  issuer_cn: string;
  issuer_country: string;
  issuer_organization: string;

  // 证书代码（PEM格式）
  pem: string;
  source: string;
}

export interface PrivateKeyEntity {
  alias: string;
  key_algorithm: string;
  key_size: number;
  pem: string;
}

export interface CSREntity {
  common_name: string;
  country: string;
  organization: string;
  email_addresses: string[];
  pem: string;
}

// Form types
export interface CreateCertificateGroupForm {
  alias: string;
  domain_name: string;
  email: string;
  organization: string;
  organization_unit: string;
  country: string;
  province: string;
  city: string;
  key_algorithm: KeyAlgorithm;
}

export interface SelfSignedCertificateForm {
  alias: string;
  expire_unit: 'day' | 'year';
  expire_value: number;
  alt_domains: string;
  alt_ips: string;
  is_ca: boolean;
}

export interface ImportCertificateForm {
  alias: string;

  // Key import options
  key_type: ImportType;
  key_file?: File;
  key_content?: string;
  key_path?: string;

  // Certificate import options
  ca_type: ImportType;
  ca_file?: File;
  ca_content?: string;
  ca_path?: string;

  // CSR import options (optional)
  csr_type?: ImportType;
  csr_file?: File;
  csr_content?: string;
  csr_path?: string;
}

// Table column types
export interface CertificateTableColumn {
  title: string;
  dataIndex: string;
  key: string;
  width?: number;
  align?: 'left' | 'center' | 'right';
  sorter?: boolean;
  render?: (value: any, record: any) => any;
}

// Operation types
export interface CertificateOperation {
  key: string;
  label: string;
  icon?: string;
  type?: 'primary' | 'danger' | 'warning';
  disabled?: boolean;
  handler: (record: any) => void;
}

// Modal types
export interface CertificateModalProps {
  visible: boolean;
  loading?: boolean;
  title?: string;
  width?: number;
}

export interface CertificateDetailModalProps extends CertificateModalProps {
  certificate?: CertificateDetailEntity;
  source?: string;
}

export interface ImportCertificateModalProps extends CertificateModalProps {
  onSubmit: (form: ImportCertificateForm) => Promise<void>;
}

export interface CreateCertificateModalProps extends CertificateModalProps {
  onSubmit: (form: CreateCertificateGroupForm) => Promise<void>;
}

export interface SelfSignedCertificateModalProps extends CertificateModalProps {
  alias: string;
  onSubmit: (form: SelfSignedCertificateForm) => Promise<void>;
}

// Utility types
export interface CertificateValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
}

export interface CertificateExpirationInfo {
  isExpired: boolean;
  isExpiringSoon: boolean;
  daysUntilExpiration: number;
  expirationDate: Date;
}
