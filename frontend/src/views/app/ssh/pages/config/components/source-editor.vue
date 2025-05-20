<template>
  <div class="source-editor-wrapper">
    <div class="source-toolbar">
      <a-button type="primary" @click="handleSave">
        {{ $t('app.ssh.source.save') }}
      </a-button>
      <a-button class="ml-2" @click="handleReset">
        {{ $t('app.ssh.source.reset') }}
      </a-button>
    </div>
    <a-textarea
      v-model="localConfig"
      :placeholder="$t('app.ssh.source.placeholder')"
      class="source-editor"
      :auto-size="{ minRows: 20, maxRows: 30 }"
    />
    <div class="source-info">
      {{ $t('app.ssh.source.info') }}
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watchEffect, defineProps, defineEmits } from 'vue';

  const props = defineProps({
    config: {
      type: String,
      required: true,
    },
    originalConfig: {
      type: String,
      required: true,
    },
  });

  const emits = defineEmits(['update:config', 'save', 'reset', 'update-form']);

  // Local copy of config for v-model
  const localConfig = ref(props.config);

  // Keep local config in sync with passed config
  watchEffect(() => {
    localConfig.value = props.config;
  });

  // Update parent value when local value changes
  watchEffect(() => {
    emits('update:config', localConfig.value);
  });

  // Handle save button click
  const handleSave = () => {
    emits('save');
  };

  // Handle reset button click
  const handleReset = () => {
    emits('reset');
  };
</script>

<style scoped lang="less">
  .source-editor-wrapper {
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .source-toolbar {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--color-border-2);

    .ml-2 {
      margin-left: 8px;
    }
  }

  .source-editor {
    width: 100%;
    font-family: monospace;
    resize: none;
    border: none;
    border-radius: 0;

    &:focus {
      box-shadow: none;
    }
  }

  .source-info {
    padding: 8px 12px;
    color: var(--color-text-3);
    font-size: 12px;
    border-top: 1px solid var(--color-border-2);
  }
</style>
