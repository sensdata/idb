import { HostGroupEntity } from '@/entity/Group';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getHostGroupListApi(params: ApiListParams) {
  return request.get<ApiListResult<HostGroupEntity>>('hosts/groups', params);
}

export type CreateHostGroupParams = Partial<HostGroupEntity>;
export function createHostGroupApi(params: CreateHostGroupParams) {
  return request.post<ApiListResult<HostGroupEntity>>('groups', params);
}

export type UpdateHostGroupParams = Partial<HostGroupEntity> & { id: number };
export function updateHostGroupApi(params: UpdateHostGroupParams) {
  return request.put<ApiListResult<HostGroupEntity>>('groups', params);
}

export function deleteHostGroupApi(id: number) {
  return request.delete<ApiListResult<HostGroupEntity>>('groups', {
    id,
  });
}
