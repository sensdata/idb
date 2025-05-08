<template>
  <div v-if="viewMode !== 'loading'" class="file-mode-controls">
    <div class="info-section">
      <span class="info-label">{{ t('app.file.editor.mode') }}:</span>
      <a-badge
        :color="getViewModeColor(viewMode)"
        :text="getViewModeText(viewMode)"
      />
      <span v-if="viewMode === 'head' || viewMode === 'tail'" class="line-info">
        ({{ lineCount }} {{ t('app.file.editor.lines') }})
      </span>
    </div>
    <div class="actions-section">
      <a-button-group size="small">
        <!-- 小文件(<100KB)只显示完整视图和实时追踪按钮 -->
        <template v-if="!isLargeFile">
          <a-button
            :type="viewMode === 'full' ? 'primary' : 'outline'"
            :disabled="viewMode === 'full'"
            :loading="loadingButton === 'full'"
            @click="handleViewModeChange('full')"
          >
            {{ t('app.file.editor.viewFull') }}
          </a-button>
          <a-button
            :type="(viewMode as string) === 'follow' ? 'primary' : 'outline'"
            :disabled="(viewMode as string) === 'follow'"
            :loading="loadingButton === 'follow'"
            @click="handleViewModeChange('follow' as ContentViewMode)"
          >
            {{ t('app.file.editor.viewFollow') }}
          </a-button>
        </template>

        <!-- 大文件(>100KB)显示查看开头、查看末尾和实时追踪按钮 -->
        <template v-else>
          <a-button
            :type="viewMode === 'head' ? 'primary' : 'outline'"
            :disabled="viewMode === 'head'"
            :loading="loadingButton === 'head'"
            @click="handleViewModeChange('head', batchSize)"
          >
            {{ t('app.file.editor.viewHead') }}
          </a-button>
          <a-button
            :type="viewMode === 'tail' ? 'primary' : 'outline'"
            :disabled="viewMode === 'tail'"
            :loading="loadingButton === 'tail'"
            @click="handleViewModeChange('tail', batchSize)"
          >
            {{ t('app.file.editor.viewTail') }}
          </a-button>
          <a-button
            :type="(viewMode as string) === 'follow' ? 'primary' : 'outline'"
            :disabled="(viewMode as string) === 'follow'"
            :loading="loadingButton === 'follow'"
            @click="handleViewModeChange('follow' as ContentViewMode)"
          >
            {{ t('app.file.editor.viewFollow') }}
          </a-button>
        </template>
      </a-button-group>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ContentViewMode } from '@/components/file/file-editor-drawer/types';

  const { t } = useI18n();

  defineProps({
    viewMode: {
      type: String as () => ContentViewMode,
      required: true,
    },
    lineCount: {
      type: Number,
      default: 0,
    },
    isLargeFile: {
      type: Boolean,
      required: true,
    },
    batchSize: {
      type: Number,
      required: true,
    },
  });

  const loadingButton = ref<ContentViewMode | null>(null);

  const emit = defineEmits<{
    (e: 'changeViewMode', mode: ContentViewMode, lines?: number): void;
  }>();

  const handleViewModeChange = (mode: ContentViewMode, lines?: number) => {
    // 设置点击的按钮为加载状态
    loadingButton.value = mode;

    // 触发模式变更事件
    emit('changeViewMode', mode, lines);
  };

  // 暴露加载按钮状态更新方法
  const clearLoadingState = () => {
    loadingButton.value = null;
  };

  // 视图模式颜色映射
  const getViewModeColor = (mode: ContentViewMode): string => {
    switch (mode) {
      case 'head':
        return '#00B42A';
      case 'tail':
        return '#165DFF';
      case 'follow':
        return '#F53F3F';
      default:
        return '#86909C';
    }
  };

  // 视图模式文本映射
  const getViewModeText = (mode: ContentViewMode): string => {
    switch (mode) {
      case 'head':
        return t('app.file.editor.viewHead');
      case 'tail':
        return t('app.file.editor.viewTail');
      case 'follow':
        return t('app.file.editor.viewFollow');
      default:
        return t('app.file.editor.viewFull');
    }
  };

  defineExpose({
    clearLoadingState,
  });
</script>

<style scoped>
  /* 文件模式控制样式 */
  .file-mode-controls {
    position: fixed;
    top: 50px;
    right: 20px;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    width: auto;
    max-width: 650px;
    padding: 6px 12px;
    background-color: var(--color-bg-1);
    border: 1px solid var(--color-border);
    border-radius: 4px;
    box-shadow: 0 2px 8px rgb(0 0 0 / 15%);
  }

  .info-section {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-right: 20px;
  }

  .info-label {
    font-weight: 500;
  }

  .line-info {
    margin-left: 4px;
    color: var(--color-text-3);
  }

  .actions-section {
    display: flex;
    align-items: center;
  }
</style>
