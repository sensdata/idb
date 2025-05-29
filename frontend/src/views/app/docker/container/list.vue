<template>
  <idb-table
    ref="gridRef"
    :columns="columns"
    :fetch="queryContainersApi"
    :filters="filters"
  >
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
      <idb-dropdown-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <logs-modal ref="logsRef" />
  <terminal-drawer ref="termRef" />
</template>

<script lang="ts" setup>
  import { h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { queryContainersApi, operateContainersApi } from '@/api/docker';
  import LogsModal from './components/logs-modal.vue';
  import TerminalDrawer from './components/terminal-drawer.vue';

  const { t } = useI18n();
  const gridRef = ref();
  const reload = () => gridRef.value?.reload();

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
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    record: any,
    operation:
      | 'start'
      | 'stop'
      | 'restart'
      | 'kill'
      | 'pause'
      | 'resume'
      | 'remove'
  ) => {
    try {
      await operateContainersApi({ names: [record.name], operation });
      Message.success(t('app.docker.container.list.operation.success'));
      reload();
    } catch (e: any) {
      Message.error(
        e.message || t('app.docker.container.list.operation.failed')
      );
    }
  };

  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const termRef = ref<InstanceType<typeof TerminalDrawer>>();
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
        logsRef.value?.connect(record.container_id);
        logsRef.value?.show();
      },
    },
    {
      text: t('app.docker.container.list.operation.start'),
      visible: record.state !== 'running',
      click: () => handleOperate(record, 'start'),
    },
    {
      text: t('app.docker.container.list.operation.stop'),
      visible: record.state === 'running',
      click: () => handleOperate(record, 'stop'),
    },
    {
      text: t('app.docker.container.list.operation.restart'),
      visible: record.state === 'running',
      click: () => handleOperate(record, 'restart'),
    },
    {
      text: t('app.docker.container.list.operation.kill'),
      visible: record.state === 'running',
      click: () => handleOperate(record, 'kill'),
    },
    {
      text: t('app.docker.container.list.operation.pause'),
      visible: record.state === 'running',
      click: () => handleOperate(record, 'pause'),
    },
    {
      text: t('app.docker.container.list.operation.resume'),
      visible: record.state === 'paused',
      click: () => handleOperate(record, 'resume'),
    },
    {
      text: t('app.docker.container.list.operation.delete'),
      confirm: t('app.docker.container.list.operation.delete.confirm'),
      click: () => handleOperate(record, 'remove'),
    },
  ];
</script>

<style scoped></style>
