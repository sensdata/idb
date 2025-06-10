<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="fetchNetworks">
    <template #leftActions>
      <a-button type="primary" @click="onCreateNetworkClick">
        {{ t('app.docker.network.list.action.create') }}
      </a-button>
    </template>
    <template #operation="{ record }">
      <idb-table-operation
        type="button"
        :options="getOperationOptions(record)"
      />
    </template>
  </idb-table>
  <create-network-drawer ref="createNetworkRef" @success="reload" />
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getNetworksApi, batchDeleteNetworkApi } from '@/api/docker';
  import IdbTable from '@/components/idb-table/index.vue';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import CreateNetworkDrawer from './components/create-network-drawer.vue';

  const { t } = useI18n();
  const gridRef = ref();
  const createNetworkRef = ref();
  const reload = () => gridRef.value?.reload();

  const fetchNetworks = async (params: any) => {
    return getNetworksApi(params);
  };

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.network.list.column.name'),
      width: 180,
    },
    {
      dataIndex: 'driver',
      title: t('app.docker.network.list.column.driver'),
      width: 120,
    },
    {
      dataIndex: 'subnet',
      title: t('app.docker.network.list.column.subnet'),
      width: 180,
    },
    {
      dataIndex: 'gateway',
      title: t('app.docker.network.list.column.gateway'),
      width: 180,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'center' as const,
      width: 180,
      slotName: 'operation',
    },
  ];

  const onCreateNetworkClick = () => createNetworkRef.value?.show();

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.network.list.operation.delete'),
      confirm: t('app.docker.network.list.operation.delete.confirm'),
      click: async () => {
        try {
          await batchDeleteNetworkApi({ force: 'false', sources: record.name });
          Message.success(
            t('app.docker.network.list.operation.delete.success')
          );
          reload();
        } catch (e: any) {
          Message.error(
            e.message || t('app.docker.network.list.operation.delete.failed')
          );
        }
      },
    },
  ];
</script>

<style scoped></style>
