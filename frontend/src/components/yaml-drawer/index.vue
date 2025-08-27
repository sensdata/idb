<template>
  <a-drawer
    v-model:visible="visible"
    :width="700"
    :title="title"
    :footer="false"
    :body-style="{ padding: '0 16px' }"
  >
    <div style="width: 100%; height: calc(100vh - 80px)">
      <code-editor
        v-model="yamlContent"
        :autofocus="true"
        :indent-with-tab="true"
        :tab-size="4"
        :extensions="yamlExtensions"
        :file="yamlFile"
      />
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import CodeEditor from '@/components/code-editor/index.vue';

  defineProps<{
    title: string;
  }>();
  const visible = defineModel<boolean>('visible', {
    default: false,
  });

  const yamlExtensions = [StreamLanguage.define(yaml)];

  // 创建一个虚拟的文件对象，用于让 code-editor 识别为 YAML 文件
  const yamlFile = computed(() => ({
    name: 'config.yaml',
    path: '/tmp/config.yaml',
  }));

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
