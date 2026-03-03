<template>
  <div>
    <!-- Docker 环境检测 -->
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />

    <idb-table
      ref="tableRef"
      :columns="columns"
      :params="params"
      :has-search="enableNameSearch"
      :fetch="queryContainersApi"
      :beforeFetchHook="beforeFetchHook"
      :afterFetchHook="afterFetchHook"
    >
      <template #leftActions>
        <a-radio-group
          v-model="stateFilter"
          type="button"
          size="small"
          @change="handleStateFilterChange"
        >
          <a-radio
            v-for="option in stateFilterOptions"
            :key="option.value"
            :value="option.value"
          >
            {{ option.label }}
          </a-radio>
        </a-radio-group>
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
        <div class="container-operation-bar">
          <a-button
            v-for="option in getPrimaryOptions(record)"
            :key="option.key"
            type="text"
            size="small"
            @click="handleOptionClick(option)"
          >
            {{ option.text }}
          </a-button>
          <a-dropdown v-if="getMoreOptions(record).length > 0" trigger="click">
            <a-button type="text" size="small">
              {{ $t('common.table.operation') }}
              <icon-down />
            </a-button>
            <template #content>
              <a-doption
                v-for="option in getMoreOptions(record)"
                :key="option.key"
                @click="handleOptionClick(option)"
              >
                <span
                  :class="{
                    'danger-option': option.status === 'danger',
                  }"
                >
                  {{ option.text }}
                </span>
              </a-doption>
            </template>
          </a-dropdown>
        </div>
      </template>
    </idb-table>
    <logs-modal ref="logsRef" />
    <terminal-drawer ref="termRef" />
    <stop-confirm-modal ref="stopConfirmRef" @confirm="afterStopConfirm" />
    <database-manager-drawer ref="databaseManagerRef" />
  </div>
</template>

<script lang="ts" setup>
  import { computed, h, onUnmounted, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRoute } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { ApiListResult } from '@/types/global';
  import {
    queryContainersApi,
    operateContainersApi,
    connectContainerUsagesFollowApi,
    pruneApi,
  } from '@/api/docker';
  import { useConfirm } from '@/composables/confirm';
  import { pick } from 'lodash';
  import { formatMemorySize } from '@/utils/format';
  import { useDatabaseManager } from '@/composables/use-database-manager';
  import DockerInstallGuide from '@/components/docker-install-guide/index.vue';
  import DatabaseManagerDrawer from '@/components/database-manager/index.vue';
  import LogsModal from './components/logs-modal.vue';
  import TerminalDrawer from './components/terminal-drawer.vue';
  import StopConfirmModal from './components/stop-confirm-modal.vue';

  const { t } = useI18n();
  const { getDatabaseType, isDatabaseCompose } = useDatabaseManager();

  const route = useRoute();
  const stateFilter = ref('all');
  const stateFilterOptions = computed(() => [
    { label: t('app.docker.container.list.state.all'), value: 'all' },
    { label: t('app.docker.container.list.state.running'), value: 'running' },
    { label: t('app.docker.container.list.state.paused'), value: 'paused' },
    {
      label: t('app.docker.container.list.state.restarting'),
      value: 'restarting',
    },
    { label: t('app.docker.container.list.state.exited'), value: 'exited' },
    { label: t('app.docker.container.list.state.dead'), value: 'dead' },
    { label: t('app.docker.container.list.state.created'), value: 'created' },
    { label: t('app.docker.container.list.state.removing'), value: 'removing' },
  ]);
  const composeId = computed(() =>
    typeof route.params.composeId === 'string' && route.params.composeId.trim()
      ? route.params.composeId.trim()
      : undefined
  );
  const enableNameSearch = computed(() => !composeId.value);
  const params = computed(() => ({
    info: composeId.value,
    state: stateFilter.value,
  }));

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.container.list.column.name'),
      width: 180,
      sortable: {
        sortDirections: ['ascend', 'descend'],
      },
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
      sortable: {
        sortDirections: ['ascend', 'descend'],
      },
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
          color: 'var(--color-text-4)',
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
        const networks = Array.isArray(record.network)
          ? record.network.filter(Boolean)
          : [];
        if (!networks.length) {
          return '-';
        }
        const preview = networks.slice(0, 2).join(', ');
        const text =
          networks.length > 2 ? `${preview} +${networks.length - 2}` : preview;
        if (networks.length <= 2) {
          return text;
        }
        return h(
          resolveComponent('a-tooltip'),
          {
            content: networks.join('\n'),
          },
          {
            default: () =>
              h(
                'span',
                {
                  class: 'ip-preview',
                },
                text
              ),
          }
        );
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
  const handleStateFilterChange = () => {
    tableRef.value?.load?.({
      state: stateFilter.value,
      page: 1,
    });
  };
  const dataRef = ref<ApiListResult<any>>();
  const currentSSE = ref<EventSource | null>(null);
  const currentSSEParams = ref<string>('');
  const lastFetchParams = ref<any>(null);

  const stopSSE = () => {
    if (currentSSE.value) {
      currentSSE.value.close();
      currentSSE.value = null;
      currentSSEParams.value = '';
    }
  };

  onUnmounted(() => {
    stopSSE();
  });

  const startSSE = (queryParams: {
    info?: string;
    state: string;
    page: number;
    page_size: number;
    order_by?: string;
  }) => {
    // 生成参数标识，用于判断参数是否变化
    const paramsKey = JSON.stringify(queryParams);

    // 如果参数相同且已有连接，不重复创建
    if (currentSSEParams.value === paramsKey && currentSSE.value) {
      return;
    }

    // 关闭旧连接
    stopSSE();

    // 创建新的 SSE 连接
    const es = connectContainerUsagesFollowApi(queryParams);

    es.addEventListener('status', (event: MessageEvent) => {
      try {
        const usagesData = JSON.parse(event.data);
        // usagesData 是 PageResult 结构，包含 items 数组
        if (usagesData?.items && dataRef.value?.items) {
          // 遍历 usages 更新对应容器的资源使用数据
          usagesData.items.forEach((usage: any) => {
            const item = dataRef.value?.items?.find(
              (i: any) => i.container_id === usage.container_id
            );
            if (item) {
              Object.assign(
                item,
                pick(usage, [
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
          });
          if (tableRef.value) {
            tableRef.value.setData(dataRef.value);
          }
        }
      } catch (e) {
        console.error('解析容器资源使用数据失败', e);
      }
    });

    es.addEventListener('error', (event: Event) => {
      console.error('SSE 连接错误', event);
    });

    currentSSE.value = es;
    currentSSEParams.value = paramsKey;
  };

  // 在 fetch 之前捕获参数
  const beforeFetchHook = (fetchParams: any) => {
    const nextParams = { ...fetchParams };

    // 通用表格搜索参数统一映射为后端容器查询参数 info
    if (composeId.value) {
      nextParams.info = composeId.value;
    } else if (typeof nextParams.search === 'string') {
      const keyword = nextParams.search.trim();
      nextParams.info = keyword || undefined;
    }
    delete nextParams.search;

    // 通用表格排序参数映射为后端容器查询排序参数
    const orderByMap: Record<string, string> = {
      name: 'name',
      state: 'state',
      create_time: 'created',
    };
    const mappedOrderBy = orderByMap[nextParams.sort_field];
    if (mappedOrderBy && ['asc', 'desc'].includes(nextParams.sort_order)) {
      nextParams.order_by = mappedOrderBy;
      nextParams.order = nextParams.sort_order;
    } else {
      delete nextParams.order_by;
      delete nextParams.order;
    }
    delete nextParams.sort_field;
    delete nextParams.sort_order;

    lastFetchParams.value = { ...nextParams };
    return nextParams;
  };

  const afterFetchHook = async (data: ApiListResult<any>) => {
    dataRef.value = data;
    // 启动 SSE 跟踪，使用捕获的查询参数
    const queryParams = lastFetchParams.value;
    if (queryParams) {
      startSSE({
        info: queryParams.info,
        state: queryParams.state || 'all',
        page: queryParams.page || 1,
        page_size: queryParams.page_size || 20,
        order_by: queryParams.order_by,
      });
    }
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
      await showErrorWithDockerCheck(e.message, e);
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
        await showErrorWithDockerCheck(
          t('app.docker.container.list.operation.failed', {
            command: result.command,
            message: result.message,
          })
        );
      }
      reload();
    } catch (e: any) {
      await showErrorWithDockerCheck(
        e.message || t('app.docker.container.list.operation.error'),
        e
      );
    }
  };

  const afterStopConfirm = async (data: { name: string; force: boolean }) => {
    await handleOperate(data.name, data.force ? 'kill' : 'stop');
  };

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 如果 Docker 状态变化，可以重新加载容器列表
    if (status === 'installed') {
      reload();
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // Docker 安装完成后重新加载容器列表
    reload();
  };

  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const termRef = ref<InstanceType<typeof TerminalDrawer>>();
  const stopConfirmRef = ref<InstanceType<typeof StopConfirmModal>>();
  const databaseManagerRef = ref<InstanceType<typeof DatabaseManagerDrawer>>();

  interface ContainerOperationOption {
    key: string;
    text: string;
    visible?: boolean;
    confirm?: string | null;
    status?: 'normal' | 'success' | 'warning' | 'danger';
    click: () => void | Promise<void>;
  }

  const handleOptionClick = async (option: ContainerOperationOption) => {
    if (option.confirm && !(await confirm(option.confirm))) {
      return;
    }
    await option.click();
  };

  // 处理数据库管理
  const handleManageDatabase = (containerName: string) => {
    const dbType = getDatabaseType(containerName);
    if (!dbType) return;

    // 立即打开 drawer，数据在 drawer 内部加载
    databaseManagerRef.value?.show(dbType, containerName);
  };

  const getOperationOptions = (record: any): ContainerOperationOption[] => {
    const manageOptions: ContainerOperationOption[] = [
      // 数据库容器显示管理按钮
      {
        key: 'manage',
        text: t('app.docker.container.list.operation.manage'),
        visible: isDatabaseCompose(record.name),
        click: () => handleManageDatabase(record.name),
      },
      {
        key: 'terminal',
        text: t('app.docker.container.list.operation.terminal'),
        click: () => {
          termRef?.value?.show(record.container_id);
        },
      },
      {
        key: 'log',
        text: t('app.docker.container.list.operation.log'),
        click: () => {
          logsRef.value?.connect(record.name);
          logsRef.value?.show();
        },
      },
    ];

    const lifecycleOptions: ContainerOperationOption[] = [
      {
        key: 'start',
        text: t('app.docker.container.list.operation.start'),
        visible: ['created', 'exited', 'dead'].includes(record.state),
        confirm: record.compose
          ? t('app.docker.container.recommendCompose')
          : null,
        click: () => handleOperate(record.name, 'start'),
      },
      {
        key: 'stop',
        text: t('app.docker.container.list.operation.stop'),
        visible: ['running', 'restarting'].includes(record.state),
        confirm: record.compose
          ? t('app.docker.container.recommendCompose')
          : null,
        click: () => stopConfirmRef.value?.show(record.name),
      },
      {
        key: 'restart',
        text: t('app.docker.container.list.operation.restart'),
        visible: ['running', 'restarting'].includes(record.state),
        confirm: record.compose
          ? t('app.docker.container.recommendCompose')
          : null,
        click: () => handleOperate(record.name, 'restart'),
      },
      {
        key: 'pause',
        text: t('app.docker.container.list.operation.pause'),
        visible: record.state === 'running',
        confirm: t('app.docker.container.list.operation.pause.confirm'),
        click: () => handleOperate(record.name, 'pause'),
      },
      {
        key: 'unpause',
        text: t('app.docker.container.list.operation.unpause'),
        visible: record.state === 'paused',
        confirm: t('app.docker.container.list.operation.unpause.confirm'),
        click: () => handleOperate(record.name, 'unpause'),
      },
    ];

    const dangerOptions: ContainerOperationOption[] = [
      {
        key: 'delete',
        text: t('app.docker.container.list.operation.delete'),
        status: 'danger',
        confirm: record.compose
          ? t('app.docker.container.recommendCompose')
          : t('app.docker.container.list.operation.delete.confirm'),
        visible: record.state !== 'removing',
        click: () => handleOperate(record.name, 'remove'),
      },
    ];

    const primaryKeyByState: Record<string, string> = {
      created: 'start',
      exited: 'start',
      dead: 'start',
      running: 'stop',
      restarting: 'stop',
      paused: 'unpause',
    };

    const visibleLifecycle = lifecycleOptions.filter((item) => item.visible);
    const preferredPrimaryKey = primaryKeyByState[record.state];
    const primaryLifecycle =
      visibleLifecycle.find((item) => item.key === preferredPrimaryKey) ||
      visibleLifecycle[0];
    const secondaryLifecycle = visibleLifecycle.filter(
      (item) => item.key !== primaryLifecycle?.key
    );

    const visibleManage = manageOptions.filter(
      (item) => item.visible !== false
    );
    const visibleDanger = dangerOptions.filter(
      (item) => item.visible !== false
    );

    const primaryAction =
      primaryLifecycle || visibleManage[0] || visibleDanger[0];
    if (!primaryAction) {
      return [];
    }

    return [
      primaryAction,
      ...secondaryLifecycle,
      ...visibleManage.filter((item) => item.key !== primaryAction.key),
      ...visibleDanger.filter((item) => item.key !== primaryAction.key),
    ];
  };

  const getPrimaryOptions = (record: any) => {
    const options = getOperationOptions(record).filter(
      (item) => item.visible !== false
    );
    return options.filter((item) => ['terminal', 'log'].includes(item.key));
  };

  const getMoreOptions = (record: any) => {
    const primaryKeys = new Set(
      getPrimaryOptions(record).map((item) => item.key)
    );
    return getOperationOptions(record).filter(
      (item) => item.visible !== false && !primaryKeys.has(item.key)
    );
  };
</script>

<style scoped>
  .container-operation-bar {
    display: flex;
    flex-wrap: nowrap;
    gap: 0.25rem;
    align-items: center;
    white-space: nowrap;
  }

  .danger-option {
    color: rgb(var(--danger-6));
  }

  .ip-preview {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: help;
  }
</style>
