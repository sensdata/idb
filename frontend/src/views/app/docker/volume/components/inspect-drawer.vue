<template>
  <a-drawer
    v-model:visible="visible"
    :width="900"
    :title="t('app.docker.volume.inspect.title')"
    :footer="false"
  >
    <a-tabs v-model:active-key="activeTab">
      <a-tab-pane
        key="friendly"
        :title="t('app.docker.volume.inspect.tab.friendly')"
      >
        <div class="inspect-grid">
          <a-card
            size="small"
            :title="t('app.docker.volume.inspect.section.basic')"
          >
            <a-descriptions :column="1" size="small">
              <a-descriptions-item
                :label="t('app.docker.volume.list.column.name')"
              >
                {{ volumeName || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.volume.list.column.driver')"
              >
                {{ driver || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.volume.list.column.mount_point')"
              >
                {{ mountpoint || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.volume.inspect.field.scope')"
              >
                {{ scope || '-' }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.volume.inspect.section.options')"
          >
            <div class="line-list">
              <div v-for="item in optionList" :key="item">{{ item }}</div>
              <div v-if="!optionList.length">-</div>
            </div>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.volume.inspect.section.labels')"
          >
            <div class="line-list">
              <div v-for="item in labelList" :key="item">{{ item }}</div>
              <div v-if="!labelList.length">-</div>
            </div>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.volume.inspect.section.usage')"
          >
            <div class="line-list">
              <div v-for="item in usageList" :key="item">{{ item }}</div>
              <div v-if="!usageList.length">-</div>
            </div>
          </a-card>
        </div>
      </a-tab-pane>
      <a-tab-pane key="raw" :title="t('app.docker.volume.inspect.tab.raw')">
        <pre class="raw-content">{{ rawContent }}</pre>
      </a-tab-pane>
    </a-tabs>
  </a-drawer>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  const { t } = useI18n();
  const visible = ref(false);
  const activeTab = ref('friendly');
  const rawContent = ref('');
  const inspectData = ref<Record<string, any>>({});

  const volumeName = computed(() => inspectData.value?.Name || '');
  const driver = computed(() => inspectData.value?.Driver || '');
  const mountpoint = computed(() => inspectData.value?.Mountpoint || '');
  const scope = computed(() => inspectData.value?.Scope || '');

  const optionList = computed<string[]>(() => {
    const options = inspectData.value?.Options || {};
    return Object.entries(options).map(([key, value]) => `${key}=${value}`);
  });

  const labelList = computed<string[]>(() => {
    const labels = inspectData.value?.Labels || {};
    return Object.entries(labels).map(([key, value]) => `${key}=${value}`);
  });

  const usageList = computed<string[]>(() => {
    const usageData = inspectData.value?.UsageData || {};
    return Object.entries(usageData).map(([key, value]) => `${key}: ${value}`);
  });

  const show = (content: string) => {
    rawContent.value = content;
    try {
      const parsed = JSON.parse(content);
      inspectData.value = parsed || {};
      rawContent.value = JSON.stringify(parsed, null, 2);
      activeTab.value = 'friendly';
    } catch {
      inspectData.value = {};
      activeTab.value = 'raw';
    }
    visible.value = true;
  };

  defineExpose({
    show,
    hide: () => {
      visible.value = false;
    },
  });
</script>

<style scoped>
  .inspect-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .line-list {
    max-height: 200px;
    overflow: auto;
    font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.5;
    word-break: break-all;
    white-space: pre-wrap;
  }

  .raw-content {
    max-height: calc(100vh - 220px);
    padding: 12px;
    margin: 0;
    overflow: auto;
    font-family: Menlo, Monaco, Consolas, 'Courier New', monospace;
    font-size: 12px;
    line-height: 1.5;
    word-break: break-all;
    white-space: pre-wrap;
    background: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  @media (width <= 768px) {
    .inspect-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
