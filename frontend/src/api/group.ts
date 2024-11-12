import { GroupEntity } from '@/entity/Group';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getGroupListApi(params: ApiListParams) {
  return request.get<ApiListResult<GroupEntity>>('groups', params);
}
