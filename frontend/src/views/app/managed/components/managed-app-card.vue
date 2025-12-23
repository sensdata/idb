<template>
  <a-card hoverable class="managed-app-card">
    <div class="card-content">
      <!-- 头部：图标和名称 -->
      <div class="card-header">
        <a-avatar
          shape="square"
          class="app-avatar"
          :size="56"
          :style="{
            backgroundColor: getHexColorByChar(app.display_name),
          }"
        >
          {{ app.display_name.charAt(0) }}
        </a-avatar>
        <div class="app-info">
          <h3 class="app-name">{{ app.display_name }}</h3>
          <div class="app-version">
            <a-tag bordered size="small">
              {{ $t('app.managed.card.version') }}: {{ app.current_version }}
            </a-tag>
          </div>
        </div>
        <div class="app-status">
          <a-tag :color="statusColor" size="small">
            <template #icon>
              <icon-check-circle v-if="containerStatus === 'running'" />
              <icon-loading v-else-if="containerStatus === 'restarting'" />
              <icon-close-circle v-else />
            </template>
            {{ statusText }}
          </a-tag>
        </div>
      </div>

      <!-- 资源使用（固定高度区域） -->
      <div class="resource-usage-wrapper">
        <div v-if="containerStatus === 'running'" class="resource-usage">
          <div class="usage-item">
            <span class="usage-label">CPU:</span>
            <a-progress
              :percent="cpuPercentDisplay"
              :stroke-width="6"
              :show-text="false"
              size="small"
              :status="cpuPercentDisplay > 80 ? 'danger' : 'normal'"
            />
            <span class="usage-value">{{ cpuPercentDisplay.toFixed(2) }}%</span>
          </div>
          <div class="usage-item">
            <span class="usage-label">
              {{ $t('app.managed.card.memory') }}:
            </span>
            <a-progress
              :percent="memoryPercentDisplay"
              :stroke-width="6"
              :show-text="false"
              size="small"
              :status="memoryPercentDisplay > 80 ? 'danger' : 'normal'"
            />
            <span class="usage-value">
              {{ formatMemorySize(memoryUsage) }} /
              {{ formatMemorySize(memoryLimit) }}
            </span>
          </div>
        </div>
        <div v-else class="resource-usage-placeholder">
          <span class="placeholder-text">
            {{ $t('app.managed.card.notRunning') }}
          </span>
        </div>
      </div>

      <!-- 操作按钮 -->
      <div class="card-actions">
        <a-button
          v-if="isDatabaseApp"
          type="primary"
          size="small"
          :disabled="!isRunning"
          @click="$emit('manage', app)"
        >
          <template #icon><icon-settings /></template>
          {{ $t('app.managed.card.manage') }}
        </a-button>
        <a-button
          size="small"
          :disabled="!isRunning"
          @click="$emit('logs', app, containerName)"
        >
          <template #icon><icon-file /></template>
          {{ $t('app.managed.card.logs') }}
        </a-button>
        <a-dropdown trigger="click">
          <a-button size="small">
            <template #icon><icon-more /></template>
          </a-button>
          <template #content>
            <a-doption
              v-if="containerStatus !== 'running'"
              @click="$emit('operate', containerName, 'start')"
            >
              <template #icon><icon-play-arrow /></template>
              {{ $t('app.managed.card.start') }}
            </a-doption>
            <a-doption
              v-if="containerStatus === 'running'"
              @click="$emit('operate', containerName, 'stop')"
            >
              <template #icon><icon-pause /></template>
              {{ $t('app.managed.card.stop') }}
            </a-doption>
            <a-doption
              v-if="containerStatus === 'running'"
              @click="$emit('operate', containerName, 'restart')"
            >
              <template #icon><icon-refresh /></template>
              {{ $t('app.managed.card.restart') }}
            </a-doption>
          </template>
        </a-dropdown>
      </div>
    </div>
  </a-card>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    IconCheckCircle,
    IconCloseCircle,
    IconLoading,
    IconSettings,
    IconFile,
    IconMore,
    IconPlayArrow,
    IconPause,
    IconRefresh,
  } from '@arco-design/web-vue/es/icon';
  import { AppSimpleEntity } from '@/entity/App';
  import { getHexColorByChar } from '@/helper/utils';
  import { formatMemorySize } from '@/utils/format';

  const { t } = useI18n();

  const props = defineProps<{
    app: AppSimpleEntity;
    containerInfo?: {
      container_id: string;
      name: string;
      state: string;
      cpu_percent?: number;
      memory_usage?: number;
      memory_limit?: number;
    };
  }>();

  defineEmits<{
    manage: [app: AppSimpleEntity];
    logs: [app: AppSimpleEntity, containerName: string];
    operate: [containerName: string, operation: 'start' | 'stop' | 'restart'];
  }>();

  const containerName = computed(
    () => props.containerInfo?.name || props.app.name
  );
  const containerStatus = computed(
    () => props.containerInfo?.state || 'unknown'
  );
  const cpuPercent = computed(() => props.containerInfo?.cpu_percent || 0);
  const memoryUsage = computed(() => props.containerInfo?.memory_usage || 0);
  const memoryLimit = computed(() => props.containerInfo?.memory_limit || 0);
  const memoryPercent = computed(() => {
    if (!memoryLimit.value) return 0;
    return (memoryUsage.value / memoryLimit.value) * 100;
  });

  // 格式化显示的百分比（保留2位小数）
  const cpuPercentDisplay = computed(() => {
    return Math.round(cpuPercent.value * 100) / 100;
  });
  const memoryPercentDisplay = computed(() => {
    return Math.round(memoryPercent.value * 100) / 100;
  });

  const statusColor = computed(() => {
    switch (containerStatus.value) {
      case 'running':
        return 'green';
      case 'exited':
      case 'dead':
        return 'red';
      case 'paused':
      case 'restarting':
        return 'orange';
      default:
        return 'gray';
    }
  });

  const statusText = computed(() => {
    const statusMap: Record<string, string> = {
      running: t('app.managed.card.status.running'),
      exited: t('app.managed.card.status.exited'),
      paused: t('app.managed.card.status.paused'),
      restarting: t('app.managed.card.status.restarting'),
      created: t('app.managed.card.status.created'),
      dead: t('app.managed.card.status.dead'),
    };
    return (
      statusMap[containerStatus.value] || t('app.managed.card.status.unknown')
    );
  });

  const isRunning = computed(() => containerStatus.value === 'running');

  const isDatabaseApp = computed(() => {
    const databaseTypes = ['mysql', 'mariadb', 'postgresql', 'redis'];
    return databaseTypes.some((type) =>
      props.app.name.toLowerCase().includes(type)
    );
  });
</script>

<style scoped lang="less">
  .managed-app-card {
    height: 100%;
    :deep(.arco-card-body) {
      height: 100%;
      padding: 1.143rem;
    }
  }

  .card-content {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .card-header {
    display: flex;
    gap: 0.857rem;
    align-items: flex-start;
  }

  .app-avatar {
    flex-shrink: 0;
    font-size: 1.714rem;
    font-weight: 600;
  }

  .app-info {
    flex: 1;
    min-width: 0;
  }

  .app-name {
    margin: 0 0 0.286rem;
    font-size: 1.143rem;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .app-version {
    :deep(.arco-tag) {
      font-size: 0.857rem;
    }
  }

  .app-status {
    flex-shrink: 0;
  }

  .resource-usage-wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: center;
    min-height: 5.714rem;
    margin: 1.143rem 0;
  }

  .resource-usage {
    display: flex;
    flex-direction: column;
    gap: 0.571rem;
    padding: 0.857rem;
    background: var(--color-fill-1);
    border-radius: 0.429rem;
  }

  .resource-usage-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 4rem;
    padding: 0.857rem;
    background: var(--color-fill-1);
    border-radius: 0.429rem;
  }

  .placeholder-text {
    font-size: 0.929rem;
    color: var(--color-text-3);
  }

  .usage-item {
    display: flex;
    gap: 0.571rem;
    align-items: center;
  }

  .usage-label {
    width: 3.571rem;
    font-size: 0.857rem;
    color: var(--color-text-3);
  }

  .usage-value {
    width: 7.143rem;
    font-size: 0.857rem;
    color: var(--color-text-2);
    text-align: right;
  }

  :deep(.arco-progress) {
    flex: 1;
  }

  .card-actions {
    display: flex;
    gap: 0.571rem;
    padding-top: 0.857rem;
    margin-top: auto;
    border-top: 1px solid var(--color-border-1);
  }
</style>
