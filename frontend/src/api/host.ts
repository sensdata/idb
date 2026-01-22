import { HostEntity, HostGroupEntity, HostStatusDo } from '@/entity/Host';
import request, { resolveApiUrl } from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';
import axios from 'axios';

export function getHostListApi(params?: ApiListParams) {
  return request.get<ApiListResult<HostEntity>>('hosts', params);
}

export function getHostInfoApi(hostId: number) {
  return request.get<HostEntity>(`hosts/${hostId}`);
}

export type CreateHostParams = Partial<HostEntity>;
export interface CreateHostResult {
  id: number;
  name: string;
}

export const createHostApi = (
  params: CreateHostParams
): Promise<CreateHostResult> => {
  return axios.post('/hosts', params);
};

export type UpdateHostParams = Partial<HostEntity> & { id: number };
export const updateHostApi = (
  params: UpdateHostParams
): Promise<CreateHostResult> => {
  return axios.put('/hosts', params);
};

export const updateHostSSHApi = (
  hostId: number,
  params: UpdateHostParams
): Promise<CreateHostResult> => {
  return axios.put(`/hosts/${hostId}/conf/ssh`, params);
};

export interface UpdateHostAgentParams {
  agent_addr: string;
  agent_port: number;
}

export const updateHostAgentApi = (
  hostId: number,
  params: UpdateHostAgentParams
): Promise<CreateHostResult> => {
  return axios.put(`/hosts/${hostId}/conf/agent`, params);
};

export function deleteHostApi(id: number) {
  return request.delete('hosts', { id });
}

export function getHostGroupListApi(params: ApiListParams) {
  return request.get<ApiListResult<HostGroupEntity>>('hosts/groups', params);
}

export type CreateHostGroupParams = Partial<HostGroupEntity>;
export function createHostGroupApi(params: CreateHostGroupParams) {
  return request.post<ApiListResult<HostGroupEntity>>('hosts/groups', params);
}

export type UpdateHostGroupParams = Partial<HostGroupEntity> & { id: number };
export function updateHostGroupApi(params: UpdateHostGroupParams) {
  return request.put<ApiListResult<HostGroupEntity>>('hosts/groups', params);
}

export function deleteHostGroupApi(id: number) {
  return request.delete<ApiListResult<HostGroupEntity>>('hosts/groups', {
    id,
  });
}

export const testHostSSHApi = (params: CreateHostParams) => {
  return axios.post('hosts/test/ssh', params);
};

export interface TestAgentResult {
  installed: boolean;
  message?: string;
}

export const testHostAgentApi = (hostId: number): Promise<TestAgentResult> => {
  return axios.get(`hosts/${hostId}/test/agent`);
};

export interface LogInfoResult {
  log_host: number;
  log_path: string;
}

export const installHostAgentApi = (hostId: number): Promise<LogInfoResult> => {
  return axios.post(`hosts/${hostId}/agent/install`, {
    upgrade: false,
  });
};

export const upgradeHostAgentApi = (hostId: number): Promise<LogInfoResult> => {
  return axios.post(`hosts/${hostId}/agent/install`, {
    upgrade: true,
  });
};

export const uninstallHostAgentApi = (
  hostId: number
): Promise<LogInfoResult> => {
  return axios.post(`hosts/${hostId}/agent/uninstall`);
};

export const getHostStatusApi = (hostId: number): Promise<HostStatusDo> => {
  return request.get(`hosts/${hostId}/status`);
};

export function connectHostStatusFollowApi(hostId: number): EventSource {
  const url = resolveApiUrl(`hosts/${hostId}/status/follow`);
  return new EventSource(url);
}

export interface HostStatusFollowItem {
  id: number;
  installed: string;
  connected: string;
  activated: boolean;
  can_upgrade: boolean;
  cpu: number;
  mem: number;
  mem_total: string;
  mem_used: string;
  disk: number;
  rx: number;
  tx: number;
}

export function connectAllHostsStatusFollowApi(ids: number[]): EventSource {
  const url = resolveApiUrl(`hosts/status/follow`, { ids: ids.join(',') });
  return new EventSource(url);
}

export const restartHostAgentApi = (hostId: number): Promise<void> => {
  return request.post(`hosts/${hostId}/agent/restart`);
};

export const activateHostApi = (hostId: number): Promise<void> => {
  return request.post(`hosts/${hostId}/activate`);
};
