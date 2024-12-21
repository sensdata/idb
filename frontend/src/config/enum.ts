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
