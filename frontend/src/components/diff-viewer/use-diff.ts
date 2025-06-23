import { ref, onUnmounted, readonly } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useConfirm } from '@/composables/confirm';
import { useLogger } from '@/composables/use-logger';
import { ParsedDiff } from './types';
import { decodeUnicodeString, parseDiffToSideBySide } from './utils';

// 提取为独立的可重用hook
export function useLoading() {
  const loading = ref(false);

  const setLoading = (value: boolean) => {
    loading.value = value;
  };

  return {
    loading: readonly(loading), // 只读，防止外部直接修改
    setLoading,
  };
}

export interface DiffOptions<P> {
  fetchDiffContent: (params: P) => Promise<string>;
  fetchCurrentVersion?: (params: P) => Promise<string>;
  restoreVersion?: (params: P) => Promise<void>;

  diffFailedMessage?: string;
  restoreSuccessMessage?: string;
  restoreFailedMessage?: string;
  restoreConfirmTitle?: string;
  restoreConfirmContent?: (version: string) => string;

  formatVersion?: (version: string) => string;

  // 使其更加明确
  versionField?: keyof P;
}

// 改进类型约束，更具体
export default function useDiff<P extends Record<string, unknown>>(
  options: DiffOptions<P>
) {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const { logError, logInfo } = useLogger('DiffViewer');

  // 使用更明确的初始值
  const diffContent = ref('');
  const diffParams = ref<P | null>(null);
  const parsedDiff = ref<ParsedDiff | null>(null);
  const restoreLoading = ref(false);
  const currentVersion = ref('');
  const onRestoreSuccess = ref<(() => void) | null>(null);

  // 提供默认格式化函数，使其更加安全
  const formatVersion = options.formatVersion ?? ((version: string) => version);

  /**
   * 重置diff相关状态 - 提前定义避免use-before-define错误
   */
  const resetDiffState = (): void => {
    parsedDiff.value = null;
    diffContent.value = '';
    currentVersion.value = '';
  };

  /**
   * 从参数中获取版本号 - 改进版本字段检测逻辑
   */
  const getVersionFromParams = (params: P): string => {
    // 如果指定了版本字段，优先使用
    if (options.versionField && params[options.versionField]) {
      return String(params[options.versionField]);
    }

    // 按优先级尝试常见的版本字段
    const versionFields = [
      'commit',
      'version',
      'revision',
      'hash',
      'id',
    ] as const;

    for (const field of versionFields) {
      const value = params[field as keyof P];
      if (value) {
        return String(value);
      }
    }

    return '';
  };

  /**
   * 处理diff内容 - 添加更好的错误处理
   */
  const processDiffContent = (content: string): void => {
    try {
      // 解码Unicode转义字符
      const decodedContent = content ? decodeUnicodeString(content) : '';

      // 解析为侧边对比格式
      parsedDiff.value = parseDiffToSideBySide(decodedContent);
      diffContent.value = decodedContent;
    } catch (error) {
      logError('处理diff内容失败', error);
      // 设置安全的默认值
      parsedDiff.value = null;
      diffContent.value = '';
      throw error; // 重新抛出以便上层处理
    }
  };

  /**
   * 获取当前版本 - 改进错误处理
   */
  const getCurrentVersion = async (params: P): Promise<string> => {
    if (!options.fetchCurrentVersion) {
      return '';
    }

    try {
      const version = await options.fetchCurrentVersion(params);
      return version || '';
    } catch (error) {
      logError('获取当前版本失败', error);
      return '';
    }
  };

  /**
   * 加载diff内容 - 优化并行请求逻辑
   */
  const loadDiffContent = async (params: P): Promise<void> => {
    try {
      setLoading(true);
      logInfo('开始加载diff内容', { params });

      // 构建并行任务
      const tasks = [options.fetchDiffContent(params)];

      if (options.fetchCurrentVersion) {
        tasks.unshift(getCurrentVersion(params));
      }

      // 并行执行所有任务
      const results = await Promise.all(tasks);

      // 处理结果
      if (options.fetchCurrentVersion) {
        const [version, content] = results;
        currentVersion.value = version;
        processDiffContent(content);
      } else {
        const [content] = results;
        processDiffContent(content);
      }

      logInfo('diff内容加载成功');
    } catch (error) {
      logError('加载diff内容失败', error);

      // 统一错误处理和状态重置
      const errorMessage =
        options.diffFailedMessage || t('common.message.operation_failed');
      Message.error(errorMessage);

      // 重置所有相关状态
      resetDiffState();
    } finally {
      setLoading(false);
    }
  };

  /**
   * 恢复到历史版本 - 改进确认逻辑
   */
  const handleRestore = async (): Promise<boolean> => {
    if (!diffParams.value || !options.restoreVersion) {
      const errorMsg = 'Missing required parameters or restore function';
      logError(errorMsg);
      Message.error(errorMsg);
      return false;
    }

    try {
      logInfo('开始恢复到历史版本', { params: diffParams.value });

      // 处理确认对话框
      if (options.restoreConfirmTitle) {
        const version = getVersionFromParams(diffParams.value);
        const formattedVersion = formatVersion(version);

        const confirmContent = options.restoreConfirmContent
          ? options.restoreConfirmContent(formattedVersion)
          : t('common.confirm.restore_content');

        const result = await confirm({
          title: options.restoreConfirmTitle,
          content: confirmContent,
        });

        if (result !== true) {
          logInfo('用户取消了恢复操作');
          return false;
        }
      }

      restoreLoading.value = true;
      await options.restoreVersion(diffParams.value);

      const successMessage =
        options.restoreSuccessMessage || t('common.message.operation_success');
      Message.success(successMessage);

      // 调用成功回调
      onRestoreSuccess.value?.();

      logInfo('恢复到历史版本成功');
      return true;
    } catch (error) {
      logError('恢复到历史版本失败', error);
      const errorMessage =
        options.restoreFailedMessage || t('common.message.operation_failed');
      Message.error(errorMessage);
      return false;
    } finally {
      restoreLoading.value = false;
    }
  };

  /**
   * 设置diff参数并加载数据
   */
  const setDiffParams = (params: P, callback?: () => void): void => {
    diffParams.value = params;
    onRestoreSuccess.value = callback || null;
    loadDiffContent(params);
  };

  /**
   * 清理状态 - 改名为更清晰的名称
   */
  const resetState = (): void => {
    currentVersion.value = '';
    parsedDiff.value = null;
    diffContent.value = '';
    diffParams.value = null;
    onRestoreSuccess.value = null;
  };

  // 组件卸载时自动清理
  onUnmounted(() => {
    resetState();
  });

  return {
    // 状态（只读）
    loading,
    diffParams: readonly(diffParams),
    parsedDiff: readonly(parsedDiff),
    currentVersion: readonly(currentVersion),
    restoreLoading: readonly(restoreLoading),

    // 方法
    formatVersion,
    setDiffParams,
    resetState,
    handleRestore,
  };
}
