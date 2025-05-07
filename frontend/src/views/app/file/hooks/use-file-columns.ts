import { formatFileSize, formatTime } from '@/utils/format';
import { FileInfoEntity } from '@/entity/FileInfo';

interface TableColumn {
  dataIndex: string;
  title: string;
  width?: number;
  ellipsis?: boolean;
  slotName?: string;
  render?: (data: { record: FileInfoEntity }) => string;
  align?: 'left' | 'center' | 'right';
}

export type TranslationFunction = (key: string) => string;

export const useFileColumns = (
  t: TranslationFunction
): { columns: TableColumn[] } => {
  const columns: TableColumn[] = [
    {
      dataIndex: 'name',
      title: t('app.file.list.column.name'),
      width: 400,
      ellipsis: true,
      slotName: 'name',
    },
    {
      dataIndex: 'size',
      title: t('app.file.list.column.size'),
      width: 80,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatFileSize(record.size);
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.file.list.column.mod_time'),
      width: 140,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      dataIndex: 'mode',
      title: t('app.file.list.column.mode'),
      width: 80,
      slotName: 'mode',
    },
    {
      dataIndex: 'user',
      title: t('app.file.list.column.user'),
      width: 80,
      slotName: 'user',
    },
    {
      dataIndex: 'group',
      title: t('app.file.list.column.group'),
      width: 80,
      slotName: 'group',
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 80,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  return { columns };
};
