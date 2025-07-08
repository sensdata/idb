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
        :color="'var(--idbturquoise-6)'"
      />
    </template>
    <template #disk="{ record }: { record: HostItem }">
      <span v-if="!record.statusReady">-</span>
      <a-progress
        v-else
        class="inline-progress"
        :percent="+(record.disk / 100).toFixed(3)"
        :color="'var(--idbturquoise-6)'"
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
      <idb-table-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <host-create ref="formRef" @ok="reload"></host-create>
  <host-edit ref="editRef" @ok="reload" />
  <ssh-terminal ref="termRef" />
  <ssh-form ref="sshFormRef" @ok="reload" />
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
    getHostStatusApi,
    restartHostAgentApi,
    getHostAgentStatusApi,
    connectHostStatusFollowApi,
  } from '@/api/host';
  import { DEFAULT_APP_ROUTE_NAME } from '@/router/constants';
  import { ApiListResult } from '@/types/global';
  import { formatTransferSpeed } from '@/utils/format';
  import { compareVersion } from '@/helper/utils';
  import UpStreamIcon from '@/assets/icons/upstream.svg';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import HostCreate from './components/create.vue';
  import SshTerminal from './components/ssh-terminal.vue';
  import HostEdit from './components/edit.vue';
  import SshForm from './components/ssh-form.vue';
  import InstallAgent from './components/install-agent.vue';
  import UninstallAgent from './components/uninstall-agent.vue';

  interface HostItem extends HostEntity {
    statusReady: boolean;
  }

  const { t } = useI18n();
  const router = useRouter();
  const termRef = ref<InstanceType<typeof SshTerminal>>();
  const editRef = ref<InstanceType<typeof HostEdit>>();
  const sshFormRef = ref<InstanceType<typeof SshForm>>();
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
          if (record.agent_status?.status === 'installed') {
            router.push({
              name: DEFAULT_APP_ROUTE_NAME,
              query: { id: record.id },
            });
          } else {
            installAgentRef.value?.confirmInstall(record.id);
          }
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
        text: t('manage.host.list.operation.installAgent'),
        visible: record.agent_status?.status !== 'installed',
        click: () => {
          installAgentRef.value?.startInstall(record.id);
        },
      },
      {
        text: t('manage.host.list.operation.upgradeAgent'),
        visible: compareVersion(record.agent_latest, record.agent_version) > 0,
        confirm: t('manage.host.list.operation.upgradeAgent.confirm'),
        click: () => {
          installAgentRef.value?.startUpgrade(record.id);
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
          } catch (error: any) {
            Message.error(error?.message);
          }
        },
      },
    ];
  };

  const tableRef = ref();
  const dataRef = ref<ApiListResult<HostItem>>();
  const isLoading = ref(false);
  const sseMap = ref<Map<number, EventSource>>(new Map());

  const fetchListStatus = async () => {
    if (!dataRef.value?.items || isLoading.value) {
      return;
    }

    isLoading.value = true;

    try {
      // 处理所有节点，每个节点只发送一个请求
      const requests = dataRef.value.items.map(async (item) => {
        // 根据节点的安装状态选择不同的API
        if (item.agent_status?.status === 'installed') {
          // 已安装代理的节点：获取监控数据
          try {
            const statusData = await getHostStatusApi(item.id);
            if (statusData) {
              // 更新监控数据
              item.cpu = statusData.cpu;
              item.disk = statusData.disk;
              item.mem = statusData.mem;
              item.mem_total = statusData.mem_total;
              item.mem_used = statusData.mem_used;
              item.rx = statusData.rx;
              item.tx = statusData.tx;
              item.statusReady = true;

              // 成功获取数据意味着代理在线
              if (item.agent_status) {
                item.agent_status.connected = 'online';
              }
            }
          } catch (error) {
            console.error('获取节点状态数据失败', item.id);
            // 如果请求失败，将代理标记为离线，但保持已安装状态
            if (item.agent_status) {
              item.agent_status.connected = 'offline';
            }
          }
        } else {
          // 未安装代理的节点：只更新代理状态
          try {
            const agentStatus = await getHostAgentStatusApi(item.id);
            if (agentStatus) {
              item.agent_status = {
                status: agentStatus.status || 'unknown',
                connected: agentStatus.connected || 'offline',
              };
            }
          } catch (error) {
            // 出错时设置默认状态
            item.agent_status = {
              status: 'unknown',
              connected: 'offline',
            };
          }
        }
      });

      // 等待所有请求完成
      await Promise.all(requests);
    } catch (err) {
      console.error(err);
    } finally {
      isLoading.value = false;
    }
  };

  const reload = () => {
    // 只重新加载表格数据，状态更新会由 dataRef 变化触发
    tableRef.value?.reload();
  };

  const stopAllSSE = () => {
    for (const es of sseMap.value.values()) {
      es.close();
    }
    sseMap.value = new Map();
  };

  onUnmounted(() => {
    stopAllSSE();
  });

  const startSSEForHosts = () => {
    if (!dataRef.value?.items) {
      return;
    }
    dataRef.value.items.forEach((item) => {
      if (
        item.agent_status?.status === 'installed' &&
        !sseMap.value.has(item.id)
      ) {
        const es = connectHostStatusFollowApi(item.id);
        es.addEventListener('status', (event) => {
          try {
            const statusData = JSON.parse(event.data);
            item.cpu = statusData.cpu;
            item.disk = statusData.disk;
            item.mem = statusData.mem;
            item.mem_total = statusData.mem_total;
            item.mem_used = statusData.mem_used;
            item.rx = statusData.rx;
            item.tx = statusData.tx;
            item.statusReady = true;
            if (item.agent_status) {
              item.agent_status.connected = 'online';
            }
            if (tableRef.value) {
              tableRef.value.setData(dataRef.value);
            }
          } catch (e) {
            console.error(e);
          }
        });
        es.onerror = () => {
          if (item.agent_status) {
            item.agent_status.connected = 'offline';
          }
        };
        sseMap.value.set(item.id, es);
      }
    });

    for (const id of sseMap.value.keys()) {
      if (!dataRef.value?.items?.find((item) => item.id === id)) {
        sseMap.value.get(id)?.close();
        sseMap.value.delete(id);
      }
    }
  };

  const afterFetchHook = async (data: ApiListResult<HostItem>) => {
    dataRef.value = data;
    await fetchListStatus();
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
