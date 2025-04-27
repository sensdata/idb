<template>
  <div class="view-controls">
    <a-space direction="vertical" size="medium" style="width: 100%">
      <!-- 视图模式选择 -->
      <a-radio-group
        v-model="currentMode"
        type="button"
        size="small"
        style="display: flex; width: 100%"
      >
        <a-radio
          v-for="option in viewOptions"
          :key="option.value"
          :value="option.value"
          style="flex: 1; text-align: center"
        >
          {{ option.label }}
        </a-radio>
      </a-radio-group>

      <!-- 行数输入，仅在头部/尾部模式下显示 -->
      <a-form-item
        v-if="currentMode !== 'full'"
        :label="t('app.file.editor.lineCount')"
      >
        <a-input-number
          v-model="tempLineCount"
          size="small"
          :min="10"
          :max="10000"
          :step="100"
          style="width: 160px"
        />
        <a-button
          type="primary"
          size="small"
          :disabled="tempLineCount === lineCount"
          style="margin-left: 8px"
          @click="applyLineCount"
        >
          {{ t('app.file.editor.apply') }}
        </a-button>
      </a-form-item>
    </a-space>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { ContentViewMode } from '@/components/file/file-editor-drawer/types';

  const { t } = useI18n();

  const props = defineProps({
    viewMode: {
      type: String as () => ContentViewMode,
      required: true,
    },
    lineCount: {
      type: Number,
      required: true,
    },
  });

  const emit = defineEmits<{
    (e: 'update:viewMode', value: ContentViewMode): void;
    (e: 'update:lineCount', value: number): void;
    (e: 'change', mode: ContentViewMode, lines: number): void;
  }>();

  const currentMode = computed({
    get: () => props.viewMode,
    set: (value: ContentViewMode) => {
      emit('update:viewMode', value);
      emit('change', value, props.lineCount);
    },
  });

  const tempLineCount = ref(props.lineCount);

  watch(
    () => props.lineCount,
    (newValue) => {
      tempLineCount.value = newValue;
    }
  );

  const viewOptions = [
    { value: 'full', label: t('app.file.editor.viewFull') },
    { value: 'head', label: t('app.file.editor.viewHead') },
    { value: 'tail', label: t('app.file.editor.viewTail') },
  ];

  function applyLineCount() {
    emit('update:lineCount', tempLineCount.value);
    emit('change', props.viewMode, tempLineCount.value);
  }
</script>

<style scoped>
  .view-controls {
    padding: 8px 12px;
    background-color: var(--color-bg-2);
    border-bottom: 1px solid var(--color-border);
  }
</style>
