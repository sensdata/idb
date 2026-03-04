<template>
  <idb-table
    ref="gridRef"
    class="service-table"
    :loading="loading"
    :params="params"
    :columns="columns"
    :fetch="fetchAndDecorateServiceList"
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
    <template #rightActions>
      <div class="right-actions">
        <a-radio-group
          v-model="sourceFilter"
          type="button"
          size="small"
          @change="handleSegmentFilterChange"
        >
          <a-radio value="all">
            {{ $t('app.service.list.filter.source.all') }}
          </a-radio>
          <a-radio value="builtin">
            {{ $t('app.service.list.filter.source.builtin') }}
          </a-radio>
          <a-radio value="custom">
            {{ $t('app.service.list.filter.source.custom') }}
          </a-radio>
        </a-radio-group>

        <a-radio-group
          v-model="runtimeFilter"
          type="button"
          size="small"
          @change="handleSegmentFilterChange"
        >
          <a-radio value="all">
            {{ $t('app.service.list.filter.runtime.all') }}
          </a-radio>
          <a-radio value="running">
            {{ $t('app.service.list.filter.runtime.running') }}
          </a-radio>
          <a-radio value="stopped">
            {{ $t('app.service.list.filter.runtime.stopped') }}
          </a-radio>
          <a-radio value="failed">
            {{ $t('app.service.list.filter.runtime.failed') }}
          </a-radio>
        </a-radio-group>

        <a-button @click="handleRefreshRuntimeStatus">
          <template #icon>
            <icon-sync />
          </template>
          {{ $t('app.service.list.action.refresh_runtime') }}
        </a-button>
      </div>
    </template>

    <template #name="{ record }: { record: ServiceEntity }">
      <div class="name-cell">
        <span class="name-text">{{ record.name }}</span>
        <a-tag
          v-if="isIdbEnabled(record)"
          color="green"
          size="small"
          class="source-tag"
        >
          {{ $t('app.service.list.source.idb_enabled') }}
        </a-tag>
      </div>
    </template>

    <template #source="{ record }: { record: ServiceEntity }">
      <div class="source-cell">
        <a-tag size="small" :color="resolveSourceColor(record)">
          {{ resolveSourceText(record) }}
        </a-tag>
      </div>
    </template>

    <template #runtime_status="{ record }: { record: ServiceEntity }">
      <div class="runtime-status-cell">
        <a-tag
          :color="resolveRuntimeStatusColor(record)"
          size="small"
          class="status-tag runtime-tag"
        >
          {{ resolveRuntimeStatusText(record) }}
        </a-tag>
        <a-button
          size="mini"
          type="text"
          :loading="Boolean(runtimeStatusLoading[buildServiceKey(record)])"
          @click="handleServiceOperate(record, SERVICE_OPERATION.Status)"
        >
          {{ $t('app.service.list.runtime.refresh') }}
        </a-button>
      </div>
    </template>

    <template #start_time="{ record }: { record: ServiceEntity }">
      <div class="start-time-cell">
        <span class="start-time">{{ resolveStartTimeText(record) }}</span>
        <span v-if="resolveUptimeText(record)" class="uptime">
          {{ resolveUptimeText(record) }}
        </span>
      </div>
    </template>

    <template #operation="{ record }: { record: ServiceEntity }">
      <div class="operation-cell">
        <idb-table-operation
          type="button"
          :options="getServiceQuickOperationOptions(record)"
        />
        <idb-table-operation
          type="button"
          :options="getServiceOperationOptions(record)"
        />
      </div>
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
  import { serviceOperateApi, syncGlobalServiceApi } from '@/api/service';
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
  const sourceFilter = ref<'all' | 'builtin' | 'custom'>('all');
  const runtimeFilter = ref<'all' | 'running' | 'stopped' | 'failed'>('all');
  const currentRecords = ref<ServiceEntity[]>([]);
  const RUNTIME_STATUS_TTL_MS = 10 * 1000;
  const RUNTIME_STATUS_CONCURRENCY = 6;
  type RuntimeStatusEntry = {
    status: string;
    activeEnterTimestamp: string;
    startAt: number | null;
    fetchedAt: number;
  };
  const runtimeStatusMap = ref<Record<string, RuntimeStatusEntry>>({});
  const runtimeStatusLoading = ref<Record<string, boolean>>({});

  const columns = [
    {
      title: t('app.service.list.columns.name'),
      dataIndex: 'name',
      slotName: 'name',
      width: 220,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('app.service.list.columns.source'),
      dataIndex: 'source',
      slotName: 'source',
      width: 140,
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
      title: t('app.service.list.columns.runtime_status'),
      dataIndex: 'runtime_status',
      slotName: 'runtime_status',
      width: 180,
    },
    {
      title: t('app.service.list.columns.start_time'),
      dataIndex: 'start_time',
      slotName: 'start_time',
      width: 220,
    },
    {
      title: t('app.service.list.columns.operation'),
      slotName: 'operation',
      width: 380,
      align: 'left' as const,
    },
  ];

  const buildServiceKey = (record: ServiceEntity): string =>
    `${record.name}::${record.source || ''}`;

  const parseRuntimeStatus = (rawStatus: string) => {
    const activeMatch = rawStatus.match(/ActiveState=([^\n\r]+)/);
    const subMatch = rawStatus.match(/SubState=([^\n\r]+)/);
    const activeEnterMatch = rawStatus.match(/ActiveEnterTimestamp=([^\n\r]+)/);
    const activeState = activeMatch?.[1]?.trim()?.toLowerCase() || '';
    const subState = subMatch?.[1]?.trim()?.toLowerCase() || '';
    const activeEnterTimestamp = activeEnterMatch?.[1]?.trim() || '';
    const parsedStartAt = Date.parse(activeEnterTimestamp);
    const startAt = Number.isNaN(parsedStartAt) ? null : parsedStartAt;
    let status = 'unknown';

    if (activeState === 'active') {
      status = 'running';
    } else if (activeState === 'activating') {
      status = 'activating';
    } else if (activeState === 'deactivating') {
      status = 'deactivating';
    } else if (activeState === 'failed') {
      status = 'failed';
    } else if (activeState === 'inactive') {
      status = 'stopped';
    } else if (subState === 'dead') {
      status = 'stopped';
    }

    return {
      status,
      activeEnterTimestamp,
      startAt,
    };
  };

  const resolveRuntimeStatusText = (record: ServiceEntity): string => {
    const runtimeInfo = runtimeStatusMap.value[buildServiceKey(record)];
    return t(`app.service.list.runtime.${runtimeInfo?.status || 'unknown'}`);
  };

  const resolveRuntimeStatusColor = (record: ServiceEntity): string => {
    const runtimeStatus =
      runtimeStatusMap.value[buildServiceKey(record)]?.status || '';
    if (runtimeStatus === 'running') return 'green';
    if (runtimeStatus === 'failed') return 'red';
    if (runtimeStatus === 'activating' || runtimeStatus === 'deactivating') {
      return 'orange';
    }
    if (runtimeStatus === 'query_failed') return 'red';
    return 'gray';
  };

  const formatDuration = (durationMs: number): string => {
    const totalSeconds = Math.floor(durationMs / 1000);
    const days = Math.floor(totalSeconds / 86400);
    const hours = Math.floor((totalSeconds % 86400) / 3600);
    const minutes = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    const parts: string[] = [];

    if (days > 0) parts.push(`${days}${t('app.service.list.uptime.day')}`);
    if (hours > 0 || days > 0)
      parts.push(`${hours}${t('app.service.list.uptime.hour')}`);
    if (minutes > 0 || hours > 0 || days > 0)
      parts.push(`${minutes}${t('app.service.list.uptime.minute')}`);
    parts.push(`${seconds}${t('app.service.list.uptime.second')}`);

    return parts.join(' ');
  };

  const resolveStartTimeText = (record: ServiceEntity): string => {
    const runtimeInfo = runtimeStatusMap.value[buildServiceKey(record)];
    if (!runtimeInfo) return t('app.service.list.start.unknown');
    if (
      runtimeInfo.status !== 'running' &&
      runtimeInfo.status !== 'activating'
    ) {
      return t('app.service.list.start.not_running');
    }
    if (!runtimeInfo.startAt) return t('app.service.list.start.unknown');
    return formatTime(new Date(runtimeInfo.startAt).toISOString());
  };

  const resolveUptimeText = (record: ServiceEntity): string => {
    const runtimeInfo = runtimeStatusMap.value[buildServiceKey(record)];
    if (!runtimeInfo || !runtimeInfo.startAt) return '';
    if (
      runtimeInfo.status !== 'running' &&
      runtimeInfo.status !== 'activating'
    ) {
      return '';
    }
    const durationMs = Date.now() - runtimeInfo.startAt;
    if (durationMs <= 0) return '';
    return formatDuration(durationMs);
  };

  const isSystemBuiltinService = (record: ServiceEntity): boolean => {
    return (
      isSystemType.value &&
      (record.source?.startsWith('/usr/lib/systemd/system') ||
        record.source?.startsWith('/lib/systemd/system'))
    );
  };

  const isSystemCustomService = (record: ServiceEntity): boolean => {
    return (
      isSystemType.value && record.source?.startsWith('/etc/systemd/system')
    );
  };

  const isIdbEnabled = (record: ServiceEntity): boolean => {
    return !isSystemType.value && Boolean(record.linked);
  };

  const matchSourceFilter = (record: ServiceEntity): boolean => {
    if (sourceFilter.value === 'all') return true;
    if (sourceFilter.value === 'builtin') return isSystemBuiltinService(record);
    return !isSystemBuiltinService(record);
  };

  const matchRuntimeFilter = (record: ServiceEntity): boolean => {
    if (runtimeFilter.value === 'all') return true;
    const runtimeStatus =
      runtimeStatusMap.value[buildServiceKey(record)]?.status || '';
    return runtimeStatus === runtimeFilter.value;
  };

  const resolveSourceText = (record: ServiceEntity): string => {
    if (isSystemBuiltinService(record)) {
      return t('app.service.list.source.system_builtin');
    }
    if (isSystemCustomService(record)) {
      return t('app.service.list.source.system_custom');
    }
    return t('app.service.list.source.user_config');
  };

  const resolveSourceColor = (record: ServiceEntity): string => {
    if (isSystemBuiltinService(record)) return 'rgb(var(--arcoblue-6))';
    if (isSystemCustomService(record)) return 'rgb(var(--orangered-6))';
    return 'rgb(var(--gray-6))';
  };

  const queryServiceRuntimeStatus = async (
    record: ServiceEntity,
    options?: {
      withToast?: boolean;
      force?: boolean;
    }
  ) => {
    const withToast = options?.withToast ?? false;
    const force = options?.force ?? false;
    const serviceKey = buildServiceKey(record);
    const cachedStatus = runtimeStatusMap.value[serviceKey];
    if (
      !force &&
      cachedStatus &&
      Date.now() - cachedStatus.fetchedAt <= RUNTIME_STATUS_TTL_MS
    ) {
      return;
    }
    if (runtimeStatusLoading.value[serviceKey]) return;

    runtimeStatusLoading.value[serviceKey] = true;
    try {
      const response = await serviceOperateApi({
        type: type.value,
        category: params.value.category,
        name: record.name,
        operation: SERVICE_OPERATION.Status,
      });
      runtimeStatusMap.value[serviceKey] = {
        ...parseRuntimeStatus(response.result || ''),
        fetchedAt: Date.now(),
      };
      if (withToast && response.result) {
        Message.info(response.result);
      }
    } catch (error) {
      runtimeStatusMap.value[serviceKey] = {
        status: 'query_failed',
        activeEnterTimestamp: '',
        startAt: null,
        fetchedAt: Date.now(),
      };
      if (withToast) {
        Message.error(t('app.service.list.runtime.query_failed'));
      }
    } finally {
      runtimeStatusLoading.value[serviceKey] = false;
    }
  };

  const runWithConcurrency = async (
    items: ServiceEntity[],
    limit: number,
    worker: (item: ServiceEntity) => Promise<void>
  ) => {
    if (!items.length) return;
    let index = 0;
    const next = async () => {
      if (index >= items.length) return;
      const current = items[index];
      index += 1;
      await worker(current);
      await next();
    };
    const workers = Array.from(
      { length: Math.min(limit, items.length) },
      async () => next()
    );
    await Promise.all(workers);
  };

  const refreshRuntimeStatusForRecords = (
    records: ServiceEntity[],
    options?: {
      force?: boolean;
      awaitCompletion?: boolean;
    }
  ) => {
    const force = options?.force ?? false;
    const awaitCompletion = options?.awaitCompletion ?? false;
    const task = runWithConcurrency(
      records,
      RUNTIME_STATUS_CONCURRENCY,
      async (record) => {
        await queryServiceRuntimeStatus(record, { force });
      }
    );
    if (awaitCompletion) return task;
    task.catch((error) => {
      logError(error);
    });
    return Promise.resolve();
  };

  const handleRefreshRuntimeStatus = async () => {
    if (!currentRecords.value.length) return;
    await refreshRuntimeStatusForRecords(currentRecords.value, {
      force: true,
      awaitCompletion: true,
    });
  };

  const handleSegmentFilterChange = () => {
    gridRef.value?.load({ page: 1 });
  };

  const fetchAndDecorateServiceList = async (queryParams: any) => {
    const page = Number(queryParams.page) || 1;
    const pageSize = Number(queryParams.page_size) || 20;
    const needClientFilter =
      sourceFilter.value !== 'all' || runtimeFilter.value !== 'all';

    const response = await fetchServiceList({
      ...queryParams,
      page: needClientFilter ? 1 : page,
      page_size: needClientFilter ? 1000 : pageSize,
    });

    let items: ServiceEntity[] = (response.items || []).filter(
      matchSourceFilter
    );

    if (runtimeFilter.value !== 'all') {
      await refreshRuntimeStatusForRecords(items, {
        awaitCompletion: true,
      });
      items = items.filter(matchRuntimeFilter);
    }

    if (needClientFilter) {
      const start = (page - 1) * pageSize;
      const end = start + pageSize;
      const pagedItems = items.slice(start, end);
      currentRecords.value = pagedItems;
      if (runtimeFilter.value === 'all') {
        refreshRuntimeStatusForRecords(pagedItems).catch((error) => {
          logError(error);
        });
      }
      return {
        ...response,
        items: pagedItems,
        total: items.length,
        page,
        page_size: pageSize,
      };
    }

    currentRecords.value = items;
    if (runtimeFilter.value === 'all') {
      refreshRuntimeStatusForRecords(items).catch((error) => {
        logError(error);
      });
    }
    return {
      ...response,
      items,
      total: items.length,
      page,
      page_size: pageSize,
    };
  };

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
    if (operation === SERVICE_OPERATION.Status) {
      await queryServiceRuntimeStatus(record, {
        withToast: true,
        force: true,
      });
      return;
    }

    try {
      const operationText = t(
        `app.service.list.operation.${operation.toLowerCase()}`
      );

      await confirm(
        t('app.service.list.confirm.operation', {
          operation: operationText,
          name: record.name,
        })
      );

      const result = await operateService(record, operation);
      if (result !== null) {
        reload();
      }
    } catch (error) {
      logError('Failed to operate service:', error);
    }
  };

  const getServiceQuickOperationOptions = (record: ServiceEntity) => {
    const quickOptions: Array<{
      text: string;
      click: () => void;
    }> = [
      {
        text: record.linked
          ? t('app.service.list.operation.disable')
          : t('app.service.list.operation.enable'),
        click: () =>
          handleServiceOperate(
            record,
            record.linked ? SERVICE_OPERATION.Disable : SERVICE_OPERATION.Enable
          ),
      },
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
        text: t('app.service.list.operation.status'),
        click: () => handleServiceOperate(record, SERVICE_OPERATION.Status),
      },
    ];

    return isSystemType.value ? quickOptions.slice(1) : quickOptions;
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
        {
          text: t('app.service.list.operation.history'),
          click: () => handleViewHistory(record),
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

  .runtime-status-cell {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .start-time-cell {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    line-height: 1.2;
  }

  .start-time {
    font-weight: 500;
  }

  .uptime {
    margin-top: 2px;
    font-size: 12px;
    color: rgb(var(--color-text-3));
  }

  .operation-cell {
    display: flex;
    flex-direction: column;
    gap: 2px;
    align-items: flex-start;
  }

  .right-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .name-cell {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .name-text {
    display: inline-block;
    max-width: 170px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .source-cell {
    display: flex;
    align-items: center;
  }

  .source-tag {
    margin: 0;
  }

  .status-tag {
    margin: 0;
    font-weight: 600;
    letter-spacing: 0.2px;
  }

  .runtime-tag {
    justify-content: center;
    min-width: 64px;
  }
</style>
