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
        row-key="pid"
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
          <a-tag color="blue">
            {{ record.port }}
          </a-tag>
        </template>

        <template #addresses="{ record }">
          <div class="addresses-list">
            <div
              v-for="(address, index) in record.addresses"
              :key="index"
              class="address-row"
            >
              {{ address }}
              <a-tag
                :color="getPortAccessInfo(record.port, address).color"
                size="small"
                class="access-status-tag"
              >
                {{ getPortAccessInfo(record.port, address).text }}
              </a-tag>
            </div>
          </div>
        </template>

        <template #actions="{ record }">
          <div class="actions-container">
            <a-button
              type="text"
              size="small"
              :disabled="isPortRuleExists(record.port)"
              @click="$emit('addPortRule', record.port)"
            >
              <template #icon>
                <icon-plus />
              </template>
              {{ $t('app.nftables.button.configPort') }}
            </a-button>
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
  import type { ProcessStatus, PortRuleSet } from '@/api/nftables';
  import ProcessInfo from './process-info.vue';
  import { getPortAccessInfo as getPortAccessInfoUtil } from '../utils/config-parser';

  interface Props {
    processData: ProcessStatus[];
    openPorts: Set<number>;
    portRules: PortRuleSet[];
    loading: boolean;
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
      width: 280,
      align: 'left',
    },
    {
      title: t('app.nftables.config.columns.port'),
      dataIndex: 'port',
      slotName: 'port',
      width: 120,
      align: 'center',
    },
    {
      title: t('app.nftables.config.columns.addresses'),
      dataIndex: 'addresses',
      slotName: 'addresses',
      width: 300,
      align: 'center',
    },
    {
      title: t('common.operation'),
      dataIndex: 'actions',
      slotName: 'actions',
      width: 160,
      align: 'center',
    },
  ];

  // 数据源格式化为idb-table需要的格式
  const processDataSource = computed<ApiListResult<ProcessStatus>>(() => ({
    items: props.processData,
    total: props.processData.length,
    page: 1,
    page_size: props.processData.length,
  }));

  // 检查端口规则是否已存在
  const isPortRuleExists = (port: number): boolean => {
    return props.portRules.some((rule) => rule.port === port);
  };

  // 获取端口访问状态信息
  const getPortAccessInfo = (port: number, address: string) => {
    return getPortAccessInfoUtil(port, address, props.openPorts, t);
  };

  defineExpose({
    tableRef,
  });
</script>

<style scoped lang="less">
  .process-section {
    background: #fff;
    border: 1px solid var(--color-border-1);
    border-radius: 6px;
    padding: 20px;
    margin-bottom: 20px;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .section-title {
        margin: 0;
        font-size: 16px;
        font-weight: 500;
        color: var(--color-text-1);
      }
    }

    .process-table {
      :deep(.arco-table-th) {
        background: var(--color-bg-2);
        font-weight: 500;
        color: var(--color-text-1);
        text-align: center;
      }

      :deep(.arco-table-td) {
        padding: 12px;
        text-align: center;
      }

      // 第一列（进程列）左对齐
      :deep(.arco-table-th:first-child) {
        text-align: left;
      }

      :deep(.arco-table-td:first-child) {
        text-align: left;
      }

      :deep(.arco-table-tr):hover {
        background: var(--color-bg-1);
      }

      .addresses-list {
        display: flex;
        flex-direction: column;
        gap: 2px;
        width: 100%;

        .address-row {
          padding: 2px 0;
          line-height: 1.4;
          color: var(--color-text-2);
          display: flex;
          align-items: center;
          gap: 8px;
          width: 100%;
          justify-content: center;

          .access-status-tag {
            margin-left: 8px;
            font-size: 10px;
            flex-shrink: 0;
          }
        }
      }

      .actions-container {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 4px;
      }
    }
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .process-section {
      padding: 16px;

      .section-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 12px;
      }
    }
  }

  @media (max-width: 480px) {
    .process-section {
      padding: 12px;

      .process-table {
        .addresses-list {
          .address-row {
            flex-direction: column;
            align-items: flex-start;
            gap: 4px;

            .access-status-tag {
              margin-left: 0;
              align-self: flex-end;
            }
          }
        }

        .actions-container {
          flex-direction: column;
          gap: 2px;
        }
      }
    }
  }
</style>
