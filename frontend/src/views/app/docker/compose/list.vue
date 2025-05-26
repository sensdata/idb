<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="queryComposeApi">
    <template #operation="{ record }">
      <idb-dropdown-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { queryComposeApi, operateComposeApi } from '@/api/docker';
  import { formatTime } from '@/utils/format';

  const { t } = useI18n();

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.compose.list.column.name'),
      width: 200,
    },
    {
      dataIndex: 'status',
      title: t('app.compose.list.column.status'),
      width: 120,
      render: ({ record }: { record: any }) => {
        const status = record.status;
        let color = '';
        let text = '';
        if (status === 'running') {
          color = 'color-success';
          text = t('app.compose.list.status.running');
        } else if (status === 'stopped') {
          color = 'color-danger';
          text = t('app.compose.list.status.stopped');
        } else {
          color = 'color-warning';
          text = t('app.compose.list.status.unknown');
        }
        return `<span class="${color}">${text}</span>`;
      },
    },
    {
      dataIndex: 'create_time',
      title: t('app.script.list.column.create_time'),
      width: 125,
      render: ({ record }: { record: any }) => {
        return formatTime(record.create_time);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 120,
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    record: any,
    operation: 'start' | 'stop' | 'down'
  ) => {
    try {
      await operateComposeApi({
        name: record.name,
        operation,
      });
      Message.success(t('app.compose.list.operation.success'));
      reload();
    } catch (e: any) {
      Message.error(e.message || t('app.compose.list.operation.failed'));
    }
  };

  const getOperationOptions = (record: any) => [
    {
      text: t('app.compose.list.operation.start'),
      visible: record.status !== 'running',
      click: async () => {
        await handleOperate(record, 'start');
      },
    },
    {
      text: t('app.compose.list.operation.stop'),
      visible: record.status === 'running',
      click: async () => {
        await handleOperate(record, 'stop');
      },
    },
    {
      text: t('app.compose.list.operation.down'),
      visible: true,
      click: async () => {
        await handleOperate(record, 'down');
      },
    },
    {
      text: t('app.compose.list.operation.edit'),
      visible: true,
      click: () => {
        // TODO: 打开编辑弹窗
      },
    },
    {
      text: t('app.compose.list.operation.delete'),
      visible: true,
      confirm: t('app.compose.list.operation.delete.confirm'),
      click: () => {
        // TODO: 删除操作
      },
    },
  ];
</script>
