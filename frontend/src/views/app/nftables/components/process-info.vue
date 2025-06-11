<template>
  <div class="process-info">
    <a-tooltip :content="description">
      <div class="process-icon" :class="iconClass">
        <!-- 如果是SVG图标组件，直接渲染组件 -->
        <component :is="icon" v-if="isSvgIcon(icon)" class="process-svg-icon" />
        <!-- Arco Design图标组件 -->
        <component :is="icon" v-else />
      </div>
    </a-tooltip>
    <div class="process-details">
      <div class="process-name">{{ process }}</div>
      <div v-if="pid" class="process-pid">PID: {{ pid }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useProcessIcons } from '../pages/process-config/hooks/use-process-icons';

  const props = defineProps<{
    process: string;
    pid?: number | string;
  }>();

  const {
    getProcessIcon,
    isSvgIcon,
    getProcessIconClass,
    getProcessTypeDescription,
  } = useProcessIcons();

  const icon = computed(() => getProcessIcon(props.process));
  const iconClass = computed(() => getProcessIconClass(props.process));
  const description = computed(() => getProcessTypeDescription(props.process));
</script>

<style scoped lang="less">
  .process-info {
    display: flex;
    align-items: center;
    gap: 8px;

    .process-icon {
      width: 32px;
      height: 32px;
      border-radius: 4px;
      background: var(--color-primary-light-1);
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--color-primary);
      font-size: 16px;

      // SVG图标组件样式
      .process-svg-icon {
        width: 22px;
        height: 22px;
      }

      // 不同进程类型的颜色
      &.process-icon-database {
        background: rgba(255, 193, 7, 0.1);
        color: #ffc107;
      }

      &.process-icon-server {
        background: rgba(40, 167, 69, 0.1);
        color: #28a745;
      }

      &.process-icon-container {
        background: rgba(0, 123, 255, 0.1);
        color: #007bff;
      }

      &.process-icon-network {
        background: rgba(23, 162, 184, 0.1);
        color: #17a2b8;
      }

      &.process-icon-security {
        background: rgba(220, 53, 69, 0.1);
        color: #dc3545;
      }

      &.process-icon-development {
        background: rgba(111, 66, 193, 0.1);
        color: #6f42c1;
      }

      &.process-icon-system {
        background: rgba(108, 117, 125, 0.1);
        color: #6c757d;
      }

      &.process-icon-default {
        background: var(--color-primary-light-1);
        color: var(--color-primary);
      }
    }

    .process-details {
      flex: 1;

      .process-name {
        font-size: 14px;
        font-weight: 500;
        color: var(--color-text-1);
        margin-bottom: 2px;
      }

      .process-pid {
        font-size: 12px;
        color: var(--color-text-3);
      }
    }
  }

  /* 响应式设计 */
  @media (max-width: 480px) {
    .process-info {
      flex-direction: column;
      text-align: center;
      gap: 8px;

      .process-icon {
        margin: 0 auto;
      }
    }
  }
</style>
