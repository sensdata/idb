<template>
  <div class="status-panel">
    <div class="status-panel__toolbar">
      <div>
        <div class="status-panel__title">
          {{ t('manage.settings.status.title') }}
        </div>
        <div class="status-panel__subtitle">
          {{ t('manage.settings.status.subtitle') }}
        </div>
      </div>
      <a-space>
        <a-tag :color="status?.checks.database_connected ? 'green' : 'red'">
          {{
            status?.checks.database_connected
              ? t('manage.settings.status.dbOk')
              : t('manage.settings.status.dbError')
          }}
        </a-tag>
        <a-tag :color="agentTagColor">
          {{ agentTagText }}
        </a-tag>
        <a-button size="small" :loading="loading" @click="fetchStatus">
          {{ t('manage.settings.status.refresh') }}
        </a-button>
      </a-space>
    </div>

    <a-spin :loading="loading">
      <div class="status-grid">
        <a-card
          class="status-card"
          :title="t('manage.settings.status.centerCard')"
        >
          <a-descriptions :column="2" bordered size="large">
            <a-descriptions-item :label="t('manage.settings.status.version')">
              {{ status?.center.version || '-' }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.pid')">
              {{ status?.center.pid || '-' }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.startedAt')">
              {{ formatTime(status?.center.started_at) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.uptime')">
              {{ formatSeconds(status?.center.uptime) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.cpu')">
              {{ formatPercent(status?.center.cpu_percent) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.cpuTime')">
              {{ formatCpuSeconds(status?.center.cpu_seconds) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.memoryRss')">
              {{ formatMemorySize(status?.center.memory_rss) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.heapAlloc')">
              {{ formatMemorySize(status?.center.heap_alloc) }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.heapSys')">
              {{ formatMemorySize(status?.center.heap_sys) }}
            </a-descriptions-item>
            <a-descriptions-item
              :label="t('manage.settings.status.stackInuse')"
            >
              {{ formatMemorySize(status?.center.stack_inuse) }}
            </a-descriptions-item>
            <a-descriptions-item
              :label="t('manage.settings.status.goroutines')"
            >
              {{ status?.center.goroutines ?? '-' }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.openFds')">
              {{ status?.center.open_fds ?? '-' }}
            </a-descriptions-item>
            <a-descriptions-item
              :label="t('manage.settings.status.listenAddress')"
            >
              {{ centerListenAddress }}
            </a-descriptions-item>
            <a-descriptions-item :label="t('manage.settings.status.accessUrl')">
              <a-link
                v-if="status?.center.access_url"
                :href="status.center.access_url"
                target="_blank"
              >
                {{ status.center.access_url }}
              </a-link>
              <span v-else>-</span>
            </a-descriptions-item>
          </a-descriptions>
        </a-card>

        <a-card
          class="status-card"
          :title="t('manage.settings.status.agentCard')"
        >
          <template v-if="status?.agent">
            <a-descriptions :column="2" bordered size="large">
              <a-descriptions-item
                :label="t('manage.settings.status.hostName')"
              >
                {{ status.agent.host_name || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.agentVersion')"
              >
                {{ status.agent.agent_version || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.agentEndpoint')"
              >
                {{ status.agent.agent_addr }}:{{ status.agent.agent_port }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.connectivity')"
              >
                <a-tag :color="agentConnectivityColor">
                  {{ agentConnectivityText }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.lastHeartbeat')"
              >
                {{ formatTime(status.agent.last_heartbeat) }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.freshness')"
              >
                <a-tag
                  :color="
                    status.checks.default_agent_fresh ? 'green' : 'orange'
                  "
                >
                  {{
                    status.checks.default_agent_fresh
                      ? t('manage.settings.status.fresh')
                      : t('manage.settings.status.stale')
                  }}
                </a-tag>
              </a-descriptions-item>
              <a-descriptions-item :label="t('manage.settings.status.cpu')">
                {{ formatPercent(status.agent.cpu) }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('manage.settings.status.memory')">
                {{ formatPercent(status.agent.memory) }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.memoryUsed')"
              >
                {{ status.agent.mem_used || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.memoryTotal')"
              >
                {{ status.agent.mem_total || '-' }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('manage.settings.status.disk')">
                {{ formatPercent(status.agent.disk) }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.hostAddr')"
              >
                {{ status.agent.host_addr || '-' }}
              </a-descriptions-item>
              <a-descriptions-item
                :label="t('manage.settings.status.bootTime')"
              >
                {{ formatTime(status.agent.boot_time) }}
              </a-descriptions-item>
              <a-descriptions-item :label="t('manage.settings.status.uptime')">
                {{ formatSeconds(status.agent.run_time) }}
              </a-descriptions-item>
            </a-descriptions>
          </template>
          <a-empty v-else :description="t('manage.settings.status.noAgent')" />
        </a-card>
      </div>

      <div class="status-panel__footer">
        {{ t('manage.settings.status.collectedAt') }}:
        {{ formatTime(status?.collected_at) }}
      </div>
    </a-spin>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { getSettingsStatusApi, type SettingsStatus } from '@/api/settings';
  import { formatMemorySize, formatSeconds, formatTime } from '@/utils/format';

  const { t } = useI18n();

  const loading = ref(false);
  const status = ref<SettingsStatus>();
  let timer: number | undefined;

  const fetchStatus = async () => {
    try {
      loading.value = true;
      status.value = await getSettingsStatusApi();
    } catch {
      Message.error(t('manage.settings.status.loadFailed'));
    } finally {
      loading.value = false;
    }
  };

  const formatPercent = (value?: number) => {
    if (value == null) {
      return '-';
    }
    return `${value.toFixed(2)}%`;
  };

  const formatCpuSeconds = (value?: number) => {
    if (value == null) {
      return '-';
    }
    return `${value.toFixed(2)}s`;
  };

  const centerListenAddress = computed(() => {
    if (!status.value) {
      return '-';
    }
    const protocol = status.value.center.https_enabled ? 'https' : 'http';
    return `${protocol}://${status.value.center.bind_ip}:${status.value.center.bind_port}`;
  });

  const agentConnectivityColor = computed(() =>
    status.value?.agent?.connected === 'online' ? 'green' : 'red'
  );

  const agentConnectivityText = computed(() =>
    status.value?.agent?.connected === 'online'
      ? t('manage.settings.status.online')
      : t('manage.settings.status.offline')
  );

  const agentTagColor = computed(() => {
    if (!status.value?.checks.default_host_found) {
      return 'gray';
    }
    return status.value.checks.default_agent_online ? 'green' : 'red';
  });

  const agentTagText = computed(() => {
    if (!status.value?.checks.default_host_found) {
      return t('manage.settings.status.noDefaultHost');
    }
    return status.value.checks.default_agent_online
      ? t('manage.settings.status.agentOk')
      : t('manage.settings.status.agentError');
  });

  onMounted(() => {
    fetchStatus();
    timer = window.setInterval(() => {
      fetchStatus();
    }, 15000);
  });

  onUnmounted(() => {
    if (timer) {
      window.clearInterval(timer);
    }
  });
</script>

<style scoped lang="less">
  .status-panel {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .status-panel__toolbar {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    justify-content: space-between;
  }

  .status-panel__title {
    font-size: 16px;
    font-weight: 600;
    color: rgb(var(--gray-10));
  }

  .status-panel__subtitle {
    margin-top: 4px;
    font-size: 13px;
    color: rgb(var(--gray-6));
  }

  .status-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
  }

  .status-card {
    min-height: 100%;
  }

  .status-panel__footer {
    font-size: 13px;
    color: rgb(var(--gray-6));
  }

  @media (width <= 900px) {
    .status-panel__toolbar {
      flex-direction: column;
      align-items: stretch;
    }
    .status-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
