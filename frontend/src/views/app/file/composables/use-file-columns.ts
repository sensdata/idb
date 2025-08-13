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
      align: 'left' as const,
    },
    {
      dataIndex: 'size',
      title: t('app.file.list.column.size'),
      width: 80,
      align: 'left' as const,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatFileSize(record.size);
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.file.list.column.mod_time'),
      width: 140,
      align: 'left' as const,
      render: ({ record }: { record: FileInfoEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      dataIndex: 'mode',
      title: t('app.file.list.column.mode'),
      width: 80,
      slotName: 'mode',
      align: 'left' as const,
    },
    {
      dataIndex: 'user',
      title: t('app.file.list.column.user'),
      width: 80,
      slotName: 'user',
      align: 'left' as const,
    },
    {
      dataIndex: 'group',
      title: t('app.file.list.column.group'),
      width: 80,
      slotName: 'group',
      align: 'left' as const,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 80,
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  return { columns };
};
