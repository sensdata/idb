<template>
  <idb-table
    ref="gridRef"
    class="service-table"
    :loading="loading"
    :params="params"
    :columns="columns"
    :fetch="fetchServiceList"
    :auto-load="false"
  >
    <template #leftActions>
      <a-button v-if="!isSystemType" type="primary" @click="handleCreate">
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('app.service.list.action.create') }}
      </a-button>
      <a-button v-if="type === SERVICE_TYPE.Global" @click="handleSyncGlobal">
        <template #icon>
          <icon-sync />
        </template>
        {{ $t('app.service.list.action.sync') }}
      </a-button>
    </template>

    <template #status="{ record }: { record: ServiceEntity }">
      <div class="status-cell">
        <a-tag
          :color="
            record.linked ? 'rgb(var(--success-6))' : 'rgb(var(--color-text-4))'
          "
          class="status-tag"
        >
          {{
            record.linked
              ? $t('app.service.list.status.activated')
              : $t('app.service.list.status.deactivated')
          }}
        </a-tag>
      </div>
    </template>

    <template #operation="{ record }: { record: ServiceEntity }">
      <idb-table-operation
        type="button"
        :options="getServiceOperationOptions(record)"
      />
    </template>
  </idb-table>

  <form-drawer ref="formRef" :type="type" @ok="handleFormOk" />
  <logs-drawer ref="logsRef" />
  <history-drawer ref="historyRef" />
</template>

<script setup lang="ts">
  import {
    GlobalComponents,
    PropType,
    ref,
    computed,
    onMounted,
    watch,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import {
    SERVICE_TYPE,
    SERVICE_ACTION,
    SERVICE_OPERATION,
  } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { ServiceEntity } from '@/entity/Service';
  import { syncGlobalServiceApi } from '@/api/service';
  import { useConfirm } from '@/composables/confirm';
  import { useLogger } from '@/composables/use-logger';
  import { useServiceList } from './composables/use-service-list';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';

  interface FormDrawerInstance extends InstanceType<typeof FormDrawer> {
    show: (params?: {
      name?: string;
      type?: SERVICE_TYPE;
      category?: string;
      isEdit?: boolean;
      isView?: boolean;
      record?: ServiceEntity;
    }) => Promise<void>;
  }

  const props = defineProps({
    type: {
      type: String as PropType<SERVICE_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { confirm } = useConfirm();
  const { logError } = useLogger('ServiceList');

  const type = computed(() => props.type);
  const isSystemType = computed(() => type.value === SERVICE_TYPE.System);

  const {
    params,
    loading,
    fetchServiceList,
    deleteService,
    toggleServiceStatus,
    operateService,
  } = useServiceList(type.value);

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<FormDrawerInstance>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const historyRef = ref<InstanceType<typeof HistoryDrawer>>();

  const columns = [
    {
      title: t('app.service.list.columns.name'),
      dataIndex: 'name',
      width: 220,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('app.service.list.columns.description'),
      dataIndex: 'content',
      width: 320,
      ellipsis: true,
      tooltip: true,
      render: ({ record }: { record: ServiceEntity }) => {
        if (!record.content) return '';
        const lines = record.content.split('\n');
        const descriptionLine = lines.find((line) =>
          line.trim().startsWith('Description=')
        );
        return descriptionLine
          ? descriptionLine.replace('Description=', '').trim()
          : '';
      },
    },
    {
      title: t('app.service.list.columns.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
    },
    {
      title: t('app.service.list.columns.size'),
      dataIndex: 'size',
      width: 100,
      render: ({ record }: { record: ServiceEntity }) => {
        return `${(record.size / 1024).toFixed(2)} KB`;
      },
    },
    {
      title: t('app.service.list.columns.mod_time'),
      dataIndex: 'mod_time',
      width: 180,
      render: ({ record }: { record: ServiceEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      title: t('app.service.list.columns.operation'),
      slotName: 'operation',
      width: 260,
      align: 'left' as const,
    },
  ];

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    if (isSystemType.value) return;
    formRef.value?.show({
      type: type.value,
      category: params.value.category,
      isEdit: false,
    });
  };

  const handleEdit = (record: ServiceEntity) => {
    formRef.value?.show({
      type: type.value,
      category: params.value.category,
      isEdit: !isSystemType.value,
      isView: isSystemType.value,
      record,
    });
  };

  const handleAction = async (
    record: ServiceEntity,
    action: SERVICE_ACTION
  ) => {
    try {
      const actionText =
        action === SERVICE_ACTION.Activate
          ? t('app.service.list.operation.activate')
          : t('app.service.list.operation.deactivate');

      await confirm(
        t('app.service.list.confirm.action', {
          action: actionText,
          name: record.name,
        })
      );

      const success = await toggleServiceStatus(record, action);
      if (success) {
        reload();
      }
    } catch (error) {
      logError('Failed to action service:', error);
    }
  };

  const handleViewLogs = (record: ServiceEntity) => {
    logsRef.value?.show({
      type: type.value,
      category: params.value.category,
      name: record.name,
    });
  };

  const handleViewHistory = (record: ServiceEntity) => {
    historyRef.value?.show({
      type: type.value,
      category: params.value.category,
      name: record.name,
    });
  };

  const handleDelete = async (record: ServiceEntity) => {
    const success = await deleteService(record);
    if (success) {
      reload();
    }
  };

  const handleServiceOperate = async (
    record: ServiceEntity,
    operation: SERVICE_OPERATION
  ) => {
    try {
      const operationText = t(
        `app.service.list.operation.${operation.toLowerCase()}`
      );

      if (operation !== SERVICE_OPERATION.Status) {
        await confirm(
          t('app.service.list.confirm.operation', {
            operation: operationText,
            name: record.name,
          })
        );
      }

      const result = await operateService(record, operation);
      if (result !== null) {
        if (operation === SERVICE_OPERATION.Status) {
          Message.info(result);
        }
        reload();
      }
    } catch (error) {
      logError('Failed to operate service:', error);
    }
  };

  const getServiceOperationOptions = (record: ServiceEntity) => {
    if (isSystemType.value) {
      return [
        {
          text: t('common.view'),
          click: () => handleEdit(record),
        },
        {
          text: t('app.service.list.operation.logs'),
          click: () => handleViewLogs(record),
        },
      ];
    }

    const options: Array<{
      text: string;
      status?: 'normal' | 'success' | 'warning' | 'danger';
      click: () => void;
    }> = [
      {
        text: t('common.edit'),
        click: () => handleEdit(record),
      },
      {
        text: record.linked
          ? t('app.service.list.operation.deactivate')
          : t('app.service.list.operation.activate'),
        click: () =>
          handleAction(
            record,
            record.linked ? SERVICE_ACTION.Deactivate : SERVICE_ACTION.Activate
          ),
      },
    ];

    options.push(
      {
        text: t('common.delete'),
        status: 'danger',
        click: () => handleDelete(record),
      },
      {
        text: t('app.service.list.operation.logs'),
        click: () => handleViewLogs(record),
      },
      {
        text: t('app.service.list.operation.history'),
        click: () => handleViewHistory(record),
      }
    );

    if (record.linked) {
      options.push(
        {
          text: t('app.service.list.operation.start'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Start),
        },
        {
          text: t('app.service.list.operation.stop'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Stop),
        },
        {
          text: t('app.service.list.operation.restart'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Restart),
        },
        {
          text: t('app.service.list.operation.reload'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Reload),
        },
        {
          text: t('app.service.list.operation.enable'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Enable),
        },
        {
          text: t('app.service.list.operation.disable'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Disable),
        },
        {
          text: t('app.service.list.operation.status'),
          click: () => handleServiceOperate(record, SERVICE_OPERATION.Status),
        }
      );
    }

    return options;
  };

  const handleSyncGlobal = async () => {
    try {
      await confirm(t('app.service.list.confirm.sync'));
      await syncGlobalServiceApi();
      Message.success(t('app.service.list.success.sync'));
      reload();
    } catch (error) {
      logError('Failed to sync global:', error);
      Message.error(t('app.service.list.error.sync'));
    }
  };

  const handleFormOk = () => {
    reload();
  };

  const resetComponentsState = () => {
    reload();
  };

  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      params.value.category = newType === SERVICE_TYPE.System ? '' : 'default';
      params.value.page = 1;
      reload();
    }
  );

  onMounted(() => {
    resetComponentsState();
  });

  defineExpose({
    resetComponentsState,
  });
</script>

<style scoped>
  .service-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    align-items: center;
  }

  .status-tag {
    margin: 0;
  }
</style>
