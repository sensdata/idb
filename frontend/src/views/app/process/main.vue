<template>
  <idb-table
    ref="gridRef"
    class="process-table"
    :loading="loading"
    :columns="columns"
    :fetch="getProcessListApi"
    :has-search="true"
  >
    <template #operation="{ record }: { record: any }">
      <idb-table-operation
        type="button"
        :options="getProcessOperationOptions(record)"
      />
    </template>
  </idb-table>
  <detail-drawer ref="detailRef" />
</template>

<script setup lang="ts">
  import { GlobalComponents, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { formatTime } from '@/utils/format';
  import { getProcessListApi, killProcessApi } from '@/api/process';
  import useLoading from '@/composables/loading';
  import { useConfirm } from '@/composables/confirm';
  import DetailDrawer from './components/detail-drawer/index.vue';

  const { t } = useI18n();

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const detailRef = ref<InstanceType<typeof DetailDrawer>>();
  const { loading, setLoading } = useLoading(false);

  const columns = [
    {
      title: t('app.process.list.columns.name'),
      dataIndex: 'name',
      width: 240,
      ellipsis: true,
    },
    {
      title: t('app.process.list.columns.pid'),
      dataIndex: 'pid',
      width: 100,
    },
    {
      title: t('app.process.list.columns.ppid'),
      dataIndex: 'ppid',
      width: 100,
    },
    {
      title: t('app.process.list.columns.cpu'),
      dataIndex: 'cpu_percent',
      width: 100,
      render: ({ record }: { record: any }) =>
        `${record.cpu_percent.toFixed(2)}%`,
    },
    {
      title: t('app.process.list.columns.memory'),
      dataIndex: 'mem_rss',
      width: 120,
      render: ({ record }: { record: any }) =>
        `${(record.mem_rss / 1024 / 1024).toFixed(2)} MB`,
    },
    {
      title: t('app.process.list.columns.user'),
      dataIndex: 'user',
      width: 120,
    },
    {
      title: t('app.process.list.columns.threads'),
      dataIndex: 'threads',
      width: 100,
    },
    {
      title: t('app.process.list.columns.startTime'),
      dataIndex: 'create_time',
      width: 180,
      render: ({ record }: { record: any }) => formatTime(record.create_time),
    },
    {
      title: t('common.table.operation'),
      dataIndex: 'operation',
      width: 160,
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  const handleDetail = (record: any) => {
    detailRef.value?.setParams({
      pid: record.pid,
    });
    detailRef.value?.load();
    detailRef.value?.show();
  };

  const { confirm } = useConfirm();
  const handleKill = async (record: any) => {
    if (
      await confirm({
        content: t('app.process.list.message.killConfirm', { pid: record.pid }),
      })
    ) {
      try {
        setLoading(true);
        await killProcessApi({
          pid: record.pid,
        });
        Message.success(t('app.process.list.message.killSuccess'));
        gridRef.value?.reload();
      } catch (error: any) {
        Message.error(error?.message);
      } finally {
        setLoading(false);
      }
    }
  };

  // 获取操作按钮配置
  const getProcessOperationOptions = (record: any) => [
    {
      text: t('app.process.list.operation.detail'),
      click: () => handleDetail(record),
    },
    {
      text: t('app.process.list.operation.kill'),
      status: 'danger' as const,
      confirm: t('app.process.list.message.killConfirm', { pid: record.pid }),
      click: () => handleKill(record),
    },
  ];
</script>

<style scoped></style>
