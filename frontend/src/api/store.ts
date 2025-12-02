import { AppEntity, AppSimpleEntity } from '@/entity/App';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getAppListApi(params?: ApiListParams) {
  return request.get<ApiListResult<AppSimpleEntity>>(
    'store/{host}/apps',
    params
  );
}

export function getAppDetailApi(params: { id: number }) {
  return request.get<AppEntity>('store/{host}/apps/detail', params);
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
  env_content: string;
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

export interface UpgradeAppParams {
  id: number;
  upgrade_version_id: number;
  compose_name: string;
}
export function upgradeAppApi(params: UpgradeAppParams) {
  return request.post<{
    log_host: number;
    log_path: string;
  }>('store/{host}/apps/upgrade', params);
}

export interface UninstallAppParams {
  id: number;
  compose_name: string;
}
export function uninstallAppApi(params: UninstallAppParams) {
  return request.post<{
    log_host: number;
    log_path: string;
  }>('store/{host}/apps/uninstall', params);
}
