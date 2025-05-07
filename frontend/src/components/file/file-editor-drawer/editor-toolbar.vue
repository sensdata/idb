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

      <a-button type="outline" size="small" @click="$emit('toggleEditMode')">
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

  const { t } = useI18n();

  defineProps({
    drawerWidth: {
      type: Number,
      required: true,
    },
    isFullScreen: {
      type: Boolean,
      required: true,
    },
    readOnly: {
      type: Boolean,
      default: true,
    },
  });

  const emit = defineEmits<{
    (e: 'update:drawer-width', width: number): void;
    (e: 'toggleFullScreen'): void;
    (e: 'toggleEditMode'): void;
  }>();

  const toggleFullScreen = () => {
    emit('toggleFullScreen');
  };
</script>

<style scoped>
  .editor-toolbar {
    margin-bottom: 8px;
    padding: 8px 0;
    border-bottom: 1px solid var(--color-border-2);
  }
</style>
