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
  import { shallowRef, nextTick, computed, watch } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { EditorView } from '@codemirror/view';
  import { basicSetup } from 'codemirror';
  import { IconLoading } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
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
  const editorView = shallowRef<EditorView>();

  // 使用编辑器配置
  const { extensions: configExtensions } = useEditorConfig(
    computed(() => props.file)
  );

  // 合并扩展 - 添加 basicSetup 作为基础配置
  const computedExtensions = computed(() => {
    return [basicSetup, ...configExtensions.value, ...props.extensions];
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
</style>
