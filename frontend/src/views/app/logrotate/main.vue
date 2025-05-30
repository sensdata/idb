<template>
  <a-tabs v-model:active-key="activeTab" lazy-load @change="handleTabChange">
    <a-tab-pane
      :key="LOGROTATE_TYPE.Local"
      :title="$t('app.logrotate.enum.type.local')"
    >
      <list
        :key="`list-${LOGROTATE_TYPE.Local}`"
        ref="localListRef"
        :type="LOGROTATE_TYPE.Local"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="LOGROTATE_TYPE.Global"
      :title="$t('app.logrotate.enum.type.global')"
    >
      <list
        :key="`list-${LOGROTATE_TYPE.Global}`"
        ref="globalListRef"
        :type="LOGROTATE_TYPE.Global"
      />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { LOGROTATE_TYPE } from '@/config/enum';
  import { ref, onMounted, nextTick } from 'vue';
  import List from './list.vue';

  // 定义 List 组件实例类型
  interface ListInstance {
    resetComponentsState?: () => void;
  }

  const activeTab = ref(LOGROTATE_TYPE.Local);
  const localListRef = ref<ListInstance | null>(null);
  const globalListRef = ref<ListInstance | null>(null);

  // 确保在tab切换时强制刷新列表
  const handleTabChange = async (key: string | number) => {
    // 使用 nextTick 确保DOM已更新
    await nextTick();

    if (key === LOGROTATE_TYPE.Local && localListRef.value) {
      // 强制重置本地列表的状态
      localListRef.value.resetComponentsState?.();
    } else if (key === LOGROTATE_TYPE.Global && globalListRef.value) {
      // 强制重置全局列表的状态
      globalListRef.value.resetComponentsState?.();
    }
  };

  onMounted(() => {
    // 确保初始加载时重置状态
    handleTabChange(activeTab.value);
  });
</script>
