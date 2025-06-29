<template>
  <idb-table ref="tableRef" :columns="columns" :fetch="fetchNetworks">
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
  <yaml-drawer ref="inspectRef" :title="$t('common.detail')" />
</template>

<script setup lang="ts">
  import { h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    getNetworksApi,
    batchDeleteNetworkApi,
    inspectApi,
  } from '@/api/docker';
  import YamlDrawer from '@/components/yaml-drawer/index.vue';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import CreateNetworkDrawer from './components/create-network-drawer.vue';

  const { t } = useI18n();
  const tableRef = ref();
  const createNetworkRef = ref();
  const inspectRef = ref<InstanceType<typeof YamlDrawer>>();
  const reload = () => tableRef.value?.reload();

  const fetchNetworks = async (params: any) => {
    return getNetworksApi(params);
  };

  async function handleInspect(record: any) {
    try {
      const data = await inspectApi({
        type: 'network',
        id: record.id!,
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
      title: t('app.docker.network.list.column.name'),
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
