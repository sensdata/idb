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
      <div v-if="pid" class="process-pid">PID:{{ pid }}</div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import { useProcessIcons } from '../pages/config/composables/use-process-icons';

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

      // 不同进程类型的颜色 - 使用品牌色系统
      &.process-icon-database {
        background: var(--idbdusk-1);
        color: var(--idbdusk-6);
      }

      &.process-icon-server {
        background: var(--idbgreen-1);
        color: var(--idbgreen-6);
      }

      &.process-icon-container {
        background: var(--idblue-1);
        color: var(--idblue-6);
      }

      &.process-icon-network {
        background: var(--idbturquoise-1);
        color: var(--idbturquoise-6);
      }

      &.process-icon-security {
        background: var(--idbred-1);
        color: var(--idbred-6);
      }

      &.process-icon-development {
        background: var(--idbautumn-1);
        color: var(--idbautumn-6);
      }

      &.process-icon-system {
        background: var(--color-fill-1);
        color: var(--color-text-3);
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
