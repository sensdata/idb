import { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';

export interface LogrotateEntity {
  id?: string;
  name: string;
  type: LOGROTATE_TYPE;
  category: string;
  path: string;
  frequency: LOGROTATE_FREQUENCY;
  count: string;
  compress: boolean;
  delayCompress: boolean;
  create: string;
  missingOk: boolean;
  notIfEmpty: boolean;
  preRotate: string;
  postRotate: string;
  linked: boolean;
  createdAt?: string;
  updatedAt?: string;
  content?: string;
}

export interface LogrotateFormField {
  key: string;
  value: string;
}

export interface LogrotateForm {
  name: string;
  type: LOGROTATE_TYPE;
  category: string;
  form: LogrotateFormField[];
}

export interface LogrotateCategory {
  name: string;
  type: LOGROTATE_TYPE;
  count?: number;
}

export interface LogrotateHistory {
  id: string;
  commit: string;
  message: string;
  author: string;
  date: string;
}

export interface LogrotateListParams {
  type: LOGROTATE_TYPE;
  category: string;
  page: number;
  page_size: number;
  host?: number;
}
