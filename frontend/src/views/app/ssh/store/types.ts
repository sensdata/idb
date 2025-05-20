// 定义SSH状态类型，确保与Arco Design组件兼容
export type SSHStatus =
  | 'running'
  | 'stopped'
  | 'starting'
  | 'stopping'
  | 'error'
  | 'unhealthy';

export interface SSHConfig {
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

export interface SSHState {
  config: SSHConfig | null;
  status: SSHStatus;
  loading: string | null;
  hostId: number | null;
}
