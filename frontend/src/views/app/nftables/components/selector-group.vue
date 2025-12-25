<template>
  <div class="selector-group" :class="{ 'selector-group--disabled': disabled }">
    <span v-if="showLabel" class="selector-label">{{ label }}</span>

    <!-- Button Group 模式 -->
    <a-button-group v-if="variant === 'button'" size="small">
      <a-button
        v-for="option in validOptions"
        :key="option.value"
        :type="modelValue === option.value ? 'primary' : 'outline'"
        :disabled="disabled || option.disabled"
        @click="handleChange(option.value)"
      >
        <component :is="option.icon" v-if="option.icon" />
        {{ option.label }}
      </a-button>
    </a-button-group>

    <!-- Radio Group 模式 -->
    <a-radio-group
      v-else
      :model-value="modelValue"
      :disabled="disabled"
      type="button"
      size="small"
      @change="(value: string) => handleChange(value)"
    >
      <a-radio
        v-for="option in validOptions"
        :key="option.value"
        :value="option.value"
        :disabled="option.disabled"
      >
        <template #radio="{ checked }">
          <a-space v-if="option.icon">
            <component :is="option.icon" :class="{ 'text-primary': checked }" />
            {{ option.label }}
          </a-space>
          <span v-else>{{ option.label }}</span>
        </template>
      </a-radio>
    </a-radio-group>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useLogger } from '@/composables/use-logger';
  import type { Component } from 'vue';

  interface Option {
    value: string;
    label: string;
    icon?: Component;
    disabled?: boolean;
  }

  interface Props {
    /** 选择器标签 */
    label?: string;
    /** 当前选中的值 */
    modelValue: string;
    /** 选项列表 */
    options: Option[];
    /** 选择器变体样式 */
    variant?: 'radio' | 'button';
    /** 是否显示标签 */
    showLabel?: boolean;
    /** 是否禁用 */
    disabled?: boolean;
  }

  interface Emits {
    /** 更新 v-model 值 */
    (e: 'update:modelValue', value: string): void;
    /** 值变化时触发 */
    (e: 'change', value: string): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    variant: 'radio',
    showLabel: true,
    disabled: false,
  });

  const emit = defineEmits<Emits>();

  const { logWarn } = useLogger('SelectorGroup');

  // 验证和过滤选项
  const validOptions = computed(() => {
    if (!Array.isArray(props.options)) {
      logWarn('options should be an array');
      return [];
    }

    return props.options.filter((option) => {
      if (
        !option ||
        typeof option.value !== 'string' ||
        typeof option.label !== 'string'
      ) {
        logWarn('Invalid option:', option);
        return false;
      }
      return true;
    });
  });

  // 检查当前值是否有效
  const isValidValue = computed(() => {
    return validOptions.value.some(
      (option) => option.value === props.modelValue
    );
  });

  const handleChange = (value: string) => {
    if (props.disabled) return;

    // 检查新值是否有效
    const option = validOptions.value.find((opt) => opt.value === value);
    if (!option || option.disabled) return;

    // 发出标准的 v-model 事件
    emit('update:modelValue', value);
    // 发出传统的 change 事件（向后兼容）
    emit('change', value);
  };

  // 开发模式下的警告
  if (process.env.NODE_ENV === 'development') {
    if (!isValidValue.value && props.modelValue) {
      logWarn(`Current value "${props.modelValue}" is not in options`);
    }
  }
</script>

<style scoped lang="less">
  .selector-group {
    display: flex;
    gap: 8px;
    align-items: center;
    &--disabled {
      pointer-events: none;
      opacity: 0.6;
    }
    .selector-label {
      font-size: 14px;
      font-weight: 500;
      color: var(--color-text-2);
      white-space: nowrap;
    }
    .text-primary {
      color: var(--color-primary-6);
    }
    :deep(.arco-btn-group) {
      display: flex;
      .arco-btn {
        display: inline-flex;
        gap: 4px;
        align-items: center;
        padding: 4px 12px;
        font-size: 12px;
        transition: all 0.2s ease;
        &:first-child {
          border-top-left-radius: 4px;
          border-bottom-left-radius: 4px;
        }
        &:last-child {
          border-top-right-radius: 4px;
          border-bottom-right-radius: 4px;
        }
        &:hover:not(:disabled) {
          transform: translateY(-1px);
        }
      }
    }
    :deep(.arco-radio-group) {
      .arco-radio-button {
        padding: 8px 16px;
        border-radius: 6px;
        transition: all 0.2s ease;
        &:first-child {
          border-top-left-radius: 6px;
          border-bottom-left-radius: 6px;
        }
        &:last-child {
          border-top-right-radius: 6px;
          border-bottom-right-radius: 6px;
        }
        &:hover:not(.arco-radio-disabled) {
          transform: translateY(-1px);
        }
      }
    }
  }
</style>
