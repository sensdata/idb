<template>
  <div class="raw-editor">
    <div class="editor-container">
      <codemirror
        :model-value="content"
        :placeholder="$t('app.logrotate.form.raw_placeholder')"
        :autofocus="true"
        :indent-with-tab="true"
        :tab-size="2"
        :extensions="extensions"
        class="raw-codemirror"
        @update:model-value="handleContentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Codemirror } from 'vue-codemirror';

  interface Props {
    content: string;
    extensions: any[];
  }

  const emit = defineEmits<{
    'update:content': [content: string];
  }>();

  defineProps<Props>();

  const handleContentChange = (value: string) => {
    emit('update:content', value);
  };
</script>

<style scoped>
  .raw-editor {
    display: flex;
    flex: 1;
    flex-direction: column;
    height: 100%;
    min-height: 400px;
  }

  .editor-container {
    flex: 1;
    min-height: 400px;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .raw-codemirror {
    width: 100%;
    height: 100%;
    min-height: 400px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', Consolas, source-code-pro,
      monospace;
    font-size: 13px;
    line-height: 1.5;
  }

  .raw-codemirror :deep(.cm-editor) {
    width: 100%;
    height: 100%;
    min-height: 400px;
  }

  .raw-codemirror :deep(.cm-scroller) {
    height: 100%;
    min-height: 400px;
    overflow: auto;
  }

  .raw-codemirror :deep(.cm-content) {
    min-height: 400px;
    padding: 12px;
  }

  .raw-codemirror :deep(.cm-focused) {
    outline: none;
  }
</style>
