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
          <a-tag class="port-tag">
            {{ record.port }}
          </a-tag>
        </template>

        <template #addresses="{ record }">
          <div class="addresses-list">
            <div
              v-for="(accessItem, index) in record.access"
              :key="index"
              class="address-row"
            >
              {{ accessItem.address }}
              <a-tag
                :color="getPortAccessInfoFromStatus(accessItem.status).color"
                size="small"
                :class="{
                  'access-status-tag': true,
                  'local-only-tag': getPortAccessInfoFromStatus(
                    accessItem.status
                  ).isLocalOnly,
                  'accessible-tag':
                    getPortAccessInfoFromStatus(accessItem.status).accessible &&
                    !getPortAccessInfoFromStatus(accessItem.status).isLocalOnly,
                  'restricted-access-tag':
                    !getPortAccessInfoFromStatus(accessItem.status)
                      .accessible &&
                    !getPortAccessInfoFromStatus(accessItem.status).isLocalOnly,
                }"
                class="font-medium"
              >
                {{ getPortAccessInfoFromStatus(accessItem.status).text }}
              </a-tag>
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
  import type { ProcessStatus, PortRuleSet } from '@/api/nftables';
  import ProcessInfo from './process-info.vue';
  import { getPortAccessInfoFromStatus as getPortAccessInfoFromStatusUtil } from '../utils/config-parser';

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
      align: 'left',
    },
    {
      title: t('app.nftables.config.columns.addresses'),
      dataIndex: 'addresses',
      slotName: 'addresses',
      width: 300,
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

  // 获取端口访问状态信息（基于API返回的状态）
  const getPortAccessInfoFromStatus = (status: string) => {
    return getPortAccessInfoFromStatusUtil(status, t);
  };

  defineExpose({
    tableRef,
  });
</script>

<style scoped lang="less">
  .process-section {
    background: #fff;
    border: 1px solid var(--color-border-1);
    border-radius: 0.429rem; /* 6px相对于14px根字体 (6/14=0.429) */
    padding: 1.429rem; /* 20px相对于14px根字体 (20/14=1.429) */
    margin-bottom: 1.429rem; /* 20px相对于14px根字体 (20/14=1.429) */

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 1.143rem; /* 16px相对于14px根字体 (16/14=1.143) */

      .section-title {
        margin: 0;
        font-size: 1.143rem; /* 16px相对于14px根字体 (16/14=1.143) */
        font-weight: 500;
        color: var(--color-text-1);
      }
    }

    .process-table {
      .addresses-list {
        display: flex;
        flex-direction: column;
        gap: 0.143rem; /* 2px相对于14px根字体 (2/14=0.143) */
        width: 100%;
        align-items: flex-start;

        .address-row {
          padding: 0.143rem 0; /* 2px相对于14px根字体 (2/14=0.143) */
          line-height: 1.4;
          color: var(--color-text-1);
          display: flex;
          align-items: center;
          gap: 0.571rem; /* 8px相对于14px根字体 (8/14=0.571) */
          width: 100%;
          justify-content: flex-start;

          .access-status-tag {
            margin-left: 0.571rem; /* 8px相对于14px根字体 (8/14=0.571) */
            font-size: 0.714rem; /* 10px相对于14px根字体 (10/14=0.714) */
            flex-shrink: 0;
          }

          .local-only-tag {
            background-color: var(--idblue-1) !important;
            color: var(--idblue-6) !important;
            padding: 0.25em 0.67em !important;
            border-radius: 0.167em;
            font-size: 0.857rem; /* 12px相对于14px根字体 (12/14=0.857) */
            line-height: 1.5;
            text-align: center;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border: none;
            font-weight: 500;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }

          .restricted-access-tag {
            background-color: var(
              --idbred-1
            ) !important; /* 浅红色背景: #FFECE8 */
            color: var(--idbred-6) !important; /* 深红色文字: #F53F3F */
            padding: 0.25em 0.67em !important;
            border-radius: 0.167em;
            font-size: 0.857rem; /* 12px相对于14px根字体 (12/14=0.857) */
            line-height: 1.5;
            text-align: center;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border: none;
            font-weight: 500;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }

          .accessible-tag {
            background-color: var(
              --idbturquoise-1
            ) !important; /* 浅碧涛青背景: #E6FBF9 */
            color: var(--idbturquoise-6) !important; /* 深碧涛青文字: #0FC6C2 */
            padding: 0.25em 0.67em !important;
            border-radius: 0.167em;
            font-size: 0.857rem; /* 12px相对于14px根字体 (12/14=0.857) */
            line-height: 1.5;
            text-align: center;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            border: none;
            font-weight: 500;

            // 覆盖 Arco Design 的默认样式
            &.arco-tag {
              min-width: auto;
              max-width: none;
            }
          }
        }
      }

      .port-tag {
        background-color: var(--idblue-1) !important;
        color: var(--idblue-6) !important;
        padding: 0.25em 0.67em !important; /* 使用em单位，相对于font-size */
        border-radius: 0.167em; /* 相对于当前字体大小 */
        font-size: 1rem; /* 14px相对于14px根字体 (14/14=1) */
        line-height: 1.5; /* 标准相对行高 */
        text-align: center;
        display: inline-flex;
        align-items: center;
        justify-content: center;
        border: none;
        font-weight: 500;

        // 覆盖 Arco Design 的默认样式
        &.arco-tag {
          min-width: auto;
          max-width: none;
        }
      }

      .actions-container {
        display: flex;
        align-items: center;
        justify-content: flex-start;
        gap: 0.571rem; /* 8px相对于14px根字体 (8/14=0.571) */

        .arco-link {
          display: inline-flex;
          align-items: center;
          gap: 0.286rem; /* 4px相对于14px根字体 (4/14=0.286) */
          font-size: 1rem; /* 14px相对于14px根字体 (14/14=1) */
        }
      }
    }
  }

  /* 响应式设计 */
  @media (max-width: 768px) {
    .process-section {
      padding: 1.143rem; /* 16px相对于14px根字体 (16/14=1.143) */

      .section-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 0.857rem; /* 12px相对于14px根字体 (12/14=0.857) */
      }
    }
  }

  @media (max-width: 480px) {
    .process-section {
      padding: 0.857rem; /* 12px相对于14px根字体 (12/14=0.857) */

      .process-table {
        .addresses-list {
          .address-row {
            flex-direction: column;
            align-items: flex-start;
            gap: 0.286rem; /* 4px相对于14px根字体 (4/14=0.286) */

            .access-status-tag {
              margin-left: 0;
              align-self: flex-start;
            }
          }
        }

        .actions-container {
          flex-direction: column;
          align-items: flex-start;
          gap: 0.286rem; /* 4px相对于14px根字体 (4/14=0.286) */
        }
      }
    }
  }
</style>
