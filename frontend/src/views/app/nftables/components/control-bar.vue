<template>
  <div
    class="control-bar"
    role="banner"
    :aria-label="$t('nftables.controlBar.title')"
  >
    <div class="control-left">
      <h3 class="control-title">{{ $t('nftables.controlBar.title') }}</h3>
      <small class="config-status-text" aria-live="polite">
        {{
          $t('nftables.controlBar.currentStatus', {
            configType: $t(`nftables.configType.${configType}`),
            configMode: $t(`nftables.configMode.${configMode}`),
          })
        }}
      </small>
    </div>
    <div class="control-right">
      <config-selector
        :config-type="configType"
        :config-mode="configMode"
        variant="button"
        :show-labels="true"
        @config-type-change="handleConfigTypeChange"
        @config-mode-change="handleConfigModeChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import ConfigSelector from './config-selector.vue';

  /**
   * 配置类型
   */
  type ConfigType = 'local' | 'global';

  /**
   * 配置模式
   */
  type ConfigMode = 'form' | 'file';

  /**
   * 组件Props
   */
  interface Props {
    /** 配置类型 */
    configType: ConfigType;
    /** 配置模式 */
    configMode: ConfigMode;
  }

  /**
   * 组件事件
   */
  interface Emits {
    /** 配置类型变更事件 */
    (e: 'configTypeChange', value: ConfigType): void;
    /** 配置模式变更事件 */
    (e: 'configModeChange', value: ConfigMode): void;
  }

  defineProps<Props>();
  const emit = defineEmits<Emits>();

  /**
   * 处理配置类型变更
   */
  const handleConfigTypeChange = (value: ConfigType): void => {
    emit('configTypeChange', value);
  };

  /**
   * 处理配置模式变更
   */
  const handleConfigModeChange = (value: ConfigMode): void => {
    emit('configModeChange', value);
  };
</script>

<style scoped lang="less">
  .control-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    margin-bottom: 16px;
    background: var(--color-bg-2);
    border-radius: 8px;
    border: 1px solid var(--color-border-2);
    transition: all 0.2s ease-in-out;

    &:hover {
      border-color: var(--color-border-3);
    }

    .control-left {
      flex: 1;
      min-width: 0; // 防止flex子项溢出

      .control-title {
        font-size: 16px;
        font-weight: 500;
        color: var(--color-text-1);
        margin: 0 0 4px 0;
        line-height: 1.4;
      }

      .config-status-text {
        color: var(--color-text-3);
        font-size: 12px;
        line-height: 1.4;
        display: block;
        word-break: break-all; // 长文本换行
      }
    }

    .control-right {
      display: flex;
      align-items: center;
      flex-shrink: 0; // 防止压缩
    }
  }

  // 响应式设计
  @media (max-width: 768px) {
    .control-bar {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;
      padding: 12px 16px;

      .control-left {
        width: 100%;
      }

      .control-right {
        align-self: stretch;
        justify-content: flex-end;
      }
    }
  }

  @media (max-width: 480px) {
    .control-bar {
      padding: 12px;

      .control-right {
        width: 100%;
        justify-content: center;
      }
    }
  }

  // 深色模式适配
  @media (prefers-color-scheme: dark) {
    .control-bar {
      border-color: var(--color-border-2);

      &:hover {
        border-color: var(--color-border-1);
      }
    }
  }

  // 高对比度模式支持
  @media (prefers-contrast: high) {
    .control-bar {
      border-color: var(--color-text-1);
      border-width: 2px;
    }
  }

  // 减少动画模式支持
  @media (prefers-reduced-motion: reduce) {
    .control-bar {
      transition: none;
    }
  }
</style>
