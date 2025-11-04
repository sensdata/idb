import request from '@/helper/api-helper';
import { ApiListParams } from '@/types/global';
import {
  DatabaseComposesResponse,
  DatabaseOperateRequest,
  DatabaseSetPortRequest,
  DatabaseGetConfResponse,
  DatabaseSetConfRequest,
  DatabaseRemoteAccessResponse,
  DatabaseSetRemoteAccessRequest,
  DatabasePasswordResponse,
  DatabaseSetPasswordRequest,
  RsyncListTaskResponse,
  RsyncTaskInfo,
  RsyncCreateTaskRequest,
  RsyncCreateTaskResponse,
} from '@/entity/Database';

// ==================== MySQL API ====================

export function getMysqlComposesApi(params: ApiListParams) {
  return request.get<DatabaseComposesResponse>('mysql/{host}', params);
}

export function mysqlOperationApi(data: DatabaseOperateRequest) {
  return request.post('mysql/{host}/operation', data);
}

export function mysqlSetPortApi(data: DatabaseSetPortRequest) {
  return request.post('mysql/{host}/port', data);
}

export function getMysqlConfApi(params: { name: string }) {
  return request.get<DatabaseGetConfResponse>('mysql/{host}/conf', params);
}

export function setMysqlConfApi(data: DatabaseSetConfRequest) {
  return request.post('mysql/{host}/conf', data);
}

export function getMysqlRemoteAccessApi(params: { name: string }) {
  return request.get<DatabaseRemoteAccessResponse>(
    'mysql/{host}/remote_access',
    params
  );
}

export function setMysqlRemoteAccessApi(data: DatabaseSetRemoteAccessRequest) {
  return request.post('mysql/{host}/remote_access', data);
}

export function getMysqlPasswordApi(params: { name: string }) {
  return request.get<DatabasePasswordResponse>('mysql/{host}/password', params);
}

export function setMysqlPasswordApi(data: DatabaseSetPasswordRequest) {
  return request.post('mysql/{host}/password', data);
}

// ==================== PostgreSQL API ====================

export function getPostgreSqlComposesApi(params: ApiListParams) {
  return request.get<DatabaseComposesResponse>('postgresql/{host}', params);
}

export function postgreSqlOperationApi(data: DatabaseOperateRequest) {
  return request.post('postgresql/{host}/operation', data);
}

export function postgreSqlSetPortApi(data: DatabaseSetPortRequest) {
  return request.post('postgresql/{host}/port', data);
}

export function getPostgreSqlConfApi(params: { name: string }) {
  return request.get<DatabaseGetConfResponse>('postgresql/{host}/conf', params);
}

export function setPostgreSqlConfApi(data: DatabaseSetConfRequest) {
  return request.post('postgresql/{host}/conf', data);
}

// ==================== Redis API ====================

export function getRedisComposesApi(params: ApiListParams) {
  return request.get<DatabaseComposesResponse>('redis/{host}', params);
}

export function redisOperationApi(data: DatabaseOperateRequest) {
  return request.post('redis/{host}/operation', data);
}

export function redisSetPortApi(data: DatabaseSetPortRequest) {
  return request.post('redis/{host}/port', data);
}

export function getRedisConfApi(params: { name: string }) {
  return request.get<DatabaseGetConfResponse>('redis/{host}/conf', params);
}

export function setRedisConfApi(data: DatabaseSetConfRequest) {
  return request.post('redis/{host}/conf', data);
}

export function getRedisRemoteAccessApi(params: { name: string }) {
  return request.get<DatabaseRemoteAccessResponse>(
    'redis/{host}/remote_access',
    params
  );
}

export function setRedisRemoteAccessApi(data: DatabaseSetRemoteAccessRequest) {
  return request.post('redis/{host}/remote_access', data);
}

export function getRedisPasswordApi(params: { name: string }) {
  return request.get<DatabasePasswordResponse>('redis/{host}/password', params);
}

export function setRedisPasswordApi(data: DatabaseSetPasswordRequest) {
  return request.post('redis/{host}/password', data);
}

// ==================== Rsync API ====================

export function getRsyncTaskListApi(params: ApiListParams) {
  return request.get<RsyncListTaskResponse>('rsync/task', params);
}

export function getRsyncTaskDetailApi(params: { id: string }) {
  return request.get<RsyncTaskInfo>('rsync/task/query', params);
}

export function createRsyncTaskApi(data: RsyncCreateTaskRequest) {
  return request.post<RsyncCreateTaskResponse>('rsync/task', data);
}

export function deleteRsyncTaskApi(params: { id: string }) {
  return request.delete('rsync/task', params);
}

export function cancelRsyncTaskApi(data: { id: string }) {
  return request.post('rsync/task/cancel', data);
}

export function retryRsyncTaskApi(data: { id: string }) {
  return request.post('rsync/task/retry', data);
}
