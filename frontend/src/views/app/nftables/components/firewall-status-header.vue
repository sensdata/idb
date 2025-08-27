<template>
  <div class="header-status">
    <a-space>
      <a-tag
        v-if="firewallStatus"
        :color="
          statusInfo.installStatusText === $t('app.nftables.status.installed')
            ? 'rgb(var(--success-6))'
            : 'rgb(var(--danger-6))'
        "
        size="small"
      >
        {{ statusInfo.installStatusText }}
      </a-tag>

      <a-tag
        v-if="firewallStatus"
        :color="statusInfo.activeSystemColor"
        size="small"
      >
        {{ statusInfo.activeSystemText }}
      </a-tag>

      <!-- 切换按钮 -->
      <a-button
        v-if="statusInfo.canSwitch"
        type="primary"
        size="mini"
        :loading="switchLoading"
        @click="$emit('switch')"
      >
        <span class="switch-button-text">{{
          statusInfo.switchButtonText
        }}</span>
      </a-button>

      <a-button
        type="text"
        size="mini"
        :loading="statusLoading"
        @click="$emit('refresh')"
      >
        <template #icon>
          <icon-refresh />
        </template>
      </a-button>
    </a-space>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconRefresh } from '@arco-design/web-vue/es/icon';
  import type { NftablesStatus } from '@/api/nftables';
  import { getFirewallStatusInfo } from '../utils/firewall-status';

  interface Props {
    firewallStatus: NftablesStatus | null;
    statusLoading: boolean;
    switchLoading: boolean;
  }

  const props = defineProps<Props>();

  defineEmits<{
    switch: [];
    refresh: [];
  }>();

  const { t } = useI18n();

  // 获取状态信息
  const statusInfo = computed(() =>
    getFirewallStatusInfo(props.firewallStatus, t)
  );
</script>

<style scoped lang="less">
  .header-status {
    display: flex;
    align-items: center;

    .switch-button-text {
      white-space: nowrap;
    }
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .header-status {
      width: 100%;
      justify-content: flex-end;

      :deep(.arco-space-item) {
        margin-bottom: 4px;
      }

      .switch-button-text {
        font-size: 12px;
      }
    }
  }
</style>
