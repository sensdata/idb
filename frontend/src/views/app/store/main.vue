<template>
  <a-spin :loading="loading" class="w-full">
    <a-tabs v-model:active-key="activeTab" lazy-load>
      <template #extra>
        <a-button type="text" size="small" @click="handleSync">
          {{ $t('app.store.app.syncAppList') }}
        </a-button>
      </template>
      <a-tab-pane key="ALL" :title="$t('app.store.app.tabs.all')">
        <list ref="listRef" />
      </a-tab-pane>
      <a-tab-pane key="installed" :title="$t('app.store.app.tabs.installed')">
        <installed-list />
      </a-tab-pane>
    </a-tabs>
  </a-spin>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { syncAppApi } from '@/api/store';
  import List from './list.vue';
  import InstalledList from './installed-list.vue';

  const listRef = ref<InstanceType<typeof List>>();

  const loading = ref(false);
  const handleSync = async () => {
    try {
      loading.value = true;
      await syncAppApi();
      Message.success('app.store.app.message.syncSuccess');
      listRef.value?.load();
    } catch (err: any) {
      Message.error(err.message);
    } finally {
      loading.value = false;
    }
  };

  const activeTab = ref('ALL');
</script>
