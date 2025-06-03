<template>
  <DiffViewer
    ref="diffViewer"
    :title="$t('app.logrotate.history.diff.title')"
    :loading="loading"
    :parsed-diff="parsedDiff"
    :restore-loading="restoreLoading"
    :current-version-label="currentVersionLabel"
    :history-version-label="historyVersionLabel"
    :restore-button-text="$t('app.logrotate.history.restore.button')"
    @close="resetState"
    @restore="handleRestore"
  />
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatCommitHash } from '@/utils/format';
  import DiffViewer from '@/components/diff-viewer/index.vue';
  import type { DiffParams, RestoreSuccessCallback } from './types';
  import useLogrotateDiff from './use-logrotate-diff';

  defineOptions({
    name: 'LogrotateDiffDrawer',
  });

  // 定义组件暴露的接口
  interface DiffDrawerExpose {
    show: (
      params: DiffParams,
      onRestoreSuccess?: RestoreSuccessCallback
    ) => void;
  }

  const { t } = useI18n();

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
  } = useLogrotateDiff();

  // 构建版本标签
  const currentVersionLabel = computed(() => {
    if (!currentVersion.value) {
      return t('app.logrotate.history.diff.current');
    }

    return `${t('app.logrotate.history.diff.current')} ${formatCommitHash(
      currentVersion.value
    )}`;
  });

  const historyVersionLabel = computed(() => {
    if (!diffParams.value) {
      return '';
    }

    return t('app.logrotate.history.diff.version', {
      commit: formatCommitHash(diffParams.value.commit),
    });
  });

  /**
   * 显示diff抽屉
   */
  const show = (
    params: DiffParams,
    restoreSuccessCallback?: RestoreSuccessCallback
  ): void => {
    setDiffParams(params, restoreSuccessCallback);
    diffViewer.value?.show();
  };

  // 暴露方法给父组件
  defineExpose<DiffDrawerExpose>({
    show,
  });
</script>
