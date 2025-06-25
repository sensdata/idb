<template>
  <a-table
    v-model:selectedKeys="selectedKeys"
    :row-key="rowKey"
    :loading="loading"
    :pagination="pagination"
    :columns="columns"
    :data="data"
    :summary="hasSummary"
    :summary-text="t('components.idbTable.summaryText')"
    :size="size"
    :row-selection="hasBatch ? rowSelection : undefined"
    v-bind="$attrs"
    @expand="onExpand"
    @change="onChange"
    @page-change="onPageChange"
    @page-size-change="onPageSizeChange"
  >
    <template v-for="(_, key) in $slots" :key="key" #[key]="slotData">
      <slot :name="key" v-bind="slotData"></slot>
    </template>
  </a-table>
</template>

<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type {
    TableChangeExtra,
    TableColumnData,
    TableRowSelection,
    TableData,
  } from '@arco-design/web-vue/es/table/interface';
  import { BaseEntity } from '@/types/global';
  import { SizeProps } from '../types';

  defineOptions({
    name: 'TableContent',
    inheritAttrs: false,
  });

  const props = defineProps<{
    rowKey: string;
    loading: boolean;
    pagination: any; // 使用any类型避免类型错误
    columns: TableColumnData[];
    data: BaseEntity[];
    summaryData?: Record<string, any>;
    size: SizeProps;
    hasBatch?: boolean;
  }>();

  const emit = defineEmits<{
    expand: [rowKey: string | number, record: any];
    change: [data: any, extra: TableChangeExtra];
    pageChange: [page: number];
    pageSizeChange: [pageSize: number];
    selectedChange: [selectedRows: BaseEntity[]];
    sortChange: [sortParams: { sort_field?: string; sort_order?: string }];
  }>();

  const { t } = useI18n();

  // 处理汇总行
  const hasSummary = computed(() => {
    if (!props.summaryData) return false;
    return () => [props.summaryData as TableData];
  });

  // 批量选择
  const rowSelection = {
    type: 'checkbox' as TableRowSelection['type'],
    showCheckedAll: true,
    onlyCurrent: false,
  };

  const selectedKeys = ref<(string | number)[]>([]);
  const selectedRows = ref<BaseEntity[]>([]);

  // 监听选择变化
  watch(selectedKeys, (newValue, oldValue) => {
    const newAdd = newValue.filter((item) => !oldValue.includes(item));
    selectedRows.value = selectedRows.value
      .concat(
        props.data.filter((item: any) => newAdd.includes(item[props.rowKey]))
      )
      .filter((item: any) => selectedKeys.value.includes(item[props.rowKey]));

    emit('selectedChange', selectedRows.value);
  });

  // 表格事件处理
  const onChange = (_: any, extra: TableChangeExtra) => {
    emit('change', _, extra);
    if (extra.type === 'sorter') {
      emit('sortChange', {
        sort_field: extra.sorter?.field,
        sort_order: extra.sorter?.direction === 'descend' ? 'desc' : 'asc',
      });
    }
  };

  const onExpand = (rowKey: string | number, record: any) => {
    emit('expand', rowKey, record);
  };

  const onPageChange = (page: number) => {
    emit('pageChange', page);
  };

  const onPageSizeChange = (pageSize: number) => {
    emit('pageSizeChange', pageSize);
  };

  // 清除选择
  const clearSelected = () => {
    selectedKeys.value = [];
    selectedRows.value = [];
    emit('selectedChange', []);
  };

  // 暴露方法
  defineExpose({
    clearSelected,
    getSelectedRows: () => selectedRows.value,
    selectedKeys,
  });
</script>

<style scoped lang="less">
  :deep(.arco-table-th) {
    &:last-child {
      .arco-table-th-item-title {
        margin-left: 16px;
      }
    }
  }
</style>
