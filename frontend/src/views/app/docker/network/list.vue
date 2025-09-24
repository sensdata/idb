<template>
  <div>
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />
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
  </div>
</template>

<script setup lang="ts">
  import { h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
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
      await showErrorWithDockerCheck(err?.message, err);
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
      align: 'left' as const,
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
          await showErrorWithDockerCheck(
            e.message || t('app.docker.network.list.operation.delete.failed'),
            e
          );
        }
      },
    },
  ];

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 如果 Docker 状态变化，可以重新加载网络列表
    if (status === 'installed') {
      reload();
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // Docker 安装完成后重新加载网络列表
    reload();
  };
</script>

<style scoped></style>
