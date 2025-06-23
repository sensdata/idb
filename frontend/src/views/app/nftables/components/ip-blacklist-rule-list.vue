<template>
  <div class="ip-blacklist-rule-list">
    <idb-table
      :columns="columns"
      :data-source="tableDataSource"
      :loading="loading"
      :has-search="false"
      :has-batch="false"
      :auto-load="false"
      :page-size="PAGE_SIZE"
      row-key="ip"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleAdd">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.nftables.button.addIP') }}
        </a-button>
      </template>

      <!-- IP地址列 -->
      <template #ip="{ record }">
        <div class="ip-address">
          <span class="ip-text">{{ record.ip }}</span>
          <a-tag
            v-if="record.type !== 'single'"
            :color="getTypeColor(record.type)"
            size="small"
            class="type-tag"
          >
            {{ getTypeText(record.type) }}
          </a-tag>
        </div>
      </template>

      <!-- 类型列 -->
      <template #type="{ record }">
        <a-tag :color="getTypeColor(record.type)" size="small">
          {{ getTypeText(record.type) }}
        </a-tag>
      </template>

      <!-- 状态列 -->
      <template #status>
        <a-tag color="red" size="small">
          {{ $t('app.nftables.ipBlacklist.action.drop') }}
        </a-tag>
      </template>

      <!-- 操作列 -->
      <template #operations="{ record }">
        <a-button
          type="text"
          size="small"
          status="danger"
          @click="handleDelete(record.ip)"
        >
          {{ $t('common.delete') }}
        </a-button>
      </template>
    </idb-table>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconPlus } from '@arco-design/web-vue/es/icon';
  import { useConfirm } from '@/composables/confirm';
  import type { Column } from '@/components/idb-table/types';
  import type { ApiListResult } from '@/types/global';
  import IdbTable from '@/components/idb-table/index.vue';
  // IP黑名单规则接口
  export interface IPBlacklistRule {
    ip: string;
    type: 'single' | 'cidr' | 'range';
    action: 'drop' | 'reject';
    description?: string;
    createdAt?: string;
  }

  interface Props {
    ipRules: IPBlacklistRule[];
    loading?: boolean;
  }

  interface Emits {
    (e: 'add'): void;
    (e: 'delete', ip: string): void;
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
    ip: 250,
    type: 120,
    status: 120,
    operations: 120,
  } as const;

  // 表格数据源
  const tableDataSource = computed(
    (): ApiListResult<IPBlacklistRule> => ({
      items: props.ipRules,
      total: props.ipRules.length,
      page: 1,
      page_size: PAGE_SIZE,
    })
  );

  // 表格列配置
  const columns = computed((): Column<IPBlacklistRule>[] => [
    {
      title: t('app.nftables.ipBlacklist.rules.ip'),
      dataIndex: 'ip',
      slotName: 'ip',
      width: COLUMN_WIDTHS.ip,
      sortable: {
        sortDirections: ['ascend', 'descend'],
      },
    },
    {
      title: t('app.nftables.ipBlacklist.rules.type'),
      dataIndex: 'type',
      slotName: 'type',
      width: COLUMN_WIDTHS.type,
      align: 'center',
    },
    {
      title: t('app.nftables.ipBlacklist.rules.action'),
      dataIndex: 'status',
      slotName: 'status',
      width: COLUMN_WIDTHS.status,
      align: 'center',
    },
    {
      title: t('components.idbTable.columns.operations'),
      slotName: 'operations',
      width: COLUMN_WIDTHS.operations,
      align: 'center',
      fixed: 'right',
    },
  ]);

  // 获取类型颜色
  const getTypeColor = (type: string): string => {
    const colorMap = {
      single: 'blue',
      cidr: 'green',
      range: 'orange',
    };
    return colorMap[type as keyof typeof colorMap] || 'gray';
  };

  // 获取类型文本
  const getTypeText = (type: string): string => {
    const textMap = {
      single: t('app.nftables.ipBlacklist.type.single'),
      cidr: t('app.nftables.ipBlacklist.type.cidr'),
      range: t('app.nftables.ipBlacklist.type.range'),
    };
    return textMap[type as keyof typeof textMap] || type.toUpperCase();
  };

  // 事件处理
  const handleAdd = (): void => {
    emit('add');
  };

  const handleDelete = async (ip: string): Promise<void> => {
    const confirmed = await confirm({
      title: t('common.confirm.delete'),
      content: t('app.nftables.ipBlacklist.rules.deleteConfirm', { ip }),
    });

    if (confirmed) {
      emit('delete', ip);
    }
  };
</script>

<style scoped lang="less">
  .ip-blacklist-rule-list {
    .ip-address {
      display: flex;
      align-items: center;
      gap: 8px;

      .ip-text {
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 13px;
      }

      .type-tag {
        margin: 0;
      }
    }
  }
</style>
