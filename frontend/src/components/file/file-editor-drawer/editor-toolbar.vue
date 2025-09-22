<template>
  <div class="editor-toolbar">
    <a-space>
      <a-button type="outline" size="small" @click="toggleFullScreen">
        <icon-fullscreen v-if="!isFullScreen" />
        <icon-fullscreen-exit v-else />
        {{
          isFullScreen
            ? t('app.file.editor.exitFullScreen')
            : t('app.file.editor.fullScreen')
        }}
      </a-button>

      <!-- 只有在完整视图模式下才显示编辑按钮 -->
      <a-button
        v-if="viewMode === 'full'"
        type="outline"
        size="small"
        @click="toggleEditMode"
      >
        <icon-edit v-if="readOnly" />
        <icon-eye v-else />
        {{
          readOnly
            ? t('app.file.editor.enableEdit')
            : t('app.file.editor.viewOnly')
        }}
      </a-button>
    </a-space>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import {
    IconFullscreen,
    IconFullscreenExit,
    IconEdit,
    IconEye,
  } from '@arco-design/web-vue/es/icon';

  interface Props {
    drawerWidth: number;
    isFullScreen: boolean;
    readOnly?: boolean;
    viewMode?: string; // 添加视图模式属性
  }

  const { t } = useI18n();

  defineProps<Props>();

  const emit = defineEmits<{
    'update:drawer-width': [width: number];
    'toggleFullScreen': [];
    'toggleEditMode': [];
  }>();

  const toggleFullScreen = () => {
    emit('toggleFullScreen');
  };

  const toggleEditMode = () => {
    emit('toggleEditMode');
  };
</script>

<style scoped>
  .editor-toolbar {
    display: flex;
    flex-shrink: 0; /* 防止收缩 */
    align-items: center;
    padding: 8px 12px;
    background-color: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .editor-toolbar :deep(.arco-space) {
    display: flex;
    align-items: center;
  }
</style>
