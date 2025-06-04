<template>
  <DiffViewer
    ref="diffViewer"
    :title="$t('app.service.history.diff.title')"
    :loading="loading"
    :parsed-diff="parsedDiff"
    :restore-loading="restoreLoading"
    :current-version-label="currentVersionLabel"
    :history-version-label="historyVersionLabel"
    :restore-button-text="$t('app.service.history.restore.button')"
    @close="resetState"
    @restore="handleRestore"
  />
</template>

<script setup lang="ts">
  /**
   * 服务配置差异对比抽屉组件
   * 用于显示服务配置文件的版本差异，并支持版本恢复功能
   */
  import { computed, ref, onUnmounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatCommitHash } from '@/utils/format';
  import { useLogger } from '@/hooks/use-logger';
  import DiffViewer from '@/components/diff-viewer/index.vue';
  import type {
    DiffParams,
    RestoreSuccessCallback,
    DiffDrawerExpose,
  } from './types';
  import useServiceDiff from './use-service-diff';

  defineOptions({
    name: 'ServiceDiffDrawer',
  });

  const { t } = useI18n();
  const { logWarn } = useLogger('ServiceDiffDrawer');

  // 改进ref的类型定义
  const diffViewer = ref<InstanceType<typeof DiffViewer> | null>(null);

  const {
    loading,
    diffParams,
    parsedDiff,
    currentVersion,
    restoreLoading,
    setDiffParams,
    resetState,
    handleRestore,
  } = useServiceDiff();

  // 构建版本标签
  const currentVersionLabel = computed(() => {
    if (!currentVersion.value) {
      return t('app.service.history.diff.current');
    }

    return `${t('app.service.history.diff.current')} ${formatCommitHash(
      currentVersion.value
    )}`;
  });

  const historyVersionLabel = computed(() => {
    if (!diffParams.value) {
      return '';
    }

    return t('app.service.history.diff.version', {
      commit: formatCommitHash(diffParams.value.commit),
    });
  });

  /**
   * 显示diff抽屉
   * @param params - 差异对比参数
   * @param restoreSuccessCallback - 恢复成功后的回调函数
   */
  const show = (
    params: DiffParams,
    restoreSuccessCallback?: RestoreSuccessCallback
  ): void => {
    if (!params) {
      logWarn('params is required');
      return;
    }

    setDiffParams(params, restoreSuccessCallback);
    diffViewer.value?.show();
  };

  // 组件卸载时清理状态
  onUnmounted(() => {
    resetState();
  });

  // 暴露方法给父组件
  defineExpose<DiffDrawerExpose>({
    show,
  });
</script>
