import { CRONTAB_KIND, CRONTAB_PERIOD_TYPE, CRONTAB_TYPE } from '@/config/enum';
import { BaseEntity } from '@/types/global';

export interface PeriodDetailDo {
  type: CRONTAB_PERIOD_TYPE;
  week: number;
  day: number;
  hour: number;
  minute: number;
  second: number;
}

// 计划任务
export interface CrontabEntity extends BaseEntity {
  id: number;
  name: string;
  type: CRONTAB_TYPE;
  kind: CRONTAB_KIND;
  content: string;
  content_mode?: 'direct' | 'script';
  disabled: boolean;
  mark: string;
  period_expression: string;
  period_details: PeriodDetailDo[];
  last_run_time: string;
  create_time: string;
  mod_time: string;
  linked: boolean; // Whether the crontab is running or not
}
