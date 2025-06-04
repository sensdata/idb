import { ref, reactive, computed } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { SERVICE_TYPE } from '@/config/enum';
import { ServiceHistoryEntity } from '@/entity/Service';
import { getServiceHistoryApi, restoreServiceApi } from '@/api/service';
import useLoading from '@/hooks/loading';
import { useConfirm } from '@/hooks/confirm';
import useCurrentHost from '@/hooks/current-host';
import { useLogger } from '@/hooks/use-logger';
import type { HistoryParams, PaginationConfig } from '../types';

// 常量定义
const DEFAULT_PAGE_SIZE = 10;
const DEFAULT_PAGE = 1;
const MIN_PAGE = 1;
const MIN_PAGE_SIZE = 1;
const COMMIT_HASH_DISPLAY_LENGTH = 8;

export function useHistoryData() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { currentHostId } = useCurrentHost();
  const { logError } = useLogger('HistoryData');

  const historyList = ref<ServiceHistoryEntity[]>([]);
  const currentParams = ref<HistoryParams>({
    type: SERVICE_TYPE.Local,
    category: '',
    name: '',
  });

  const pagination = reactive<PaginationConfig>({
    current: DEFAULT_PAGE,
    pageSize: DEFAULT_PAGE_SIZE,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  // 计算属性：检查是否有数据
  const hasData = computed(() => historyList.value.length > 0);

  // 计算属性：检查是否正在加载
  const isLoading = computed(() => loading.value);

  // 验证必要参数
  const validateParams = (): string | null => {
    const hostId = currentHostId.value;
    if (!hostId) {
      return t('app.service.history.validation.host_required');
    }
    if (!currentParams.value.category || !currentParams.value.name) {
      return t('app.service.history.validation.category_name_required');
    }
    return null;
  };

  // 重置数据状态
  const resetData = (): void => {
    historyList.value = [];
    pagination.total = 0;
  };

  // 处理API错误
  const handleApiError = (
    error: unknown,
    logMessage: string,
    i18nKey: string
  ): void => {
    const errorMessage = error instanceof Error ? error.message : String(error);
    logError(`${logMessage} ${errorMessage}`, error);
    Message.error(t(i18nKey));
  };

  // 加载历史记录
  const loadHistory = async (): Promise<void> => {
    const validationError = validateParams();
    if (validationError) {
      Message.error(validationError);
      return;
    }

    try {
      setLoading(true);
      const response = await getServiceHistoryApi({
        type: currentParams.value.type,
        category: currentParams.value.category,
        name: currentParams.value.name,
        page: pagination.current,
        pageSize: pagination.pageSize,
      });

      historyList.value = response.items || [];
      pagination.total = response.total || 0;
    } catch (error) {
      handleApiError(
        error,
        'Failed to load history:',
        'app.service.history.message.load_failed'
      );
      // 重置数据以防止显示过期信息
      resetData();
    } finally {
      setLoading(false);
    }
  };

  // 初始化数据
  const initializeHistory = (params: HistoryParams): void => {
    currentParams.value = { ...params };
    pagination.current = DEFAULT_PAGE;
    loadHistory();
  };

  // 处理分页变化
  const handlePageChange = (page: number): void => {
    if (page < MIN_PAGE) {
      logError('Invalid page number:', page);
      return;
    }
    pagination.current = page;
    loadHistory();
  };

  const handlePageSizeChange = (pageSize: number): void => {
    if (pageSize < MIN_PAGE_SIZE) {
      logError('Invalid page size:', pageSize);
      return;
    }
    pagination.pageSize = pageSize;
    pagination.current = DEFAULT_PAGE;
    loadHistory();
  };

  // 验证记录数据
  const validateRecord = (record: ServiceHistoryEntity): string | null => {
    if (!record) {
      return t('app.service.history.validation.invalid_record');
    }
    if (!record.commit) {
      return t('app.service.history.validation.invalid_commit');
    }
    return null;
  };

  // 恢复版本
  const handleRestore = async (
    record: ServiceHistoryEntity
  ): Promise<boolean> => {
    const validationError = validateParams();
    if (validationError) {
      Message.error(validationError);
      return false;
    }

    const recordError = validateRecord(record);
    if (recordError) {
      Message.error(recordError);
      return false;
    }

    try {
      const confirmed = await confirm({
        title: t('app.service.history.restore.title'),
        content: t('app.service.history.restore.content', {
          commit: record.commit.substring(0, COMMIT_HASH_DISPLAY_LENGTH),
        }),
      });

      if (!confirmed) {
        return false;
      }

      await restoreServiceApi({
        type: currentParams.value.type,
        category: currentParams.value.category,
        name: currentParams.value.name,
        commit: record.commit,
      });

      Message.success(t('app.service.history.message.restore_success'));
      return true;
    } catch (error) {
      handleApiError(
        error,
        'Failed to restore:',
        'app.service.history.message.restore_failed'
      );
      return false;
    }
  };

  // 刷新当前页数据
  const refresh = (): void => {
    if (currentParams.value.category && currentParams.value.name) {
      loadHistory();
    }
  };

  return {
    // 状态
    historyList,
    pagination,
    loading,
    currentParams,

    // 计算属性
    hasData,
    isLoading,

    // 方法
    initializeHistory,
    handlePageChange,
    handlePageSizeChange,
    handleRestore,
    loadHistory,
    refresh,
  };
}
