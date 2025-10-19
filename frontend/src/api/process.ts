import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export function getProcessListApi(params: ApiListParams & { search?: string }) {
  return request.get<ApiListResult<any>>('process/{host}', {
    page: params.page,
    page_size: params.page_size,
    order_by: params.order_by || 'cpu',
    order: params.order || 'desc',
    name: params.search || params.name,
    pid: params.pid,
    user: params.user,
  });
}

export function getProcessDetailApi(params: { pid: string }) {
  return request.get<any>('process/{host}/detail', {
    pid: params.pid,
  });
}

export function killProcessApi(params: { pid: string }) {
  return request.delete('process/{host}', {
    pid: params.pid,
  });
}
