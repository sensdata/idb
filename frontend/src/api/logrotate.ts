import { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';
import {
  LogrotateEntity,
  LogrotateForm,
  LogrotateCategory,
  LogrotateHistory,
  LogrotateListParams,
  LogrotateFormField,
} from '@/entity/Logrotate';
import { ApiListResult } from '@/types/global';
import request from '@/helper/api-helper';
import { createLogger } from '@/utils/logger';

export type { LogrotateListParams };

// 创建日志实例
const logger = createLogger('LogrotateAPI');

// 常量定义
const DEFAULT_VALUES = {
  FREQUENCY: LOGROTATE_FREQUENCY.Daily,
  COUNT: 'rotate 7',
  CREATE: 'create 0644 root root',
} as const;

// logrotate配置关键字正则表达式
const LOGROTATE_CONFIG_KEYWORDS =
  /^\s*(daily|weekly|monthly|yearly|hourly|rotate|compress|delaycompress|missingok|notifempty|create|prerotate|postrotate|endscript)/;

// 后端 API 响应类型定义
interface LogrotateItemResponse {
  source: string;
  name: string;
  extension: string;
  content: string;
  size: number;
  mod_time: string;
  linked: boolean;
}

interface LogrotateHistoryResponse {
  commit_hash: string;
  author: string;
  email: string;
  date: string;
  message: string;
}

interface LogrotateCategoryResponse {
  source: string;
  name: string;
  extension: string;
  content: string;
  size: number;
  mod_time: string;
  linked: boolean;
}

// API接口类型定义
interface ServiceFormField {
  key: string;
  label: string;
  name: string;
  type: string;
  value: string;
  default?: string;
  hint?: string;
  options?: string[];
  required?: boolean;
  validation?: {
    maxLength?: number;
    maxValue?: number;
    minLength?: number;
    minValue?: number;
    pattern?: string;
  };
}

interface ServiceForm {
  fields: ServiceFormField[];
}

interface CreateServiceForm {
  name: string;
  type: string;
  category?: string;
  form: Array<{ key: string; value: string }>;
}

interface UpdateServiceForm {
  type: string;
  category: string;
  new_category: string;
  name: string;
  new_name: string;
  form: Array<{ key: string; value: string }>;
}

interface CreateLogrotateCategory {
  type: string;
  category: string;
}

interface UpdateLogrotateCategory {
  type: string;
  category: string;
  new_name?: string;
}

interface CreateLogrotateFile {
  type: string;
  category: string;
  name: string;
  content?: string;
}

interface UpdateLogrotateFile {
  type: string;
  category: string;
  name: string;
  content: string;
  new_category?: string;
  new_name?: string;
}

interface RestoreLogrotateFile {
  type: string;
  category: string;
  name: string;
  commit_hash: string;
}

interface ServiceActivate {
  type: string;
  category?: string;
  name: string;
  action: 'activate' | 'deactivate';
}

interface PageResult<T> {
  items: T[];
  total: number;
}

// 插件信息接口
interface LogrotatePluginInfo {
  name: string;
  version: string;
  description?: string;
  author?: string;
  [key: string]: unknown;
}

// 插件菜单接口
interface LogrotatePluginMenu {
  id: string;
  name: string;
  path: string;
  icon?: string;
  children?: LogrotatePluginMenu[];
  [key: string]: unknown;
}

// 历史查询参数接口
interface LogrotateHistoryParams {
  type: LOGROTATE_TYPE;
  category: string;
  name: string;
  page: number;
  pageSize: number;
  host: number;
}

// 历史差异查询参数接口
interface LogrotateHistoryDiffParams {
  type: LOGROTATE_TYPE;
  category: string;
  name: string;
  commit: string;
  host: number;
}

// 恢复参数接口
interface LogrotateRestoreParams {
  type: LOGROTATE_TYPE;
  category: string;
  name: string;
  commit: string;
  host: number;
}

/**
 * 从logrotate配置内容中解析日志路径
 * @param content - logrotate配置文件内容
 * @returns 解析出的日志文件路径
 */
const parseLogPathFromContent = (content: string): string => {
  if (!content?.trim()) {
    return '';
  }

  try {
    // 查找第一个 { 之前的内容，这通常是日志文件路径
    const lines = content.split('\n');
    for (const line of lines) {
      const trimmedLine = line.trim();
      if (trimmedLine && !trimmedLine.startsWith('#')) {
        // 找到第一个非注释行
        const braceIndex = trimmedLine.indexOf('{');
        if (braceIndex !== -1) {
          // 如果这行包含 {，则提取 { 之前的部分作为路径
          return trimmedLine.substring(0, braceIndex).trim();
        }
        if (
          !trimmedLine.includes('}') &&
          !LOGROTATE_CONFIG_KEYWORDS.test(trimmedLine)
        ) {
          // 如果这行不包含配置关键字，可能是路径
          return trimmedLine;
        }
      }
    }
  } catch (error) {
    logger.logError('Failed to parse log path from content:', error);
  }

  return '';
};

/**
 * 转换函数：将后端ServiceFormField转换为前端LogrotateEntity
 * @param name - 配置名称
 * @param type - logrotate类型
 * @param category - 分类
 * @param fields - 后端字段数组
 * @returns 转换后的LogrotateEntity对象
 */
const convertServiceFormToLogrotateEntity = (
  name: string,
  type: LOGROTATE_TYPE,
  category: string,
  fields: ServiceFormField[]
): LogrotateEntity => {
  const entity: LogrotateEntity = {
    name,
    type,
    category,
    path: '',
    frequency: DEFAULT_VALUES.FREQUENCY,
    count: DEFAULT_VALUES.COUNT,
    compress: false,
    delayCompress: false,
    create: DEFAULT_VALUES.CREATE,
    missingOk: false,
    notIfEmpty: false,
    preRotate: '',
    postRotate: '',
    linked: false,
  };

  // 根据字段映射设置值
  fields.forEach((field) => {
    switch (field.key) {
      case 'Path':
        entity.path = field.value || '';
        break;
      case 'Frequency':
        entity.frequency =
          (field.value as LOGROTATE_FREQUENCY) || DEFAULT_VALUES.FREQUENCY;
        break;
      case 'Count':
        entity.count = field.value || DEFAULT_VALUES.COUNT;
        break;
      case 'Compress':
        entity.compress = field.value === 'true';
        break;
      case 'DelayCompress':
        entity.delayCompress = field.value === 'true';
        break;
      case 'Create':
        entity.create = field.value || DEFAULT_VALUES.CREATE;
        break;
      case 'MissingOk':
        entity.missingOk = field.value === 'true';
        break;
      case 'NotIfEmpty':
        entity.notIfEmpty = field.value === 'true';
        break;
      case 'PreRotate':
        entity.preRotate = field.value || '';
        break;
      case 'PostRotate':
        entity.postRotate = field.value || '';
        break;
      default:
        // Handle unknown field keys
        break;
    }
  });

  return entity;
};

/**
 * 转换函数：将LogrotateFormField转换为后端KeyValue格式
 * @param fields - 前端表单字段数组
 * @returns 转换后的键值对数组
 */
const convertLogrotateFormFieldsToKeyValue = (
  fields: LogrotateFormField[]
): Array<{ key: string; value: string }> => {
  return fields.map((field) => ({
    key: field.key,
    value: field.value,
  }));
};

/**
 * 获取logrotate配置列表
 * @param params - 查询参数
 * @returns logrotate配置列表
 */
export const getLogrotateListApi = async (
  params: LogrotateListParams
): Promise<ApiListResult<LogrotateEntity>> => {
  const response = await request.get<PageResult<LogrotateItemResponse>>(
    'logrotate/{host}',
    {
      type: params.type,
      category: params.category,
      page: params.page,
      page_size: params.page_size,
      host: params.host,
    }
  );

  // 转换后端数据为前端格式
  const items: LogrotateEntity[] = response.items.map(
    (item: LogrotateItemResponse) => ({
      id: item.source,
      name: item.name,
      type: params.type,
      category: params.category,
      path: parseLogPathFromContent(item.content || ''),
      frequency: DEFAULT_VALUES.FREQUENCY,
      count: DEFAULT_VALUES.COUNT,
      compress: false,
      delayCompress: false,
      create: DEFAULT_VALUES.CREATE,
      missingOk: false,
      notIfEmpty: false,
      preRotate: '',
      postRotate: '',
      linked: item.linked || false,
      createdAt: item.mod_time,
      updatedAt: item.mod_time,
      content: item.content || '',
    })
  );

  return {
    items,
    total: response.total,
    page: params.page,
    page_size: params.page_size,
  };
};

/**
 * 获取logrotate配置详情
 * @param type - logrotate类型
 * @param category - 分类
 * @param name - 配置名称
 * @param host - 主机ID
 * @returns logrotate配置详情
 */
export const getLogrotateDetailApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  host: number
): Promise<LogrotateEntity> => {
  const response = await request.get<ServiceForm>('logrotate/{host}/form', {
    type,
    category,
    name,
    host,
  });

  return convertServiceFormToLogrotateEntity(
    name,
    type,
    category,
    response.fields
  );
};

/**
 * 创建logrotate配置
 * @param data - 配置数据
 * @param host - 主机ID
 */
export const createLogrotateApi = async (
  data: LogrotateForm,
  host: number
): Promise<void> => {
  const requestData: CreateServiceForm = {
    name: data.name,
    type: data.type,
    category: data.category,
    form: convertLogrotateFormFieldsToKeyValue(data.form),
  };

  await request.post('logrotate/{host}/form', { ...requestData, host });
};

/**
 * 更新logrotate配置
 * @param type - logrotate类型
 * @param category - 分类
 * @param name - 配置名称
 * @param data - 更新数据
 * @param host - 主机ID
 */
export const updateLogrotateApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  data: Record<string, any>,
  host: number
): Promise<void> => {
  // 使用Record<string, any>类型来支持动态添加字段
  const requestData: Record<string, any> = {
    type,
    category,
    name,
    form: data.form,
  };

  // 添加new_name字段(如果存在)
  if (data.new_name) {
    requestData.new_name = data.new_name;
  } else {
    requestData.new_name = name; // 后端要求此字段，未更改时使用原名
  }

  // 添加new_category字段(如果存在)
  if (data.new_category) {
    requestData.new_category = data.new_category;
  } else {
    requestData.new_category = category; // 后端要求此字段，未更改时使用原分类
  }

  await request.put('logrotate/{host}/form', { ...requestData, host });
};

/**
 * 删除logrotate配置
 * @param type - logrotate类型
 * @param category - 分类
 * @param name - 配置名称
 * @param host - 主机ID
 */
export const deleteLogrotateApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  host: number
): Promise<void> => {
  await request.delete('logrotate/{host}', {
    type,
    category,
    name,
    host,
  });
};

/**
 * 激活/停用logrotate配置
 * @param type - logrotate类型
 * @param category - 分类
 * @param name - 配置名称
 * @param action - 操作类型（activate/deactivate）
 * @param host - 主机ID
 */
export const activateLogrotateApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  action: 'activate' | 'deactivate',
  host: number
): Promise<void> => {
  const requestData: ServiceActivate = {
    type,
    category,
    name,
    action,
  };

  await request.post('logrotate/{host}/activate', { ...requestData, host });
};

export const getLogrotateContentApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  host: number
): Promise<string> => {
  return request.get<string>('logrotate/{host}/raw', {
    type,
    category,
    name,
    host,
  });
};

export const createLogrotateRawApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  content: string,
  host: number
): Promise<void> => {
  const requestData: CreateLogrotateFile = {
    type,
    category,
    name,
    content,
  };

  await request.post('logrotate/{host}/raw', { ...requestData, host });
};

export const updateLogrotateContentApi = async (
  type: LOGROTATE_TYPE,
  category: string,
  name: string,
  content: string,
  host: number
): Promise<void> => {
  const requestData: UpdateLogrotateFile = {
    type,
    category,
    name,
    content,
  };

  await request.put('logrotate/{host}/raw', { ...requestData, host });
};

export const getLogrotateHistoryApi = async (
  params: LogrotateHistoryParams
): Promise<ApiListResult<LogrotateHistory>> => {
  const response = await request.get<PageResult<LogrotateHistoryResponse>>(
    'logrotate/{host}/log',
    {
      type: params.type,
      category: params.category,
      name: params.name,
      page: params.page,
      page_size: params.pageSize,
      host: params.host,
    }
  );

  // 转换后端数据为前端格式
  const items: LogrotateHistory[] = response.items.map(
    (item: LogrotateHistoryResponse) => ({
      id: item.commit_hash,
      commit: item.commit_hash,
      message: item.message,
      author: item.author,
      date: item.date,
    })
  );

  return {
    items,
    total: response.total,
    page: params.page,
    page_size: params.pageSize,
  };
};

export const getLogrotateHistoryDiffApi = async (
  params: LogrotateHistoryDiffParams
): Promise<string> => {
  return request.get<string>('logrotate/{host}/diff', {
    type: params.type,
    category: params.category,
    name: params.name,
    commit: params.commit,
    host: params.host,
  });
};

export const restoreLogrotateApi = async (
  params: LogrotateRestoreParams
): Promise<void> => {
  const requestData: RestoreLogrotateFile = {
    type: params.type,
    category: params.category,
    name: params.name,
    commit_hash: params.commit,
  };

  await request.put('logrotate/{host}/restore', {
    ...requestData,
    host: params.host,
  });
};

export const getLogrotateCategoriesApi = async (
  type: LOGROTATE_TYPE,
  page: number,
  pageSize: number,
  host: number
): Promise<ApiListResult<LogrotateCategory>> => {
  const response = await request.get<PageResult<LogrotateCategoryResponse>>(
    'logrotate/{host}/category',
    {
      type,
      page,
      page_size: pageSize,
      host,
    }
  );

  // 转换后端数据为前端格式
  // 分类 API 返回的是目录列表（GitFile 结构）
  const items: LogrotateCategory[] = response.items.map(
    (item: LogrotateCategoryResponse) => ({
      name: item.name,
      type,
      count: 0, // 目录的 size 字段不代表文件数量，这里设为 0
    })
  );

  return {
    items,
    total: response.total,
    page,
    page_size: pageSize,
  };
};

export const createLogrotateCategoryApi = async (
  type: LOGROTATE_TYPE,
  name: string,
  host: number
): Promise<void> => {
  const requestData: CreateLogrotateCategory = {
    type,
    category: name,
  };

  await request.post('logrotate/{host}/category', { ...requestData, host });
};

export const updateLogrotateCategoryApi = async (
  type: LOGROTATE_TYPE,
  oldName: string,
  newName: string,
  host: number
): Promise<void> => {
  const requestData: UpdateLogrotateCategory = {
    type,
    category: oldName,
    new_name: newName,
  };

  await request.put('logrotate/{host}/category', { ...requestData, host });
};

export const deleteLogrotateCategoryApi = async (
  type: LOGROTATE_TYPE,
  name: string,
  host: number
): Promise<void> => {
  await request.delete('logrotate/{host}/category', {
    type,
    category: name,
    host,
  });
};

export const syncLogrotateApi = async (host: number): Promise<void> => {
  await request.post('logrotate/{host}/sync', { host });
};

// 获取插件信息
export const getLogrotatePluginInfoApi =
  async (): Promise<LogrotatePluginInfo> => {
    return request.get('logrotate/info');
  };

// 获取插件菜单
export const getLogrotatePluginMenuApi = async (): Promise<
  LogrotatePluginMenu[]
> => {
  return request.get('logrotate/menu');
};
