<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="fetchVolumes">
    <template #leftActions>
      <a-button type="primary" @click="onCreateVolumeClick">
        {{ t('app.docker.volume.list.action.create') }}
      </a-button>
    </template>
    <template #operation="{ record }">
      <idb-table-operation
        type="button"
        :options="getOperationOptions(record)"
      />
    </template>
  </idb-table>
  <create-volume-drawer ref="createVolumeRef" @success="reload" />
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { formatTime } from '@/utils/format';
  import { getVolumesApi, batchDeleteVolumeApi } from '@/api/docker';
  import IdbTable from '@/components/idb-table/index.vue';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import CreateVolumeDrawer from './components/create-volume-drawer.vue';

  const { t } = useI18n();
  const gridRef = ref();
  const createVolumeRef = ref();
  const reload = () => gridRef.value?.reload();

  const fetchVolumes = async (params: any) => {
    return getVolumesApi(params);
  };

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.volume.list.column.name'),
      width: 180,
    },
    {
      dataIndex: 'driver',
      title: t('app.docker.volume.list.column.driver'),
      width: 100,
    },
    {
      dataIndex: 'mount_point',
      title: t('app.docker.volume.list.column.mount_point'),
      width: 320,
    },
    {
      dataIndex: 'created_at',
      title: t('app.docker.volume.list.column.created'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return formatTime(record.created_at);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'center' as const,
      width: 100,
      slotName: 'operation',
    },
  ];

  const onCreateVolumeClick = () => createVolumeRef.value?.show();

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.volume.list.operation.delete'),
      confirm: t('app.docker.volume.list.operation.delete.confirm'),
      click: async () => {
        try {
          await batchDeleteVolumeApi({ force: 'false', sources: record.name });
          Message.success(t('app.docker.volume.list.operation.delete.success'));
          reload();
        } catch (e: any) {
          Message.error(
            e.message || t('app.docker.volume.list.operation.delete.failed')
          );
        }
      },
    },
  ];
</script>

<style scoped></style>
