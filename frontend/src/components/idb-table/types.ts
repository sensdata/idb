import { ComponentProps } from '@/types/utils';
import { BaseEntity, ApiListParams, ApiListResult } from '@/types/global';
import {
  SelectOptionData,
  SelectOptionGroup,
  TableColumnData,
  Table,
} from '@arco-design/web-vue';

type TableProps = ComponentProps<typeof Table>;

export interface CreateButton {
  text?: string;
  onClick: () => void;
}

export type DownloadFn = (params: Record<string, any>) => Promise<any>;

type FilterType =
  | 'input'
  | 'input-number'
  | 'select'
  | 'range-picker'
  | 'date-picker'
  | 'month-picker'
  | 'component'; // 自定义组件

// 过滤项
export interface FilterItem {
  // 过滤项字段，同时也作为key，会作为默认的请求参数字段
  field: string;
  // 过滤项的label
  label: string;
  type: FilterType;
  // 自定义组件
  component?: any;
  options?: Array<SelectOptionData | SelectOptionGroup>;
  // 默认值
  defaultValue?: any;
  // 多选, 仅select需要
  multiple?: boolean;
  placeholder?: string;
  allowClear?: boolean;
  // 回调函数，用以将值转换为请求参数，eg. value => ({ [key]: value })
  toParams?: (value: any, options: FilterItem) => Record<string, any>;
  // 组件的其它配置项
  [key: string]: any;
}

export type SizeProps = 'mini' | 'small' | 'medium' | 'large';
export type Column = TableColumnData & { checked?: boolean };

export interface Props extends /* @vue-ignore */ TableProps {
  // 过滤项
  filters?: FilterItem[];
  // 过滤项label对齐方式
  filterLabelAlign?: 'left' | 'right';
  // 下载
  download?: DownloadFn;
  // 接口
  fetch?: <T extends BaseEntity>(
    params: ApiListParams
  ) => Promise<ApiListResult<T>>;
  // 每页条数
  pageSize?: number;
  // 请求参数
  params?: Record<string, any>;
  // 表格列
  columns: TableColumnData[];
  // 自动加载
  autoLoad?: boolean;
  // 是否有选择列(即批量操作)
  hasBatch?: boolean;
  // 搜索
  allowSearch?: boolean;
}
