<template>
  <div class="port-rule-list">
    <idb-table
      :columns="columns"
      :data-source="tableDataSource"
      :loading="loading"
      :has-search="false"
      :has-batch="false"
      :auto-load="false"
      :page-size="PAGE_SIZE"
      row-key="port"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleAdd">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.nftables.button.addRule') }}
        </a-button>
      </template>

      <!-- 协议列 -->
      <template #protocol="{ record }">
        <a-tag :color="getProtocolColor(record.protocol)" size="small">
          {{ getProtocolText(record.protocol) }}
        </a-tag>
      </template>

      <!-- 端口列 -->
      <template #port="{ record }">
        <span class="port-number">
          {{
            Array.isArray(record.port) ? record.port.join(', ') : record.port
          }}
        </span>
      </template>

      <!-- 动作列 -->
      <template #action="{ record }">
        <a-tag :color="getActionColor(record.action)" size="small">
          {{ getActionText(record.action) }}
        </a-tag>
      </template>

      <!-- 来源列 -->
      <template #source="{ record }">
        <div v-if="record.source" class="source-info">
          <icon-user class="source-icon" />
          <span>{{ record.source }}</span>
        </div>
        <span v-else class="text-gray-400">-</span>
      </template>

      <!-- 高级规则列 -->
      <template #advancedRules="{ record }">
        <div
          v-if="record.rules && record.rules.length > 0"
          class="advanced-rules"
        >
          <a-space direction="vertical" size="mini">
            <a-tag
              v-for="(rule, index) in record.rules"
              :key="index"
              :color="getAdvancedRuleColor(rule.type)"
              size="small"
            >
              {{ getAdvancedRuleText(rule) }}
            </a-tag>
          </a-space>
        </div>
        <span v-else class="text-gray-400">-</span>
      </template>

      <!-- 描述列 -->
      <template #description="{ record }">
        <span class="description-text">
          {{ record.description || '' }}
        </span>
      </template>

      <!-- 操作列 -->
      <template #operations="{ record }">
        <div class="operation">
          <a-button type="text" size="small" @click="handleEdit(record)">
            {{ $t('common.edit') }}
          </a-button>
          <a-button
            type="text"
            size="small"
            status="danger"
            @click="handleDelete(record)"
          >
            {{ $t('common.delete') }}
          </a-button>
        </div>
      </template>
    </idb-table>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconPlus, IconUser } from '@arco-design/web-vue/es/icon';
  import { useConfirm } from '@/composables/confirm';
  import type { PortRule } from '@/api/nftables';
  import type { Column } from '@/components/idb-table/types';
  import type { ApiListResult } from '@/types/global';
  import IdbTable from '@/components/idb-table/index.vue';

  interface Props {
    portRules: PortRule[];
    loading?: boolean;
  }

  interface Emits {
    (e: 'add'): void;
    (e: 'edit', rule: PortRule): void;
    (e: 'delete', rule: PortRule): void;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  const emit = defineEmits<Emits>();

  // 定义插槽类型
  defineSlots<{
    leftActions?: () => any;
  }>();

  const { t } = useI18n();
  const { confirm } = useConfirm();

  // 常量定义
  const PAGE_SIZE = 20;
  const COLUMN_WIDTHS = {
    protocol: 100,
    port: 120,
    action: 100,
    source: 150,
    operations: 120,
  } as const;

  // 表格数据源
  const tableDataSource = computed(
    (): ApiListResult<PortRule> => ({
      items: props.portRules,
      total: props.portRules.length,
      page: 1,
      page_size: PAGE_SIZE,
    })
  );

  // 表格列配置
  const columns = computed((): Column<PortRule>[] => [
    {
      title: t('app.nftables.config.rules.protocol'),
      dataIndex: 'protocol',
      slotName: 'protocol',
      width: COLUMN_WIDTHS.protocol,
      align: 'center',
    },
    {
      title: t('app.nftables.config.rules.port'),
      dataIndex: 'port',
      slotName: 'port',
      width: COLUMN_WIDTHS.port,
      align: 'center',
      sortable: {
        sortDirections: ['ascend', 'descend'],
      },
    },
    {
      title: t('app.nftables.config.rules.action'),
      dataIndex: 'action',
      slotName: 'action',
      width: COLUMN_WIDTHS.action,
      align: 'center',
    },
    {
      title: t('app.nftables.config.rules.source'),
      dataIndex: 'source',
      slotName: 'source',
      width: COLUMN_WIDTHS.source,
    },
    {
      title: t('app.nftables.form.advancedRules'),
      dataIndex: 'rules',
      slotName: 'advancedRules',
      width: 180,
    },
    {
      title: t('app.nftables.config.rules.description'),
      dataIndex: 'description',
      slotName: 'description',
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('components.idbTable.columns.operations'),
      slotName: 'operations',
      width: COLUMN_WIDTHS.operations,
      align: 'center',
      fixed: 'right',
    },
  ]);

  // 获取协议颜色
  const getProtocolColor = (protocol?: string): string => {
    if (!protocol) return 'gray';
    const colorMap = {
      tcp: 'blue',
      udp: 'purple',
      both: 'cyan',
    } as const;
    return colorMap[protocol as keyof typeof colorMap] || 'gray';
  };

  // 获取协议文本
  const getProtocolText = (protocol?: string): string => {
    if (!protocol) return t('app.nftables.config.rules.unknown');
    return protocol.toUpperCase();
  };

  // 获取动作颜色
  const getActionColor = (action?: string): string => {
    if (!action) return 'gray';
    const colorMap = {
      accept: 'green',
      drop: 'red',
      reject: 'orange',
    } as const;
    return colorMap[action as keyof typeof colorMap] || 'gray';
  };

  // 获取动作文本
  const getActionText = (action?: string): string => {
    if (!action) return t('app.nftables.config.rules.unknown');
    const actionMap = {
      accept: () => t('app.nftables.config.rules.allow'),
      drop: () => t('app.nftables.config.rules.deny'),
      reject: () => t('app.nftables.config.rules.deny'),
    } as const;
    return (
      actionMap[action as keyof typeof actionMap]?.() || action.toUpperCase()
    );
  };

  // 获取高级规则颜色
  const getAdvancedRuleColor = (type: string): string => {
    const colorMap = {
      default: 'blue',
      rate_limit: 'orange',
      concurrent_limit: 'purple',
    } as const;
    return colorMap[type as keyof typeof colorMap] || 'gray';
  };

  // 获取高级规则文本
  const getAdvancedRuleText = (rule: any): string => {
    const typeMap = {
      default: () => t('app.nftables.form.basicRule'),
      rate_limit: () =>
        `${t('app.nftables.form.rateLimit')}: ${rule.rate || 'N/A'}`,
      concurrent_limit: () =>
        `${t('app.nftables.form.concurrentLimit')}: ${rule.count || 'N/A'}`,
    } as const;
    return typeMap[rule.type as keyof typeof typeMap]?.() || rule.type;
  };

  // 事件处理
  const handleAdd = () => {
    emit('add');
  };

  const handleEdit = (rule: PortRule) => {
    emit('edit', rule);
  };

  const handleDelete = async (rule: PortRule) => {
    const portText = Array.isArray(rule.port)
      ? rule.port.join(', ')
      : rule.port;
    const confirmed = await confirm(
      t('app.nftables.rules.deleteConfirm', { port: portText })
    );
    if (confirmed) {
      emit('delete', rule);
    }
  };
</script>

<style scoped lang="less">
  .port-rule-list {
    .port-number {
      font-size: 16px;
      font-weight: 600;
      color: var(--color-text-1);
    }

    .source-info {
      display: flex;
      align-items: center;
      gap: 4px;
      color: var(--color-text-2);

      .source-icon {
        width: 14px;
        height: 14px;
      }
    }

    .description-text {
      color: var(--color-text-2);
    }

    .advanced-rules {
      .arco-space {
        width: 100%;
      }

      .arco-tag {
        max-width: 160px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .empty-state {
      padding: 40px 20px;
      text-align: center;
    }

    // 自定义表格样式
    :deep(.arco-table) {
      .arco-table-td {
        .arco-tag {
          margin: 0;
        }
      }
    }

    // 操作列样式
    .operation :deep(.arco-btn-size-small) {
      padding-right: 4px;
      padding-left: 4px;
    }

    // 响应式设计
    @media (max-width: 768px) {
      :deep(.arco-table) {
        .arco-table-container {
          overflow-x: auto;
        }
      }
    }
  }
</style>
