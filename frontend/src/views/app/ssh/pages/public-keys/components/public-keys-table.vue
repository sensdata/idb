<template>
  <div class="public-keys-container">
    <idb-table
      ref="tableRef"
      :loading="loading"
      :dataSource="dataSource"
      row-key="key"
      :pagination="false"
      :columns="columns"
      @reload="handleReload"
    >
      <template #leftActions>
        <a-button
          type="primary"
          class="generate-key-btn"
          @click="handleShowAddModal"
        >
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.ssh.publicKeys.addPublicKey') }}
        </a-button>
      </template>

      <template #operations="{ record }">
        <a-space>
          <a-button
            type="text"
            status="danger"
            @click="() => handleRemove(record)"
          >
            {{ $t('app.ssh.publicKeys.remove') }}
          </a-button>
        </a-space>
      </template>
    </idb-table>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconPlus } from '@arco-design/web-vue/es/icon';
  import type { Column } from '@/components/idb-table/types';
  import type { ApiListResult } from '@/types/global';
  import { AuthKeyInfo } from '@/views/app/ssh/types';

  // Props 定义
  interface Props {
    loading: boolean;
    keys: AuthKeyInfo[];
  }

  const props = defineProps<Props>();

  // 事件定义
  const emit = defineEmits<{
    (e: 'reload'): void;
    (e: 'showAddModal'): void;
    (e: 'remove', record: AuthKeyInfo): void;
  }>();

  // 国际化
  const { t } = useI18n();

  // 表格引用
  const tableRef = ref<HTMLElement | null>(null);

  // 事件处理函数
  const handleReload = () => emit('reload');
  const handleShowAddModal = () => emit('showAddModal');
  const handleRemove = (record: AuthKeyInfo) => emit('remove', record);

  // 表格数据源
  const dataSource = computed<ApiListResult<AuthKeyInfo>>(() => ({
    items: props.keys,
    total: props.keys.length,
    page: 1,
    page_size: 20,
  }));

  // 表格列定义
  const columns = ref<Column[]>([
    {
      title: t('app.ssh.publicKeys.columns.algorithm'),
      dataIndex: 'algorithm',
      width: 100,
    },
    {
      title: t('app.ssh.publicKeys.columns.key'),
      dataIndex: 'key',
      ellipsis: true,
    },
    {
      title: t('app.ssh.publicKeys.columns.comment'),
      dataIndex: 'comment',
      width: 150,
    },
    {
      title: t('app.ssh.publicKeys.columns.operations'),
      dataIndex: 'operations',
      width: 100,
      slotName: 'operations',
    },
  ]);

  // 暴露方法给父组件使用
  defineExpose({
    tableRef,
  });
</script>

<style scoped lang="less">
  .public-keys-container {
    padding: 0;
  }

  .generate-key-btn {
    background-color: #6c52fa;

    &:hover,
    &:focus {
      background-color: #8b74ff;
    }
  }
</style>
