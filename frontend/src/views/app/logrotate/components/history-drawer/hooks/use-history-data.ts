import { ref, reactive } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { LOGROTATE_TYPE } from '@/config/enum';
import { LogrotateHistory } from '@/entity/Logrotate';
import { getLogrotateHistoryApi, restoreLogrotateApi } from '@/api/logrotate';
import useLoading from '@/hooks/loading';
import { useConfirm } from '@/hooks/confirm';
import useCurrentHost from '@/hooks/current-host';
import type { HistoryParams, PaginationConfig } from '../types';

export function useHistoryData() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { currentHostId } = useCurrentHost();

  const historyList = ref<LogrotateHistory[]>([]);
  const currentParams = ref<HistoryParams>({
    type: LOGROTATE_TYPE.Local,
    category: '',
    name: '',
  });

  const pagination = reactive<PaginationConfig>({
    current: 1,
    pageSize: 10,
    total: 0,
    showTotal: true,
    showPageSize: true,
  });

  // 验证必要参数
  const validateParams = (): string | null => {
    const hostId = currentHostId.value;
    if (!hostId) {
      return 'Host ID is required';
    }
    if (!currentParams.value.category || !currentParams.value.name) {
      return 'Category and name are required';
    }
    return null;
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
      const response = await getLogrotateHistoryApi({
        type: currentParams.value.type,
        category: currentParams.value.category,
        name: currentParams.value.name,
        page: pagination.current,
        pageSize: pagination.pageSize,
        host: currentHostId.value!,
      });

      historyList.value = response.items || [];
      pagination.total = response.total || 0;
    } catch (error) {
      console.error('Failed to load history:', error);
      Message.error(t('app.logrotate.history.message.load_failed'));
      // 重置数据以防止显示过期信息
      historyList.value = [];
      pagination.total = 0;
    } finally {
      setLoading(false);
    }
  };

  // 初始化数据
  const initializeHistory = (params: HistoryParams): void => {
    currentParams.value = { ...params };
    pagination.current = 1;
    loadHistory();
  };

  // 处理分页变化
  const handlePageChange = (page: number): void => {
    if (page < 1) return;
    pagination.current = page;
    loadHistory();
  };

  const handlePageSizeChange = (pageSize: number): void => {
    if (pageSize < 1) return;
    pagination.pageSize = pageSize;
    pagination.current = 1;
    loadHistory();
  };

  // 恢复版本
  const handleRestore = async (record: LogrotateHistory): Promise<boolean> => {
    const validationError = validateParams();
    if (validationError) {
      Message.error(validationError);
      return false;
    }

    if (!record.commit) {
      Message.error('Invalid commit hash');
      return false;
    }

    try {
      const confirmed = await confirm({
        title: t('app.logrotate.history.restore.title'),
        content: t('app.logrotate.history.restore.content', {
          commit: record.commit.substring(0, 8),
        }),
      });

      if (!confirmed) {
        return false;
      }

      await restoreLogrotateApi({
        type: currentParams.value.type,
        category: currentParams.value.category,
        name: currentParams.value.name,
        commit: record.commit,
        host: currentHostId.value!,
      });

      Message.success(t('app.logrotate.history.message.restore_success'));
      return true;
    } catch (error) {
      console.error('Failed to restore:', error);
      Message.error(t('app.logrotate.history.message.restore_failed'));
      return false;
    }
  };

  return {
    // 状态
    historyList,
    pagination,
    loading,

    // 方法
    initializeHistory,
    handlePageChange,
    handlePageSizeChange,
    handleRestore,
  };
}
