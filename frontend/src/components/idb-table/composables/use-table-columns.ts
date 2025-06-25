import { computed, shallowRef, watch } from 'vue';
import type { Column, SizeProps } from '../types';

export function useTableColumns(columns: any) {
  // 列设置 - 使用shallowRef优化性能
  const cloneColumns = shallowRef<Column[]>([]);
  const showColumns = shallowRef<Column[]>([]);
  const size = shallowRef<SizeProps>('medium');

  const visibleColumns = computed(() => {
    const checkedDataIndex = showColumns.value
      .filter((item) => item.checked)
      .map((item) => item.dataIndex);

    return checkedDataIndex.map((dataIndex) =>
      cloneColumns.value.find((item) => item.dataIndex === dataIndex)
    );
  });

  // 优化列处理，减少不必要的深拷贝
  watch(
    columns,
    (val) => {
      // 只在必要时进行深拷贝
      const newColumns = val.map((column: any) => ({
        ...column,
        checked: true,
      }));
      cloneColumns.value = newColumns;
      showColumns.value = [...newColumns];
    },
    { deep: true, immediate: true }
  );

  const handleToggleColumn = (checked: boolean, column: Column) => {
    // 找到对应的列并更新checked状态
    const targetColumn = showColumns.value.find(
      (item) => item.dataIndex === column.dataIndex
    );
    if (targetColumn) {
      targetColumn.checked = checked;
      // 触发响应式更新
      showColumns.value = [...showColumns.value];
    }
  };

  const handleColumnsReordered = (newColumns: Column[]) => {
    showColumns.value = newColumns;
  };

  const handleSelectDensity = (val: SizeProps) => {
    size.value = val;
  };

  return {
    cloneColumns,
    showColumns,
    visibleColumns,
    size,
    handleToggleColumn,
    handleColumnsReordered,
    handleSelectDensity,
  };
}
