export enum AUTH_MODE {
  Password = 'password',
  PrivateKey = 'privateKey',
}

export enum SCRIPT_TYPE {
  Local = 'local',
  Global = 'global',
}

export enum CRONTAB_TYPE {
  Local = 'local',
  Global = 'global',
}

export enum CRONTAB_KIND {
  Shell = 'shell',
}

export enum CRONTAB_PERIOD_TYPE {
  MONTHLY = 'MONTHLY', // 每月
  WEEKLY = 'WEEKLY', // 每周
  DAILY = 'DAILY', // 每天
  HOURLY = 'HOURLY', // 每小时
  EVERY_N_DAYS = 'EVERY_N_DAYS', // 每N天
  EVERY_N_HOURS = 'EVERY_N_HOURS', // 每N小时
  EVERY_N_MINUTES = 'EVERY_N_MINUTES', // 每N分钟
  EVERY_N_SECONDS = 'EVERY_N_SECONDS', // 每N秒
}

export enum TASK_STATUS {
  Created = 'created',
  Running = 'running',
  Success = 'success',
  Failed = 'failed',
  Canceled = 'canceled',
}

export enum LOGROTATE_TYPE {
  Local = 'local',
  Global = 'global',
}

export enum LOGROTATE_MODE {
  Form = 'form',
  Raw = 'raw',
}

export enum LOGROTATE_FREQUENCY {
  Daily = 'daily',
  Weekly = 'weekly',
  Monthly = 'monthly',
  Yearly = 'yearly',
}

export enum SERVICE_TYPE {
  Local = 'local',
  Global = 'global',
}

export enum SERVICE_MODE {
  Form = 'form',
  Raw = 'raw',
}

export enum SERVICE_OPERATION {
  Start = 'start',
  Stop = 'stop',
  Restart = 'restart',
  Enable = 'enable',
  Disable = 'disable',
  Reload = 'reload',
  Status = 'status',
}

export enum SERVICE_ACTION {
  Activate = 'activate',
  Deactivate = 'deactivate',
}

export enum COMPOSE_STATUS {
  Running = 'running', // 所有容器运行中
  Exited = 'exited', // 所有容器退出
  Partial = 'partial', // 混合状态（有的 exited，有的 running）
  Paused = 'paused', // 所有容器暂停
  Restarting = 'restarting', // 所有容器都在重启
  Removing = 'removing', // 所有容器都在被删除
  Dead = 'dead', // 所有容器都是 dead
  Mixed = 'mixed', // 多种状态混合（running + restarting + paused 等）
  Unknown = 'unknown', // 状态未知或无容器
}
