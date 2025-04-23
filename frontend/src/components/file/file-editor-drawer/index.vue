<template>
  <a-drawer
    :visible="visible"
    :width="drawerWidth"
    :closable="true"
    :unmount-on-close="false"
    class="file-editor-drawer"
    @cancel="handleCancel"
  >
    <template #title>
      <div class="drawer-title">
        <span>{{ file ? file.name : t('app.file.editor.title') }}</span>
      </div>
    </template>

    <div class="resize-handle" @mousedown="startResize"></div>

    <a-spin :loading="loading" class="file-editor-container">
      <EditorToolbar
        :drawer-width="drawerWidth"
        :is-full-screen="isFullScreen"
        @update:drawer-width="setDrawerWidth"
        @toggle-full-screen="toggleFullScreen"
      />

      <div class="codemirror-wrapper">
        <codemirror
          v-model="content"
          :autofocus="true"
          :indent-with-tab="true"
          :tab-size="2"
          :extensions="extensions"
          @ready="handleReady"
        />
      </div>
    </a-spin>

    <template #footer>
      <EditorFooter
        :file="file"
        :is-edited="isEdited"
        @cancel="handleCancel"
        @save="handleSave"
      />
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, shallowRef } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Codemirror } from 'vue-codemirror';
  import { useConfirm } from '@/hooks/confirm';

  import useFileEditor from './hooks/use-file-editor';
  import useEditorConfig from './hooks/use-editor-config';
  import useDrawerResize from './hooks/use-drawer-resize';

  import EditorToolbar from './editor-toolbar.vue';
  import EditorFooter from './editor-footer.vue';

  const { t } = useI18n();
  const visible = ref(false);
  const editorView = shallowRef();

  const {
    file,
    content,
    loading,
    isEdited,
    setFile: loadFile,
    saveFile,
  } = useFileEditor();

  const {
    drawerWidth,
    isFullScreen,
    setDrawerWidth,
    startResize,
    toggleFullScreen,
  } = useDrawerResize();
  const { extensions } = useEditorConfig(file);

  const { confirm } = useConfirm();

  const emit = defineEmits<{
    (e: 'ok'): void;
  }>();

  const handleReady = (payload: any) => {
    editorView.value = payload.view;
  };

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
    content.value = '';
  };

  const handleCancel = async () => {
    if (isEdited.value) {
      const confirmed = await confirm({
        title: t('app.file.editor.confirmTitle'),
        content: t('app.file.editor.confirmContent'),
        okText: t('common.discard'),
        cancelText: t('common.form.cancelText'),
      });
      if (confirmed) {
        hide();
      }
    } else {
      hide();
    }
  };

  const handleSave = async () => {
    const success = await saveFile();
    if (success) {
      emit('ok');
    }
  };

  defineExpose({
    show,
    hide,
    setFile: loadFile,
  });
</script>

<style scoped>
  .file-editor-drawer :deep(.arco-drawer-body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 0;
  }

  .file-editor-container {
    position: relative;
    display: flex;
    flex: 1;
    flex-direction: column;
    width: 100%;
    overflow: hidden;
  }

  .codemirror-wrapper {
    position: relative;
    flex: 1;
    overflow: hidden;
  }

  .drawer-title {
    font-weight: 500;
  }

  .resize-handle {
    position: absolute;
    top: 0;
    left: 0;
    z-index: 100;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 5px;
    height: 100%;
    background: linear-gradient(to right, transparent, rgb(0 0 0 / 5%));
    cursor: col-resize;
  }

  .resize-handle::after {
    width: 2px;
    height: 30px;
    background-color: rgb(0 0 0 / 20%);
    border-radius: 2px;
    content: '';
  }

  .resize-handle:hover {
    background: linear-gradient(to right, transparent, rgb(0 0 0 / 12%));
  }

  .resize-handle:hover::after {
    width: 3px;
    background-color: rgb(0 0 0 / 40%);
  }

  :deep(.cm-editor) {
    width: 100%;
    height: 100%;
  }

  :deep(.cm-scroller) {
    width: 100%;
    height: 100%;
    overflow: auto;
  }

  :deep(.cm-content) {
    min-height: 100%;
  }

  .file-editor-drawer :deep(.arco-drawer-footer) {
    padding: 12px 20px;
    border-top: 1px solid var(--color-border);
  }
</style>
