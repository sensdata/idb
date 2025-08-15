<template>
  <a-modal
    :visible="visible"
    :unmount-on-close="true"
    :mask-closable="false"
    :ok-text="$t('app.ssh.unsavedChanges.discard')"
    :cancel-text="$t('app.ssh.unsavedChanges.cancel')"
    @cancel="$emit('cancel')"
    @ok="$emit('confirm')"
  >
    <template #title>{{ $t('app.ssh.unsavedChanges.title') }}</template>
    <p>{{ $t('app.ssh.unsavedChanges.content') }}</p>
  </a-modal>
</template>

<script lang="ts" setup>
  defineProps<{
    visible: boolean;
  }>();

  defineEmits<{
    (e: 'cancel'): void;
    (e: 'confirm'): void;
  }>();
</script>
