<template>
  <div class="file-selector">
    <a-input-group compact class="w-full">
      <a-input
        v-model="inputValue"
        :placeholder="
          placeholder || t('components.file.fileSelector.placeholder')
        "
        :disabled="disabled"
        :readonly="readonly"
        :error="error"
        :allow-clear="true"
        @change="handleInputChange"
      />
      <a-popover
        v-model:popup-visible="isPopoverVisible"
        :trigger="['click']"
        :unmount-on-close="false"
        :click-outside-to-close="true"
        position="right"
        class="file-selector-popover"
      >
        <a-button :disabled="disabled" @click="handleOpenSelector">
          <template #icon>
            <icon-folder />
          </template>
        </a-button>
        <template #content>
          <file-browser
            :initial-path="currentPath"
            :type="props.type"
            :host="props.host"
            @select="handleFileSelect"
            @cancel="closePopover"
          />
        </template>
      </a-popover>
    </a-input-group>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type {
    FileSelectorItem as FileItem,
    FileSelectProps,
    FileSelectEmits,
  } from './types';
  import { FileSelectType } from './types';
  import FileBrowser from './file-browser.vue';

  const props = withDefaults(defineProps<FileSelectProps>(), {
    modelValue: '',
    placeholder: '',
    disabled: false,
    readonly: false,
    error: false,
    allowCreate: false,
    initialPath: '/',
    type: FileSelectType.FILE,
  });

  const emit = defineEmits<FileSelectEmits>();
  const { t } = useI18n();

  const inputValue = ref<string>(props.modelValue);
  const isPopoverVisible = ref<boolean>(false);
  const currentPath = ref<string>(props.initialPath);

  const closePopover = (): void => {
    isPopoverVisible.value = false;
  };

  watch(
    () => props.modelValue,
    (newValue: string) => {
      inputValue.value = newValue;
    }
  );

  const handleInputChange = (value: string): void => {
    emit('update:modelValue', value);
    emit('change', value);
  };

  const handleOpenSelector = (): void => {
    if (!props.disabled) {
      isPopoverVisible.value = true;
    }
  };

  const handleFileSelect = (file: FileItem): void => {
    inputValue.value = file.path;
    emit('update:modelValue', file.path);
    emit('change', file.path);
    closePopover();
  };

  defineExpose({
    closePopover: () => {
      isPopoverVisible.value = false;
    },
  });
</script>

<style lang="less" scoped>
  .file-selector {
    display: inline-block;
    width: 100%;

    :deep(.arco-input-group) {
      display: flex;
      gap: 0.5rem;

      .arco-input-wrapper {
        flex: 1;
        border-radius: 0.375rem !important;
      }

      .arco-btn {
        border-radius: 0.375rem !important;
        flex-shrink: 0;
      }
    }
  }

  :deep(.file-selector-popover) {
    padding: 0;
    .arco-popover-content {
      padding: 0;
    }
  }
</style>
