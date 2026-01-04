import { PeriodDetailDo } from '@/entity/Crontab';
import { CRONTAB_PERIOD_TYPE } from '@/config/enum';

export const usePeriodUtils = () => {
  // 将周期详情转换为标准cron表达式
  const convertPeriodToCronExpression = (
    periodDetails: PeriodDetailDo[]
  ): string => {
    if (!periodDetails || periodDetails.length === 0) {
      return '* * * * *';
    }

    const period = periodDetails[0];

    let minute = '0';
    let hour = '0';
    let day = '*';
    const month = '*';
    let week = '*';

    switch (period.type) {
      case CRONTAB_PERIOD_TYPE.MONTHLY:
        minute = period.minute.toString();
        hour = period.hour.toString();
        day = period.day.toString();
        break;

      case CRONTAB_PERIOD_TYPE.WEEKLY:
        minute = period.minute.toString();
        hour = period.hour.toString();
        week = period.week.toString();
        break;

      case CRONTAB_PERIOD_TYPE.DAILY:
        minute = period.minute.toString();
        hour = period.hour.toString();
        break;

      case CRONTAB_PERIOD_TYPE.HOURLY:
        minute = period.minute.toString();
        hour = '*';
        break;

      case CRONTAB_PERIOD_TYPE.EVERY_N_DAYS:
        minute = period.minute.toString();
        hour = period.hour.toString();
        if (period.day === 1) {
          day = '*';
        } else {
          day = `*/${period.day}`;
        }
        break;

      case CRONTAB_PERIOD_TYPE.EVERY_N_HOURS:
        minute = period.minute.toString();
        if (period.hour === 1) {
          hour = '*';
        } else {
          hour = `*/${period.hour}`;
        }
        break;

      case CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES:
        if (period.minute === 1) {
          minute = '*';
        } else {
          minute = `*/${period.minute}`;
        }
        hour = '*';
        break;

      default:
        minute = '*';
        hour = '*';
        break;
    }

    return `${minute} ${hour} ${day} ${month} ${week}`;
  };

  // 解析cron表达式为周期详情
  const parseCronExpression = (
    cronExpression: string
  ): PeriodDetailDo | null => {
    if (!cronExpression) return null;

    const parts = cronExpression.trim().split(/\s+/);
    if (parts.length < 5) return null;

    const [minute, hour, day, month, week] = parts;

    const parseValue = (value: string): number | null => {
      if (value === '*') return null;
      if (value.includes('/')) {
        const valueParts = value.split('/');
        if (valueParts.length === 2 && valueParts[0] === '*') {
          return parseInt(valueParts[1], 10);
        }
      }
      const num = parseInt(value, 10);
      return Number.isNaN(num) ? null : num;
    };

    const periodDetail: PeriodDetailDo = {
      type: CRONTAB_PERIOD_TYPE.DAILY,
      day: 1,
      week: 1,
      hour: 0,
      minute: 0,
      second: 0,
    };

    const minuteValue = parseValue(minute);
    if (minuteValue !== null) {
      periodDetail.minute = minuteValue;
    }

    const hourValue = parseValue(hour);
    if (hourValue !== null) {
      periodDetail.hour = hourValue;
    }

    const dayValue = parseValue(day);
    if (dayValue !== null) {
      periodDetail.day = dayValue;
    }

    const weekValue = parseValue(week);
    if (weekValue !== null) {
      periodDetail.week = weekValue;
    }

    // 根据cron表达式格式确定周期类型
    // 优先检测包含 */N 的间隔类型，因为它们更具体
    if (
      minute.startsWith('*/') &&
      hour === '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      // 每隔N分钟: */N * * * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES;
    } else if (
      minute !== '*' &&
      hour.startsWith('*/') &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      // 每隔N小时: M */N * * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_HOURS;
    } else if (
      minute !== '*' &&
      hour !== '*' &&
      day.startsWith('*/') &&
      month === '*' &&
      week === '*'
    ) {
      // 每隔N天: M H */N * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_DAYS;
    } else if (
      minute !== '*' &&
      hour !== '*' &&
      day === '*' &&
      month === '*' &&
      week !== '*'
    ) {
      // 每周: M H * * W
      periodDetail.type = CRONTAB_PERIOD_TYPE.WEEKLY;
    } else if (
      minute !== '*' &&
      hour !== '*' &&
      day !== '*' &&
      month === '*' &&
      week === '*'
    ) {
      // 每月: M H D * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.MONTHLY;
    } else if (
      minute !== '*' &&
      hour !== '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      // 每天: M H * * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.DAILY;
    } else if (
      minute !== '*' &&
      hour === '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      // 每小时: M * * * *
      periodDetail.type = CRONTAB_PERIOD_TYPE.HOURLY;
    }

    return periodDetail;
  };

  return {
    convertPeriodToCronExpression,
    parseCronExpression,
  };
};
