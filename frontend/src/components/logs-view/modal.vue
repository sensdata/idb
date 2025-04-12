<template>
  <a-modal
    v-model:visible="visible"
    :title="title"
    :footer="false"
    :width="width"
  >
    <logs-view :content="contentRef" />
  </a-modal>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import LogsView from './index.vue';

  defineProps({
    width: {
      type: Number,
      default: 1000,
    },
    title: {
      type: String,
      default: '',
    },
  });

  const visible = ref(false);
  const contentRef = ref('');
  function setContent(content: string) {
    contentRef.value = content;
  }
  function show() {
    visible.value = true;
  }

  defineExpose({
    show,
    setContent,
  });
</script>
