<template>
  <div class="diff-header">
    <div class="version-header">
      <div class="version-column">
        <a-tag color="green" size="large">
          {{ currentVersionLabel }}
        </a-tag>
      </div>
      <div class="version-column">
        <div class="version-info">
          <a-tag color="blue" size="large">
            {{ historyVersionLabel }}
          </a-tag>
          <a-button
            v-if="showRestoreButton"
            type="primary"
            size="small"
            :loading="restoreLoading"
            class="restore-button"
            @click="handleRestore"
          >
            {{ computedRestoreButtonText }}
          </a-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';

  defineOptions({
    name: 'VersionHeader',
  });

  const { t } = useI18n();

  const props = withDefaults(
    defineProps<{
      currentVersionLabel: string;
      historyVersionLabel: string;
      restoreLoading?: boolean;
      showRestoreButton?: boolean;
      restoreButtonText?: string;
    }>(),
    {
      restoreLoading: false,
      showRestoreButton: true,
      restoreButtonText: undefined,
    }
  );

  const emit = defineEmits<{
    (e: 'restore'): void;
  }>();

  // 计算恢复按钮文本，支持自定义文本和国际化
  const computedRestoreButtonText = computed(
    () => props.restoreButtonText || t('common.restore', 'Restore')
  );

  const handleRestore = () => {
    emit('restore');
  };
</script>

<style scoped>
  .diff-header {
    padding-bottom: 16px;
    margin-bottom: 16px;
    border-bottom: 1px solid var(--color-border-2);
  }

  .version-header {
    display: flex;
    width: 100%;
  }

  .version-column {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
  }

  .version-info {
    display: flex;
    flex-direction: column;
    gap: 8px;
    align-items: center;
  }

  .restore-button {
    height: 28px;
    padding: 0 12px;
    font-size: 12px;
  }
</style>
