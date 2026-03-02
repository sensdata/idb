<template>
  <div class="settings-modals-container">
    <a-modal
      :visible="portModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.portModal.save')"
      :cancel-text="$t('app.ssh.portModal.cancel')"
      @cancel="$emit('update:portModalVisible', false)"
      @ok="handlePortSave"
    >
      <template #title>{{ $t('app.ssh.portModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{ $t('app.ssh.portModal.port') }}</span>
          <div class="modal-input-wrapper">
            <a-input v-model="newPort" placeholder="22" />
            <div class="modal-field-description">
              {{ $t('app.ssh.portModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      :visible="listenModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.listenModal.save')"
      :cancel-text="$t('app.ssh.listenModal.cancel')"
      @cancel="$emit('update:listenModalVisible', false)"
      @ok="handleListenSave"
    >
      <template #title>{{ $t('app.ssh.listenModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{
            $t('app.ssh.listenModal.address')
          }}</span>
          <div class="modal-input-wrapper">
            <a-select
              v-model="newListenAddress"
              allow-search
              allow-create
              :loading="listenAddressLoading"
              placeholder="0.0.0.0"
            >
              <a-option
                v-for="option in listenAddressOptions"
                :key="option.value"
                :value="option.value"
              >
                <div class="listen-option-row">
                  <a-tag size="small" :color="option.color">
                    {{ option.tag }}
                  </a-tag>
                  <span class="listen-option-value">{{ option.value }}</span>
                </div>
              </a-option>
            </a-select>
            <div class="modal-field-description">
              {{ $t('app.ssh.listenModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      :visible="rootModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.rootModal.save')"
      :cancel-text="$t('app.ssh.rootModal.cancel')"
      @cancel="$emit('update:rootModalVisible', false)"
      @ok="handleRootSave"
    >
      <template #title>{{ $t('app.ssh.rootModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{ $t('app.ssh.rootModal.label') }}</span>
          <div class="modal-input-wrapper">
            <a-switch v-model="newRootEnabled" />
            <div class="modal-field-description">
              {{ $t('app.ssh.rootModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import { getSysInfoNetworkApi } from '@/api/sysinfo';

  const { t } = useI18n();

  const props = defineProps<{
    portModalVisible: boolean;
    listenModalVisible: boolean;
    rootModalVisible: boolean;
    port: string;
    listenAddress: string;
    rootEnabled: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:portModalVisible', value: boolean): void;
    (e: 'update:listenModalVisible', value: boolean): void;
    (e: 'update:rootModalVisible', value: boolean): void;
    (e: 'savePort', port: string): void;
    (e: 'saveListen', address: string): void;
    (e: 'saveRoot', enabled: boolean): void;
  }>();

  // 表单值
  const newPort = ref<string>(props.port);
  const newListenAddress = ref<string>(props.listenAddress);
  const newRootEnabled = ref<boolean>(props.rootEnabled);
  const listenAddressLoading = ref(false);
  const rawListenAddresses = ref<string[]>([
    '0.0.0.0',
    '127.0.0.1',
    '::',
    '::1',
  ]);

  type ListenOptionTagType =
    | 'current'
    | 'recommended'
    | 'private'
    | 'public'
    | 'loopback'
    | 'ipv6'
    | 'linklocal';

  interface ListenAddressOption {
    value: string;
    tag: string;
    color: string;
    priority: number;
  }

  const listenAddressOptions = ref<ListenAddressOption[]>([]);

  const normalizeAddress = (value: string): string => {
    return (value || '').trim();
  };

  const isIPv4 = (value: string): boolean => {
    return /^(25[0-5]|2[0-4]\d|1?\d?\d)(\.(25[0-5]|2[0-4]\d|1?\d?\d)){3}$/.test(
      value
    );
  };

  const isIPv6 = (value: string): boolean => {
    return value.includes(':');
  };

  const isLoopback = (value: string): boolean => {
    return value === '::1' || value.startsWith('127.');
  };

  const isAnyAddress = (value: string): boolean => {
    return value === '0.0.0.0' || value === '::';
  };

  const isPrivateIPv4 = (value: string): boolean => {
    return (
      value.startsWith('10.') ||
      value.startsWith('192.168.') ||
      /^172\.(1[6-9]|2\d|3[0-1])\./.test(value)
    );
  };

  const isLinkLocalIPv4 = (value: string): boolean => {
    return value.startsWith('169.254.');
  };

  const isLinkLocalIPv6 = (value: string): boolean => {
    return value.toLowerCase().startsWith('fe80:');
  };

  const isUlaIPv6 = (value: string): boolean => {
    const v = value.toLowerCase();
    return v.startsWith('fc') || v.startsWith('fd');
  };

  const getListenOptionMeta = (
    ip: string
  ): {
    type: ListenOptionTagType;
    priority: number;
  } => {
    const normalized = normalizeAddress(ip);
    if (normalized === normalizeAddress(newListenAddress.value)) {
      return { type: 'current', priority: 0 };
    }
    if (isAnyAddress(normalized)) {
      return { type: 'recommended', priority: 1 };
    }
    if (isLoopback(normalized)) {
      return { type: 'loopback', priority: 6 };
    }
    if (isIPv4(normalized)) {
      if (isPrivateIPv4(normalized)) {
        return { type: 'private', priority: 2 };
      }
      if (isLinkLocalIPv4(normalized)) {
        return { type: 'linklocal', priority: 7 };
      }
      return { type: 'public', priority: 4 };
    }
    if (isIPv6(normalized)) {
      if (isLinkLocalIPv6(normalized)) {
        return { type: 'linklocal', priority: 7 };
      }
      if (isUlaIPv6(normalized)) {
        return { type: 'private', priority: 3 };
      }
      return { type: 'ipv6', priority: 5 };
    }
    return { type: 'public', priority: 8 };
  };

  const buildListenAddressOptions = (
    addresses: string[]
  ): ListenAddressOption[] => {
    const colorMap: Record<ListenOptionTagType, string> = {
      current: 'rgb(var(--primary-6))',
      recommended: 'rgb(var(--success-6))',
      private: 'rgb(var(--arcoblue-6))',
      public: 'rgb(var(--warning-6))',
      loopback: 'rgb(var(--gray-6))',
      ipv6: 'rgb(var(--purple-6))',
      linklocal: 'rgb(var(--danger-6))',
    };

    const labelMap: Record<ListenOptionTagType, string> = {
      current: t('app.ssh.listenModal.option.current'),
      recommended: t('app.ssh.listenModal.option.recommended'),
      private: t('app.ssh.listenModal.option.private'),
      public: t('app.ssh.listenModal.option.public'),
      loopback: t('app.ssh.listenModal.option.loopback'),
      ipv6: t('app.ssh.listenModal.option.ipv6'),
      linklocal: t('app.ssh.listenModal.option.linklocal'),
    };

    return addresses
      .map((value) => normalizeAddress(value))
      .filter(Boolean)
      .map((value) => {
        const meta = getListenOptionMeta(value);
        return {
          value,
          tag: labelMap[meta.type],
          color: colorMap[meta.type],
          priority: meta.priority,
        };
      })
      .sort((a, b) => {
        if (a.priority !== b.priority) {
          return a.priority - b.priority;
        }
        return a.value.localeCompare(b.value, undefined, {
          numeric: true,
          sensitivity: 'base',
        });
      });
  };

  const refreshListenAddressOptions = () => {
    listenAddressOptions.value = buildListenAddressOptions(
      rawListenAddresses.value
    );
  };

  const fetchLocalListenAddressOptions = async (): Promise<void> => {
    try {
      listenAddressLoading.value = true;
      const res = await getSysInfoNetworkApi();
      const set = new Set(rawListenAddresses.value.map(normalizeAddress));

      (res.networks || []).forEach((network) => {
        (network.address || []).forEach((addr) => {
          const ip = normalizeAddress(addr.ip);
          if (!ip) return;
          if (isIPv4(ip) || isIPv6(ip)) {
            set.add(ip);
          }
        });
      });

      rawListenAddresses.value = [...set];
      refreshListenAddressOptions();
    } catch (error) {
      Message.warning(t('app.ssh.listenModal.loadIpsFailed'));
    } finally {
      listenAddressLoading.value = false;
    }
  };

  // 当属性变化时更新表单值
  watch(
    () => props.port,
    (value) => {
      newPort.value = value;
    }
  );

  watch(
    () => props.listenAddress,
    (value) => {
      newListenAddress.value = value;
      const current = normalizeAddress(value);
      if (current && !rawListenAddresses.value.includes(current)) {
        rawListenAddresses.value = [...rawListenAddresses.value, current];
      }
      refreshListenAddressOptions();
    }
  );

  watch(
    () => props.listenModalVisible,
    (visible) => {
      if (visible) {
        fetchLocalListenAddressOptions();
      }
    }
  );

  watch(
    () => props.rootEnabled,
    (value) => {
      newRootEnabled.value = value;
    }
  );

  watch(newListenAddress, (value) => {
    const current = normalizeAddress(value);
    if (current && !rawListenAddresses.value.includes(current)) {
      rawListenAddresses.value = [...rawListenAddresses.value, current];
    }
    refreshListenAddressOptions();
  });

  refreshListenAddressOptions();

  // 弹窗保存处理函数
  const handlePortSave = () => {
    emit('update:portModalVisible', false);
    emit('savePort', newPort.value);
  };

  const handleListenSave = () => {
    emit('update:listenModalVisible', false);
    emit('saveListen', newListenAddress.value);
  };

  const handleRootSave = () => {
    emit('update:rootModalVisible', false);
    emit('saveRoot', newRootEnabled.value);
  };
</script>

<style scoped lang="less">
  .modal-form-wrapper {
    padding: 0 20px;
  }

  .modal-form-item {
    display: flex;
    margin-bottom: 20px;
  }

  .modal-label {
    width: 80px;
    margin-right: 20px;
    font-weight: 500;
    line-height: 32px;
    color: var(--color-text-1);
    text-align: right;
  }

  .modal-input-wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .modal-field-description {
    margin-top: 4px;
    font-size: 12px;
    color: var(--color-text-3);
  }

  .listen-option-row {
    display: flex;
    gap: 8px;
    align-items: center;
    justify-content: space-between;
  }

  .listen-option-value {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
