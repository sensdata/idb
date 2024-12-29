<template>
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
  />
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { javascript } from '@codemirror/lang-javascript';
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
  const emit = defineEmits(['update:modelValue']);

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

  const extensions = [javascript(), oneDark];
</script>
