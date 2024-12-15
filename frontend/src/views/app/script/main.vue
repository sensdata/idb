<template>
  <a-tabs v-model:active-key="activeTab" lazy-load>
    <a-tab-pane
      :key="ScriptType.Local"
      :title="$t('app.script.enum.type.local')"
    >
      <list :type="ScriptType.Local" />
    </a-tab-pane>
    <a-tab-pane
      :key="ScriptType.Global"
      :title="$t('app.script.enum.type.global')"
    >
      <list :type="ScriptType.Global" />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { ScriptType } from '@/config/enum';
  import { ref } from 'vue';
  import List from './list.vue';

  const activeTab = ref(ScriptType.Local);
</script>
