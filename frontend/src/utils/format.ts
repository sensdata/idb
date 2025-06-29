import { useI18n } from 'vue-i18n';
import { truncate } from 'lodash';

export function formatFileSize(
  size?: number,
  fixed = 1,
  useAbbrUnit = false
): string {
  if (size == null) {
    return '-';
  }

  if (size < 1024) {
    return size + 'B';
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(fixed) + (useAbbrUnit ? 'K' : 'KB');
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(fixed) + (useAbbrUnit ? 'M' : 'MB');
  }
  size /= 1024;
  return size.toFixed(fixed) + (useAbbrUnit ? 'G' : 'GB');
}

export const formatMemorySize = (size?: number, fixed = 1) =>
  formatFileSize(size, fixed, true);

export function formatBandwidth(bps?: number, fixed = 1): string {
  if (bps == null) {
    return '-';
  }

  if (bps < 1000) {
    return `${bps.toFixed(0)} bps`;
  }
  bps /= 1000;
  if (bps < 1000) {
    return `${bps.toFixed(fixed)} Kbps`;
  }
  bps /= 1000;
  if (bps < 1000) {
    return `${bps.toFixed(fixed)} Mbps`;
  }
  bps /= 1000;
  return `${bps.toFixed(fixed)} Gbps`;
}

export function formatTransferSpeed(bytesPerSecond?: number): string {
  if (bytesPerSecond == null) {
    return '-';
  }

  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond.toFixed(0)} B/s`;
  }
  bytesPerSecond /= 1024;
  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond.toFixed(bytesPerSecond >= 100 ? 0 : 1)} K/s`;
  }
  bytesPerSecond /= 1024;
  if (bytesPerSecond < 1024) {
    return `${bytesPerSecond.toFixed(bytesPerSecond >= 100 ? 0 : 1)} M/s`;
  }
  bytesPerSecond /= 1024;
  return `${bytesPerSecond.toFixed(bytesPerSecond >= 100 ? 0 : 1)} G/s`;
}

export function formatTime(time?: string | number): string {
  if (!time) {
    return '-';
  }
  if (
    typeof time === 'number' &&
    time.toString().length <= 11 &&
    time.toString().length >= 10
  ) {
    time *= 1e3;
  }
  const date = new Date(time);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }

  return date.toLocaleString();
}

export function formatTimeWithoutSeconds(time?: string | number): string {
  if (!time) {
    return '-';
  }
  if (
    typeof time === 'number' &&
    time.toString().length <= 11 &&
    time.toString().length >= 10
  ) {
    time *= 1e3;
  }
  const date = new Date(time);
  if (Number.isNaN(date.getTime())) {
    return '-';
  }

  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
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

/**
 * 格式化提交哈希，显示指定长度
 * @param commit - 完整的提交哈希
 * @param length - 要显示的长度，默认8位
 * @returns 格式化后的提交哈希
 */
export function formatCommitHash(commit: string, length = 8): string {
  if (!commit || typeof commit !== 'string') {
    return '';
  }
  return truncate(commit, { length, omission: '' });
}
