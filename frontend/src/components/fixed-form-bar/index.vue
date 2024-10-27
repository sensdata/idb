<template>
  <fixed-footer-bar>
    <template #right>
      <a-space>
        <a-button v-if="hasCancel" @click="emit('cancel')">{{
          $t('common.form.cancelText')
        }}</a-button>
        <a-button type="primary" :loading="loading" @click="emit('submit')">
          {{ $t('common.form.submitText') }}
        </a-button>
      </a-space>
    </template>
  </fixed-footer-bar>
</template>

<script lang="ts" setup>
  import 'vue-i18n';

  const emit = defineEmits(['submit', 'cancel']);

  defineProps({
    loading: {
      type: Boolean,
      default: false,
    },
    hasCancel: {
      type: Boolean,
      default: true,
    },
  });
</script>
