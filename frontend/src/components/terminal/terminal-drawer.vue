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

<style>
  /* 
 * 临时解决方案：修复 arco-drawer-body 的左边距问题
 * 
 * 问题描述：
 * - Arco Design Vue 的 Drawer 组件目前只支持 drawer-style 属性
 * - 不支持 body-style 属性来直接设置 drawer body 的样式
 * - 使用 :deep() 深度选择器无法有效覆盖 arco-drawer-body 的样式
 * 
 * 相关 Issue：
 * https://github.com/arco-design/arco-design-vue/issues/3184
 * 
 * 临时方案：
 * 使用全局样式，但通过 .terminal-drawer 类选择器限制影响范围，
 * 确保只影响终端抽屉组件，不会影响项目中的其他 drawer 组件
 * 
 * TODO：
 * 当 Arco Design Vue 官方支持 body-style 属性时，应移除此临时方案
 */

  .terminal-drawer .arco-drawer-body {
    padding: 0 !important;
  }
</style>
