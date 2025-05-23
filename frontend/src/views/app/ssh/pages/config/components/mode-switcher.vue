<template>
  <div class="mode-switch">
    <a-radio-group :model-value="mode" type="button" @change="handleModeChange">
      <a-radio value="visual">{{ $t('app.ssh.mode.visual') }}</a-radio>
      <a-radio value="source">{{ $t('app.ssh.mode.source') }}</a-radio>
    </a-radio-group>
  </div>
</template>

<script lang="ts" setup>
  import { defineProps, defineEmits } from 'vue';
  import type { ConfigMode } from '../types';

  defineProps<{
    mode: ConfigMode;
  }>();

  const emit = defineEmits<{
    (e: 'modeChange', mode: ConfigMode): void;
  }>();

  const handleModeChange = (value: string | number | boolean): void => {
    emit('modeChange', value as ConfigMode);
  };
</script>

<style scoped lang="less">
  .mode-switch {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 16px;
  }
</style>
