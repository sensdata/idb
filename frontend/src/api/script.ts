import { ScriptEntity } from '@/entity/Script';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface ScriptListApiParams extends ApiListParams {
  path?: string;
  // todo: 接口差参数
  show_hidden?: boolean;
}
export function getScriptListApi(params?: ScriptListApiParams) {
  return request
    .get<ApiListResult<ScriptEntity>>('scripts/:host', params)
    .then((res: any) => {
      return {
        total: res?.item_total,
        items: res?.items,
        page: params?.page || 1,
        page_size: params?.page_size || 20,
      };
    });
}
