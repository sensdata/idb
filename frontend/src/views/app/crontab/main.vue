<template>
  <idb-table
    ref="gridRef"
    class="crontab-table"
    :loading="loading"
    :params="params"
    :columns="columns"
    :fetch="getCrontabListApi"
  >
    <template #leftActions>
      <a-button type="primary" @click="handleCreate">
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('app.crontab.list.action.create') }}
      </a-button>
    </template>
    <template #operation="{ record }: { record: CrontabEntity }">
      <div class="operation">
        <a-button type="text" size="small" @click="handleEdit(record)">
          {{ $t('common.edit') }}
        </a-button>
        <a-button type="text" size="small" @click="handleRun(record)">
          {{ $t('app.crontab.list.operation.run') }}
        </a-button>
        <a-button type="text" size="small" @click="handleLog(record)">
          {{ $t('app.crontab.list.operation.log') }}
        </a-button>
        <a-button
          type="text"
          size="small"
          status="danger"
          @click="handleDelete(record)"
        >
          {{ $t('common.delete') }}
        </a-button>
      </div>
    </template>
  </idb-table>
  <form-drawer ref="formRef" @ok="reload" />
  <logs-drawer ref="logsRef" />
</template>

<script setup lang="ts">
  import { GlobalComponents, PropType, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { CrontabEntity } from '@/entity/Crontab';
  import {
    deleteCrontabApi,
    getCrontabListApi,
    runCrontabApi,
  } from '@/api/crontab';
  import useLoading from '@/hooks/loading';
  import { useConfirm } from '@/hooks/confirm';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';

  const props = defineProps({
    type: {
      type: String as PropType<SCRIPT_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<InstanceType<typeof FormDrawer>>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const selectedCat = ref(null);
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const params = ref({
    type: props.type,
    category: selectedCat.value,
  });

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.crontab.list.column.name'),
      width: 150,
      slotName: 'name',
    },
    {
      dataIndex: 'status',
      title: t('app.crontab.list.column.status'),
      width: 100,
      slotName: 'status',
    },
    {
      dataIndex: 'period',
      title: t('app.crontab.list.column.period'),
      width: 120,
    },
    {
      dataIndex: 'last_run_time',
      title: t('app.crontab.list.column.last_run_time'),
      width: 160,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTime(record.last_run_time);
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.crontab.list.column.mod_time'),
      width: 160,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      dataIndex: 'create_time',
      title: t('app.crontab.list.column.create_time'),
      width: 125,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTime(record.create_time);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 210,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    formRef.value?.show();
  };
  const handleEdit = (record: CrontabEntity) => {
    formRef.value?.setParams({
      id: record.id,
    });
    formRef.value?.load();
    formRef.value?.show();
  };
  const handleRun = async (record: CrontabEntity) => {
    setLoading(true);
    try {
      await runCrontabApi(record);
      Message.success(t('app.crontab.list.message.run_success'));
    } finally {
      setLoading(false);
    }
  };
  const handleLog = (record: CrontabEntity) => {
    logsRef.value?.show(record.id);
  };
  const handleDelete = async (record: CrontabEntity) => {
    if (
      await confirm({
        content: t('app.crontab.list.delete.confirm', { name: record.name }),
      })
    ) {
      setLoading(true);
      try {
        await deleteCrontabApi(record);
        Message.success(t('app.crontab.list.message.delete_success'));
      } finally {
        setLoading(false);
        reload();
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
