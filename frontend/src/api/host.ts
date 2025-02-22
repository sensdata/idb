import { HostEntity, HostStatusDo } from '@/entity/Host';
import request from '@/helper/api-helper';
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

export function deleteHostApi(id: number) {
  return request.delete('hosts/delete', { id });
}

export interface TestSSHResult {
  success: boolean;
  message?: string;
}

export const testHostSSHApi = (
  params: CreateHostParams
): Promise<TestSSHResult> => {
  return axios.post('hosts/test/ssh', params);
};

export interface TestAgentResult {
  installed: boolean;
  message?: string;
}

export interface InstallAgentResult {
  success: boolean;
  message?: string;
}

export const testHostAgentApi = (hostId: number): Promise<TestAgentResult> => {
  return axios.get(`hosts/${hostId}/test/agent`);
};

export const installHostAgentApi = (
  hostId: number
): Promise<InstallAgentResult> => {
  return axios.post(`hosts/${hostId}/agent/install`);
};

export const getHostStatusApi = (hostId: number): Promise<HostStatusDo> => {
  return request.get(`hosts/${hostId}/status`);
};

export const restartHostAgentApi = (hostId: number): Promise<void> => {
  return request.post(`hosts/${hostId}/agent/restart`);
};
