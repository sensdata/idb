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

export function getCrontabDetailApi(params: { id: number }) {
  return request.get<CrontabEntity>('crontab/{host}/detail', params);
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
}
export function createUpdateCrontabRawApi(
  params: CreateUpdateCrontabRawParams
) {
  return request.post('crontab/{host}/raw', params);
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
