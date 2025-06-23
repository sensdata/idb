import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { formatCommitHash } from '@/utils/format';
import {
  getLogrotateHistoryDiffApi,
  restoreLogrotateApi,
  getLogrotateHistoryApi,
} from '@/api/logrotate';
import useCurrentHost from '@/composables/current-host';
import useDiff from '@/components/diff-viewer/use-diff';
import { DiffParams } from './types';

export default function useLogrotateDiff() {
  const { t } = useI18n();
  const { currentHostId } = useCurrentHost();

  // 使用计算属性确保响应性
  const isHostAvailable = computed(() => !!currentHostId.value);

  // 当前模块特有的获取当前版本的方法
  const getCurrentCommit = async (params: DiffParams): Promise<string> => {
    if (!isHostAvailable.value) {
      throw new Error('Host ID is required but not available');
    }

    try {
      const historyResult = await getLogrotateHistoryApi({
        type: params.type,
        category: params.category,
        name: params.name,
        page: 1,
        pageSize: 1,
        host: currentHostId.value!,
      });

      // 第一条记录就是当前版本
      if (historyResult?.items?.length > 0) {
        return historyResult.items[0].commit;
      }

      return '';
    } catch (error) {
      console.error('Failed to get current commit:', error);
      throw new Error(t('app.logrotate.history.message.get_current_failed'));
    }
  };

  // 获取diff内容的方法
  const fetchDiffContent = async (params: DiffParams): Promise<string> => {
    if (!isHostAvailable.value) {
      throw new Error('Host ID is required but not available');
    }

    try {
      const content = await getLogrotateHistoryDiffApi({
        type: params.type,
        category: params.category,
        name: params.name,
        commit: params.commit,
        host: currentHostId.value!,
      });

      return content || '';
    } catch (error) {
      console.error('Failed to fetch diff content:', error);
      throw new Error(t('app.logrotate.history.message.diff_failed'));
    }
  };

  // 恢复版本的方法
  const restoreVersion = async (params: DiffParams): Promise<void> => {
    if (!isHostAvailable.value) {
      throw new Error('Host ID is required but not available');
    }

    try {
      await restoreLogrotateApi({
        type: params.type,
        category: params.category,
        name: params.name,
        commit: params.commit,
        host: currentHostId.value!,
      });
    } catch (error) {
      console.error('Failed to restore version:', error);
      throw new Error(t('app.logrotate.history.message.restore_failed'));
    }
  };

  // 使用通用的diff hook，注入特定的业务逻辑
  const diffHook = useDiff<DiffParams>({
    fetchDiffContent,
    fetchCurrentVersion: getCurrentCommit,
    restoreVersion,
    formatVersion: formatCommitHash,

    // 本地化消息
    diffFailedMessage: t('app.logrotate.history.message.diff_failed'),
    restoreSuccessMessage: t('app.logrotate.history.message.restore_success'),
    restoreFailedMessage: t('app.logrotate.history.message.restore_failed'),
    restoreConfirmTitle: t('app.logrotate.history.restore.title'),
    restoreConfirmContent: (commit) =>
      t('app.logrotate.history.restore.content', { commit }),
  });

  return {
    ...diffHook,
    isHostAvailable,
  };
}
