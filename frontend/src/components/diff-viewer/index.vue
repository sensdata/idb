<template>
  <a-drawer
    v-model:visible="visible"
    :title="title"
    :width="width"
    :footer="false"
    unmount-on-close
  >
    <div class="diff-drawer">
      <a-spin :loading="loading" style="width: 100%">
        <VersionHeader
          :current-version-label="currentVersionLabel"
          :history-version-label="historyVersionLabel"
          :restore-loading="restoreLoading"
          :show-restore-button="showRestoreButton"
          @restore="handleRestore"
        />

        <DiffView :parsed-diff="parsedDiff" />
      </a-spin>

      <!-- 底部操作栏 -->
      <div class="diff-footer">
        <div class="footer-actions">
          <a-button @click="handleClose">
            {{ $t('common.cancel') }}
          </a-button>
        </div>
      </div>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useLogger } from '@/hooks/use-logger';
  import DiffView from './diff-view.vue';
  import VersionHeader from './version-header.vue';
  import type { ParsedDiff, DiffViewerExpose } from './types';

  defineOptions({
    name: 'DiffViewer',
  });

  interface Props {
    title?: string;
    width?: number;
    currentVersionLabel?: string;
    historyVersionLabel?: string;
    showRestoreButton?: boolean;
    loading?: boolean;
    parsedDiff: ParsedDiff | null;
    restoreLoading?: boolean;
    loadDiffContent?: () => Promise<void>;
  }

  const { t } = useI18n();
  const { logError } = useLogger('DiffViewer');

  const props = withDefaults(defineProps<Props>(), {
    title: '',
    width: 1000,
    currentVersionLabel: '',
    historyVersionLabel: '',
    showRestoreButton: true,
    loading: false,
    restoreLoading: false,
    loadDiffContent: undefined,
  });

  // 计算属性处理i18n默认值
  const title = computed(() => props.title || t('diff-viewer.title'));
  const currentVersionLabel = computed(
    () => props.currentVersionLabel || t('diff-viewer.currentVersion')
  );
  const historyVersionLabel = computed(
    () => props.historyVersionLabel || t('diff-viewer.historyVersion')
  );

  interface Emits {
    (e: 'close'): void;
    (e: 'restore'): void;
  }

  const emit = defineEmits<Emits>();

  const visible = ref(false);
  const onRestoreSuccessCallback = ref<(() => void) | null>(null);

  /**
   * 显示diff抽屉
   */
  const show = async (restoreSuccessCallback?: () => void): Promise<void> => {
    onRestoreSuccessCallback.value = restoreSuccessCallback || null;
    visible.value = true;

    if (props.loadDiffContent) {
      try {
        await props.loadDiffContent();
      } catch (error) {
        logError(t('diff-viewer.loadError'), error);
      }
    }
  };

  /**
   * 关闭抽屉
   */
  const handleClose = (): void => {
    visible.value = false;
    emit('close');
  };

  /**
   * 处理恢复操作
   */
  const handleRestore = (): void => {
    emit('restore');
  };

  /**
   * 执行恢复成功回调
   */
  const executeRestoreSuccessCallback = (): void => {
    if (onRestoreSuccessCallback.value) {
      onRestoreSuccessCallback.value();
    }
  };

  // 监听visible变化，当抽屉关闭时清理回调
  watch(visible, (newVisible) => {
    if (!newVisible) {
      onRestoreSuccessCallback.value = null;
    }
  });

  // 暴露方法给父组件
  defineExpose<DiffViewerExpose>({
    show,
    executeRestoreSuccessCallback,
  });
</script>

<style scoped>
  .diff-drawer {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .diff-footer {
    flex-shrink: 0;
    padding: 16px 0 0 0;
    border-top: 1px solid var(--color-border-2);
  }

  .footer-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
</style>
