import { HostEntity } from '@/entity/host';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getHostListApi(params: ApiListParams) {
  return request.get<ApiListResult<HostEntity>>('hosts', params);
}

export function getHostOptionsApi() {
  return request.get<{ items: HostEntity[] }>('hosts/options');
}

export type CreateHostParams = Partial<HostEntity>;
export function createHostApi(data: CreateHostParams) {
  return request.post<HostEntity>('hosts', data);
}

export function deleteHostApi(ids: number[]) {
  return request.delete('host/delete', { ids });
}
