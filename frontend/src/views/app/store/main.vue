<template>
  <a-spin :loading="loading" class="w-full">
    <!-- Docker 环境检测 -->
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />

    <a-tabs v-model:active-key="activeTab" lazy-load destroy-on-hide>
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
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { syncAppApi } from '@/api/store';
  import DockerInstallGuide from '@/components/docker-install-guide/index.vue';
  import List from './list.vue';
  import InstalledList from './installed-list.vue';

  const { t } = useI18n();
  const listRef = ref<InstanceType<typeof List>>();

  const loading = ref(false);
  const handleSync = async () => {
    try {
      loading.value = true;
      await syncAppApi();
      Message.success(t('app.store.app.message.syncSuccess'));
      listRef.value?.load();
    } catch (err: any) {
      await showErrorWithDockerCheck(err.message, err);
    } finally {
      loading.value = false;
    }
  };

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 可以根据状态变化做一些处理，比如刷新应用列表等
    if (status === 'installed') {
      // Docker 安装完成后可以执行相关操作
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // 可以刷新页面或重新加载数据
  };

  const activeTab = ref('ALL');
</script>
