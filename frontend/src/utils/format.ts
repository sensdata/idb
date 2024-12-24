import { useI18n } from 'vue-i18n';

export function formatFileSize(size?: number): string {
  if (size == null) {
    return '-';
  }

  if (size < 1024) {
    return size + 'B';
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(2) + 'KB';
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(2) + 'MB';
  }
  size /= 1024;
  return size.toFixed(2) + 'GB';
}

export function formatTime(time?: string): string {
  if (!time) {
    return '-';
  }
  const date = new Date(time);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }

  return date.toLocaleString();
}

export function formatSeconds(seconds?: number) {
  if (seconds == null) {
    return '-';
  }

  const { t } = useI18n();

  const days = Math.floor(seconds / (24 * 60 * 60));
  seconds %= 24 * 60 * 60;
  const hours = Math.floor(seconds / (60 * 60));
  seconds %= 60 * 60;
  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  let str = '';
  if (days > 0) {
    str += t('common.timeFormat.days', { days });
  }
  if (hours > 0 || str) {
    str += t('common.timeFormat.hours', { hours });
  }
  if (minutes > 0 || str) {
    str += t('common.timeFormat.minutes', { minutes });
  }
  str += t('common.timeFormat.seconds', { seconds: remainingSeconds });

  return str;
}
