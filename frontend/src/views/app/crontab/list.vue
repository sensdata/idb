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
          {{ $t('app.crontab.list.status.running') }}
        </span>
        <span
          v-else-if="record.linked === false"
          class="status-tag"
          style="color: rgb(var(--color-text-4))"
        >
          {{ $t('app.crontab.list.status.not_running') }}
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

  <a-modal
    v-model:visible="operateResultVisible"
    :title="operateResultTitle"
    :footer="false"
    :width="760"
    :mask-closable="true"
  >
    <a-typography-paragraph :copyable="true" class="operate-result-content">
      {{ operateResultText }}
    </a-typography-paragraph>
  </a-modal>
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
    operateCrontabApi,
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
  const operateResultVisible = ref(false);
  const operateResultTitle = ref('');
  const operateResultText = ref('');
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

  const showOperateResult = (
    operation: 'test' | 'execute',
    result: string
  ): void => {
    operateResultTitle.value = t(
      operation === 'test'
        ? 'app.crontab.system.operate.result.test_title'
        : 'app.crontab.system.operate.result.execute_title'
    );
    operateResultText.value =
      result?.trim() || t('app.crontab.system.operate.result.empty');
    operateResultVisible.value = true;
  };

  const handleSystemOperate = async (
    record: CrontabEntity,
    operation: 'test' | 'execute'
  ): Promise<void> => {
    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.crontab.list.message.no_host_selected'));
        return;
      }

      if (operation === 'execute') {
        const confirmed = await confirm({
          title: t('app.crontab.system.operate.execute_confirm.title'),
          content: t('app.crontab.system.operate.execute_confirm.content', {
            name: record.name,
          }),
        });
        if (!confirmed) {
          return;
        }
      }

      const response = await operateCrontabApi({
        type: params.value.type,
        category: record.category || '',
        name: record.name,
        operation,
        host: currentHostId.value,
      });

      Message.success(
        operation === 'test'
          ? t('app.crontab.system.operate.test_success')
          : t('app.crontab.system.operate.execute_success')
      );
      showOperateResult(operation, response.result);
    } catch (error: any) {
      Message.error(
        error?.message ||
          (operation === 'test'
            ? t('app.crontab.system.operate.test_failed')
            : t('app.crontab.system.operate.execute_failed'))
      );
    }
  };

  const getCrontabOperationOptions = (record: CrontabEntity) => {
    if (isSystemType.value) {
      return [
        {
          text: t('app.crontab.system.operate.test'),
          click: () => handleSystemOperate(record, 'test'),
        },
        {
          text: t('app.crontab.system.operate.execute'),
          click: () => handleSystemOperate(record, 'execute'),
        },
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

  .operate-result-content {
    max-height: 480px;
    overflow: auto;
    white-space: pre-wrap;
  }
</style>
