export interface PmaComposeInfo {
  name: string;
  version: string;
  port: string;
  status: string;
  container?: string;
}

export interface PmaComposesResponse {
  total: number;
  composes: PmaComposeInfo[];
}

export interface PmaOperateRequest {
  name: string;
  operation: 'start' | 'stop' | 'restart';
}

export interface PmaSetPortRequest {
  name: string;
  port: string;
}

export interface PmaServerInfo {
  verbose: string;
  host: string;
  port: string;
}

export interface PmaGetServersResponse {
  total: number;
  servers: PmaServerInfo[];
}

export interface PmaGetServersParams {
  // phpMyAdmin compose 名称
  name: string;
  page: number;
  page_size: number;
}

export interface PmaAddOrUpdateServerRequest {
  name: string;
  verbose: string;
  host: string;
  port: string;
}

export interface PmaRemoveServerRequest {
  name: string;
  host: string;
  port: string;
}
