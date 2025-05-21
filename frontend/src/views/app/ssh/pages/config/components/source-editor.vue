<template>
  <div class="source-editor-wrapper">
    <div class="source-toolbar">
      <a-button
        type="primary"
        :disabled="!hasChanges || loading"
        :loading="loading"
        @click="handleSave"
      >
        {{ $t('app.ssh.source.save') }}
      </a-button>
      <a-button class="ml-2" :disabled="loading" @click="handleReset">
        {{ $t('app.ssh.source.reset') }}
      </a-button>
    </div>
    <div class="editor-container">
      <codemirror
        v-model="localConfig"
        :placeholder="$t('app.ssh.source.placeholder')"
        :indent-with-tab="true"
        :tabSize="2"
        :extensions="extensions"
        class="source-editor"
        :disabled="loading"
      />
    </div>
    <div class="source-info">
      {{ $t('app.ssh.source.info') }}
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch, defineProps, defineEmits, computed } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { StreamLanguage } from '@codemirror/language';
  import { shell } from '@codemirror/legacy-modes/mode/shell';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();

  const props = defineProps({
    config: {
      type: String,
      required: true,
    },
    originalConfig: {
      type: String,
      required: true,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  });

  const emits = defineEmits(['update:config', 'save', 'reset', 'update-form']);

  const localConfig = ref(props.config);

  const extensions = [StreamLanguage.define(shell), oneDark];

  // 计算是否有未保存的更改
  const hasChanges = computed(() => {
    return localConfig.value !== props.originalConfig;
  });

  // 监听 props.config 变化，更新 localConfig
  watch(
    () => props.config,
    (val) => {
      localConfig.value = val;
    }
  );

  // 监听 localConfig 变化，向父组件同步
  watch(localConfig, (val) => {
    emits('update:config', val);
  });

  // 处理保存按钮点击
  const handleSave = () => {
    if (!localConfig.value.trim()) {
      Message.warning(t('app.ssh.source.emptyConfig'));
      return;
    }
    emits('save');
  };

  // 处理重置按钮点击
  const handleReset = () => {
    localConfig.value = props.originalConfig;
    emits('reset');
  };

  // 检查是否有未保存的更改的公共方法
  const checkUnsavedChanges = () => {
    return hasChanges.value;
  };

  defineExpose({
    checkUnsavedChanges,
  });
</script>

<style scoped lang="less">
  .source-editor-wrapper {
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
    display: flex;
    flex-direction: column;
    height: 500px;
  }

  .source-toolbar {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-border-2);
    flex-shrink: 0;

    .ml-2 {
      margin-left: 8px;
    }
  }

  .editor-container {
    flex: 1;
    overflow: hidden;
    min-height: 0;
    position: relative;
  }

  .source-editor {
    width: 100%;
    height: 100%;
    font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', 'source-code-pro',
      monospace;
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
  }

  .source-editor :deep(.cm-editor) {
    height: 100%;
  }

  .source-editor :deep(.cm-scroller) {
    overflow: auto;
    padding-bottom: 24px;
  }

  .source-info {
    padding: 8px 12px;
    color: var(--color-text-3);
    font-size: 12px;
    border-top: 1px solid var(--color-border-2);
    flex-shrink: 0;
  }
</style>
