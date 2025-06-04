<template>
  <idb-table
    ref="gridRef"
    :loading="loading"
    :columns="columns"
    :fetch="queryComposeApi"
  >
    <template #operation="{ record }">
      <idb-table-operation
        type="button"
        :options="getOperationOptions(record)"
      />
    </template>
  </idb-table>
  <edit-drawer ref="editRef" />
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    queryComposeApi,
    operateComposeApi,
    deleteComposeApi,
  } from '@/api/docker';
  import EditDrawer from './components/edit-drawer.vue';

  const { t } = useI18n();

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const loading = ref(false);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.compose.list.column.name'),
      width: 200,
    },
    {
      dataIndex: 'status',
      title: t('app.compose.list.column.container_status'),
      width: 120,
      render: ({ record }: { record: any }) => {
        return [
          record.containers.filter((item: any) => item.state === 'running')
            .length,
          record.container_number,
        ].join('/');
      },
    },
    {
      dataIndex: 'created_at',
      title: t('app.compose.list.column.created_at'),
      width: 160,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'center' as const,
      width: 150,
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    record: any,
    operation: 'start' | 'stop' | 'restart' | 'down' | 'up'
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

  const editRef = ref<InstanceType<typeof EditDrawer>>();
  const getOperationOptions = (record: any) => [
    {
      text: t('app.compose.list.operation.edit'),
      click: () => {
        editRef.value?.setParams({ name: record.name });
        editRef.value?.load();
        editRef.value?.show();
      },
    },
    {
      text: t('app.compose.list.operation.start'),
      click: async () => {
        await handleOperate(record, 'start');
      },
    },
    {
      text: t('app.compose.list.operation.stop'),
      click: async () => {
        await handleOperate(record, 'stop');
      },
    },
    {
      text: t('app.compose.list.operation.restart'),
      click: async () => {
        await handleOperate(record, 'restart');
      },
    },
    {
      text: t('app.compose.list.operation.down'),
      click: async () => {
        await handleOperate(record, 'down');
      },
    },
    {
      text: t('app.compose.list.operation.delete'),
      confirm: t('app.compose.list.operation.delete.confirm'),
      click: async () => {
        loading.value = true;
        try {
          await deleteComposeApi(record.name);
          Message.success(t('common.message.operationSuccess'));
        } catch (err: any) {
          Message.error(err.message || t('common.message.operationError'));
        } finally {
          loading.value = false;
        }
      },
    },
  ];
</script>
