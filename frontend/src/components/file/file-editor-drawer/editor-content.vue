<template>
  <div class="editor-content">
    <div class="codemirror-wrapper">
      <codemirror
        :model-value="modelValue"
        :autofocus="true"
        :indent-with-tab="true"
        :tab-size="2"
        :extensions="extensions"
        :disabled="readOnly"
        @update:model-value="$emit('update:modelValue', $event)"
        @ready="handleReady"
      />
    </div>

    <!-- 文件加载中 -->
    <div v-if="loading && !modelValue" class="loading-container">
      <a-spin :size="24">
        <template #icon>
          <icon-loading />
        </template>
        <template #tip>
          <span class="loading-text">加载文件内容中...</span>
        </template>
      </a-spin>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { shallowRef, nextTick } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { Extension } from '@codemirror/state';
  import { IconLoading } from '@arco-design/web-vue/es/icon';

  const props = defineProps({
    modelValue: {
      type: String,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
    extensions: {
      type: Array as () => Extension[],
      required: true,
    },
    isPartialView: {
      type: Boolean,
      default: false,
    },
    readOnly: {
      type: Boolean,
      default: true,
    },
  });

  const emit = defineEmits(['update:modelValue', 'editorReady']);

  const editorView = shallowRef();

  const handleReady = (payload: any) => {
    editorView.value = payload.view;

    // 将编辑器实例传递给父组件
    emit('editorReady', payload.view);

    // 确保编辑器初始化后，如果处于部分查看模式，滚动到顶部
    if (props.isPartialView && payload.view && payload.view.scrollDOM) {
      nextTick(() => {
        payload.view.scrollDOM.scrollTop = 0;
      });
    }
  };
</script>

<style scoped>
  .editor-content {
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
  }

  .loading-text {
    margin-top: 8px;
    color: var(--color-text-2);
    font-size: 14px;
  }
</style>
