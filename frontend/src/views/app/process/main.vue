<template>
  <idb-table
    ref="gridRef"
    class="process-table"
    :loading="loading"
    :columns="columns"
    :fetch="getProcessListApi"
  >
    <template #operation="{ record }: { record: any }">
      <div class="operation">
        <a-button type="text" size="small" @click="handleDetail(record)">
          {{ $t('app.process.list.operation.detail') }}
        </a-button>
        <a-button
          type="text"
          size="small"
          status="danger"
          @click="handleKill(record)"
        >
          {{ $t('app.process.list.operation.kill') }}
        </a-button>
      </div>
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
  import useLoading from '@/hooks/loading';
  import { useConfirm } from '@/hooks/confirm';
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
      title: t('app.process.list.columns.status'),
      dataIndex: 'status',
      width: 100,
    },
    {
      title: t('app.process.list.columns.cpu'),
      dataIndex: 'cpu',
      width: 100,
      render: ({ record }: { record: any }) => `${record.cpu}%`,
    },
    {
      title: t('app.process.list.columns.memory'),
      dataIndex: 'memory',
      width: 120,
      render: ({ record }: { record: any }) =>
        `${(record.memory / 1024 / 1024).toFixed(2)} MB`,
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
      dataIndex: 'startTime',
      width: 180,
      render: ({ record }: { record: any }) => formatTime(record.startTime),
    },
    {
      title: t('common.table.operation'),
      dataIndex: 'operation',
      width: 160,
      align: 'center' as const,
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
</script>

<style scoped>
  .operation :deep(.arco-btn-size-small) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
