import { HostEntity, HostGroupEntity } from '@/entity/Host';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';
import axios from 'axios';

export function getHostListApi(params?: ApiListParams) {
  return request.get<ApiListResult<HostEntity>>('hosts', params);
}

export function getHostGroupListApi(params: ApiListParams) {
  return request.get<ApiListResult<HostGroupEntity>>('hosts/groups', params);
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

export function deleteHostApi(ids: number[]) {
  return request.delete('host/delete', { ids });
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

export interface HostStatusResult {
  cpu: number;
  disk: number;
  mem: number;
  mem_total: string;
  mem_used: string;
  rx: number;
  tx: number;
}

export const getHostStatusApi = (hostId: number): Promise<HostStatusResult> => {
  return request.get(`hosts/${hostId}/status`);
};
