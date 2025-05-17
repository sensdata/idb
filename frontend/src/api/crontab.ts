import { CrontabEntity } from '@/entity/Crontab';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';
import { CRONTAB_TYPE } from '@/config/enum';

export interface CrontabListApiParams extends ApiListParams {
  type: CRONTAB_TYPE;
  category: string;
}
export function getCrontabListApi(params: CrontabListApiParams) {
  return request.get<ApiListResult<CrontabEntity>>('crontab/{host}', params);
}

export interface CrontabDetailApiParams {
  type: CRONTAB_TYPE;
  category: string;
  name: string;
}

export function getCrontabDetailApi(params: CrontabDetailApiParams) {
  return request.get<CrontabEntity>('crontab/{host}/raw', params);
}

export function createCrontabApi(data: Partial<CrontabEntity>) {
  return request.post('crontab/{host}', data);
}

export function updateCrontabApi(data: Partial<CrontabEntity>) {
  return request.put('crontab/{host}', data);
}

export interface CreateUpdateCrontabRawParams {
  type: CRONTAB_TYPE;
  category: string;
  name: string;
  content: string;
  isEdit?: boolean;
}
export function createUpdateCrontabRawApi(
  params: CreateUpdateCrontabRawParams
) {
  const { isEdit, ...requestParams } = params;
  return isEdit
    ? request.put('crontab/{host}/raw', requestParams)
    : request.post('crontab/{host}/raw', requestParams);
}

export interface CrontabCategoryListApiParams extends ApiListParams {
  type: CRONTAB_TYPE;
}
export function getCrontabCategoryListApi(
  params: CrontabCategoryListApiParams
) {
  return request.get<
    ApiListResult<{
      mod_time: string;
      name: string;
      size: number;
      source: string;
    }>
  >('crontab/{host}/category', params);
}

export function createCrontabCategoryApi(params: {
  type: CRONTAB_TYPE;
  category: string;
}) {
  return request.post('crontab/{host}/category', params);
}

export function updateCrontabCategoryApi(params: {
  type: CRONTAB_TYPE;
  category: string;
  new_name: string;
}) {
  return request.put('crontab/{host}/category', params);
}

export function deleteCrontabCategoryApi(params: {
  type: CRONTAB_TYPE;
  category: string;
}) {
  return request.delete('crontab/{host}/category', params);
}

export interface CrontabVersionsApiParams extends ApiListParams {
  id: number;
}
export function getCrontabVersionListApi(params: CrontabVersionsApiParams) {
  return request.get('crontab/{host}/versions', params);
}

export interface RunCrontabParams {
  id: number;
  type: CRONTAB_TYPE;
  category: string;
  name: string;
}

export function runCrontabApi(params: RunCrontabParams) {
  return request.post('crontab/{host}/run', params);
}

export interface CrontabRunRecordsApiParams extends ApiListParams {
  id: number;
}
export function getCrontabRecordsApi(params: CrontabRunRecordsApiParams) {
  return request.get('crontab/{host}/records', params);
}

export interface CrontabRunLogApiParams extends ApiListParams {
  id: number;
  record_id: number;
}
export function getCrontabRunLogApi(params: CrontabRunLogApiParams) {
  return request.get('crontab/{host}/run/log', params);
}

export interface DeleteCrontabParams {
  type: CRONTAB_TYPE;
  category: string;
  name: string;
}

export function deleteCrontabApi(params: DeleteCrontabParams) {
  return request.delete('crontab/{host}', params);
}
