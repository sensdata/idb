<template>
  <a-button
    :size="size"
    :type="type"
    :disabled="disabled"
    @click="handleCategoryManage"
  >
    <template #icon>
      <icon-settings />
    </template>
  </a-button>

  <!-- 分类管理抽屉 -->
  <category-manage-drawer
    ref="categoryManageRef"
    :config="config"
    @ok="handleManageOk"
  />
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { IconSettings } from '@arco-design/web-vue/es/icon';
  import CategoryManageDrawer from '@/components/idb-tree/components/category-manage-drawer.vue';
  import type {
    CategoryManageButtonProps,
    CategoryManageButtonEmits,
    CategoryManageButtonExposed,
  } from './types';

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const props = withDefaults(defineProps<CategoryManageButtonProps>(), {
    size: 'medium',
    type: 'secondary',
    disabled: false,
  });

  const emit = defineEmits<CategoryManageButtonEmits>();

  // 组件引用
  const categoryManageRef = ref<InstanceType<typeof CategoryManageDrawer>>();

  // 处理分类管理按钮点击
  const handleCategoryManage = () => {
    emit('manage');
    categoryManageRef.value?.show();
  };

  // 处理分类管理完成
  const handleManageOk = () => {
    emit('ok');
  };

  // 暴露方法
  const show = () => {
    categoryManageRef.value?.show();
  };

  const hide = () => {
    // category-manage-drawer 组件暂时没有 hide 方法
    // 可以通过内部状态控制
  };

  defineExpose<CategoryManageButtonExposed>({
    show,
    hide,
  });
</script>

<style scoped>
  /* 如果需要特定样式，可以在这里添加 */
</style>
