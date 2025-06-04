import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { formatCommitHash } from '@/utils/format';
import { useLogger } from '@/hooks/use-logger';
import { getServiceDiffApi, restoreServiceApi } from '@/api/service';
import {
  parseDiffToSideBySide,
  decodeUnicodeString,
} from '@/components/diff-viewer/utils';
import { useConfirm } from '@/hooks/confirm';
import useLoading from '@/hooks/loading';
import type { ParsedDiff } from '@/components/diff-viewer/types';
import type { DiffParams, RestoreSuccessCallback } from './types';

export default function useServiceDiff() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { logError } = useLogger('useServiceDiff');

  const diffParams = ref<DiffParams | null>(null);
  const parsedDiff = ref<ParsedDiff | null>(null);
  const currentVersion = ref<string>('');
  const restoreLoading = ref(false);
  const onRestoreSuccess = ref<RestoreSuccessCallback | null>(null);

  /**
   * 加载diff内容
   */
  const loadDiffContent = async (params: DiffParams): Promise<void> => {
    try {
      setLoading(true);

      const diffContent = await getServiceDiffApi({
        type: params.type,
        category: params.category,
        name: params.name,
        commit: params.commit,
      });

      // 先解码Unicode字符串，然后解析diff内容为侧边对比格式
      const decodedContent = decodeUnicodeString(diffContent);
      parsedDiff.value = parseDiffToSideBySide(decodedContent);
    } catch (error) {
      logError('Failed to load diff content:', error);
      Message.error(t('app.service.history.diff.error.load'));
      parsedDiff.value = null;
    } finally {
      setLoading(false);
    }
  };

  /**
   * 设置diff参数并加载数据
   */
  const setDiffParams = (
    params: DiffParams,
    callback?: RestoreSuccessCallback
  ): void => {
    diffParams.value = params;
    onRestoreSuccess.value = callback || null;
    loadDiffContent(params);
  };

  /**
   * 重置状态
   */
  const resetState = (): void => {
    diffParams.value = null;
    parsedDiff.value = null;
    currentVersion.value = '';
    onRestoreSuccess.value = null;
  };

  /**
   * 处理恢复操作
   */
  const handleRestore = async (): Promise<void> => {
    if (!diffParams.value) {
      return;
    }

    try {
      const confirmed = await confirm({
        title: t('app.service.history.restore.title'),
        content: t('app.service.history.restore.content', {
          commit: formatCommitHash(diffParams.value.commit),
        }),
      });

      if (!confirmed) {
        return;
      }

      restoreLoading.value = true;

      await restoreServiceApi({
        type: diffParams.value.type,
        category: diffParams.value.category,
        name: diffParams.value.name,
        commit: diffParams.value.commit,
      });

      Message.success(t('app.service.history.restore.success'));

      // 调用成功回调
      if (onRestoreSuccess.value) {
        onRestoreSuccess.value();
      }
    } catch (error) {
      logError('Failed to restore:', error);
      Message.error(t('app.service.history.restore.error'));
    } finally {
      restoreLoading.value = false;
    }
  };

  return {
    loading,
    diffParams,
    parsedDiff,
    currentVersion,
    restoreLoading,
    setDiffParams,
    resetState,
    handleRestore,
  };
}
