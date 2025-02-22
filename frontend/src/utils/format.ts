import { useI18n } from 'vue-i18n';

export function formatFileSize(size?: number, fixed = 1): string {
  if (size == null) {
    return '-';
  }

  if (size < 1024) {
    return size + 'B';
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(fixed) + 'KB';
  }
  size /= 1024;
  if (size < 1024) {
    return size.toFixed(fixed) + 'MB';
  }
  size /= 1024;
  return size.toFixed(fixed) + 'GB';
}

export const formatMemorySize = formatFileSize;

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
