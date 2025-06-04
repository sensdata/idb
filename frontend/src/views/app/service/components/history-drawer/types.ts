import { SERVICE_TYPE } from '@/config/enum';

export interface HistoryParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
}

export interface PaginationConfig {
  current: number;
  pageSize: number;
  total: number;
  showTotal: boolean;
  showPageSize: boolean;
}

export interface HistoryDrawerExpose {
  show: (params: HistoryParams) => void;
}

export interface RestoreParams extends HistoryParams {
  commit: string;
  host: string;
}

export interface HistoryApiParams extends HistoryParams {
  page: number;
  pageSize: number;
  host: string;
}
