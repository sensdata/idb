<template>
  <idb-table
    ref="tableRef"
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
    <template #activated="{ record }: { record: HostItem }">
      <div
        v-if="record.statusReady && record.agent_status?.connected === 'online'"
      >
        <div v-if="record.activated" class="color-success">
          {{ $t('manage.host.list.activated.yes') }}
        </div>
        <div v-else class="color-warning">
          {{ $t('manage.host.list.activated.no') }}
        </div>
      </div>
      <div v-else class="color-text-3">-</div>
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
      <span
        v-if="
          !record.statusReady || record.agent_status?.connected !== 'online'
        "
        >-</span
      >
      <a-progress
        v-else
        class="inline-progress"
        :percent="toProgressPercent(record.cpu)"
      />
    </template>
    <template #memory="{ record }: { record: HostItem }">
      <span
        v-if="
          !record.statusReady || record.agent_status?.connected !== 'online'
        "
        >-</span
      >
      <a-progress
        v-else
        class="inline-progress"
        :percent="toProgressPercent(record.mem)"
        :color="'var(--idbturquoise-6)'"
      />
    </template>
    <template #disk="{ record }: { record: HostItem }">
      <span
        v-if="
          !record.statusReady || record.agent_status?.connected !== 'online'
        "
        >-</span
      >
      <a-progress
        v-else
        class="inline-progress"
        :percent="toProgressPercent(record.disk)"
        :color="getDiskUsageColor(record.disk)"
      />
    </template>
    <template #network="{ record }: { record: HostItem }">
      <span
        v-if="
          !record.statusReady || record.agent_status?.connected !== 'online'
        "
        >-</span
      >
      <div v-else class="network-cell">
        <up-stream-icon class="network-icon" />
        <span>{{ formatTransferSpeed(record.tx) }}</span>
        <down-stream-icon class="network-icon downstream" />
        <span>{{ formatTransferSpeed(record.rx) }}</span>
      </div>
    </template>
    <template #operation="{ record }: { record: HostItem }">
      <idb-table-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <host-create ref="formRef" @ok="reload"></host-create>
  <host-edit ref="editRef" @ok="reload" />
  <ssh-terminal ref="termRef" />
  <ssh-form ref="sshFormRef" @ok="reload" />
  <agent-form ref="agentFormRef" @ok="reload" />
  <install-agent ref="installAgentRef" @ok="reload" />
  <uninstall-agent ref="uninstallAgentRef" @ok="reload" />
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
    restartHostAgentApi,
    connectAllHostsStatusFollowApi,
    activateHostApi,
    type HostStatusFollowItem,
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
  import AgentForm from './components/agent-form.vue';
  import InstallAgent from './components/install-agent.vue';
  import UninstallAgent from './components/uninstall-agent.vue';

  interface HostItem extends HostEntity {
    statusReady: boolean;
    can_upgrade: boolean;
  }

  const { t } = useI18n();
  const router = useRouter();
  const termRef = ref<InstanceType<typeof SshTerminal>>();
  const editRef = ref<InstanceType<typeof HostEdit>>();
  const sshFormRef = ref<InstanceType<typeof SshForm>>();
  const agentFormRef = ref<InstanceType<typeof AgentForm>>();
  const installAgentRef = ref<InstanceType<typeof InstallAgent>>();
  const uninstallAgentRef = ref<InstanceType<typeof UninstallAgent>>();

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
      dataIndex: 'activated',
      title: t('manage.host.list.column.activated'),
      width: 100,
      slotName: 'activated',
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

  const tableRef = ref();
  const dataRef = ref<ApiListResult<HostItem>>();
  const sseRef = ref<EventSource>();

  const getOperationOptions = (record: HostItem) => {
    return [
      {
        text: t('manage.host.list.operation.activate'),
        visible: record.statusReady && !record.activated && !record.default,
        click: async () => {
          try {
            await activateHostApi(record.id);
            Message.success(t('manage.host.list.activate.success'));
            // 刷新表格以获取最新的 agent_version/agent_latest 等字段，避免需要手动刷新页面
            tableRef.value?.reload();
          } catch (error: any) {
            Message.error(
              error?.message || t('manage.host.list.activate.error')
            );
          }
        },
      },
      {
        text: t('manage.host.list.operation.upgradeAgent'),
        visible: record.can_upgrade,
        confirm: t('manage.host.list.operation.upgradeAgent.confirm'),
        click: () => {
          installAgentRef.value?.startUpgrade(record.id);
        },
      },
      {
        text: t('manage.host.list.operation.goto'),
        visible:
          record.agent_status?.status === 'installed' && !record.can_upgrade,
        click: () => {
          router.push({
            name: DEFAULT_APP_ROUTE_NAME,
            query: { id: record.id },
          });
        },
      },
      {
        text: t('manage.host.list.operation.goto'),
        visible: record.agent_status?.status !== 'installed',
        click: () => {
          installAgentRef.value?.confirmInstall(record.id);
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
        text: t('manage.host.list.operation.updateAgent'),
        click: () => {
          const form = agentFormRef.value;
          form?.reset();
          form?.load(record.id);
          form?.show();
        },
      },
      {
        text: t('manage.host.list.operation.installAgent'),
        visible: record.agent_status?.status !== 'installed',
        click: () => {
          installAgentRef.value?.startInstall(record.id);
        },
      },

      {
        text: t('manage.host.list.operation.uninstallAgent'),
        confirm: t('manage.host.list.uninstallAgent.confirm'),
        visible: record.agent_status?.status === 'installed',
        click: () => {
          uninstallAgentRef.value?.startUninstall(record.id);
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
            tableRef.value?.reload();
          } catch (error: any) {
            Message.error(error?.message);
          }
        },
      },
    ];
  };

  const toProgressPercent = (value: number) => +(value / 100).toFixed(3);

  const getDiskUsageColor = (value: number) => {
    const percent = toProgressPercent(value);
    if (percent >= 0.9) {
      return 'rgb(var(--red-6))';
    }
    if (percent >= 0.8) {
      return 'rgb(var(--orange-6))';
    }
    return 'var(--idbturquoise-6)';
  };

  const reload = () => {
    // 只重新加载表格数据，状态更新会由 dataRef 变化触发
    tableRef.value?.reload();
  };

  const stopAllSSE = () => {
    if (sseRef.value) {
      sseRef.value.close();
      sseRef.value = undefined;
    }
  };

  onUnmounted(() => {
    stopAllSSE();
  });

  const autoActivateDefaultHosts = async () => {
    if (!dataRef.value?.items) {
      return;
    }

    const autoActivatePromises = dataRef.value.items
      .filter((item) => item.default && item.statusReady && !item.activated)
      .map(async (item) => {
        try {
          await activateHostApi(item.id);
          item.activated = true; // Update the local state immediately
        } catch (error) {
          console.error(
            `Failed to auto-activate default host ${item.id}:`,
            error
          );
        }
      });

    if (autoActivatePromises.length > 0) {
      await Promise.all(autoActivatePromises);
    }
  };

  const handleAllHostsStatusUpdate = (event: Event) => {
    try {
      const statusList: HostStatusFollowItem[] = JSON.parse(
        (event as MessageEvent).data
      );

      if (!dataRef.value?.items) {
        return;
      }

      // 更新每个主机的状态
      statusList.forEach((statusData) => {
        const item = dataRef.value?.items?.find((i) => i.id === statusData.id);
        if (item) {
          // 更新监控数据
          item.activated = statusData.activated;
          item.cpu = statusData.cpu;
          item.disk = statusData.disk;
          item.mem = statusData.mem;
          item.mem_total = statusData.mem_total;
          item.mem_used = statusData.mem_used;
          item.rx = statusData.rx;
          item.tx = statusData.tx;
          item.statusReady = true;
          item.can_upgrade = statusData.can_upgrade;

          // 更新 agent 状态
          if (item.agent_status) {
            item.agent_status.status = statusData.installed;
            item.agent_status.connected = statusData.connected;
          }
        }
      });

      tableRef.value?.setData(dataRef.value);

      // 自动激活默认主机（如果尚未激活）
      autoActivateDefaultHosts();
    } catch (e) {
      console.error('Failed to parse hosts status:', e);
    }
  };

  const startSSEForHosts = () => {
    if (!dataRef.value?.items || dataRef.value.items.length === 0) {
      return;
    }

    // 停止之前的 SSE 连接
    stopAllSSE();

    // 收集所有主机的 ID
    const hostIds = dataRef.value.items.map((item) => item.id);

    // 建立统一的 SSE 连接
    const es = connectAllHostsStatusFollowApi(hostIds);
    es.addEventListener('status', handleAllHostsStatusUpdate);
    es.onerror = (err) => {
      console.error('SSE error:', err);
    };
    sseRef.value = es;
  };

  const afterFetchHook = async (data: ApiListResult<HostItem>) => {
    dataRef.value = data;

    // 启动 SSE 连接，数据会通过 SSE 实时推送
    startSSEForHosts();
    return data;
  };

  const formRef = ref<InstanceType<typeof HostCreate>>();
  const handleCreate = () => {
    const form = formRef.value;
    form?.reset();
    form?.loadOptions();
    form?.show();
  };
</script>

<style scoped>
  .inline-progress {
    width: 100%;
  }

  .inline-progress :deep(.arco-progress-line) {
    display: flex;
    align-items: center;
  }

  .inline-progress :deep(.arco-progress-line-wrapper) {
    flex: 1;
    min-width: 0;
  }

  .inline-progress :deep(.arco-progress-line-text, .arco-progress-steps-text) {
    width: 48px;
    margin-left: 10px;
    font-variant-numeric: tabular-nums;
    text-align: right;
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
