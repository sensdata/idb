import { ScriptType } from '@/config/enum';
import { ScriptEntity } from '@/entity/Script';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface ScriptListApiParams extends ApiListParams {
  type?: ScriptType;
  category?: string;
}
export function getScriptListApi(params: ScriptListApiParams) {
  return request.get<ApiListResult<ScriptEntity>>('scripts/:host', params);
}

export function getScriptDetailApi(params: { id: number }) {
  return request.get<ScriptEntity>('scripts/:host/detail', params);
}

export interface ScriptCategoryListApiParams extends ApiListParams {
  type: ScriptType;
}
export function getScriptCategoryListApi(params: ScriptCategoryListApiParams) {
  return request.get<ApiListResult<string>>('scripts/:host/category', params);
}

export interface CreateScriptApiParams {
  name: string;
  type: ScriptType;
  category?: string | null;
  content: string;
  mark?: string;
}
export function createScriptApi(data: CreateScriptApiParams) {
  return request.post('scripts/:host', data);
}

export interface UpdateScriptApiParams {
  id: number;
  name?: string;
  category?: string | null;
  content: string;
}
export function updateScriptApi(data: UpdateScriptApiParams) {
  return request.put('scripts/:host', data);
}

export interface ScriptVersionsApiParams extends ApiListParams {
  id: number;
}
export function getScriptVersionListApi(params: ScriptVersionsApiParams) {
  return request.get('scripts/:host/versions', params);
}

export function runScriptApi(params: { id: number }) {
  return request.post('scripts/:host/run', params);
}

export interface ScriptRunRecordsApiParams extends ApiListParams {
  id: number;
}
export function getScriptRecordsApi(params: ScriptRunRecordsApiParams) {
  return request.get('scripts/:host/records', params);
}

export interface ScriptRunLogApiParams extends ApiListParams {
  id: number;
  record_id: number;
}
export function getScriptRunLogApi(params: ScriptRunLogApiParams) {
  return request.get('scripts/:host/run/log', params);
}

export function deleteScriptApi(params: { id: number }) {
  return request.delete('scripts/:host', params);
}

export interface RestoreScriptApiParams {
  id: number;
  commit_hash: string;
}
export function restoreScriptApi(params: RestoreScriptApiParams) {
  return request.put('scripts/:host/restore', params);
}
