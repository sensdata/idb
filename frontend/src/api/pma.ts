import request from '@/helper/api-helper';
import { ApiListParams } from '@/types/global';
import {
  PmaComposesResponse,
  PmaOperateRequest,
  PmaSetPortRequest,
  PmaGetServersResponse,
  PmaGetServersParams,
  PmaAddOrUpdateServerRequest,
  PmaRemoveServerRequest,
} from '@/entity/Pma';

export function getPmaComposesApi(params: ApiListParams) {
  return request.get<PmaComposesResponse>('pma/{host}', params);
}

export function pmaOperationApi(data: PmaOperateRequest) {
  return request.post('pma/{host}/operation', data);
}

export function pmaSetPortApi(data: PmaSetPortRequest) {
  return request.post('pma/{host}/port', data);
}

export function getPmaServersApi(params: PmaGetServersParams) {
  return request.get<PmaGetServersResponse>('pma/{host}/servers', params);
}

export function pmaAddServerApi(data: PmaAddOrUpdateServerRequest) {
  return request.post('pma/{host}/server', data);
}

export function pmaUpdateServerApi(data: PmaAddOrUpdateServerRequest) {
  return request.put('pma/{host}/server', data);
}

export function pmaRemoveServerApi(params: PmaRemoveServerRequest) {
  return request.delete('pma/{host}/server', params);
}
