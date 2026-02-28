<template>
  <div class="empty-state">
    <div class="empty-card">
      <div class="empty-icon-wrap">
        <icon-code-square class="empty-icon" />
      </div>
      <div class="empty-title">
        {{ $t('components.terminal.workspace.noSessions') }}
      </div>
      <div class="empty-desc">
        {{
          $t('components.terminal.workspace.noSessionsHint', {
            host: hostName || '-',
          })
        }}
      </div>
      <div class="empty-actions">
        <a-button type="outline" @click="handleAttachSession">
          {{ $t('components.terminal.session.attach') }}
        </a-button>
        <a-button type="primary" @click="handleCreateSession">
          {{ $t('components.terminal.session.createFirst') }}
        </a-button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useLogger } from '@/composables/use-logger';

  /**
   * 组件事件定义
   */
  interface Props {
    hostName?: string;
  }

  interface Emits {
    /** 创建会话事件 */
    (e: 'createSession'): void;
    /** 打开连接已有会话 */
    (e: 'attachSession'): void;
  }

  defineProps<Props>();
  const emit = defineEmits<Emits>();
  const { logDebug, logError } = useLogger('EmptyState');

  /**
   * 处理创建会话按钮点击事件
   */
  function handleCreateSession(): void {
    try {
      logDebug('Creating new terminal session');
      emit('createSession');
    } catch (error) {
      logError('Failed to emit createSession event:', error);
    }
  }

  function handleAttachSession(): void {
    try {
      logDebug('Opening attach session dialog');
      emit('attachSession');
    } catch (error) {
      logError('Failed to emit attachSession event:', error);
    }
  }
</script>

<style scoped>
  .empty-state {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    padding: 32px;
    background: var(--color-bg-1);
  }

  .empty-card {
    width: min(560px, 100%);
    padding: 28px 24px;
    text-align: center;
    background: linear-gradient(
      180deg,
      var(--color-fill-1) 0%,
      rgb(255 255 255 / 0%) 100%
    );
    border: 1px dashed var(--color-border-3);
    border-radius: 12px;
  }

  .empty-icon-wrap {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 56px;
    height: 56px;
    margin-bottom: 12px;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-2);
    border-radius: 50%;
  }

  .empty-icon {
    font-size: 26px;
    color: var(--color-primary-6);
  }

  .empty-title {
    margin-bottom: 6px;
    font-size: 16px;
    font-weight: 600;
    color: var(--color-text-1);
  }

  .empty-desc {
    max-width: 460px;
    margin: 0 auto;
    font-size: 13px;
    line-height: 1.6;
    color: var(--color-text-3);
  }

  .empty-actions {
    display: flex;
    gap: 10px;
    justify-content: center;
    margin-top: 16px;
  }

  @media (width <= 768px) {
    .empty-state {
      padding: 16px;
    }
    .empty-card {
      padding: 20px 16px;
    }
    .empty-actions {
      flex-direction: column;
    }
  }
</style>
