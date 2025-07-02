<template>
  <div class="tab-title-container" @contextmenu.prevent="showContextMenu">
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
    <a-dropdown
      v-if="!item.isRenaming"
      position="bottom"
      trigger="click"
      @select="handleAction"
    >
      <span
        class="tab-menu-btn"
        :title="$t('components.terminal.session.actions')"
        @click.stop
      >
        <icon-down />
      </span>
      <template #content>
        <a-doption value="rename">
          <template #icon>
            <icon-edit />
          </template>
          {{ $t('components.terminal.session.rename') }}
        </a-doption>
        <a-doption value="detach">
          <template #icon>
            <icon-disconnect />
          </template>
          {{ $t('components.terminal.session.detach') }}
        </a-doption>
        <a-doption value="quit">
          <template #icon>
            <icon-close />
          </template>
          {{ $t('components.terminal.session.quit') }}
        </a-doption>
      </template>
    </a-dropdown>

    <a-dropdown
      v-model:popup-visible="contextMenuVisible"
      position="bottom"
      :trigger="[]"
      @select="handleAction"
    >
      <span style="display: none"></span>
      <template #content>
        <a-doption value="rename">
          <template #icon>
            <icon-edit />
          </template>
          {{ $t('components.terminal.session.rename') }}
        </a-doption>
        <a-doption value="detach">
          <template #icon>
            <icon-disconnect />
          </template>
          {{ $t('components.terminal.session.detach') }}
        </a-doption>
        <a-doption value="quit">
          <template #icon>
            <icon-close />
          </template>
          {{ $t('components.terminal.session.quit') }}
        </a-doption>
      </template>
    </a-dropdown>
  </div>
</template>

<script setup lang="ts">
  import { ref, nextTick, computed } from 'vue';
  import { useLogger } from '@/composables/use-logger';

  export interface TermSessionItem {
    key: string;
    type: 'attach' | 'start';
    hostId: number;
    hostName: string;
    title: string;
    sessionId?: string;
    sessionName?: string; // 服务器端的会话名称，不应该被用户重命名影响
    originalSessionName?: string; // 保存原始的服务器会话名称，用于恢复会话
    termRef?: any;
    isRenaming?: boolean;
    renameValue?: string;
    isCustomTitle?: boolean; // 标记标题是否为用户自定义
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
          try {
            renameInputRef.value.focus();
            // 检查是否有select方法再调用
            if (typeof renameInputRef.value.select === 'function') {
              renameInputRef.value.select();
            }
          } catch (error) {
            logError('Failed to focus/select rename input:', error);
          }
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

  const contextMenuVisible = ref(false);

  function showContextMenu(): void {
    contextMenuVisible.value = true;
  }
</script>

<style scoped>
  .tab-title-container {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .tab-title {
    color: var(
      --color-primary-6
    ) !important; /* 默认使用primary color - 符合MasterGo设计 */

    cursor: pointer;
    user-select: none;
    transition: color 0.2s ease;
  }

  .tab-title:hover {
    color: var(--color-primary-5) !important; /* 悬停状态使用primary-5颜色 */
  }

  .rename-input {
    width: 120px;
    min-width: 80px;
  }

  .tab-action-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px; /* 与menu-btn保持一致 */
    height: 16px; /* 与menu-btn保持一致 */
    margin-left: 4px; /* 添加左边距 */
    font-size: 10px;
    color: var(--color-text-2); /* 统一使用text-2颜色变量 */
    cursor: pointer;
    border-radius: 4px; /* 与menu-btn保持一致 */
    transition: all 0.15s ease;
  }

  .tab-action-btn:hover {
    color: var(--color-text-1);
    background-color: var(--color-fill-2); /* 使用更明显的背景色 */
    box-shadow: 0 2px 4px rgb(0 0 0 / 10%); /* 添加阴影 */
    transform: scale(1.1); /* 轻微放大效果 */
  }

  .tab-menu-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px; /* 增加可点击区域 */
    height: 16px; /* 增加可点击区域 */
    margin-left: 4px; /* 添加左边距，避免与文字太近 */
    font-size: 10px;
    color: var(--color-text-2); /* 使用系统text-2颜色变量 - 符合MasterGo设计 */
    cursor: pointer;
    border-radius: 4px; /* 增加圆角 */
    transition: all 0.15s ease;
  }

  .tab-menu-btn:hover {
    color: var(--color-text-1);
    background-color: var(--color-fill-2); /* 使用更明显的背景色 */
    box-shadow: 0 2px 4px rgb(0 0 0 / 10%); /* 添加阴影 */
    transform: scale(1.1); /* 轻微放大效果 */
  }
</style>
