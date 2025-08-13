<template>
  <idb-table
    ref="tableRef"
    :columns="columns"
    :params="params"
    :filters="filters"
    :fetch="queryContainersApi"
    :afterFetchHook="afterFetchHook"
  >
    <template #leftActions>
      <a-button type="primary" @click="handlePrune">
        {{ t('app.docker.container.list.action.prune') }}
      </a-button>
    </template>
    <template #usage="{ record }">
      <div v-if="record.cpu_percent != null" class="text-[13px]">
        <span>{{ $t('app.docker.container.cpu') }}: </span>
        <span> {{ record.cpu_percent.toFixed(1) }}% </span>
      </div>
      <div v-if="record.memory_usage != null" class="text-[13px]">
        <span>{{ $t('app.docker.container.memory') }}: </span>
        <span>
          {{ formatMemorySize(record.memory_usage) }} /
          {{ formatMemorySize(record.memory_limit) }}
        </span>
      </div>
    </template>
    <template #ports="{ record }">
      <div
        v-for="(port, index) in record.ports"
        :key="port"
        :class="{
          'mt-2': index > 0,
        }"
      >
        <a-tag bordered>{{ port }}</a-tag>
      </div>
    </template>
    <template #operation="{ record }">
      <idb-table-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <logs-modal ref="logsRef" />
  <terminal-drawer ref="termRef" />
  <stop-confirm-modal ref="stopConfirmRef" @confirm="afterStopConfirm" />
</template>

<script lang="ts" setup>
  import { h, onUnmounted, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { ApiListResult } from '@/types/global';
  import {
    queryContainersApi,
    operateContainersApi,
    getContainerUsageApi,
    connectContainerUsageFollowApi,
    pruneApi,
  } from '@/api/docker';
  import { useConfirm } from '@/composables/confirm';
  import { pick } from 'lodash';
  import { formatMemorySize } from '@/utils/format';
  import LogsModal from './components/logs-modal.vue';
  import TerminalDrawer from './components/terminal-drawer.vue';
  import StopConfirmModal from './components/stop-confirm-modal.vue';

  const { t } = useI18n();

  const route = useRoute();
  const params = {
    composeId: +route.params.composeId || undefined,
  };

  const filters = [
    {
      label: t('app.docker.container.list.filter.state'),
      field: 'state',
      type: 'select' as const,
      defaultValue: 'all',
      options: [
        {
          label: t('app.docker.container.list.state.all'),
          value: 'all',
        },
        {
          label: t('app.docker.container.list.state.created'),
          value: 'created',
        },
        {
          label: t('app.docker.container.list.state.running'),
          value: 'running',
        },
        {
          label: t('app.docker.container.list.state.paused'),
          value: 'paused',
        },
        {
          label: t('app.docker.container.list.state.restarting'),
          value: 'restarting',
        },
        {
          label: t('app.docker.container.list.state.removing'),
          value: 'removing',
        },
        {
          label: t('app.docker.container.list.state.exited'),
          value: 'exited',
        },
        {
          label: t('app.docker.container.list.state.dead'),
          value: 'dead',
        },
      ],
    },
  ];

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.container.list.column.name'),
      width: 180,
    },
    {
      dataIndex: 'image_name',
      title: t('app.docker.container.list.column.image'),
      width: 180,
    },
    {
      dataIndex: 'state',
      title: t('app.docker.container.list.column.state'),
      width: 110,
      render: ({ record }: { record: any }) => {
        const stateMap = {
          created: {
            color: 'orange',
            text: t('app.docker.container.list.state.created'),
          },
          running: {
            color: 'green',
            text: t('app.docker.container.list.state.running'),
          },
          paused: {
            color: 'orange',
            text: t('app.docker.container.list.state.paused'),
          },
          restarting: {
            color: 'orange',
            text: t('app.docker.container.list.state.restarting'),
          },
          removing: {
            color: 'orange',
            text: t('app.docker.container.list.state.removing'),
          },
          exited: {
            color: 'red',
            text: t('app.docker.container.list.state.exited'),
          },
          dead: {
            color: 'red',
            text: t('app.docker.container.list.state.dead'),
          },
        };

        const { color, text } = stateMap[
          record.state as keyof typeof stateMap
        ] || {
          color: '#ccc',
          text: record.state,
        };

        return h(
          resolveComponent('a-tag'),
          { color },
          {
            default: () => text,
          }
        );
      },
    },
    {
      dataIndex: 'usage',
      title: t('app.docker.container.list.column.usage'),
      width: 160,
      slotName: 'usage',
    },
    {
      dataIndex: 'network',
      title: t('app.docker.container.list.column.ip'),
      width: 140,
      render: ({ record }: { record: any }) => {
        return record.network.join(',');
      },
    },
    {
      dataIndex: 'ports',
      title: t('app.docker.container.list.column.ports'),
      width: 180,
      slotName: 'ports',
    },
    {
      dataIndex: 'run_time',
      title: t('app.docker.container.list.column.uptime'),
      width: 180,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 120,
      fixed: 'right' as const,
      slotName: 'operation',
    },
  ];

  const tableRef = ref();
  const reload = () => tableRef.value?.reload();
  const dataRef = ref<ApiListResult<any>>();
  const isLoading = ref(false);
  const sseMap = ref<Map<number, EventSource>>(new Map());
  const fetchUsage = async () => {
    if (!dataRef.value?.items || isLoading.value) {
      return;
    }
    isLoading.value = true;
    try {
      const requests = dataRef.value.items.map(async (item) => {
        try {
          const statusData = await getContainerUsageApi(item.container_id);
          if (statusData) {
            Object.assign(
              item,
              pick(statusData, [
                'cpu_total_usage',
                'system_usage',
                'cpu_percent',
                'per_cpu_usage',
                'memory_usage',
                'memory_limit',
                'memory_percent',
              ])
            );
          }
        } catch (error) {
          console.error('获取容器状态数据失败', item.container_id);
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
      if (!sseMap.value.has(item.container_id)) {
        const es = connectContainerUsageFollowApi(item.container_id);
        es.addEventListener('status', (event) => {
          try {
            const statusData = JSON.parse(event.data);
            Object.assign(
              item,
              pick(statusData, [
                'cpu_total_usage',
                'system_usage',
                'cpu_percent',
                'per_cpu_usage',
                'memory_usage',
                'memory_limit',
                'memory_percent',
              ])
            );
            if (tableRef.value) {
              tableRef.value.setData(dataRef.value);
            }
          } catch (e) {
            console.error(e);
          }
        });
        es.addEventListener('error', (event) => {
          console.error(event);
        });
        sseMap.value.set(item.container_id, es);
      }
    });

    for (const id of sseMap.value.keys()) {
      if (!dataRef.value?.items?.find((item) => item.container_id === id)) {
        sseMap.value.get(id)?.close();
        sseMap.value.delete(id);
      }
    }
  };

  const afterFetchHook = async (data: ApiListResult<any>) => {
    dataRef.value = data;
    await fetchUsage();
    startSSEForHosts();
    return data;
  };

  const { confirm } = useConfirm();
  const handlePrune = async () => {
    if (!(await confirm(t('app.docker.container.prune.confirm')))) {
      return;
    }
    try {
      await pruneApi({ type: 'container', with_tag_all: true });
      Message.success(t('app.docker.container.prune.success'));
      reload();
    } catch (e: any) {
      Message.error(e.message);
    }
  };

  const handleOperate = async (
    name: string,
    operation:
      | 'start'
      | 'stop'
      | 'restart'
      | 'kill'
      | 'pause'
      | 'unpause'
      | 'remove'
  ) => {
    try {
      const result = await operateContainersApi({ names: [name], operation });
      if (result.success) {
        Message.success(
          t('app.docker.container.list.operation.success', {
            command: result.command,
          })
        );
      } else {
        Message.error(
          t('app.docker.container.list.operation.failed', {
            command: result.command,
            message: result.message,
          })
        );
      }
      reload();
    } catch (e: any) {
      Message.error(
        e.message || t('app.docker.container.list.operation.error')
      );
    }
  };

  const afterStopConfirm = async (data: { name: string; force: boolean }) => {
    await handleOperate(data.name, data.force ? 'kill' : 'stop');
  };

  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const termRef = ref<InstanceType<typeof TerminalDrawer>>();
  const stopConfirmRef = ref<InstanceType<typeof StopConfirmModal>>();
  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.container.list.operation.terminal'),
      click: () => {
        termRef?.value?.show(record.container_id);
      },
    },
    {
      text: t('app.docker.container.list.operation.log'),
      click: () => {
        logsRef.value?.connect(record.name);
        logsRef.value?.show();
      },
    },
    {
      text: t('app.docker.container.list.operation.start'),
      visible: record.state !== 'running',
      confirm: record.compose
        ? t('app.docker.container.recommendCompose')
        : null,
      click: () => handleOperate(record.name, 'start'),
    },
    {
      text: t('app.docker.container.list.operation.stop'),
      visible: record.state === 'running',
      confirm: record.compose
        ? t('app.docker.container.recommendCompose')
        : null,
      click: () => stopConfirmRef.value?.show(record.container_id),
    },
    {
      text: t('app.docker.container.list.operation.restart'),
      visible: record.state === 'running',
      confirm: record.compose
        ? t('app.docker.container.recommendCompose')
        : null,
      click: () => handleOperate(record.name, 'restart'),
    },
    {
      text: t('app.docker.container.list.operation.pause'),
      visible: record.state === 'running',
      confirm: t('app.docker.container.list.operation.pause.confirm'),
      click: () => handleOperate(record.name, 'pause'),
    },
    {
      text: t('app.docker.container.list.operation.unpause'),
      visible: record.state === 'paused',
      confirm: t('app.docker.container.list.operation.unpause.confirm'),
      click: () => handleOperate(record.name, 'unpause'),
    },
    {
      text: t('app.docker.container.list.operation.delete'),
      // confirm: t('app.docker.container.list.operation.delete.confirm'),
      confirm: record.compose
        ? t('app.docker.container.recommendCompose')
        : t('app.docker.container.list.operation.delete.confirm'),
      click: () => handleOperate(record.name, 'remove'),
    },
  ];
</script>

<style scoped></style>
