<template>
  <div class="config-selector">
    <a-space :size="16">
      <!-- 配置类型选择器 -->
      <selector-group
        :label="$t('app.nftables.config.configScope')"
        :model-value="configType"
        :options="configTypeOptions"
        :variant="variant"
        :show-label="showLabels"
        @change="handleConfigTypeChange"
      />

      <!-- 配置模式选择器 -->
      <selector-group
        :label="$t('app.nftables.config.editMode')"
        :model-value="configMode"
        :options="configModeOptions"
        :variant="variant"
        :show-label="showLabels"
        @change="handleConfigModeChange"
      />
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconEdit as IconForm, IconCode } from '@arco-design/web-vue/es/icon';
  import SelectorGroup from './selector-group.vue';

  interface Props {
    configType: 'local' | 'global';
    configMode: 'form' | 'file';
    variant?: 'radio' | 'button';
    showLabels?: boolean;
  }

  interface Emits {
    (e: 'configTypeChange', value: 'local' | 'global'): void;
    (e: 'configModeChange', value: 'form' | 'file'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    variant: 'radio',
    showLabels: true,
  });

  const emit = defineEmits<Emits>();
  const { t } = useI18n();

  // 配置类型选项
  const configTypeOptions = computed(() => [
    {
      value: 'local',
      label: t('app.nftables.config.type.local'),
    },
    {
      value: 'global',
      label: t('app.nftables.config.type.global'),
    },
  ]);

  // 配置模式选项
  const configModeOptions = computed(() => [
    {
      value: 'form',
      label: t('app.nftables.config.visualMode'),
      icon: IconForm,
    },
    {
      value: 'file',
      label: t('app.nftables.config.fileMode'),
      icon: IconCode,
    },
  ]);

  const handleConfigTypeChange = (value: string) => {
    emit('configTypeChange', value as 'local' | 'global');
  };

  const handleConfigModeChange = (value: string) => {
    emit('configModeChange', value as 'form' | 'file');
  };
</script>

<style scoped lang="less">
  .config-selector {
    display: flex;
    align-items: center;
    gap: 16px;
  }
</style>
