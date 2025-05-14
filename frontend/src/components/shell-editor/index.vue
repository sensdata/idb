<template>
  <div class="shell-editor-container">
    <codemirror
      v-model="value"
      :autofocus="true"
      :indent-with-tab="true"
      :tabSize="4"
      :lineWrapping="true"
      :matchBrackets="true"
      theme="cobalt"
      :styleActiveLine="true"
      :extensions="extensions"
      :disabled="readonly"
      @focus="handleFocus"
      @blur="handleBlur"
    />
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { StreamLanguage } from '@codemirror/language';
  import { shell } from '@codemirror/legacy-modes/mode/shell';
  import { oneDark } from '@codemirror/theme-one-dark';

  const props = defineProps({
    readonly: {
      type: Boolean,
      default: true,
    },
    modelValue: {
      type: String,
      default: '',
    },
    defaultValue: {
      type: String,
      default: '',
    },
  });
  const emit = defineEmits(['update:modelValue', 'focus', 'blur']);

  const value = ref(props.modelValue || props.defaultValue);
  watch(
    () => props.modelValue,
    (newVal: string) => {
      value.value = newVal;
    },
    {
      immediate: true,
    }
  );

  watch(
    () => props.defaultValue,
    (newVal: string) => {
      if (!value.value) {
        value.value = newVal;
      }
    },
    {
      immediate: true,
    }
  );

  watch(value, (newVal: string) => {
    emit('update:modelValue', newVal);
  });

  const handleFocus = () => {
    emit('focus', value.value);
  };

  const handleBlur = () => {
    emit('blur', value.value);
  };

  const extensions = [StreamLanguage.define(shell), oneDark];
</script>

<style scoped>
  .shell-editor-container {
    min-width: 100%;
    min-height: 150px;
    border: 1px solid var(--color-border);
    border-radius: 2px;
  }

  .shell-editor-container :deep(.cm-editor) {
    height: 100%;
    min-height: 150px;
  }

  .shell-editor-container :deep(.cm-scroller) {
    min-height: 150px;
  }
</style>
