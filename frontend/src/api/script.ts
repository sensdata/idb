import { SCRIPT_TYPE } from '@/config/enum';
import { ScriptEntity } from '@/entity/Script';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface ScriptListApiParams extends ApiListParams {
  type?: SCRIPT_TYPE;
  category?: string;
}
export function getScriptListApi(params: ScriptListApiParams) {
  return request.get<ApiListResult<ScriptEntity>>('scripts/{host}', params);
}

export function getScriptDetailApi(params: {
  name: string;
  category: string;
  type: SCRIPT_TYPE;
}) {
  return request.get<ScriptEntity>('scripts/{host}/detail', params);
}

export interface ScriptCategoryListApiParams extends ApiListParams {
  type: SCRIPT_TYPE;
}
export function getScriptCategoryListApi(params: ScriptCategoryListApiParams) {
  return request.get<
    ApiListResult<{
      mod_time: string;
      name: string;
      size: number;
      source: string;
    }>
  >('scripts/{host}/category', params);
}

export function createScriptCategoryApi(params: {
  type: SCRIPT_TYPE;
  category: string;
}) {
  return request.post('scripts/{host}/category', params);
}

export function updateScriptCategoryApi(params: {
  type: SCRIPT_TYPE;
  category: string;
  new_name: string;
}) {
  return request.put('scripts/{host}/category', params);
}

export function deleteScriptCategoryApi(params: {
  type: SCRIPT_TYPE;
  category: string;
}) {
  return request.delete('scripts/{host}/category', params);
}

export interface CreateScriptApiParams {
  name: string;
  type: SCRIPT_TYPE;
  category?: string;
  content: string;
  mark?: string;
}
export function createScriptApi(data: CreateScriptApiParams) {
  return request.post('scripts/{host}', data);
}

export interface UpdateScriptApiParams {
  name: string;
  type: SCRIPT_TYPE;
  category: string;
  new_name: string;
  new_category: string;
  content: string;
}
export function updateScriptApi(data: UpdateScriptApiParams) {
  return request.put('scripts/{host}', data);
}

export interface ScriptVersionsApiParams extends ApiListParams {
  name: string;
  type: SCRIPT_TYPE;
  category: string;
}
export function getScriptVersionListApi(params: ScriptVersionsApiParams) {
  return request.get('scripts/{host}/log', params);
}

export interface RestoreScriptVersionsApiParams {
  name: string;
  type: SCRIPT_TYPE;
  category: string;
  commit_hash: string;
}
export function restoreScriptVersionApi(
  params: RestoreScriptVersionsApiParams
) {
  return request.put('scripts/{host}/log', params);
}

export function runScriptApi(params: { host_id: number; script_path: string }) {
  return request.post<{
    task_id: string;
    log_path: string;
    start: string;
    end: string;
    out: string;
    err: string;
  }>('scripts/{host}/run', params);
}

export interface ScriptRunRecordsApiParams extends ApiListParams {
  path: string;
}
export function getScriptRecordsApi(params: ScriptRunRecordsApiParams) {
  return request.get('scripts/{host}/run/logs', params);
}

export interface ScriptRunLogApiParams extends ApiListParams {
  path: string;
}
export function getScriptRunLogApi(params: ScriptRunLogApiParams) {
  return request.get('scripts/{host}/run/logs/detail', params);
}

export function deleteScriptApi(params: { id: number }) {
  return request.delete('scripts/{host}', params);
}

export interface RestoreScriptApiParams {
  id: number;
  commit_hash: string;
}
export function restoreScriptApi(params: RestoreScriptApiParams) {
  return request.put('scripts/{host}/restore', params);
}
