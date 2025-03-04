<template>
  <idb-table
    ref="gridRef"
    :columns="columns"
    :fetch="getHostListApi"
    :afterFetchHook="afterFetchHook"
  >
    <template #leftActions>
      <a-button type="primary" @click="handleCreate">
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('manage.host.list.action.add') }}
      </a-button>
    </template>
    <template #agent="{ record }: { record: HostItem }">
      <div
        v-if="record.agent_status?.status !== 'installed'"
        class="color-danger"
        >{{ $t('manage.host.list.agent.uninstalled') }}</div
      >
      <div
        v-else-if="record.agent_status?.connected === 'online'"
        class="color-success"
        >{{ $t('manage.host.list.agent.online') }}</div
      >
      <div
        v-else-if="record.agent_status?.connected === 'offline'"
        class="color-danger"
        >{{ $t('manage.host.list.agent.offline') }}</div
      >
    </template>
    <template #name="{ record }: { record: HostItem }">
      <div>{{ record.addr }}</div>
      <div>
        <span>{{ record.name }}</span>
        <span v-if="record.default" class="cyan-6">
          ({{ $t('manage.host.list.name_default') }})
        </span>
      </div>
    </template>
    <template #cpu="{ record }: { record: HostItem }">
      <span v-if="!record.statusReady">-</span>
      <a-progress
        v-else
        class="inline-progress"
        :percent="+(record.cpu / 100).toFixed(3)"
      />
    </template>
    <template #memory="{ record }: { record: HostItem }">
      <span v-if="!record.statusReady">-</span>
      <a-progress
        v-else
        class="inline-progress"
        :percent="+(record.mem / 100).toFixed(3)"
        color="#0FC6C2"
      />
    </template>
    <template #disk="{ record }: { record: HostItem }">
      <span v-if="!record.statusReady">-</span>
      <a-progress
        v-else
        class="inline-progress"
        :percent="+(record.disk / 100).toFixed(3)"
        color="#0FC6C2"
      />
    </template>
    <template #network="{ record }: { record: HostItem }">
      <span v-if="!record.statusReady">-</span>
      <div v-else class="network-cell">
        <up-stream-icon class="network-icon" />
        <span>{{ formatTransferSpeed(record.tx) }}</span>
        <down-stream-icon class="network-icon downstream" />
        <span>{{ formatTransferSpeed(record.rx) }}</span>
      </div>
    </template>
    <template #operation="{ record }: { record: HostItem }">
      <idb-dropdown-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <host-create ref="formRef" @ok="reload"></host-create>
  <host-edit ref="editRef" @ok="reload" />
  <ssh-terminal ref="termRef" />
  <ssh-form ref="sshFormRef" @ok="reload" />
</template>

<script lang="ts" setup>
  import { onUnmounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { HostEntity } from '@/entity/Host';
  import {
    deleteHostApi,
    getHostListApi,
    getHostStatusApi,
    restartHostAgentApi,
  } from '@/api/host';
  import { DEFAULT_APP_ROUTE_NAME } from '@/router/constants';
  import { ApiListResult } from '@/types/global';
  import { formatTransferSpeed } from '@/utils/format';
  import UpStreamIcon from '@/assets/icons/upstream.svg';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import HostCreate from './components/create.vue';
  import SshTerminal from './components/ssh-terminal.vue';
  import HostEdit from './components/edit.vue';
  import SshForm from './components/ssh-form.vue';

  interface HostItem extends HostEntity {
    statusReady: boolean;
  }

  const { t } = useI18n();
  const router = useRouter();
  const termRef = ref<InstanceType<typeof SshTerminal>>();
  const editRef = ref<InstanceType<typeof HostEdit>>();
  const sshFormRef = ref<InstanceType<typeof SshForm>>();

  const columns = [
    {
      dataIndex: 'name',
      title: t('manage.host.list.column.name'),
      width: 150,
      slotName: 'name',
    },
    {
      dataIndex: 'group_name',
      title: t('manage.host.list.column.group_name'),
      width: 125,
      render: ({ record }: { record: HostItem }) => {
        return record.group
          ? record.group?.group_name
          : t('manage.host.list.group_default');
      },
    },
    {
      dataIndex: 'agent_status',
      title: t('manage.host.list.column.agent'),
      width: 120,
      slotName: 'agent',
    },
    {
      dataIndex: 'cpu',
      title: t('manage.host.list.column.cpu'),
      width: 110,
      slotName: 'cpu',
    },
    {
      dataIndex: 'memory',
      title: t('manage.host.list.column.memory'),
      width: 110,
      slotName: 'memory',
    },
    {
      dataIndex: 'disk',
      title: t('manage.host.list.column.disk'),
      width: 110,
      slotName: 'disk',
    },
    {
      dataIndex: 'network',
      title: t('manage.host.list.column.network'),
      width: 160,
      slotName: 'network',
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      slotName: 'operation',
    },
  ];

  const getOperationOptions = (record: HostItem) => {
    return [
      {
        text: t('manage.host.list.operation.goto'),
        click: () => {
          router.push({
            name: DEFAULT_APP_ROUTE_NAME,
            query: { id: record.id },
          });
        },
      },
      {
        text: t('manage.host.list.operation.setting'),
        click: () => {
          const form = editRef.value;
          form?.reset();
          form?.loadOptions();
          form?.setData(record);
          form?.show();
        },
      },
      {
        text: t('manage.host.list.operation.updateSSH'),
        click: () => {
          const form = sshFormRef.value;
          form?.reset();
          form?.load(record.id);
          form?.show();
        },
      },
      {
        text: t('manage.host.list.operation.sshTerminal'),
        click: () => {
          termRef.value?.show(record.id);
        },
      },
      {
        text: t('manage.host.list.operation.restart'),
        confirm: t('manage.host.list.restart.confirm'),
        click: async () => {
          try {
            await restartHostAgentApi(record.id);
            Message.success(t('manage.host.list.restart.success'));
          } catch (error: any) {
            Message.error(error?.message);
          }
        },
      },
      {
        text: t('manage.host.list.operation.delete'),
        confirm: t('manage.host.list.delete.confirm'),
        click: async () => {
          try {
            await deleteHostApi(record.id);
            Message.success(t('manage.host.list.delete.success'));
          } catch (error: any) {
            Message.error(error?.message);
          }
        },
      },
    ];
  };

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const dataRef = ref<ApiListResult<HostItem>>();

  const fetchListStatus = async () => {
    if (!dataRef.value?.items) {
      return;
    }
    await Promise.all(
      dataRef.value?.items.map((item) =>
        getHostStatusApi(item.id).then((statusData) => {
          Object.assign(item, statusData, {
            statusReady: true,
          });
        })
      )
    );
    gridRef.value?.setData(dataRef.value);
  };

  const timerRef = ref<number>();
  const stopAutoFetchStatus = () => {
    if (timerRef.value) {
      clearInterval(timerRef.value);
    }
  };
  const startAutoFetchStatus = () => {
    stopAutoFetchStatus();
    fetchListStatus();
    timerRef.value = window.setInterval(() => {
      fetchListStatus();
    }, 5000);
  };

  const afterFetchHook = (data: ApiListResult<HostItem>) => {
    dataRef.value = data;
    startAutoFetchStatus();
    return data;
  };

  const formRef = ref<InstanceType<typeof HostCreate>>();
  const handleCreate = () => {
    const form = formRef.value;
    form?.reset();
    form?.loadOptions();
    form?.show();
  };

  onUnmounted(() => {
    stopAutoFetchStatus();
  });
</script>

<style scoped>
  .inline-progress :deep(.arco-progress-line-text, .arco-progress-steps-text) {
    min-width: 0;
    margin-left: 10px;
  }

  .network-cell {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: flex-start;
    min-width: 0;
    color: var(--color-text-3);
  }

  .network-icon {
    width: 12px;
    height: 12px;
    margin-right: 2px;
  }

  .network-icon.downstream {
    margin-left: 8px;
  }
</style>
