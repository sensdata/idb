import { ref, toRaw, watch, isRef, type Ref } from 'vue';
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
  loading?: boolean | Ref<boolean | undefined>;
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
  // 支持传入 Ref<boolean>，以便父组件的 loading 变化被追踪
  watch(
    () => (isRef(options.loading) ? options.loading.value : options.loading),
    (val) => {
      if (val !== undefined) {
        setLoading(!!val);
      }
    },
    { immediate: true }
  );

  const renderData = ref<BaseEntity[]>([]);
  const summaryData = ref<Record<string, any>>();

  const setData = (data: ApiListResult<any>) => {
    renderData.value = data.items || [];
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

    // 注意：不要在此处强制关闭 loading。
    // 当外部通过 props.loading 控制加载状态（例如父组件在请求中），
    // 在这里关闭会导致加载动效过早消失并显示“暂无数据”。
    // 加载流程由：
    // - 内部请求：load() 的 finally 中关闭
    // - 外部数据源：由父组件传入的 loading 控制
  };

  const load = async (newParams?: Partial<ApiListParams>) => {
    if (!fetch) {
      return;
    }

    // 防止重复调用：如果已经在加载中且没有新参数，则跳过
    // 但如果有新参数，允许重新加载（这对于初始化很重要）
    if (loading.value && !newParams) {
      logDebug('🔍 load function called but already loading, skipping:', {
        hasNewParams: !!newParams,
        newParams,
        currentLoading: loading.value,
        timestamp: new Date().toISOString(),
      });
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
      Message.error({
        content: t('components.idbTable.error.loadFailed', {
          error: errorMessage,
        }),
        duration: 5000,
      });
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
