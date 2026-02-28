<template>
  <div class="code-editor">
    <div class="codemirror-wrapper">
      <codemirror
        :model-value="modelValue"
        :autofocus="autofocus"
        :indent-with-tab="indentWithTab"
        :tab-size="tabSize"
        :extensions="computedExtensions"
        :disabled="readOnly"
        @update:model-value="$emit('update:modelValue', $event)"
        @ready="handleReady"
      />
    </div>

    <!-- 编辑器加载中 -->
    <div v-if="loading && !modelValue" class="loading-container">
      <a-spin :size="24">
        <template #icon>
          <icon-loading />
        </template>
        <template #tip>
          <span class="loading-text">{{ defaultLoadingText }}</span>
        </template>
      </a-spin>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { shallowRef, nextTick, computed, watch, onUnmounted } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { EditorView } from '@codemirror/view';
  import { basicSetup } from 'codemirror';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { IconLoading } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import useThemes from '@/composables/themes';
  import useEditorConfig from './composables/use-editor-config';
  import type { EditorProps, EditorEmits } from './types';

  const props = withDefaults(defineProps<EditorProps>(), {
    loading: false,
    readOnly: false,
    autofocus: true,
    indentWithTab: true,
    tabSize: 2,
    file: null,
    extensions: () => [],
    isPartialView: false,
    loadingText: '',
  });

  const emit = defineEmits<EditorEmits>();

  const { t } = useI18n();
  const { isDark } = useThemes();
  const editorView = shallowRef<EditorView>();

  // 使用编辑器配置
  const { extensions: configExtensions } = useEditorConfig(
    computed(() => props.file)
  );

  // 合并扩展 - 添加 basicSetup 和主题配置
  const computedExtensions = computed(() => {
    const themeExtensions = isDark.value ? [oneDark] : [];
    return [
      basicSetup,
      ...themeExtensions,
      ...configExtensions.value,
      ...props.extensions,
    ];
  });

  // 默认加载文本
  const defaultLoadingText = computed(() => {
    return (
      props.loadingText || t('app.file.editor.loadingContent', '加载内容中...')
    );
  });

  const handleReady = (payload: any) => {
    editorView.value = payload.view;

    // 将编辑器实例传递给父组件
    emit('editorReady', { view: payload.view });

    // 确保编辑器初始化后，如果处于部分查看模式，滚动到顶部
    if (props.isPartialView && payload.view && payload.view.scrollDOM) {
      nextTick(() => {
        payload.view.scrollDOM.scrollTop = 0;
      });
    }
  };

  const handleContentDoubleClick = () => {
    emit('contentDoubleClick');
  };

  // 监听文件变化，重置滚动位置
  watch(
    () => props.file,
    () => {
      if (editorView.value) {
        nextTick(() => {
          if (editorView.value?.scrollDOM) {
            editorView.value.scrollDOM.scrollTop = 0;
          }
        });
      }
    }
  );

  watch(
    () => editorView.value,
    (view, oldView) => {
      if (oldView?.dom) {
        oldView.dom.removeEventListener('dblclick', handleContentDoubleClick);
      }
      if (view?.dom) {
        view.dom.addEventListener('dblclick', handleContentDoubleClick);
      }
    }
  );

  onUnmounted(() => {
    if (editorView.value?.dom) {
      editorView.value.dom.removeEventListener(
        'dblclick',
        handleContentDoubleClick
      );
    }
  });

  // 暴露编辑器实例
  defineExpose({
    editorView: computed(() => editorView.value),
    focus: () => {
      editorView.value?.focus();
    },
    scrollToTop: () => {
      if (editorView.value) {
        editorView.value.dispatch({
          effects: EditorView.scrollIntoView(0),
        });
      }
    },
    scrollToBottom: () => {
      if (editorView.value) {
        const { state } = editorView.value;
        editorView.value.dispatch({
          effects: EditorView.scrollIntoView(state.doc.length),
        });
      }
    },
  });
</script>

<style scoped>
  .code-editor {
    position: relative;
    display: flex;
    flex: 1;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  .codemirror-wrapper {
    position: relative;
    flex: 1;
    overflow: hidden;
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

  :deep(.cm-focused) {
    outline: none;
  }

  :deep(.cm-editor.cm-focused) {
    outline: none;
  }

  .loading-container {
    position: absolute;
    top: 0;
    left: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    background-color: rgb(255 255 255 / 80%);
    backdrop-filter: blur(2px);
  }

  .loading-text {
    margin-top: 8px;
    font-size: 14px;
    color: var(--color-text-2);
  }

  /* 深色模式适配 */
  @media (prefers-color-scheme: dark) {
    .loading-container {
      background-color: var(--color-bg-1);
      opacity: 0.6;
    }
  }

  /* 暗黑模式下的编辑器样式 */
  body[arco-theme='dark'] {
    :deep(.cm-editor) {
      color: var(--color-text-1);
      background-color: var(--color-bg-2);
    }
    :deep(.cm-gutters) {
      background-color: var(--color-bg-3);
      border-right: 1px solid var(--color-border-2);
    }
    :deep(.cm-lineNumbers .cm-gutterElement) {
      color: var(--color-text-3);
    }
    :deep(.cm-activeLineGutter) {
      background-color: var(--color-bg-3);
    }
    :deep(.cm-activeLine) {
      background-color: var(--color-fill-1);
    }
    :deep(.cm-content) {
      color: var(--color-text-1);
      background-color: var(--color-bg-2);
    }
    :deep(.cm-cursor) {
      border-left-color: var(--color-text-1);
    }
    :deep(.cm-selectionBackground) {
      background-color: var(--color-fill-3) !important;
    }
    .loading-container {
      background-color: rgb(var(--color-bg-1) / 80%);
    }
  }
</style>
