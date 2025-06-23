<template>
  <div class="permission-input">
    <!-- 简洁预览模式 -->
    <div class="permission-preview-compact">
      <div class="preview-content">
        <div class="preview-label">
          {{ $t('app.logrotate.permission.preview') }}:
        </div>
        <div class="preview-value-with-comment">
          <code class="preview-value">
            create {{ permission.mode }} {{ permission.user }}
            {{ permission.group }}
          </code>
          <span class="preview-comment">// {{ permission.description }}</span>
        </div>
      </div>
      <a-button type="outline" size="small" @click="openModal">
        <template #icon>
          <icon-settings />
        </template>
        {{ $t('app.logrotate.permission.settings') }}
      </a-button>
    </div>

    <!-- 详细设置弹窗 -->
    <permission-modal
      v-model:visible="showModal"
      :permission="permission"
      @confirm="handleConfirm"
    />
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { IconSettings } from '@arco-design/web-vue/es/icon';
  import PermissionModal from './permission-modal.vue';
  import { usePermission } from './composables/use-permission';
  import type { PermissionConfig } from './types';

  interface Props {
    modelValue?: string;
  }

  interface Emits {
    (e: 'update:modelValue', value: string): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: 'create 0644 root root',
  });

  const emit = defineEmits<Emits>();

  // 使用权限管理hook
  const { permission, updateFromString, toString } = usePermission(
    props.modelValue
  );

  // 弹窗状态
  const showModal = ref(false);

  // 打开弹窗
  const openModal = () => {
    showModal.value = true;
  };

  // 确认修改
  const handleConfirm = (newPermission: PermissionConfig) => {
    const permissionString = toString(newPermission);
    emit('update:modelValue', permissionString);
    updateFromString(permissionString);
    showModal.value = false;
  };

  // 监听props变化
  watch(
    () => props.modelValue,
    (newValue: string) => {
      if (newValue) {
        updateFromString(newValue);
      }
    },
    { immediate: true }
  );
</script>

<style scoped>
  .permission-input {
    width: 100%;
  }

  .permission-preview-compact {
    display: flex;
    gap: 12px;
    align-items: flex-start;
    padding: 12px 16px;
    background-color: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 6px;
  }

  .preview-content {
    flex: 1;
  }

  .preview-label {
    display: block;
    margin-bottom: 4px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .preview-value {
    display: inline;
    padding: 4px 8px;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 13px;
    color: var(--color-text-1);
    background-color: var(--color-bg-3);
    border-radius: 3px;
  }

  .preview-value-with-comment {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    margin-bottom: 6px;
  }

  .preview-comment {
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 12px;
    font-style: italic;
    color: var(--color-text-3);
  }
</style>
