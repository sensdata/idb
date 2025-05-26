import { AppEntity, AppSimpleEntity } from '@/entity/App';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getAppListApi(params?: ApiListParams) {
  return request.get<ApiListResult<AppSimpleEntity>>('store/apps', params);
}

export function getAppDetail(params: { id: number }) {
  return request.get<AppEntity>('store/apps/detail', params);
}

export function syncAppApi() {
  return request.post('/store/apps/sync');
}

export function getInstalledAppListApi(params?: ApiListParams) {
  return request.get<ApiListResult<AppSimpleEntity>>(
    'store/{host}/apps/installed',
    params
  );
}

export interface InstallAppParams {
  id: number;
  version_id: number;
  compose_content: string;
  extra_params: Array<{
    key: string;
    value: string;
  }>;
  form_params: Array<{
    key: string;
    value: string;
  }>;
}
export function installAppApi(params: InstallAppParams) {
  return request.post<{
    log_host: number;
    log_path: string;
  }>('store/{host}/apps/install', params);
}

export function uninstallAppApi(params: { id: number }) {
  return request.post('store/{host}/apps/uninstall', params);
}
