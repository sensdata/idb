<template>
  <div class="config-status-banner">
    <a-alert
      :type="isCurrentConfigActive ? 'success' : 'warning'"
      :show-icon="true"
      :closable="false"
    >
      <template #icon>
        <icon-check v-if="isCurrentConfigActive" />
        <icon-exclamation-circle v-else />
      </template>
      <template #title>
        <a-space>
          <span>{{ configTypeLabel }}</span>
          <a-tag
            :color="isCurrentConfigActive ? 'green' : 'orange'"
            size="small"
          >
            {{ statusLabel }}
          </a-tag>
        </a-space>
      </template>
      <div class="status-description">
        <div>{{ configDescription }}</div>
        <div v-if="!isCurrentConfigActive" class="activate-hint">
          {{ activateHint }}
        </div>
      </div>
      <template #action>
        <a-button
          v-if="!isCurrentConfigActive"
          type="primary"
          size="small"
          :loading="loading"
          @click="$emit('activate')"
        >
          <template #icon>
            <icon-play-arrow-fill />
          </template>
          {{ $t('app.nftables.button.activateConfig') }}
        </a-button>
      </template>
    </a-alert>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    IconExclamationCircle,
    IconCheck,
    IconPlayArrowFill,
  } from '@arco-design/web-vue/es/icon';
  import type { ConfigFileDetail } from '@/api/nftables';

  interface Props {
    configType: 'local' | 'global';
    activeConfigType: 'local' | 'global';
    defaultFileDetail?: ConfigFileDetail | null;
    loading?: boolean;
  }

  interface Emits {
    (e: 'activate'): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
    defaultFileDetail: null,
  });

  defineEmits<Emits>();

  const { t } = useI18n();

  // 计算属性 - 修改激活状态判断逻辑，只有当 linked 为 true 时才显示"当前已激活"
  const isCurrentConfigActive = computed(() => {
    // 首先检查配置类型是否匹配
    const configTypeMatches = props.configType === props.activeConfigType;

    // 如果没有文件详情，使用原来的逻辑
    if (!props.defaultFileDetail) {
      return configTypeMatches;
    }

    // 只有当配置类型匹配且 linked 为 true 时，才认为当前配置已激活
    return configTypeMatches && props.defaultFileDetail.linked === true;
  });

  const configTypeLabel = computed(() =>
    props.configType === 'global'
      ? t('app.nftables.config.type.global')
      : t('app.nftables.config.type.local')
  );

  const statusLabel = computed(() =>
    isCurrentConfigActive.value
      ? t('app.nftables.status.currentActive')
      : t('app.nftables.status.inactive')
  );

  const configDescription = computed(() =>
    props.configType === 'global'
      ? t('app.nftables.config.globalDescription')
      : t('app.nftables.config.localDescription')
  );

  const activateHint = computed(() =>
    t('app.nftables.config.activateHint', {
      viewing:
        props.configType === 'global'
          ? t('app.nftables.config.type.global')
          : t('app.nftables.config.type.local'),
      active:
        props.activeConfigType === 'global'
          ? t('app.nftables.config.type.global')
          : t('app.nftables.config.type.local'),
    })
  );
</script>

<style scoped lang="less">
  .config-status-banner {
    margin-bottom: 20px;

    .status-description {
      margin-top: 8px;
      line-height: 1.5;

      .activate-hint {
        margin-top: 4px;
        font-size: 13px;
        color: var(--color-text-3);
      }
    }

    :deep(.arco-alert-action) {
      margin-top: 8px;
    }
  }
</style>
