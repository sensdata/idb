/**
 * SSH 模块的统一类型定义文件
 */

// SSH配置相关类型

/**
 * SSH Configuration Response Interface
 */
export interface SSHConfigResponse {
  port: string;
  listen_address: string;
  permit_root_login: string;
  password_authentication: string;
  pubkey_authentication: string;
  use_dns: string;
  auto_start: boolean;
}

/**
 * SSH Configuration Content Response Interface
 */
export interface SSHConfigContentResponse {
  content: string;
}

/**
 * SSH Configuration Interface (for use in the component)
 */
export interface SSHConfig {
  port: string;
  listenAddress: string;
  permitRootLogin: string;
  passwordAuth: boolean;
  keyAuth: boolean;
  reverseLookup: boolean;
}

// SSHKey 相关类型定义

export type EncryptionMode = 'rsa' | 'ed25519' | 'ecdsa' | 'dsa';
export type KeyBits = 1024 | 2048;

export interface SSHKeyRecord {
  key_name: string;
  key_bits: number;
  fingerprint: string;
  has_private_key: boolean;
  private_key_path: string;
  status: string;
  user: string;
}

export interface GenerateKeyForm {
  key_name: string;
  encryption_mode: EncryptionMode;
  key_bits: KeyBits;
  password: string;
  enable: boolean;
}

export enum SSHKeyStatus {
  ENABLED = 'enabled',
  DISABLED = 'disabled',
}

// Store相关类型定义

// 定义SSH状态类型，确保与Arco Design组件兼容
export type SSHStatus =
  | 'running'
  | 'stopped'
  | 'starting'
  | 'stopping'
  | 'error'
  | 'unhealthy'
  | 'loading';

// 服务端SSH配置接口
export interface SSHStoreConfig {
  status: string;
  port: string;
  listen_address: string;
  permit_root_login: string;
  password_authentication: string;
  pubkey_authentication: string;
  use_dns: string;
  auto_start: boolean;
  message?: string;
}

// 前端表单配置接口，与API响应格式不同
export interface SSHFormConfig {
  port: string;
  listenAddress: string;
  permitRootLogin: string;
  passwordAuth: boolean;
  keyAuth: boolean;
  reverseLookup: boolean;
}

export interface SSHState {
  config: SSHStoreConfig | null;
  status: SSHStatus;
  loading: string | null;
  hostId: number | null;
  lastFetch: number;
  formConfig: SSHFormConfig;
  isConfigFetching: boolean;
  requestId: number;
}
