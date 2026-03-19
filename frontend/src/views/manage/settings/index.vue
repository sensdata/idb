<template>
  <div class="settings-page">
    <a-tabs v-model:active-key="activeTab" lazy-load>
      <a-tab-pane key="basic" :title="t('manage.settings.tabs.basic')">
        <SettingsForm />
      </a-tab-pane>
      <a-tab-pane key="status" :title="t('manage.settings.tabs.status')">
        <StatusPanel />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import SettingsForm from './components/settings-form.vue';
  import StatusPanel from './components/status-panel.vue';

  const { t } = useI18n();
  const activeTab = ref('basic');
</script>

<style scoped lang="less">
  .settings-page {
    :deep(.arco-tabs-content) {
      padding-top: 20px;
    }
  }
</style>
