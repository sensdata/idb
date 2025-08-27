<template>
  <code-editor
    ref="editorRef"
    v-model="value"
    :autofocus="true"
    :indent-with-tab="true"
    :tab-size="4"
    :extensions="extensions"
    :read-only="true"
    :file="logFile"
  />
</template>

<script lang="ts" setup>
  import { ref, watch, computed } from 'vue';
  import { StreamLanguage } from '@codemirror/language';
  import { simpleMode } from '@codemirror/legacy-modes/mode/simple-mode';
  import CodeEditor from '@/components/code-editor/index.vue';

  const props = defineProps({
    content: {
      type: String,
      default: '',
    },
  });

  const editorRef = ref<InstanceType<typeof CodeEditor>>();
  const scrollBottom = () => {
    if (editorRef.value?.editorView) {
      const view = editorRef.value.editorView;
      view.dispatch({
        effects: view.state.reconfigure({
          scrollIntoView: view.state.doc.length,
        }),
      });
    }
  };

  const value = ref(props.content);
  watch(
    () => props.content,
    (newVal) => {
      value.value = newVal;
    },
    {
      immediate: true,
    }
  );

  // 创建日志文件对象，让 code-editor 使用日志语法高亮
  const logFile = computed(() => ({
    name: 'application.log',
    path: '/tmp/application.log',
  }));

  // 使用简单的日志语法高亮
  const logSyntax = {
    start: [
      // 时间戳
      { regex: /\d{4}-\d{2}-\d{2}[\sT]\d{2}:\d{2}:\d{2}/, token: 'number' },
      // 错误级别
      { regex: /\b(?:ERROR|FATAL|CRITICAL)\b/i, token: 'keyword' },
      { regex: /\b(?:WARN|WARNING)\b/i, token: 'variable' },
      { regex: /\b(?:INFO|INFORMATION)\b/i, token: 'comment' },
      { regex: /\b(?:DEBUG|TRACE)\b/i, token: 'string' },
      // IP 地址
      { regex: /\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/, token: 'def' },
    ],
  };

  const extensions = [StreamLanguage.define(simpleMode(logSyntax))];

  defineExpose({
    scrollBottom,
  });
</script>
