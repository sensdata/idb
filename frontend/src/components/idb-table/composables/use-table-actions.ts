import { ApiListParams } from '@/types/global';

interface UseTableActionsOptions {
  load: (params?: Partial<ApiListParams>) => Promise<void>;
  pagination: any;
  urlSync: boolean;
  updatePagination?: (page?: number, pageSize?: number) => void;
  emit: any;
}

export function useTableActions(options: UseTableActionsOptions) {
  const { load, pagination, urlSync, updatePagination, emit } = options;

  // 搜索
  const onSearch = (value: string) => {
    load({
      search: value,
      page: 1,
    });
    emit('search', value);
  };

  // 排序
  const onSortChange = (sortParams: {
    sort_field?: string;
    sort_order?: string;
  }) => {
    load(sortParams);
    emit('sortChange', sortParams);
  };

  // 翻页
  const onPageChange = (page: number) => {
    if (urlSync && updatePagination) {
      // URL同步模式下，更新URL并触发事件
      updatePagination(page);
      emit('pageChange', page);
      return;
    }

    // 非URL同步模式的原有逻辑
    pagination.current = page;
    load({ page });
    emit('pageChange', page);
  };

  const onPageSizeChange = (pageSize: number) => {
    if (urlSync && updatePagination) {
      // URL同步模式下，更新URL并触发事件
      updatePagination(1, pageSize);
      emit('pageSizeChange', pageSize);
      return;
    }

    // 非URL同步模式的原有逻辑
    pagination.pageSize = pageSize;
    pagination.current = 1; // 重置到第一页
    load({
      page: 1,
      page_size: pageSize,
    });
    emit('pageSizeChange', pageSize);
  };

  // 过滤
  const onFilter = (filterParams: any) => {
    load({
      ...filterParams,
      page: 1,
    });
    emit('filter', filterParams);
  };

  const onFilterReady = (filterParams: any, autoLoad: boolean) => {
    if (!autoLoad) {
      return;
    }
    load({
      ...filterParams,
      page: 1,
    });
    emit('filterReady', filterParams);
  };

  return {
    onSearch,
    onSortChange,
    onPageChange,
    onPageSizeChange,
    onFilter,
    onFilterReady,
  };
}
