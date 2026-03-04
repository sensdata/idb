<template>
  <idb-table
    ref="gridRef"
    class="crontab-table"
    :loading="loading"
    :params="params"
    :columns="columns"
    :fetch="fetchCrontabList"
    :auto-load="false"
  >
    <template #leftActions>
      <a-button v-if="!isSystemType" type="primary" @click="handleCreate">
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('app.crontab.list.action.create') }}
      </a-button>
    </template>

    <template #status="{ record }: { record: CrontabEntity }">
      <div class="status-cell">
        <span
          v-if="record.linked === true"
          class="status-tag"
          style="color: rgb(var(--success-6))"
        >
          生效中
        </span>
        <span
          v-else-if="record.linked === false"
          class="status-tag"
          style="color: rgb(var(--color-text-4))"
        >
          未激活
        </span>
        <span v-else class="status-tag">
          {{ record.linked }}
        </span>
      </div>
    </template>

    <template #operation="{ record }: { record: CrontabEntity }">
      <idb-table-operation
        type="button"
        :options="getCrontabOperationOptions(record)"
      />
    </template>
  </idb-table>

  <form-drawer ref="formRef" :type="type" @ok="handleFormOk" />
  <logs-drawer ref="logsRef" />
</template>

<script setup lang="ts">
  import {
    GlobalComponents,
    PropType,
    ref,
    onMounted,
    computed,
    watch,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { formatTimeWithoutSeconds } from '@/utils/format';
  import { CrontabEntity } from '@/entity/Crontab';
  import {
    deleteCrontabApi,
    getCrontabListApi,
    actionCrontabApi,
  } from '@/api/crontab';
  import useLoading from '@/composables/loading';
  import { useConfirm } from '@/composables/confirm';
  import usetCurrentHost from '@/composables/current-host';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import { useCronDescription } from './components/form-drawer/composables/use-cron-description';

  const DEFAULT_CRONTAB_CATEGORY = 'default';

  interface FormDrawerInstance extends InstanceType<typeof FormDrawer> {
    show: (params?: {
      name?: string;
      type?: CRONTAB_TYPE;
      category?: string;
      isEdit?: boolean;
      isView?: boolean;
      record?: CrontabEntity;
    }) => Promise<void>;
  }

  const props = defineProps({
    type: {
      type: String as PropType<CRONTAB_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { getCronDescriptionFromContent } = useCronDescription();
  const isSystemType = computed(() => props.type === CRONTAB_TYPE.System);
  const { currentHostId } = usetCurrentHost();

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<FormDrawerInstance>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();

  const params = ref<{
    type: CRONTAB_TYPE;
    category: string;
    page: number;
    page_size: number;
  }>({
    type: props.type,
    category: isSystemType.value ? '' : DEFAULT_CRONTAB_CATEGORY,
    page: 1,
    page_size: 20,
  });

  function extractPeriodFromRecord(record: CrontabEntity): string {
    if (!record.content) {
      return record.period_expression || '';
    }

    const description = getCronDescriptionFromContent(record.content);
    if (description) return description;

    return record.period_expression || '';
  }

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.crontab.list.column.name'),
      width: 150,
    },
    {
      dataIndex: 'status',
      title: t('app.crontab.list.column.status'),
      width: 120,
      align: 'left' as const,
      slotName: 'status',
    },
    {
      dataIndex: 'period',
      title: t('app.crontab.list.column.period'),
      width: 150,
      align: 'left' as const,
      render: ({ record }: { record: CrontabEntity }) =>
        extractPeriodFromRecord(record),
    },
    {
      dataIndex: 'mod_time',
      title: t('app.crontab.list.column.mod_time'),
      width: 160,
      align: 'left' as const,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTimeWithoutSeconds(record.mod_time);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 210,
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  const fetchCrontabList = async (fetchParams: Record<string, unknown>) => {
    if (!currentHostId.value) {
      console.warn('hostId is undefined, skipping API request');
      return Promise.resolve({
        items: [],
        total: 0,
        page: 1,
        page_size: params.value.page_size,
      });
    }

    return getCrontabListApi({
      ...fetchParams,
      type: params.value.type,
      category: isSystemType.value
        ? ''
        : params.value.category || DEFAULT_CRONTAB_CATEGORY,
    });
  };

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    if (isSystemType.value) return;
    formRef.value?.show({
      type: params.value.type,
      category: DEFAULT_CRONTAB_CATEGORY,
    });
  };

  const handleFormOk = async () => {
    reload();
  };

  const handleEdit = async (record: CrontabEntity) => {
    if (!formRef.value) return;
    try {
      if (isSystemType.value) {
        formRef.value.show({
          name: record.name,
          type: params.value.type,
          category: '',
          isEdit: true,
          isView: true,
        });
        return;
      }
      formRef.value.show({
        record,
        isEdit: true,
        category:
          record.category || params.value.category || DEFAULT_CRONTAB_CATEGORY,
      });
    } catch (error) {
      Message.error(t('app.crontab.list.message.edit_error'));
    }
  };

  const handleAction = async (
    record: CrontabEntity,
    action: 'activate' | 'deactivate'
  ) => {
    try {
      await actionCrontabApi({
        type: params.value.type,
        category:
          record.category || params.value.category || DEFAULT_CRONTAB_CATEGORY,
        name: record.name,
        action,
      });
      const messageKey =
        action === 'activate'
          ? 'app.crontab.list.message.activate_success'
          : 'app.crontab.list.message.deactivate_success';
      Message.success(t(messageKey));
      // eslint-disable-next-line no-promise-executor-return
      await new Promise((resolve) => setTimeout(resolve, 1000));
      if (gridRef.value) {
        await gridRef.value.load();
      }
    } catch (err) {
      if (err instanceof Error) {
        Message.error(err.message);
      } else {
        Message.error(String(err));
      }
    }
  };

  const handleDelete = async (record: CrontabEntity) => {
    if (
      await confirm({
        content: t('app.crontab.list.delete.confirm', { name: record.name }),
      })
    ) {
      setLoading(true);
      try {
        await deleteCrontabApi({
          name: record.name,
          type: params.value.type,
          category:
            record.category ||
            params.value.category ||
            DEFAULT_CRONTAB_CATEGORY,
        });
        Message.success(t('app.crontab.list.message.delete_success'));
        reload();
      } catch (err) {
        if (err instanceof Error) {
          Message.error(err.message);
        } else {
          Message.error(String(err));
        }
      } finally {
        setLoading(false);
      }
    }
  };

  const getCrontabOperationOptions = (record: CrontabEntity) => {
    if (isSystemType.value) {
      return [
        {
          text: t('common.view'),
          click: () => handleEdit(record),
        },
      ];
    }

    return [
      {
        text: t('common.edit'),
        click: () => handleEdit(record),
      },
      {
        text: record.linked
          ? t('app.crontab.list.operation.deactivate')
          : t('app.crontab.list.operation.activate'),
        click: () =>
          handleAction(record, record.linked ? 'deactivate' : 'activate'),
      },
      {
        text: t('common.delete'),
        status: 'danger' as const,
        confirm: t('app.crontab.list.delete.confirm', { name: record.name }),
        click: () => handleDelete(record),
      },
    ];
  };

  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      params.value.page = 1;
      params.value.page_size = 20;
      params.value.category =
        newType === CRONTAB_TYPE.System ? '' : DEFAULT_CRONTAB_CATEGORY;
      reload();
    }
  );

  onMounted(() => {
    params.value.category = isSystemType.value ? '' : DEFAULT_CRONTAB_CATEGORY;
    reload();
  });

  defineExpose({
    resetComponentsState: () => {
      params.value.page = 1;
      params.value.page_size = 20;
      params.value.category = isSystemType.value
        ? ''
        : DEFAULT_CRONTAB_CATEGORY;
      reload();
    },
  });
</script>

<style scoped>
  .crontab-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }

  .status-tag {
    padding: 4px 8px;
    font-size: 12px;
    border-radius: 4px;
  }
</style>
