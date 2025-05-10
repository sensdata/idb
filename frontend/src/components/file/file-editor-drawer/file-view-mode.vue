<template>
  <div v-if="viewMode !== 'loading'" class="file-mode-controls">
    <div class="info-section">
      <span class="info-label">{{ t('app.file.editor.mode') }}:</span>
      <a-badge
        :color="getViewModeColor(viewMode)"
        :text="getViewModeText(viewMode)"
      />
    </div>

    <div class="actions-section">
      <!-- 行数输入框 - 直接在工具栏上显示 -->
      <div v-if="isLargeFile" class="line-input-section">
        <a-input-number
          v-model="tempLineCount"
          :min="1"
          :max="10000"
          size="small"
          style="width: 80px"
          @input="handleLineCountInput"
          @keyup.enter="handleJumpAction"
        />
        <span class="lines-label">{{ t('app.file.editor.lines') }}</span>
        <!-- 跳转按钮 - 只有当行数改变时才可点击 -->
        <a-button
          type="primary"
          size="small"
          :disabled="!lineCountChanged"
          @click="handleJumpAction"
        >
          {{ t('app.file.editor.jump') }}
        </a-button>
      </div>

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
            @click="() => handleViewModeChange('head')"
          >
            {{ t('app.file.editor.viewHead') }}
          </a-button>
          <a-button
            :type="viewMode === 'tail' ? 'primary' : 'outline'"
            :disabled="viewMode === 'tail'"
            :loading="loadingButton === 'tail'"
            @click="() => handleViewModeChange('tail')"
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
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ContentViewMode } from '@/components/file/file-editor-drawer/types';

  const { t } = useI18n();

  const props = defineProps({
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

  const emit = defineEmits<{
    (e: 'changeViewMode', mode: ContentViewMode, lines?: number): void;
  }>();

  const loadingButton = ref<ContentViewMode | null>(null);
  const tempLineCount = ref(100); // 默认显示100行
  const originalLineCount = ref(props.lineCount); // 记录原始行数

  // 检查行数是否发生了变化
  const lineCountChanged = computed(() => {
    return tempLineCount.value !== originalLineCount.value;
  });

  // 处理输入框值变化 - 这是新增的函数，用于实时响应输入
  const handleLineCountInput = (value: number | undefined) => {
    // 输入时立即更新tempLineCount值
    if (value !== undefined) {
      tempLineCount.value = value;
    }
  };

  // 视图模式切换函数 - 先定义这个函数，避免被使用前未定义错误
  const handleViewModeChange = (mode: ContentViewMode, lines?: number) => {
    // 设置点击的按钮为加载状态
    loadingButton.value = mode;

    // 触发模式变更事件
    emit('changeViewMode', mode, lines);
  };

  // 处理跳转按钮点击或Enter按键事件
  const handleJumpAction = () => {
    // 如果行数没有变化，不做任何操作
    if (!lineCountChanged.value) return;

    // 触发当前模式下的行数跳转
    handleViewModeChange(props.viewMode, tempLineCount.value);

    // 更新originalLineCount以便重新计算lineCountChanged状态
    originalLineCount.value = tempLineCount.value;
  };

  // 当lineCount变化时，更新tempLineCount和originalLineCount
  watch(
    () => props.lineCount,
    (newValue) => {
      if (newValue > 0) {
        tempLineCount.value = newValue;
        originalLineCount.value = newValue;
      }
    },
    { immediate: true }
  );

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
    gap: 8px;
    align-items: center;
  }

  .line-input-section {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .lines-label {
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
