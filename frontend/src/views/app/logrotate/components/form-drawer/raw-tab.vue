<template>
  <div class="raw-editor">
    <div class="editor-container">
      <CodeEditor
        :model-value="content"
        :file="logrotateFile"
        :autofocus="true"
        :indent-with-tab="true"
        :tab-size="2"
        @update:model-value="handleContentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import CodeEditor from '@/components/code-editor/index.vue';

  interface Props {
    content: string;
    extensions?: any[]; // 保留 extensions 属性以保持向后兼容，但不再使用
  }

  const emit = defineEmits<{
    'update:content': [content: string];
  }>();

  defineProps<Props>();

  // 创建虚拟文件对象，用于自动语法高亮
  const logrotateFile = computed(() => ({
    name: 'logrotate',
    path: '/etc/logrotate.d/logrotate',
  }));

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

  /* CodeEditor 组件样式调整 */
  .editor-container :deep(.code-editor) {
    width: 100%;
    height: 100%;
    min-height: 400px;
  }

  .editor-container :deep(.cm-editor) {
    width: 100%;
    height: 100%;
    min-height: 400px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', Consolas, source-code-pro,
      monospace;
    font-size: 13px;
    line-height: 1.5;
  }

  .editor-container :deep(.cm-scroller) {
    height: 100%;
    min-height: 400px;
    overflow: auto;
  }

  .editor-container :deep(.cm-content) {
    min-height: 400px;
    padding: 12px;
  }

  .editor-container :deep(.cm-focused) {
    outline: none;
  }
</style>
