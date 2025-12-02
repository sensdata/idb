// MySQL, PostgreSQL, Redis 通用类型
export interface DatabaseComposeInfo {
  name: string;
  version: string;
  port: string;
  status: string;
  container?: string;
}

export interface DatabaseComposesResponse {
  total: number;
  composes: DatabaseComposeInfo[];
}

export interface DatabaseOperateRequest {
  name: string;
  operation: 'start' | 'stop' | 'restart';
}

export interface DatabaseSetPortRequest {
  name: string;
  port: string;
}

export interface DatabaseGetConfResponse {
  path: string;
  content: string;
}

export interface DatabaseSetConfRequest {
  name: string;
  content: string;
}

export interface DatabaseRemoteAccessResponse {
  remote_access: boolean;
}

export interface DatabaseSetRemoteAccessRequest {
  name: string;
  remote_access: boolean;
}

export interface DatabasePasswordResponse {
  password: string;
}

export interface DatabaseSetPasswordRequest {
  name: string;
  new_pass: string;
}

export interface DatabaseConnectionEndpoint {
  host: string;
  port: string;
}

export interface DatabaseConnectionInfo {
  container_connection: DatabaseConnectionEndpoint;
  public_connection: DatabaseConnectionEndpoint;
}

export interface DatabaseConnectionInfoResponse {
  container_connection: DatabaseConnectionEndpoint;
  public_connection: DatabaseConnectionEndpoint;
}

// Rsync 相关类型
export interface RsyncHost {
  id: number;
  host: string;
  port: number;
  user: string;
  auth_mode: string;
  key_path?: string;
  password?: string;
}

export interface RsyncTaskInfo {
  id: string;
  src: string;
  dst: string;
  cache_dir: string;
  mode: 'copy' | 'incremental';
  status: string;
  progress: number;
  step: string;
  start_time: string;
  end_time: string;
  error: string;
  last_log: string;
}

export interface RsyncListTaskResponse {
  total: number;
  tasks: RsyncTaskInfo[];
}

export interface RsyncCreateTaskRequest {
  src_host_id: number;
  src: string;
  dst_host_id: number;
  dst: string;
  mode: 'copy' | 'incremental';
}

export interface RsyncCreateTaskResponse {
  id: string;
}

export interface RsyncQueryTaskRequest {
  id: string;
}

export interface RsyncCancelTaskRequest {
  id: string;
}

export interface RsyncDeleteTaskRequest {
  id: string;
}

export interface RsyncRetryTaskRequest {
  id: string;
}

// Rsync 客户端相关类型
export interface RsyncClientTask {
  id: string;
  name: string;
  direction: 'remote_to_local' | 'local_to_remote';
  local_path: string;
  remote_type: 'ssh' | 'rsync';
  remote_host: string;
  remote_port: number;
  username: string;
  password?: string;
  ssh_private_key?: string;
  auth_mode: 'password' | 'anonymous' | 'private_key';
  remote_path: string;
  module?: string;
  created_at: string;
  updated_at: string;
  state: string;
  last_error?: string;
  attempt: number;
}

export interface RsyncClientCreateTaskRequest {
  name: string;
  direction: 'remote_to_local' | 'local_to_remote';
  local_path: string;
  remote_type: 'ssh' | 'rsync';
  remote_host: string;
  remote_port: number;
  username: string;
  auth_mode: 'password' | 'anonymous' | 'private_key';
  password?: string;
  ssh_private_key?: string;
  remote_path: string;
  module?: string;
  enqueue: boolean;
}

export interface RsyncClientListTaskResponse {
  total: number;
  tasks: RsyncClientTask[];
}

export interface RsyncTaskLog {
  id: string;
  path: string;
}

export interface RsyncClientTaskLogListResponse {
  total: number;
  logs: RsyncTaskLog[];
}
