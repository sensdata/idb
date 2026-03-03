<template>
  <a-drawer
    v-model:visible="visible"
    :width="900"
    :title="t('app.docker.image.inspect.title')"
    :footer="false"
  >
    <a-tabs v-model:active-key="activeTab">
      <a-tab-pane
        key="friendly"
        :title="t('app.docker.image.inspect.tab.friendly')"
      >
        <div class="inspect-grid">
          <a-card
            size="small"
            :title="t('app.docker.image.inspect.section.basic')"
          >
            <a-descriptions :column="1" size="small">
              <a-descriptions-item label="ID">{{
                imageId || '-'
              }}</a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.image.list.column.created')"
              >
                {{ createdAt || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.image.list.column.size')"
              >
                {{ imageSize || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.image.inspect.field.os')"
              >
                {{ os || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('app.docker.image.inspect.field.architecture')"
              >
                {{ architecture || '-' }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <a-card size="small" :title="t('app.docker.image.list.column.tags')">
            <a-space wrap>
              <a-tag v-for="tag in repoTags" :key="tag" bordered>{{
                tag
              }}</a-tag>
              <span v-if="!repoTags.length">-</span>
            </a-space>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.image.inspect.section.command')"
          >
            <a-descriptions :column="1" size="small">
              <a-descriptions-item label="Entrypoint">
                {{ entrypointText || '-' }}
              </a-descriptions-item>
              <a-descriptions-item label="Cmd">{{
                cmdText || '-'
              }}</a-descriptions-item>
              <a-descriptions-item label="User">{{
                configUser || '-'
              }}</a-descriptions-item>
              <a-descriptions-item label="WorkingDir">
                {{ workingDir || '-' }}
              </a-descriptions-item>
            </a-descriptions>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.image.inspect.section.env')"
          >
            <div class="line-list">
              <div v-for="item in envList" :key="item">{{ item }}</div>
              <div v-if="!envList.length">-</div>
            </div>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.image.inspect.section.labels')"
          >
            <div class="line-list">
              <div v-for="item in labelList" :key="item">{{ item }}</div>
              <div v-if="!labelList.length">-</div>
            </div>
          </a-card>

          <a-card
            size="small"
            :title="t('app.docker.image.inspect.section.layers')"
          >
            <div class="line-list">
              <div v-for="item in layerList" :key="item">{{ item }}</div>
              <div v-if="!layerList.length">-</div>
            </div>
          </a-card>
        </div>
      </a-tab-pane>

      <a-tab-pane key="raw" :title="t('app.docker.image.inspect.tab.raw')">
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
  const inspectData = ref<any>({});

  const config = computed(() => inspectData.value?.Config || {});
  const rootFs = computed(() => inspectData.value?.RootFS || {});

  const imageId = computed(() => inspectData.value?.Id || '');
  const createdAt = computed(() => inspectData.value?.Created || '');
  const imageSize = computed(() =>
    inspectData.value?.Size != null ? `${inspectData.value.Size} B` : ''
  );
  const os = computed(() => inspectData.value?.Os || '');
  const architecture = computed(() => inspectData.value?.Architecture || '');

  const repoTags = computed<string[]>(() =>
    Array.isArray(inspectData.value?.RepoTags) ? inspectData.value.RepoTags : []
  );
  const entrypointText = computed(() =>
    Array.isArray(config.value?.Entrypoint)
      ? config.value.Entrypoint.join(' ')
      : config.value?.Entrypoint || ''
  );
  const cmdText = computed(() =>
    Array.isArray(config.value?.Cmd)
      ? config.value.Cmd.join(' ')
      : config.value?.Cmd || ''
  );
  const configUser = computed(() => config.value?.User || '');
  const workingDir = computed(() => config.value?.WorkingDir || '');

  const envList = computed<string[]>(() =>
    Array.isArray(config.value?.Env) ? config.value.Env : []
  );
  const labelList = computed<string[]>(() => {
    const labels = config.value?.Labels || {};
    return Object.entries(labels).map(([key, value]) => `${key}=${value}`);
  });
  const layerList = computed<string[]>(() =>
    Array.isArray(rootFs.value?.Layers) ? rootFs.value.Layers : []
  );

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
