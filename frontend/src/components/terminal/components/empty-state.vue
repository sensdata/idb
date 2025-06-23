<template>
  <div class="empty-state">
    <a-empty :description="$t('components.terminal.workspace.noSessions')">
      <template #image>
        <icon-code-square />
      </template>
      <a-button type="primary" @click="handleCreateSession">
        {{ $t('components.terminal.session.createFirst') }}
      </a-button>
    </a-empty>
  </div>
</template>

<script setup lang="ts">
  import { useLogger } from '@/composables/use-logger';

  /**
   * 组件事件定义
   */
  interface Emits {
    /** 创建会话事件 */
    (e: 'createSession'): void;
  }

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
</style>
