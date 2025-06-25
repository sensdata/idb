import { ref } from 'vue';
import { BaseEntity } from '@/types/global';

export function useTableSelection(rowKey: string) {
  const selectedRows = ref<BaseEntity[]>([]);
  const selectedRowKeys = ref<number[]>([]);

  const onSelectedChange = (rows: BaseEntity[]) => {
    selectedRows.value = rows;
    selectedRowKeys.value = rows.map((item: any) => item[rowKey]) as number[];
  };

  const clearSelected = () => {
    selectedRows.value = [];
    selectedRowKeys.value = [];
  };

  const getSelectedRows = () => selectedRows.value;

  return {
    selectedRows,
    selectedRowKeys,
    onSelectedChange,
    clearSelected,
    getSelectedRows,
  };
}
