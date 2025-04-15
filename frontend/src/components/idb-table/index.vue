<template>
  <table-filters
    :label-align="filterLabelAlign"
    :items="filters"
    @filter="onFilter"
    @init="onFilterReady"
  />
  <a-divider v-if="filters.length" style="margin-top: 0" />
  <a-row v-if="hasToolbar" align="center" style="margin-bottom: 16px">
    <a-col :span="12">
      <a-space>
        <slot name="leftActions" />
      </a-space>
    </a-col>
    <a-col
      :span="12"
      style="display: flex; align-items: center; justify-content: end"
    >
      <slot name="rightActions" />
      <a-input-search
        v-if="hasSearch"
        v-model="searchValue"
        class="w-[240px] mr-4"
        :placeholder="$t('components.idbTable.search.placeholder')"
        :loading="loading"
        allow-clear
        @clear="() => onSearch('')"
        @search="onSearch"
        @press-enter="onSearchEnter"
      />
      <a-button v-if="download">
        <template #icon>
          <icon-download />
        </template>
        {{ $t('components.idbTable.actions.download') }}
      </a-button>
      <a-tooltip :content="$t('components.idbTable.actions.refresh')">
        <div class="action-icon" @click="reload"
          ><icon-refresh size="18"
        /></div>
      </a-tooltip>
      <a-dropdown @select="handleSelectDensity">
        <a-tooltip :content="$t('components.idbTable.actions.density')">
          <div class="action-icon"><icon-line-height size="18" /></div>
        </a-tooltip>
        <template #content>
          <a-doption
            v-for="item in densityList"
            :key="item.value"
            :value="item.value"
            :class="{ active: item.value === size }"
          >
            <span>{{ item.name }}</span>
          </a-doption>
        </template>
      </a-dropdown>
      <a-tooltip :content="$t('components.idbTable.actions.columnSetting')">
        <a-popover
          trigger="click"
          position="br"
          @popup-visible-change="popupVisibleChange"
        >
          <div class="action-icon"><icon-settings size="18" /></div>
          <template #content>
            <div id="tableSetting">
              <div
                v-for="item in showColumns"
                :key="item.dataIndex"
                class="setting"
              >
                <div style="margin-right: 4px; cursor: move">
                  <icon-drag-arrow />
                </div>
                <div>
                  <a-checkbox
                    v-model="item.checked"
                    @change="handleToggleColumn($event, item as Column)"
                  >
                    {{ item.title }}
                  </a-checkbox>
                </div>
              </div>
            </div>
          </template>
        </a-popover>
      </a-tooltip>
    </a-col>
  </a-row>
  <a-table
    v-model:selectedKeys="selectedRowKeys"
    :row-key="rowKey"
    :loading="loading"
    :pagination="pagination"
    :columns="(visibleColumns as TableColumnData[])"
    :data="renderData"
    :summary="!summaryData ? false : () => [summaryData!]"
    :summary-text="t('components.idbTable.summaryText')"
    :size="size"
    :row-selection="hasBatch ? rowSelection : undefined"
    v-bind="$attrs"
    @expand="onExpand"
    @change="onChange"
    @page-change="onPageChange"
    @page-size-change="onPageSizeChange"
  >
    <template v-for="(_, key) in tableSlots" :key="key" #[key]="slotData">
      <slot :name="key" v-bind="slotData"></slot>
    </template>
  </a-table>
  <fixed-footer-bar v-if="hasBatch && selectedRows.length > 0">
    <template #left>
      <span>
        {{ $t('components.idbTable.batch.selectedPrefix') }}
        <strong class="selected-count">{{ selectedRows.length }}</strong>
        {{ $t('components.idbTable.batch.selectedSuffix') }}
      </span>
      <a-button type="text" class="cancel-selected" @click="cancelSelected">{{
        $t('components.idbTable.batch.cancelSelected')
      }}</a-button>
    </template>
    <template #right>
      <slot
        name="batch"
        :selected-rows="selectedRows"
        :selected-row-keys="selectedRowKeys"
      />
    </template>
  </fixed-footer-bar>
</template>

<script lang="ts" setup>
  import {
    computed,
    ref,
    reactive,
    toRaw,
    watch,
    nextTick,
    useSlots,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/hooks/loading';
  import { ApiListParams, ApiListResult, BaseEntity } from '@/types/global';
  import type {
    TableChangeExtra,
    TableColumnData,
    TableRowSelection,
  } from '@arco-design/web-vue/es/table/interface';
  import { cloneDeep, omit } from 'lodash';
  import Sortable from 'sortablejs';
  import TableFilters from './table-filters.vue';
  import { Props, Column, SizeProps } from './types';

  defineOptions({
    inheritAttrs: false,
  });

  defineSlots<{
    // 左侧操作按钮插槽
    leftActions: (props: any) => any;
    // 批量操作按钮插槽
    batch: <T extends BaseEntity>(props: {
      selectedRows: T[];
      selectedRowKeys?: number[];
    }) => any;
    // 右侧操作按钮插槽
    rightActions: (props: any) => any;
    [key: string]: (props: any) => any;
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
  });

  const { t } = useI18n();

  const rowKey = computed(() => props.rowKey ?? 'id');

  const slots = useSlots();
  const tableSlots = computed(() =>
    omit(slots, ['leftActions', 'batch', 'rightActions'])
  );

  // 列设置
  const cloneColumns = ref<Column[]>([]);
  const showColumns = ref<Column[]>([]);
  const visibleColumns = computed(() => {
    const checkedDataIndex = showColumns.value
      .filter((item) => item.checked)
      .map((item) => item.dataIndex);

    return checkedDataIndex.map((dataIndex) =>
      cloneColumns.value.find((item) => item.dataIndex === dataIndex)
    );
  });
  watch(
    () => props.columns,
    (val) => {
      cloneColumns.value = cloneDeep(val);
      cloneColumns.value.forEach((item, index) => {
        item.checked = true;
      });
      showColumns.value = cloneDeep(cloneColumns.value);
    },
    { deep: true, immediate: true }
  );
  const handleToggleColumn = (
    checked: boolean | (string | boolean | number)[],
    column: Column
  ) => {
    column.checked = checked as boolean;
  };
  const exchangeArray = <T extends Array<any>>(
    array: T,
    beforeIdx: number,
    newIdx: number,
    isDeep = false
  ): T => {
    const newArray = isDeep ? cloneDeep(array) : array;
    if (beforeIdx > -1 && newIdx > -1) {
      // 先替换后面的，然后拿到替换的结果替换前面的
      newArray.splice(
        beforeIdx,
        1,
        newArray.splice(newIdx, 1, newArray[beforeIdx]).pop()
      );
    }
    return newArray;
  };
  const popupVisibleChange = (val: boolean) => {
    if (val) {
      nextTick(() => {
        const el = document.getElementById('tableSetting') as HTMLElement;
        const sortable = new Sortable(el, {
          onEnd(e: any) {
            const { oldIndex, newIndex } = e;
            exchangeArray(showColumns.value, oldIndex, newIndex);
          },
        });
      });
    }
  };

  // 批量操作
  const rowSelection = reactive({
    type: 'checkbox' as TableRowSelection['type'],
    showCheckedAll: true,
    onlyCurrent: false,
  });
  const selectedRows = ref<BaseEntity[]>([]);
  const selectedRowKeys = ref<number[]>([]);
  watch(selectedRowKeys, (newValue, oldValue) => {
    const newAdd = newValue.filter((item) => !oldValue.includes(item));
    selectedRows.value = selectedRows.value
      .concat(
        renderData.value.filter((item: any) =>
          newAdd.includes(item[rowKey.value])
        )
      )
      .filter((item: any) =>
        selectedRowKeys.value.includes(item[rowKey.value])
      );

    emit('selectedChange', selectedRows.value);
  });
  const cancelSelected = () => {
    selectedRows.value = [];
    selectedRowKeys.value = [];
    emit('selectedChange', selectedRows.value);
  };

  // 调整密度
  const size = ref<SizeProps>('medium');
  const densityList = computed(() => [
    { name: t('components.idbTable.size.mini'), value: 'mini' },
    { name: t('components.idbTable.size.small'), value: 'small' },
    { name: t('components.idbTable.size.medium'), value: 'medium' },
    { name: t('components.idbTable.size.large'), value: 'large' },
  ]);
  const handleSelectDensity = (
    val: string | number | Record<string, any> | undefined,
    e: Event
  ) => {
    size.value = val as SizeProps;
  };

  // 数据加载和参数处理
  const { loading, setLoading } = useLoading(false);
  watch(
    () => props.loading,
    (val) => {
      if (val !== undefined) {
        setLoading(val);
      }
    },
    {
      immediate: true,
    }
  );

  const params = reactive<ApiListParams>({
    page: 1,
    page_size: props.pageSize,
  });
  watch(
    () => props.params,
    () => {
      Object.assign(params, props.params);
    },
    {
      deep: true,
      immediate: true,
    }
  );
  const pagination = reactive({
    current: 1,
    pageSize: props.pageSize,
    pageSizeOptions: [10, 20, 50, 100],
    total: 1,
    showTotal: true,
    showPageSize: true,
    showJumper: true,
  });
  const renderData = ref<BaseEntity[]>([]);
  const summaryData = ref<BaseEntity>();
  const load = async (newParams?: Partial<ApiListParams>) => {
    if (!props.fetch) {
      return;
    }
    setLoading(true);
    try {
      Object.assign(params, newParams);
      let rawParams = toRaw(params);
      if (props.beforeFetchHook) {
        rawParams = props.beforeFetchHook(rawParams);
      }
      let data = await props.fetch(rawParams);
      if (props.afterFetchHook) {
        data = props.afterFetchHook(data);
      }
      setData(data);
    } finally {
      setLoading(false);
    }
  };
  const setData = (data: ApiListResult<any>) => {
    renderData.value = data.items;
    if (data.amount) {
      (data.amount as any)[rowKey.value] = t('components.idbTable.summaryText');
    }
    summaryData.value = data.amount;
    if (data.total) {
      pagination.total = data.total;
    }
    if (data.page) {
      pagination.current = data.page;
    }
    if (data.page_size) {
      pagination.pageSize = data.page_size;
    }
  };
  const reload = () => {
    load();
    emit('reload');
  };

  const onChange = (_: any, extra: TableChangeExtra) => {
    if (extra.type === 'sorter') {
      const sortParams = {
        sort_field: extra.sorter?.field,
        sort_order: extra.sorter?.direction === 'descend' ? 'desc' : 'asc',
      };
      load(sortParams);
      emit('sortChange', {
        sort_field: extra.sorter?.field,
        sort_order: extra.sorter?.direction === 'descend' ? 'desc' : 'asc',
      });
    }
  };
  // 翻页
  const onPageChange = (page: number) => {
    pagination.current = page;
    load({
      page,
    });
    emit('pageChange', page);
  };
  const onPageSizeChange = (pageSize: number) => {
    pagination.pageSize = pageSize;
    load({
      page_size: pageSize,
    });
    emit('pageSizeChange', pageSize);
  };

  // 搜索
  const searchValue = ref('');
  const onSearch = (value: string) => {
    searchValue.value = value;
    load({
      search: value,
      page: 1,
    });
    emit('search', value);
  };
  const onSearchEnter = () => {
    onSearch(searchValue.value);
  };

  // 过滤
  const onFilter = (filterParams: any) => {
    load({
      ...filterParams,
      page: 1,
    });
    emit('filter', filterParams);
  };
  const onFilterReady = (filterParams: any) => {
    if (!props.autoLoad) {
      return;
    }
    load({
      ...filterParams,
      page: 1,
    });
    emit('filterReady', filterParams);
  };

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

  defineExpose({
    load,
    setData,
    reload,
    setLoading,
    clearSelected: cancelSelected,
    getSelectedRows: () => selectedRows.value,
  });
</script>

<script lang="ts">
  export default {
    name: 'IdbTable',
  };
</script>

<style scoped lang="less">
  :deep(.arco-table-th) {
    &:last-child {
      .arco-table-th-item-title {
        margin-left: 16px;
      }
    }
  }

  .action-icon {
    margin-left: 12px;
    cursor: pointer;
  }

  .active {
    color: #0960bd;
    background-color: #e3f4fc;
  }

  .setting {
    display: flex;
    align-items: center;
    min-width: 160px;
  }

  .selected-count {
    margin: 0 5px;
    color: rgb(var(--primary-6));
    font-weight: bold;
  }

  .cancel-selected {
    margin-left: 20px;
  }
</style>
