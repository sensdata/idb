<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="fetchContainers">
    <template #operation="{ record }">
      <idb-dropdown-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { queryContainersApi, operateContainersApi } from '@/api/docker';
  import { formatTime } from '@/utils/format';

  const { t } = useI18n();
  const gridRef = ref();
  const reload = () => gridRef.value?.reload();

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.container.list.column.name'),
      width: 180,
    },
    {
      dataIndex: 'image',
      title: t('app.docker.container.list.column.image'),
      width: 180,
    },
    {
      dataIndex: 'status',
      title: t('app.docker.container.list.column.status'),
      width: 110,
      render: ({ record }: { record: any }) => {
        const status = record.status;
        let color = '';
        let text = '';
        if (status === 'running') {
          color = 'color-success';
          text = t('app.docker.container.list.status.running');
        } else if (status === 'exited') {
          color = 'color-danger';
          text = t('app.docker.container.list.status.stopped');
        } else if (status === 'paused') {
          color = 'color-warning';
          text = t('app.docker.container.list.status.paused');
        } else {
          color = 'color-warning';
          text = t('app.docker.container.list.status.unknown');
        }
        return `<span class="${color}">${text}</span>`;
      },
    },
    {
      dataIndex: 'resource',
      title: t('app.docker.container.list.column.resource'),
      width: 120,
      render: ({ record }: { record: any }) => {
        // 假设 record.resource 是字符串，如 "0.2 CPU, 128MB RAM"
        return record.resource || '-';
      },
    },
    {
      dataIndex: 'ip',
      title: t('app.docker.container.list.column.ip'),
      width: 140,
      render: ({ record }: { record: any }) => record.ip || '-',
    },
    {
      dataIndex: 'ports',
      title: t('app.docker.container.list.column.ports'),
      width: 140,
      render: ({ record }: { record: any }) =>
        Array.isArray(record.ports)
          ? record.ports
              .map(
                (p: any) =>
                  `${p.host_ip || ''}:${p.host_port}->${p.container_port}/${
                    p.protocol
                  }`
              )
              .join('<br>')
          : '-',
    },
    {
      dataIndex: 'uptime',
      title: t('app.docker.container.list.column.uptime'),
      width: 140,
      render: ({ record }: { record: any }) => formatTime(record.uptime),
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 180,
      slotName: 'operation',
    },
  ];

  const fetchContainers = async (params: any) => {
    // 这里假设 params 里有 host，实际可根据业务调整
    const res = await queryContainersApi({ ...params, state: '' });
    return res;
  };

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

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.container.list.operation.terminal'),
      click: () =>
        Message.info(t('app.docker.container.list.operation.terminal.todo')),
    },
    {
      text: t('app.docker.container.list.operation.log'),
      click: () =>
        Message.info(t('app.docker.container.list.operation.log.todo')),
    },
    {
      text: t('app.docker.container.list.operation.start'),
      visible: record.status !== 'running',
      click: () => handleOperate(record, 'start'),
    },
    {
      text: t('app.docker.container.list.operation.stop'),
      visible: record.status === 'running',
      click: () => handleOperate(record, 'stop'),
    },
    {
      text: t('app.docker.container.list.operation.restart'),
      visible: record.status === 'running',
      click: () => handleOperate(record, 'restart'),
    },
    {
      text: t('app.docker.container.list.operation.kill'),
      visible: record.status === 'running',
      click: () => handleOperate(record, 'kill'),
    },
    {
      text: t('app.docker.container.list.operation.pause'),
      visible: record.status === 'running',
      click: () => handleOperate(record, 'pause'),
    },
    {
      text: t('app.docker.container.list.operation.resume'),
      visible: record.status === 'paused',
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
