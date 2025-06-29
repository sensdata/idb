<template>
  <a-drawer
    v-model:visible="visible"
    :width="700"
    :title="title"
    :footer="false"
    :body-style="{ padding: '0 16px' }"
  >
    <codemirror
      :model-value="yamlContent"
      theme="cobalt"
      :tab-size="4"
      style="width: 100%; height: calc(100vh - 80px)"
      :extensions="yamlExtensions"
      autofocus
      indent-with-tab
      line-wrapping
      match-brackets
      style-active-line
    />
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import { oneDark } from '@codemirror/theme-one-dark';

  defineProps<{
    title: string;
  }>();
  const visible = defineModel<boolean>('visible', {
    default: false,
  });

  const yamlExtensions = [StreamLanguage.define(yaml), oneDark];

  const yamlContent = ref('');
  defineExpose({
    show: () => {
      visible.value = true;
    },
    hide: () => {
      visible.value = false;
    },
    setContent: (content: string) => {
      yamlContent.value = content;
    },
  });
</script>
