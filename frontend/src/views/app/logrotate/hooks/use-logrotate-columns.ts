import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { LogrotateEntity } from '@/entity/Logrotate';
import { COLUMN_WIDTHS } from '../constants';

export function useLogrotateColumns() {
  const { t } = useI18n();

  // 格式化日期
  const formatDate = (dateString: string | undefined): string => {
    if (!dateString) return '-';
    try {
      const date = new Date(dateString);
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    } catch (error) {
      return dateString;
    }
  };

  // 格式化轮转数量，去掉 "rotate " 前缀
  const formatCount = (count: string | undefined): string => {
    if (!count) return '';
    return count.replace(/^rotate\s+/, '');
  };

  // 表格列定义
  const columns = computed(() => [
    {
      title: t('app.logrotate.list.column.name'),
      dataIndex: 'name',
      width: COLUMN_WIDTHS.NAME,
    },
    {
      title: t('app.logrotate.list.column.path'),
      dataIndex: 'path',
      width: COLUMN_WIDTHS.PATH,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('app.logrotate.list.column.frequency'),
      dataIndex: 'frequency',
      width: COLUMN_WIDTHS.FREQUENCY,
      slotName: 'frequency',
    },
    {
      title: t('app.logrotate.list.column.count'),
      dataIndex: 'count',
      width: COLUMN_WIDTHS.COUNT,
      slotName: 'count',
    },
    {
      title: t('app.logrotate.list.column.status'),
      slotName: 'status',
      width: COLUMN_WIDTHS.STATUS,
    },
    {
      title: t('app.logrotate.list.column.updated_at'),
      dataIndex: 'updatedAt',
      width: COLUMN_WIDTHS.UPDATED_AT,
      render: ({ record }: { record: LogrotateEntity }) =>
        formatDate(record.updatedAt),
    },
    {
      title: t('common.operation'),
      dataIndex: 'operation',
      slotName: 'operation',
      width: COLUMN_WIDTHS.OPERATION,
      fixed: 'right' as const,
      align: 'center' as const,
    },
  ]);

  return {
    columns,
    formatDate,
    formatCount,
  };
}
