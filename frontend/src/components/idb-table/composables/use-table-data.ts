import { ref, toRaw, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import useLoading from '@/composables/loading';
import { useLogger } from '@/composables/use-logger';
import { ApiListParams, ApiListResult, BaseEntity } from '@/types/global';

interface UseTableDataOptions {
  fetch?: (params: ApiListParams) => Promise<ApiListResult<any>>;
  beforeFetchHook?: (params: ApiListParams) => ApiListParams;
  afterFetchHook?: (data: ApiListResult<any>) => Promise<ApiListResult<any>>;
  rowKey: string;
  loading?: boolean;
  pagination: any;
  params: any;
  urlSync: boolean;
  updatePagination?: (page?: number, pageSize?: number) => void;
}

export function useTableData(options: UseTableDataOptions) {
  const { t } = useI18n();
  const { logDebug, logError } = useLogger('TableData');
  const {
    fetch,
    beforeFetchHook,
    afterFetchHook,
    rowKey,
    pagination,
    params,
    urlSync,
    updatePagination,
  } = options;

  const { loading, setLoading } = useLoading(true);

  // 监听外部loading状态
  watch(
    () => options.loading,
    (val) => {
      if (val !== undefined) {
        setLoading(val);
      }
    },
    { immediate: true }
  );

  const renderData = ref<BaseEntity[]>([]);
  const summaryData = ref<Record<string, any>>();

  const setData = (data: ApiListResult<any>) => {
    renderData.value = data.items;
    if (data.amount) {
      (data.amount as any)[rowKey] = t('components.idbTable.summaryText');
    }
    summaryData.value = data.amount;
    if (data.total) {
      pagination.total = data.total;
    }
    // 服务器返回的分页信息是准确的，应该信任并使用
    if (data.page) {
      pagination.current = data.page;
    }
    if (data.page_size) {
      pagination.pageSize = data.page_size;
    }

    // 设置数据后关闭loading状态
    setLoading(false);
  };

  const load = async (newParams?: Partial<ApiListParams>) => {
    if (!fetch) {
      return;
    }

    logDebug('🔍 load function called:', {
      hasNewParams: !!newParams,
      newParams,
      currentLoading: loading.value,
      timestamp: new Date().toISOString(),
    });

    setLoading(true);
    logDebug('🔍 setLoading(true) called, loading state:', loading.value);

    try {
      // 合并新参数
      if (newParams) {
        Object.assign(params, newParams);

        // 处理分页参数
        if (
          urlSync &&
          updatePagination &&
          (newParams.page !== undefined || newParams.page_size !== undefined)
        ) {
          updatePagination(newParams.page, newParams.page_size);
        } else if (!urlSync) {
          // 非URL同步模式下直接更新分页UI
          if (newParams.page !== undefined) {
            pagination.current = newParams.page;
          }
          if (newParams.page_size !== undefined) {
            pagination.pageSize = newParams.page_size;
          }
        }
      }

      let rawParams = toRaw(params);
      if (beforeFetchHook) {
        rawParams = beforeFetchHook(rawParams);
      }

      logDebug('🔍 calling fetch with params:', rawParams);
      let data = await fetch(rawParams);
      if (afterFetchHook) {
        data = await afterFetchHook(data);
      }
      logDebug('🔍 fetch completed, setting data');
      setData(data);
    } catch (error) {
      // 错误处理
      const errorMessage =
        error instanceof Error ? error.message : String(error);

      // 如果错误消息是"OK"，说明可能是参数问题导致的，不显示错误弹窗
      if (errorMessage !== 'OK') {
        Message.error({
          content: t('components.idbTable.error.loadFailed', {
            error: errorMessage,
          }),
          duration: 5000,
        });
      }

      // 保持现有数据，不清空
      renderData.value = renderData.value || [];
      logError('🔍 fetch failed:', error);
    } finally {
      setLoading(false);
      logDebug('🔍 setLoading(false) called, loading state:', loading.value);
    }
  };

  const reload = () => {
    load();
  };

  return {
    loading,
    setLoading,
    renderData,
    summaryData,
    load,
    setData,
    reload,
  };
}
