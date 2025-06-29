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
  <down-confirm-modal ref="downConfirmRef" @confirm="afterDownConfirm" />
</template>

<script lang="ts" setup>
  import { ref, h, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import {
    queryComposeApi,
    operateComposeApi,
    deleteComposeApi,
  } from '@/api/docker';
  import EditDrawer from './components/edit-drawer.vue';
  import DownConfirmModal from './components/down-confirm-modal.vue';

  const { t } = useI18n();
  const router = useRouter();

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const loading = ref(false);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.compose.list.column.name'),
      width: 200,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              router.push(`/app/docker/container/${record.container_number}`);
            },
            hoverable: false,
          },
          {
            default: () => record.name,
          }
        );
      },
    },
    {
      dataIndex: 'status',
      title: t('app.docker.compose.list.column.container_status'),
      width: 120,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              router.push(`/app/docker/container/${record.container_number}`);
            },
            hoverable: false,
          },
          {
            default: () =>
              [
                record.containers.filter(
                  (item: any) => item.state === 'running'
                ).length,
                record.container_number,
              ].join('/'),
          }
        );
      },
    },
    {
      dataIndex: 'created_at',
      title: t('app.docker.compose.list.column.created_at'),
      width: 160,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'center' as const,
      width: 180,
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    name: string,
    operation: 'start' | 'stop' | 'restart' | 'up' | 'down',
    params?: {
      delete_volumes?: boolean;
    }
  ) => {
    try {
      const result = await operateComposeApi({
        name,
        operation,
        ...params,
      });
      if (result.success) {
        Message.success(
          t('app.docker.compose.list.operation.success', {
            command: result.command,
          })
        );
      } else {
        Message.error(
          t('app.docker.compose.list.operation.failed', {
            command: result.command,
            message: result.message,
          })
        );
      }
      reload();
    } catch (e: any) {
      Message.error(e.message || t('app.docker.compose.list.operation.error'));
    }
  };

  const afterDownConfirm = async (params: {
    name: string;
    delete_volumes: boolean;
  }) => {
    await handleOperate(params.name, 'down', {
      delete_volumes: params.delete_volumes,
    });
  };

  const editRef = ref<InstanceType<typeof EditDrawer>>();
  const downConfirmRef = ref<InstanceType<typeof DownConfirmModal>>();
  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.compose.list.operation.edit'),
      click: () => {
        editRef.value?.setParams({ name: record.name });
        editRef.value?.load();
        editRef.value?.show();
      },
    },
    {
      text: t('app.docker.compose.list.operation.start'),
      click: async () => {
        await handleOperate(record.name, 'start');
      },
    },
    {
      text: t('app.docker.compose.list.operation.stop'),
      confirm: t('app.docker.compose.list.operation.stop.confirm'),
      click: async () => {
        await handleOperate(record.name, 'stop');
      },
    },
    {
      text: t('app.docker.compose.list.operation.restart'),
      click: async () => {
        await handleOperate(record.name, 'restart');
      },
    },
    {
      text: t('app.docker.compose.list.operation.up'),
      click: async () => {
        await handleOperate(record.name, 'up');
      },
    },
    {
      text: t('app.docker.compose.list.operation.down'),
      click: async () => {
        downConfirmRef.value?.show(record.name);
      },
    },
    {
      text: t('app.docker.compose.list.operation.delete'),
      confirm: t('app.docker.compose.list.operation.delete.confirm'),
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
