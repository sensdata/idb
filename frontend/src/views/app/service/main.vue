<template>
  <a-tabs v-model:active-key="activeTab" lazy-load @change="handleTabChange">
    <a-tab-pane
      :key="SERVICE_TYPE.Local"
      :title="$t('app.service.enum.type.local')"
    >
      <list
        :key="`list-${SERVICE_TYPE.Local}`"
        ref="localListRef"
        :type="SERVICE_TYPE.Local"
      />
    </a-tab-pane>
    <a-tab-pane
      :key="SERVICE_TYPE.Global"
      :title="$t('app.service.enum.type.global')"
    >
      <list
        :key="`list-${SERVICE_TYPE.Global}`"
        ref="globalListRef"
        :type="SERVICE_TYPE.Global"
      />
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { SERVICE_TYPE } from '@/config/enum';
  import { ref, onMounted, nextTick } from 'vue';
  import List from './list.vue';

  // 定义 List 组件实例类型
  interface ListInstance {
    resetComponentsState?: () => void;
  }

  const activeTab = ref(SERVICE_TYPE.Local);
  const localListRef = ref<ListInstance | null>(null);
  const globalListRef = ref<ListInstance | null>(null);

  // 确保在tab切换时刷新当前激活的列表
  const handleTabChange = async (key: string | number) => {
    // 使用 nextTick 确保DOM已更新
    await nextTick();

    // 切换tab时，刷新对应的列表以确保数据最新
    if (key === SERVICE_TYPE.Local && localListRef.value) {
      localListRef.value.resetComponentsState?.();
    } else if (key === SERVICE_TYPE.Global && globalListRef.value) {
      globalListRef.value.resetComponentsState?.();
    }
  };

  onMounted(() => {
    // 确保初始加载时重置状态
    handleTabChange(activeTab.value);
  });
</script>
