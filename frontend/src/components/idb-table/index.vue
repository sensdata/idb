<template>
  <table-filters
    :label-align="filterLabelAlign"
    :items="filters"
    @filter="onFilter"
    @init="onFilterReady"
  />
  <a-divider v-if="filters.length" style="margin-top: 0" />

  <!-- 表格工具栏 -->
  <table-toolbar
    :has-toolbar="hasToolbar"
    :has-search="hasSearch"
    :loading="loading"
    :download="download"
    :size="size"
    :show-columns="showColumns"
    @search="onSearch"
    @refresh="reload"
    @toggle-column="handleToggleColumn"
    @select-density="handleSelectDensity"
    @columns-reordered="handleColumnsReordered"
  >
    <template #leftActions>
      <slot name="leftActions" />
    </template>
    <template #rightActions>
      <slot name="rightActions" />
    </template>
  </table-toolbar>

  <!-- 表格内容 -->
  <table-content
    ref="tableContentRef"
    :row-key="rowKey"
    :loading="loading"
    :pagination="pagination"
    :columns="(visibleColumns as TableColumnData[])"
    :data="renderData"
    :summary-data="summaryData"
    :size="size"
    :has-batch="hasBatch"
    v-bind="$attrs"
    @expand="onExpand"
    @page-change="onPageChange"
    @page-size-change="onPageSizeChange"
    @selected-change="onSelectedChange"
    @sort-change="onSortChange"
  >
    <template v-for="(_, key) in tableSlots" :key="key" #[key]="slotData">
      <slot :name="key" v-bind="slotData"></slot>
    </template>
  </table-content>

  <!-- 批量操作栏 -->
  <batch-action-bar
    :has-batch="hasBatch"
    :selected-rows="selectedRows"
    :selected-row-keys="selectedRowKeys"
    @cancel-selected="cancelSelected"
  >
    <template #batch="slotProps">
      <slot
        name="batch"
        :selected-rows="slotProps.selectedRows"
        :selected-row-keys="slotProps.selectedRowKeys"
      />
    </template>
  </batch-action-bar>
</template>

<script lang="ts" setup>
  import { computed, ref, reactive, watch, useSlots, onMounted } from 'vue';
  import { omit } from 'lodash';
  import { ApiListParams, BaseEntity } from '@/types/global';
  import type { TableColumnData } from '@arco-design/web-vue/es/table/interface';
  import TableFilters from './table-filters.vue';
  import TableToolbar from './components/table-toolbar.vue';
  import TableContent from './components/table-content.vue';
  import BatchActionBar from './components/batch-action-bar.vue';
  import useUrlPaginationSync from './composables/use-url-pagination-sync';
  import { useTableColumns } from './composables/use-table-columns';
  import { useTableData } from './composables/use-table-data';
  import { useTableSelection } from './composables/use-table-selection';
  import { useTableActions } from './composables/use-table-actions';
  import { Props } from './types';

  defineOptions({
    inheritAttrs: false,
  });

  defineSlots<{
    // 左侧操作按钮插槽
    leftActions?: () => any;
    // 批量操作按钮插槽
    batch?: (props: {
      selectedRows: BaseEntity[];
      selectedRowKeys?: number[];
    }) => any;
    // 右侧操作按钮插槽
    rightActions?: () => any;
    // 表格列插槽 - 动态插槽用于表格单元格渲染
    [key: string]: ((props?: any) => any) | undefined;
  }>();

  const emit = defineEmits([
    'selectedChange',
    'filter',
    'filterReady',
    'reload',
    'search',
    'pageChange',
    'pageSizeChange',
    'sortChange',
  ]);

  const props = withDefaults(defineProps<Props>(), {
    filters: () => [],
    pageSize: 20,
    hasSearch: false,
    hasToolbar: true,
    autoLoad: true,
    urlSync: false,
    urlParamNames: () => ({ page: 'page', pageSize: 'pageSize' }),
  });

  const tableContentRef = ref<InstanceType<typeof TableContent>>();

  const rowKey = computed(() => props.rowKey ?? 'id');

  const slots = useSlots();
  const tableSlots = computed(() =>
    omit(slots, ['leftActions', 'batch', 'rightActions'])
  );

  // 使用URL同步composable
  const {
    pagination: urlPagination,
    params,
    updatePagination,
  } = props.urlSync
    ? useUrlPaginationSync({
        pageSize: props.pageSize,
        urlParamNames: props.urlParamNames,
      })
    : {
        pagination: reactive({ current: 1, pageSize: props.pageSize }),
        params: reactive<ApiListParams>({ page: 1, page_size: props.pageSize }),
        updatePagination: () => params,
      };

  // 完整分页配置
  const pagination = reactive({
    ...urlPagination,
    pageSizeOptions: [10, 20, 50, 100],
    total: 1,
    showTotal: true,
    showPageSize: true,
    showJumper: true,
  });

  // 使用列管理 composable
  const {
    showColumns,
    visibleColumns,
    size,
    handleToggleColumn,
    handleColumnsReordered,
    handleSelectDensity,
  } = useTableColumns(() => props.columns);

  // 使用数据管理 composable
  const {
    loading,
    setLoading,
    renderData,
    summaryData,
    load,
    setData,
    reload: reloadData,
  } = useTableData({
    fetch: props.fetch,
    beforeFetchHook: props.beforeFetchHook,
    afterFetchHook: props.afterFetchHook,
    rowKey: rowKey.value,
    loading: props.loading,
    pagination,
    params,
    urlSync: props.urlSync,
    updatePagination,
  });

  // 使用选择管理 composable
  const {
    selectedRows,
    selectedRowKeys,
    onSelectedChange: handleSelectedChange,
    getSelectedRows,
  } = useTableSelection(rowKey.value);

  // 使用操作管理 composable
  const {
    onSearch,
    onSortChange,
    onPageChange,
    onPageSizeChange,
    onFilter,
    onFilterReady: handleFilterReady,
  } = useTableActions({
    load,
    pagination,
    urlSync: props.urlSync,
    updatePagination,
    emit,
  });

  // 事件处理器
  const reload = () => {
    reloadData();
    emit('reload');
  };

  const onSelectedChange = (rows: any[]) => {
    handleSelectedChange(rows);
    emit('selectedChange', selectedRows.value);
  };

  const cancelSelected = () => {
    if (tableContentRef.value) {
      tableContentRef.value.clearSelected();
    }
  };

  const onExpand = (rowKey: string | number, record: any) => {
    if (props.onExpand) {
      props.onExpand(rowKey, record);
    }
  };

  const onFilterReady = (filterParams: any) => {
    handleFilterReady(filterParams, props.autoLoad);
  };

  // 监听参数变化
  watch(
    () => props.params,
    (newPropsParams: ApiListParams | undefined) => {
      if (!newPropsParams) return;

      if (props.urlSync) {
        // URL同步模式下，接受父组件传入的分页参数
        // 但需要同步到UI状态
        Object.assign(params, newPropsParams);
        if (newPropsParams.page) {
          pagination.current = newPropsParams.page;
        }
        if (newPropsParams.page_size) {
          pagination.pageSize = newPropsParams.page_size;
        }
      } else {
        // 非URL同步模式下，分页参数由组件内部管理
        const { page, page_size: pageSize, ...otherParams } = newPropsParams;
        Object.assign(params, otherParams);
      }
    },
    {
      deep: true,
      // URL同步模式下不使用immediate，避免与父组件的显式调用产生冲突
      immediate: !props.urlSync,
    }
  );

  // 监听数据源变化
  watch(
    () => props.dataSource,
    (val) => {
      if (val) {
        setData(val);
      }
    },
    {
      immediate: true,
    }
  );

  // 组件挂载时初始化
  onMounted(() => {
    // 启用了URL同步的表格，等待父组件明确调用load
    // 这样可以确保路径导航完成后再加载数据
    if (props.urlSync) {
      return;
    }

    // 正常的自动加载逻辑（非URL同步模式）
    if (props.autoLoad) {
      load();
    }
  });

  defineExpose({
    load,
    setData,
    reload,
    setLoading,
    clearSelected: cancelSelected,
    getSelectedRows,
    getData: () => renderData.value,
  });
</script>

<script lang="ts">
  export default {
    name: 'IdbTable',
  };
</script>
