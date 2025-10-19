import { useI18n } from 'vue-i18n';
import { PeriodDetailDo } from '@/entity/Crontab';
import { CRONTAB_PERIOD_TYPE } from '@/config/enum';

export const usePeriodUtils = () => {
  const { t } = useI18n();

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

  // 生成格式化的周期描述文本
  const generateFormattedPeriodComment = (
    periodDetails: PeriodDetailDo[]
  ): string => {
    if (!periodDetails || periodDetails.length === 0) {
      return (
        t('app.crontab.period.execution_period') +
        ': ' +
        t('app.crontab.period.description.daily', { time: '12:00 AM' })
      );
    }

    const formatTime = (hour: number, minute: number) => {
      const isPM = hour >= 12;
      const hour12 = hour % 12 || 12;
      return `${hour12}:${minute.toString().padStart(2, '0')} ${
        isPM ? 'PM' : 'AM'
      }`;
    };

    const weekDayKeys = [
      '',
      'app.crontab.enum.week.monday',
      'app.crontab.enum.week.tuesday',
      'app.crontab.enum.week.wednesday',
      'app.crontab.enum.week.thursday',
      'app.crontab.enum.week.friday',
      'app.crontab.enum.week.saturday',
      'app.crontab.enum.week.sunday',
    ];

    const period = periodDetails[0];
    let readablePeriodDesc = '';

    switch (period.type) {
      case CRONTAB_PERIOD_TYPE.MONTHLY:
        readablePeriodDesc = t('app.crontab.period.description.monthly', {
          day: period.day,
          time: formatTime(period.hour, period.minute),
        });
        break;
      case CRONTAB_PERIOD_TYPE.WEEKLY:
        readablePeriodDesc = t('app.crontab.period.description.weekly', {
          weekday: t(weekDayKeys[period.week]),
          time: formatTime(period.hour, period.minute),
        });
        break;
      case CRONTAB_PERIOD_TYPE.DAILY:
        readablePeriodDesc = t('app.crontab.period.description.daily', {
          time: formatTime(period.hour, period.minute),
        });
        break;
      case CRONTAB_PERIOD_TYPE.HOURLY:
        readablePeriodDesc = t('app.crontab.period.description.hourly', {
          minute: period.minute,
        });
        break;
      case CRONTAB_PERIOD_TYPE.EVERY_N_DAYS:
        readablePeriodDesc = t('app.crontab.period.description.every_n_days', {
          day: period.day,
          time: formatTime(period.hour, period.minute),
        });
        break;
      case CRONTAB_PERIOD_TYPE.EVERY_N_HOURS:
        readablePeriodDesc = t('app.crontab.period.description.every_n_hours', {
          hour: period.hour,
          minute: period.minute,
        });
        break;
      case CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES:
        readablePeriodDesc = t(
          'app.crontab.period.description.every_n_minutes',
          {
            minute: period.minute,
          }
        );
        break;
      default:
        readablePeriodDesc = '';
        break;
    }

    return t('app.crontab.period.execution_period') + ': ' + readablePeriodDesc;
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
    if (
      minute !== '*' &&
      hour !== '*' &&
      day !== '*' &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.MONTHLY;
    }
    if (
      minute !== '*' &&
      hour !== '*' &&
      day === '*' &&
      month === '*' &&
      week !== '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.WEEKLY;
    }
    if (
      minute !== '*' &&
      hour !== '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.DAILY;
    }
    if (
      minute !== '*' &&
      hour === '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.HOURLY;
    }
    if (
      minute !== '*' &&
      hour !== '*' &&
      day.startsWith('*/') &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_DAYS;
    }
    if (
      minute !== '*' &&
      hour.startsWith('*/') &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_HOURS;
    }
    if (
      minute.startsWith('*/') &&
      hour === '*' &&
      day === '*' &&
      month === '*' &&
      week === '*'
    ) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES;
    }
    if (weekValue !== null && week !== '*') {
      periodDetail.type = CRONTAB_PERIOD_TYPE.WEEKLY;
    } else if (dayValue !== null && day !== '*') {
      periodDetail.type = CRONTAB_PERIOD_TYPE.MONTHLY;
    } else if (hourValue !== null && minuteValue !== null) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.DAILY;
    } else if (minuteValue !== null) {
      periodDetail.type = CRONTAB_PERIOD_TYPE.HOURLY;
    }

    return periodDetail;
  };

  return {
    convertPeriodToCronExpression,
    generateFormattedPeriodComment,
    parseCronExpression,
  };
};
