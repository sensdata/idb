export interface AnyObject {
  [key: string]: unknown;
}

export interface Options {
  value: unknown;
  label: string;
}

export interface NodeOptions extends Options {
  children?: NodeOptions[];
}

export interface GetParams {
  body: null;
  type: string;
  url: string;
}

export interface PostData {
  body: string;
  type: string;
  url: string;
}

export interface Pagination {
  current: number;
  pageSize: number;
  total?: number;
}

export type TimeRanger = [string, string];

export interface GeneralChart {
  xAxis: string[];
  data: Array<{ name: string; value: number[] }>;
}

export interface ApiListParams {
  page?: number;
  page_size?: number;
  [key: string]: any;
}

export interface ApiListResult<T> {
  page: number;
  page_size: number;
  total?: number;
  items: T[];
  amount?: T;
}

export interface BaseEntity {
  id: number;
}
