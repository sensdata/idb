<template>
  <div class="loading-state" role="status" :aria-label="loadingText">
    <div class="loading-content">
      <a-spin :size="spinSize" />
      <div class="loading-text">
        {{ loadingText }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';

  interface Props {
    /** 加载文本的国际化key，默认为终端会话创建文本 */
    textKey?: string;
    /** 加载动画大小 */
    size?: number;
  }

  const props = withDefaults(defineProps<Props>(), {
    textKey: 'components.terminal.workspace.creatingSession',
    size: 32,
  });

  const { t } = useI18n();

  const loadingText = computed(() => t(props.textKey));
  const spinSize = computed(() => props.size);
</script>

<style scoped>
  .loading-state {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    min-height: 200px; /* 确保有足够的高度 */
    padding: 32px;
    background: var(--color-bg-1);
  }

  .loading-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
    align-items: center;
  }

  .loading-text {
    max-width: 300px; /* 防止文本过长 */
    margin-top: 8px;
    font-size: 14px;
    color: var(--color-text-2);
    text-align: center;
  }

  /* 响应式设计 */
  @media (width <= 768px) {
    .loading-state {
      padding: 16px;
    }
    .loading-text {
      font-size: 12px;
    }
  }
</style>
