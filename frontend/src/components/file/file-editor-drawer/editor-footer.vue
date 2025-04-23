<template>
  <div class="drawer-footer">
    <div class="drawer-footer-info">
      <span v-if="file">{{ file.path }}</span>
    </div>
    <div class="drawer-footer-actions">
      <a-button @click="$emit('cancel')">{{
        t('common.form.cancelText')
      }}</a-button>
      <a-button type="primary" :disabled="!isEdited" @click="$emit('save')">
        {{ t('common.save') }}
      </a-button>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import { FileItem } from '@/views/app/file/types/file-item';

  const { t } = useI18n();

  defineProps({
    file: {
      type: Object as () => FileItem | null,
      default: null,
    },
    isEdited: {
      type: Boolean,
      required: true,
    },
  });

  defineEmits(['cancel', 'save']);
</script>

<style scoped>
  .drawer-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .drawer-footer-info {
    color: var(--color-text-3);
    font-size: 12px;
  }

  .drawer-footer-actions {
    display: flex;
    gap: 8px;
  }
</style>
