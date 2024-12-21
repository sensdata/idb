import { CRONTAB_TYPE } from '@/config/enum';
import { CrontabEntity } from '@/entity/Crontab';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface CrontabListApiParams extends ApiListParams {
  type?: CRONTAB_TYPE;
}
export function getCrontabListApi(params: CrontabListApiParams) {
  return request.get<ApiListResult<CrontabEntity>>('crontab/:host', params);
}

export function getCrontabDetailApi(params: { id: number }) {
  return request.get<CrontabEntity>('crontab/:host/detail', params);
}

export function createCrontabApi(data: Partial<CrontabEntity>) {
  return request.post('crontab/:host', data);
}

export function updateCrontabApi(data: Partial<CrontabEntity>) {
  return request.put('crontab/:host', data);
}

export interface CrontabVersionsApiParams extends ApiListParams {
  id: number;
}
export function getCrontabVersionListApi(params: CrontabVersionsApiParams) {
  return request.get('crontab/:host/versions', params);
}

export function runCrontabApi(params: { id: number }) {
  return request.post('crontab/:host/run', params);
}

export interface CrontabRunRecordsApiParams extends ApiListParams {
  id: number;
}
export function getCrontabRecordsApi(params: CrontabRunRecordsApiParams) {
  return request.get('crontab/:host/records', params);
}

export interface CrontabRunLogApiParams extends ApiListParams {
  id: number;
  record_id: number;
}
export function getCrontabRunLogApi(params: CrontabRunLogApiParams) {
  return request.get('crontab/:host/run/log', params);
}

export function deleteCrontabApi(params: { id: number }) {
  return request.delete('crontab/:host', params);
}
