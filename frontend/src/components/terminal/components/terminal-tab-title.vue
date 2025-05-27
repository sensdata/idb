<template>
  <div class="tab-title-container">
    <span v-if="!item.isRenaming" class="tab-title" @dblclick="handleRename">
      {{ item.title }}
    </span>
    <a-input
      v-else
      ref="renameInputRef"
      v-model="renameValue"
      size="mini"
      class="rename-input"
      :placeholder="$t('components.terminal.session.renamePlaceholder')"
      @blur="finishRename"
      @keyup.enter="finishRename"
      @keyup.esc="cancelRename"
    />
    <a-dropdown position="bottom" @select="handleAction">
      <span class="arco-icon-hover tab-action-btn">
        <icon-more />
      </span>
      <template #content>
        <a-doption value="rename">
          {{ $t('components.terminal.session.rename') }}
        </a-doption>
        <a-doption value="detach">
          {{ $t('components.terminal.session.detach') }}
        </a-doption>
        <a-doption value="quit">
          {{ $t('components.terminal.session.quit') }}
        </a-doption>
      </template>
    </a-dropdown>
  </div>
</template>

<script setup lang="ts">
  import { ref, nextTick, computed } from 'vue';
  import { useLogger } from '@/hooks/use-logger';

  export interface TermSessionItem {
    key: string;
    type: 'attach' | 'start';
    hostId: number;
    hostName: string;
    title: string;
    sessionId?: string;
    sessionName?: string;
    termRef?: any;
    isRenaming?: boolean;
    renameValue?: string;
  }

  interface Props {
    item: TermSessionItem;
  }

  interface Emits {
    (e: 'rename', item: TermSessionItem): void;
    (
      e: 'action',
      item: TermSessionItem,
      action: 'rename' | 'quit' | 'detach'
    ): void;
  }

  type ActionType = 'rename' | 'quit' | 'detach';

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  const { logWarn, logError } = useLogger('TerminalTabTitle');

  const renameInputRef = ref<HTMLInputElement>();
  const renameValue = ref('');

  // 计算属性：验证重命名值是否有效
  const isRenameValueValid = computed(() => {
    return renameValue.value && renameValue.value.trim().length > 0;
  });

  // 取消重命名
  function cancelRename(): void {
    if (!props.item) {
      logWarn('Terminal tab item is not available');
      return;
    }

    try {
      const updatedItem: TermSessionItem = {
        ...props.item,
        isRenaming: false,
      };
      emit('rename', updatedItem);
    } catch (error) {
      logError('Failed to cancel rename operation:', error);
    } finally {
      renameValue.value = '';
    }
  }

  // 开始重命名
  function handleRename(): void {
    if (!props.item) {
      logWarn('Terminal tab item is not available');
      return;
    }

    try {
      // 通过emit通知父组件更新状态，避免直接修改props
      emit('action', props.item, 'rename');
      renameValue.value = props.item.title || '';

      nextTick(() => {
        if (renameInputRef.value) {
          renameInputRef.value.focus();
          renameInputRef.value.select();
        }
      }).catch((error) => {
        logError('Failed to focus rename input:', error);
      });
    } catch (error) {
      logError('Failed to start rename operation:', error);
    }
  }

  // 完成重命名
  function finishRename(): void {
    if (!props.item) {
      logWarn('Terminal tab item is not available');
      return;
    }

    try {
      if (isRenameValueValid.value) {
        // 创建新的item对象，避免直接修改props
        const updatedItem: TermSessionItem = {
          ...props.item,
          title: renameValue.value.trim(),
          isRenaming: false,
        };
        emit('rename', updatedItem);
      } else {
        // 取消重命名 - 恢复原始状态
        cancelRename();
      }
    } catch (error) {
      logError('Failed to finish rename operation:', error);
      // 发生错误时取消重命名
      cancelRename();
    } finally {
      renameValue.value = '';
    }
  }

  // 验证操作类型
  function isValidAction(action: string): action is ActionType {
    return ['rename', 'quit', 'detach'].includes(action);
  }

  // 处理操作
  function handleAction(
    action: string | number | Record<string, any> | undefined
  ): void {
    if (!props.item) {
      logWarn('Terminal tab item is not available');
      return;
    }

    try {
      const actionStr = String(action);

      if (!isValidAction(actionStr)) {
        logWarn('Invalid action type:', actionStr);
        return;
      }

      if (actionStr === 'rename') {
        handleRename();
      } else {
        emit('action', props.item, actionStr);
      }
    } catch (error) {
      logError('Failed to handle action:', error);
    }
  }
</script>

<style scoped>
  .tab-title-container {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .tab-title {
    cursor: pointer;
    user-select: none;
    transition: color 0.2s ease;
  }

  .tab-title:hover {
    color: var(--color-text-1);
  }

  .rename-input {
    width: 120px;
    min-width: 80px;
  }

  .tab-action-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    font-size: 12px;
    color: var(--color-text-3);
    cursor: pointer;
    border-radius: 2px;
    transition: all 0.1s ease;
  }

  .tab-action-btn:hover {
    color: var(--color-text-1);
    background-color: var(--color-fill-3);
  }
</style>
