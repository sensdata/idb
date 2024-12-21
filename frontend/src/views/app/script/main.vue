<template>
  <a-tabs v-model:active-key="activeTab" lazy-load>
    <a-tab-pane
      :key="SCRIPT_TYPE.Local"
      :title="$t('app.script.enum.type.local')"
    >
      <list :type="SCRIPT_TYPE.Local" />
    </a-tab-pane>
    <a-tab-pane
      :key="SCRIPT_TYPE.Global"
      :title="$t('app.script.enum.type.global')"
    >
      <list :type="SCRIPT_TYPE.Global" />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { SCRIPT_TYPE } from '@/config/enum';
  import { ref } from 'vue';
  import List from './list.vue';

  const activeTab = ref(SCRIPT_TYPE.Local);
</script>
