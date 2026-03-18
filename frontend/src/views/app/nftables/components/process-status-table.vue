<template>
  <div class="process-section">
    <div class="section-header">
      <h3 class="section-title">{{
        $t('app.nftables.config.processStatusList')
      }}</h3>
    </div>

    <div class="process-table">
      <idb-table
        ref="tableRef"
        :loading="loading"
        :data-source="processDataSource"
        row-key="rowKey"
        :pagination="false"
        :columns="processColumns"
        :has-search="false"
        :has-toolbar="true"
        @reload="$emit('reload')"
      >
        <template #process="{ record }">
          <process-info :process="record.process" :pid="record.pid" />
        </template>

        <template #port="{ record }">
          <a-tag class="port-tag">
            {{ record.port }}
          </a-tag>
        </template>

        <template #addresses="{ record }">
          <div class="addresses-list">
            <div class="addresses-grid">
              <div
                v-for="family in getAddressFamilies(record.access)"
                :key="family.type"
                class="family-column"
              >
                <div class="address-family-title">
                  {{ family.label }}
                </div>
                <div v-if="family.items.length === 0" class="address-empty">
                  -
                </div>
                <div v-else class="address-items">
                  <div
                    v-for="(item, index) in family.items"
                    :key="`${family.type}-${item.address}-${index}`"
                    class="address-row"
                  >
                    <a-tag size="small" class="address-type-tag">
                      {{ item.groupShortLabel }}
                    </a-tag>
                    <span class="address-text">{{ item.address }}</span>
                    <a-tag
                      :color="item.statusInfo.color"
                      size="small"
                      :class="{
                        'access-status-tag': true,
                        'local-only-tag': item.statusInfo.isLocalOnly,
                        'accessible-tag':
                          item.statusInfo.accessible &&
                          !item.statusInfo.isLocalOnly,
                        'restricted-access-tag':
                          !item.statusInfo.accessible &&
                          !item.statusInfo.isLocalOnly,
                      }"
                      class="font-medium"
                    >
                      {{ item.statusInfo.text }}
                    </a-tag>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </template>

        <template #actions="{ record }">
          <div class="actions-container">
            <a-link
              :disabled="isPortRuleExists(record.port)"
              @click="$emit('addPortRule', record.port)"
            >
              <icon-plus />
              {{ $t('app.nftables.button.configPort') }}
            </a-link>
            <a-button
              v-if="isPortRuleExists(record.port)"
              type="text"
              size="small"
              @click="$emit('editPortRule', record.port)"
            >
              <template #icon>
                <icon-edit />
              </template>
              {{ $t('common.edit') }}
            </a-button>
          </div>
        </template>
      </idb-table>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconPlus, IconEdit } from '@arco-design/web-vue/es/icon';
  import type { Column } from '@/components/idb-table/types';
  import type { ApiListResult } from '@/types/global';
  import type {
    ProcessStatus,
    PortRangeRule,
    PortAccessStatus,
  } from '@/api/nftables';
  import ProcessInfo from './process-info.vue';
  import { getPortAccessInfoFromStatus as getPortAccessInfoFromStatusUtil } from '../utils/config-parser';

  interface Props {
    processData: ProcessStatus[];
    portRules: PortRangeRule[];
    loading: boolean;
  }
  interface ProcessTableRow extends ProcessStatus {
    rowKey: string;
  }

  type AddressGroupType = 'external' | 'internal' | 'local';
  type AddressFamilyType = 'ipv4' | 'ipv6';

  interface GroupedAccessItem {
    address: string;
    status: PortAccessStatus['status'];
    groupType: AddressGroupType;
    groupShortLabel: string;
    statusInfo: ReturnType<typeof getPortAccessInfoFromStatusUtil>;
  }

  interface AddressFamily {
    label: string;
    type: AddressFamilyType;
    items: GroupedAccessItem[];
  }

  const props = defineProps<Props>();

  defineEmits<{
    reload: [];
    addPortRule: [port: number];
    editPortRule: [port: number];
  }>();

  const { t } = useI18n();
  const tableRef = ref();

  // 表格列定义
  const processColumns: Column[] = [
    {
      title: t('app.nftables.config.columns.process'),
      dataIndex: 'process',
      slotName: 'process',
      width: 220,
      align: 'left',
    },
    {
      title: t('app.nftables.config.columns.port'),
      dataIndex: 'port',
      slotName: 'port',
      width: 120,
      align: 'left',
    },
    {
      title: t('app.nftables.config.columns.addresses'),
      dataIndex: 'addresses',
      slotName: 'addresses',
      width: 520,
      align: 'left',
    },
    {
      title: t('common.operation'),
      dataIndex: 'actions',
      slotName: 'actions',
      width: 160,
      align: 'left',
    },
  ];

  // 数据源格式化为idb-table需要的格式
  const processRows = computed<ProcessTableRow[]>(() =>
    props.processData.map((item, index) => ({
      ...item,
      rowKey: `${item.process}-${item.pid}-${item.port}-${index}`,
    }))
  );

  const processDataSource = computed<ApiListResult<ProcessTableRow>>(() => ({
    items: processRows.value,
    total: processRows.value.length,
    page: 1,
    page_size: processRows.value.length,
  }));

  // 检查端口规则是否已存在
  const isPortRuleExists = (port: number): boolean => {
    return props.portRules.some(
      (rule) => port >= rule.port_start && port <= rule.port_end
    );
  };

  // 获取端口访问状态信息（基于API返回的状态）
  const getPortAccessInfoFromStatus = (status: string) => {
    return getPortAccessInfoFromStatusUtil(status, t);
  };

  const isLocalAddress = (address: string): boolean => {
    const value = address.trim().toLowerCase();
    return (
      value === 'localhost' ||
      value === '::1' ||
      value === '[::1]' ||
      value === '127.0.0.1' ||
      value.startsWith('127.')
    );
  };

  const isInternalAddress = (address: string): boolean => {
    const value = address.trim().toLowerCase();
    const v4 = value.match(/^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/);

    if (v4) {
      const first = Number(v4[1]);
      const second = Number(v4[2]);
      if (first === 10) return true;
      if (first === 192 && second === 168) return true;
      if (first === 172 && second >= 16 && second <= 31) return true;
      if (first === 169 && second === 254) return true;
    }

    // IPv6 ULA/link-local
    return (
      value.startsWith('fc') ||
      value.startsWith('fd') ||
      value.startsWith('fe80')
    );
  };

  const classifyAddressType = (address: string): AddressGroupType => {
    if (isLocalAddress(address)) return 'local';
    if (isInternalAddress(address)) return 'internal';
    return 'external';
  };

  const isIPv4Address = (address: string): boolean => {
    return /^(\d{1,3}\.){3}\d{1,3}$/.test(address.trim().toLowerCase());
  };

  const classifyAddressFamily = (address: string): AddressFamilyType => {
    if (isIPv4Address(address)) return 'ipv4';
    return 'ipv6';
  };

  const groupShortLabelByType = (type: AddressGroupType): string => {
    if (type === 'external')
      return t('app.nftables.config.addressType.external');
    if (type === 'internal')
      return t('app.nftables.config.addressType.internal');
    return t('app.nftables.config.addressType.local');
  };

  const familyLabelByType = (type: AddressFamilyType): string => {
    if (type === 'ipv4') return t('app.nftables.config.addressFamily.ipv4');
    return t('app.nftables.config.addressFamily.ipv6');
  };

  const getAddressFamilies = (access: PortAccessStatus[]): AddressFamily[] => {
    const families: Record<AddressFamilyType, GroupedAccessItem[]> = {
      ipv4: [],
      ipv6: [],
    };

    for (const item of access || []) {
      const groupType = classifyAddressType(item.address);
      const familyType = classifyAddressFamily(item.address);
      families[familyType].push({
        address: item.address,
        status: item.status,
        groupType,
        groupShortLabel: groupShortLabelByType(groupType),
        statusInfo: getPortAccessInfoFromStatus(item.status),
      });
    }

    return (['ipv4', 'ipv6'] as AddressFamilyType[]).map((type) => ({
      type,
      label: familyLabelByType(type),
      items: families[type],
    }));
  };

  defineExpose({
    tableRef,
  });
</script>

<style scoped lang="less">
  .process-section {
    padding: 1.429rem;
    margin-bottom: 1.429rem;
    background: var(--color-bg-2);
    border: 1px solid var(--color-border-1);
    border-radius: 0.429rem;
    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 1.143rem;
      .section-title {
        margin: 0;
        font-size: 1.143rem;
        font-weight: 500;
        color: var(--color-text-1);
      }
    }
    .process-table {
      .addresses-list {
        width: 100%;
        .addresses-grid {
          display: grid;
          grid-template-columns: repeat(2, minmax(0, 1fr));
          gap: 8px 12px;
          width: 100%;
        }
        .address-family-title {
          margin-bottom: 4px;
          font-size: 11px;
          font-weight: 500;
          color: var(--color-text-4);
        }
        .family-column {
          min-width: 0;
        }
        .address-items {
          display: flex;
          flex-direction: column;
          gap: 4px;
          width: 100%;
        }
        .address-empty {
          font-size: 12px;
          color: var(--color-text-3);
        }
        .address-row {
          display: grid;
          grid-template-columns: auto minmax(0, 1fr) auto;
          gap: 6px;
          align-items: center;
          max-width: 100%;
          padding: 2px 8px;
          line-height: 1.2;
          color: var(--color-text-1);
          background: var(--color-fill-1);
          border-radius: 4px;
          .address-text {
            flex: 1;
            min-width: 0;
            overflow: hidden;
            text-overflow: ellipsis;
            font-size: 12px;
            white-space: nowrap;
          }
          .address-type-tag {
            flex-shrink: 0;
            color: var(--color-text-3);
            background: var(--color-fill-2);
            border: 1px solid var(--color-border-2);
          }
          .access-status-tag {
            flex-shrink: 0;
            margin-left: 0;
            font-size: 0.714rem; /* 10px */
          }
          .local-only-tag {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 0.25em 0.67em !important;
            font-size: 0.857rem;
            font-weight: 500;
            line-height: 1.5;
            color: var(--idblue-6) !important;
            text-align: center;
            background-color: var(--idblue-1) !important;
            border: none;
            border-radius: 0.167em;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }
          .restricted-access-tag {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 0.25em 0.67em !important;
            font-size: 0.857rem;
            font-weight: 500;
            line-height: 1.5;
            color: var(--idbred-6) !important; /* 深红色文字: #F53F3F */
            text-align: center;
            background-color: var(
              --idbred-1
            ) !important; /* 浅红色背景: #FFECE8 */

            border: none;
            border-radius: 0.167em;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }
          .accessible-tag {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 0.25em 0.67em !important;
            font-size: 0.857rem;
            font-weight: 500;
            line-height: 1.5;
            color: var(--idbturquoise-6) !important; /* 深碧涛青文字: #0FC6C2 */
            text-align: center;
            background-color: var(
              --idbturquoise-1
            ) !important; /* 浅碧涛青背景: #E6FBF9 */

            border: none;
            border-radius: 0.167em;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }
        }
      }
      .port-tag {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding: 0.25em 0.67em !important; /* 使用em单位，相对于font-size */
        font-size: 1rem;
        font-weight: 500;
        line-height: 1.5; /* 标准相对行高 */
        color: var(--idblue-6) !important;
        text-align: center;
        background-color: var(--idblue-1) !important;
        border: none;
        border-radius: 0.167em; /* 相对于当前字体大小 */

        // 覆盖 Arco Design 的默认样式
        &.arco-tag {
          min-width: auto;
          max-width: none;
        }
      }
      .actions-container {
        display: flex;
        gap: 0.571rem;
        align-items: center;
        justify-content: flex-start;
        .arco-link {
          display: inline-flex;
          gap: 0.286rem;
          align-items: center;
          font-size: 1rem;
          text-decoration: none;
          &:hover {
            text-decoration: none;
          }
        }
        .arco-btn {
          height: auto;
          padding: 0;
          font-size: 1rem;
        }
      }
    }
  }

  /* 响应式设计 */
  @media (width <= 768px) {
    .process-section {
      padding: 1.143rem;
      .section-header {
        flex-direction: column;
        gap: 0.857rem;
        align-items: flex-start;
      }
      .process-table {
        .actions-container {
          flex-direction: column;
          gap: 0.286rem;
          align-items: flex-start;
        }
      }
    }
  }

  @media (width <= 480px) {
    .process-section {
      padding: 0.857rem;
      .process-table {
        .addresses-list {
          .addresses-grid {
            grid-template-columns: 1fr;
          }
        }
        .actions-container {
          flex-direction: column;
          gap: 0.286rem;
          align-items: flex-start;
        }
      }
    }
  }
</style>
