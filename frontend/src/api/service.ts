import { ServiceEntity, ServiceHistoryEntity } from '@/entity/Service';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';
import { SERVICE_TYPE, SERVICE_OPERATION, SERVICE_ACTION } from '@/config/enum';

export interface ServiceListApiParams extends ApiListParams {
  type: SERVICE_TYPE;
  category: string;
  host?: number;
}

export function getServiceListApi(params: ServiceListApiParams) {
  return request.get<ApiListResult<ServiceEntity>>('services/{host}', params);
}

export interface ServiceDetailApiParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
}

export function getServiceDetailApi(params: ServiceDetailApiParams) {
  return request.get<string>('services/{host}/raw', params);
}

export interface ServiceFormApiParams {
  type: SERVICE_TYPE;
  category: string;
  name?: string;
}

export interface ServiceFormField {
  name: string;
  label: string;
  key: string;
  type: string;
  default: string;
  value: string;
  required: boolean;
  hint: string;
  options?: string[];
  validation?: {
    min_length?: number;
    max_length?: number;
    pattern?: string;
  };
}

export interface ServiceForm {
  fields: ServiceFormField[];
}

export function getServiceFormApi(params: ServiceFormApiParams) {
  return request.get<ServiceForm>('services/{host}/form', params);
}

export interface CreateServiceRawParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  content: string;
}

export function createServiceRawApi(params: CreateServiceRawParams) {
  return request.post('services/{host}/raw', params);
}

export interface UpdateServiceRawParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  new_name: string;
  content: string;
}

export function updateServiceRawApi(params: UpdateServiceRawParams) {
  return request.put('services/{host}/raw', params);
}

export interface CreateServiceFormParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  form: Array<{ key: string; value: string }>;
}

export function createServiceFormApi(params: CreateServiceFormParams) {
  return request.post('services/{host}/form', params);
}

export interface UpdateServiceFormParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  new_name: string;
  new_category: string;
  form: Array<{ key: string; value: string }>;
}

export function updateServiceFormApi(params: UpdateServiceFormParams) {
  return request.put('services/{host}/form', params);
}

export interface ServiceCategoryListApiParams extends ApiListParams {
  type: SERVICE_TYPE;
  host?: number;
}

export function getServiceCategoryListApi(
  params: ServiceCategoryListApiParams
) {
  return request.get<
    ApiListResult<{
      mod_time: string;
      name: string;
      size: number;
      source: string;
    }>
  >('services/{host}/category', params);
}

export function createServiceCategoryApi(params: {
  type: SERVICE_TYPE;
  category: string;
}) {
  return request.post('services/{host}/category', params);
}

export function updateServiceCategoryApi(params: {
  type: SERVICE_TYPE;
  category: string;
  new_name: string;
}) {
  return request.put('services/{host}/category', params);
}

export function deleteServiceCategoryApi(params: {
  type: SERVICE_TYPE;
  category: string;
}) {
  return request.delete('services/{host}/category', params);
}

export interface ServiceHistoryApiParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  page: number;
  pageSize: number;
}

// 后端历史响应类型
interface ServiceHistoryResponse {
  commit_hash: string;
  author: string;
  date: string;
  message: string;
  changes?: number;
}

// 分页结果类型
interface PageResult<T> {
  items: T[];
  total: number;
}

export function getServiceHistoryApi(
  params: ServiceHistoryApiParams
): Promise<ApiListResult<ServiceHistoryEntity>> {
  return request
    .get<PageResult<ServiceHistoryResponse>>('services/{host}/history', {
      type: params.type,
      category: params.category,
      name: params.name,
      page: params.page,
      page_size: params.pageSize,
    })
    .then((response) => {
      // 转换后端数据为前端格式
      const items: ServiceHistoryEntity[] = response.items.map(
        (item: ServiceHistoryResponse) => ({
          commit: item.commit_hash,
          author: item.author,
          date: item.date,
          message: item.message,
          changes: item.changes || 0,
        })
      );

      return {
        items,
        total: response.total,
        page: params.page,
        page_size: params.pageSize,
      };
    });
}

export interface ServiceDiffApiParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  commit: string;
}

export function getServiceDiffApi(params: ServiceDiffApiParams) {
  return request.get<string>('services/{host}/diff', params);
}

export interface RestoreServiceParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  commit: string;
}

export function restoreServiceApi(params: RestoreServiceParams) {
  return request.put('services/{host}/restore', {
    type: params.type,
    category: params.category,
    name: params.name,
    commit_hash: params.commit, // 后端期望 commit_hash 而不是 commit
  });
}

export interface ServiceActivateParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  action: SERVICE_ACTION;
}

export function serviceActivateApi(params: ServiceActivateParams) {
  return request.post('services/{host}/activate', params);
}

export interface ServiceOperateParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  operation: SERVICE_OPERATION;
}

export interface ServiceOperateResult {
  result: string;
}

export function serviceOperateApi(params: ServiceOperateParams) {
  return request.post<ServiceOperateResult>('services/{host}/operate', params);
}

export interface DeleteServiceParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
}

export function deleteServiceApi(params: DeleteServiceParams) {
  return request.delete('services/{host}', params);
}

export function syncGlobalServiceApi() {
  return request.post('services/{host}/sync');
}

export interface ServiceLogStreamParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  follow?: boolean;
  tail?: number;
  since?: string;
}

// 构建服务日志流的URL
export function buildServiceLogStreamUrl(
  params: ServiceLogStreamParams
): string {
  const { type, category, name, follow, tail, since } = params;

  const searchParams = new URLSearchParams({
    type,
    category,
    name,
  });

  if (follow !== undefined) {
    searchParams.append('follow', follow.toString());
  }

  if (tail !== undefined) {
    searchParams.append('tail', tail.toString());
  }

  if (since) {
    searchParams.append('since', since);
  }

  // 注意：这里需要与实际的API路径匹配
  return `services/{host}/logs/tail?${searchParams.toString()}`;
}
