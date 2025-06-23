<template>
  <div class="source-mode">
    <source-editor
      ref="sourceEditorRef"
      v-model:config="modelValue"
      :original-config="originalConfig"
      :loading="loading"
      @save="$emit('save')"
      @reset="$emit('reset')"
      @update-form="$emit('updateForm')"
    />

    <div v-if="loading" class="source-loading-overlay">
      <a-spin :size="36" />
      <span class="loading-text">{{ $t('app.ssh.savingConfig') }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { defineProps, defineEmits, defineExpose, ref, computed } from 'vue';
  import { useLogger } from '@/composables/use-logger';
  import SourceEditor from './source-editor.vue';
  import { EditorRefType } from '../types';

  const { logError } = useLogger('SourceConfigPanel');

  const props = defineProps<{
    config: string;
    originalConfig: string;
    loading: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:config', value: string): void;
    (e: 'save'): void;
    (e: 'reset'): void;
    (e: 'updateForm'): void;
  }>();

  // 计算属性，用于双向绑定配置数据
  const modelValue = computed({
    get: () => props.config,
    set: (value: string) => emit('update:config', value),
  });

  // 引用编辑器组件的实例
  const sourceEditorRef = ref<EditorRefType | null>(null);

  // 暴露检查未保存更改的方法给父组件
  defineExpose({
    checkUnsavedChanges: () => {
      try {
        if (
          sourceEditorRef.value &&
          typeof sourceEditorRef.value.checkUnsavedChanges === 'function'
        ) {
          return sourceEditorRef.value.checkUnsavedChanges();
        }
        return false;
      } catch (error: unknown) {
        logError(
          '检查未保存更改时出错:',
          error instanceof Error ? error.message : error
        );
        return false;
      }
    },
  });
</script>

<style scoped lang="less">
  .source-mode {
    position: relative;
    margin-top: 16px;
  }

  .source-loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 255, 255, 0.8);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    padding: 20px;
    z-index: 10;

    .loading-text {
      margin-top: 16px;
      color: var(--color-text-2);
      font-size: 14px;
    }
  }
</style>
