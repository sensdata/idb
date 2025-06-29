<template>
  <idb-table ref="tableRef" :columns="columns" :fetch="fetchVolumes">
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
  <yaml-drawer ref="inspectRef" :title="$t('common.detail')" />
</template>

<script setup lang="ts">
  import { h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { formatTime } from '@/utils/format';
  import {
    getVolumesApi,
    batchDeleteVolumeApi,
    inspectApi,
  } from '@/api/docker';
  import YamlDrawer from '@/components/yaml-drawer/index.vue';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import CreateVolumeDrawer from './components/create-volume-drawer.vue';

  const { t } = useI18n();
  const tableRef = ref();
  const inspectRef = ref<InstanceType<typeof YamlDrawer>>();
  const createVolumeRef = ref();
  const reload = () => tableRef.value?.reload();

  const fetchVolumes = async (params: any) => {
    return getVolumesApi(params);
  };

  async function handleInspect(record: any) {
    try {
      const data = await inspectApi({
        type: 'volume',
        id: record.name!,
      });
      inspectRef.value?.setContent(
        JSON.stringify(JSON.parse(data.info), null, 2)
      );
      inspectRef.value?.show();
    } catch (err: any) {
      Message.error(err?.message);
    }
  }

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.volume.list.column.name'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              handleInspect(record);
            },
            hoverable: false,
          },
          {
            default: () => {
              return record.name;
            },
          }
        );
      },
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
