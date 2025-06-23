<template>
  <a-drawer
    :visible="visible"
    :footer="false"
    height="95vh"
    class="terminal-drawer"
    placement="bottom"
    @update:visible="handleVisibleChange"
  >
    <template #title>
      <span>{{ $t('components.terminal.title') }}</span>
    </template>
    <terminal-workspace v-if="visible" ref="workspaceRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useLogger } from '@/composables/use-logger';
  import TerminalWorkspace from './terminal-workspace.vue';

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError } = useLogger('TerminalDrawer');

  // Props 定义
  defineProps<{
    visible: boolean;
  }>();

  // Emits 定义
  const emit = defineEmits<{
    'update:visible': [value: boolean];
  }>();

  // 组件引用
  const workspaceRef = ref<InstanceType<typeof TerminalWorkspace>>();

  // 初始化工作区
  const initializeWorkspace = async () => {
    try {
      // 等待下一个tick确保组件已经渲染
      await nextTick();

      if (workspaceRef.value?.reinitialize) {
        await workspaceRef.value.reinitialize();
      }
    } catch (error) {
      logError('Failed to initialize terminal workspace:', error);
      Message.error(t('components.terminal.workspace.initializeFailed'));
    }
  };

  // 处理弹窗显示状态变化
  const handleVisibleChange = async (value: boolean) => {
    emit('update:visible', value);

    // 当弹窗打开时，重新初始化工作区
    if (value) {
      await initializeWorkspace();
    }
  };
</script>

<style scoped>
  .terminal-drawer :deep(.arco-drawer-title) {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 98%;
  }

  .terminal-drawer :deep(.arco-drawer-body) {
    height: 100%;
    padding: 0;
  }
</style>
