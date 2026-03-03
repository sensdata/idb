<template>
  <div>
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />
    <idb-table
      ref="gridRef"
      :loading="loading"
      :columns="columns"
      :fetch="queryComposeApi"
    >
      <template #leftActions>
        <a-button type="primary" @click="createRef?.show()">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.docker.compose.list.action.add') }}
        </a-button>
      </template>
      <template #operation="{ record }">
        <div class="compose-operation-bar">
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
    <edit-drawer ref="editRef" @ok="reload" />
    <create-drawer ref="createRef" @success="reload" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, h, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { useConfirm } from '@/composables/confirm';
  import { COMPOSE_STATUS } from '@/config/enum';
  import {
    queryComposeApi,
    operateComposeApi,
    deleteComposeApi,
  } from '@/api/docker';
  import { createFileRoute } from '@/utils/file-route';
  import LogsModal from './components/logs-modal.vue';
  import EditDrawer from './components/edit-drawer.vue';
  import CreateDrawer from './components/create-drawer.vue';

  const { t } = useI18n();
  const router = useRouter();
  const { confirm } = useConfirm();

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const loading = ref(false);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.compose.list.column.name'),
      width: 200,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              router.push({
                name: 'container',
                params: {
                  composeId: record.name,
                },
              });
            },
            hoverable: false,
          },
          {
            default: () => record.name,
          }
        );
      },
    },
    {
      dataIndex: 'status',
      title: t('app.docker.compose.list.column.container_status'),
      width: 120,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              router.push({
                name: 'container',
                params: {
                  composeId: record.name,
                },
              });
            },
            hoverable: false,
          },
          {
            default: () =>
              [
                record.containers.filter(
                  (item: any) => item.state === 'running'
                ).length,
                record.container_number,
              ].join('/'),
          }
        );
      },
    },
    {
      dataIndex: 'created_at',
      title: t('app.docker.compose.list.column.created_at'),
      width: 160,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'left' as const,
      width: 160,
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    name: string,
    operation: 'start' | 'stop' | 'restart' | 'up' | 'pull' | 'down',
    params?: {
      remove_volumes?: boolean;
    }
  ) => {
    try {
      const result = await operateComposeApi({
        name,
        operation,
        ...params,
      });
      if (result.success) {
        Message.success(
          t('app.docker.compose.list.operation.success', {
            command: result.command,
          })
        );
      } else {
        await showErrorWithDockerCheck(
          t('app.docker.compose.list.operation.failed', {
            command: result.command,
            message: result.message,
          })
        );
      }
      reload();
    } catch (e: any) {
      await showErrorWithDockerCheck(
        e.message || t('app.docker.compose.list.operation.error'),
        e
      );
    }
  };

  const handleDownWithVolumes = async (name: string) => {
    if (
      !(await confirm(
        t('app.docker.compose.list.operation.downWithVolumes.confirm2')
      ))
    ) {
      return;
    }
    await handleOperate(name, 'down', {
      remove_volumes: true,
    });
  };

  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const editRef = ref<InstanceType<typeof EditDrawer>>();
  const createRef = ref<InstanceType<typeof CreateDrawer>>();

  interface ComposeOperationOption {
    key: string;
    text: string;
    visible?: boolean;
    confirm?: string | null;
    status?: 'normal' | 'success' | 'warning' | 'danger';
    click: () => void | Promise<void>;
  }

  const handleOptionClick = async (option: ComposeOperationOption) => {
    if (option.confirm && !(await confirm(option.confirm))) {
      return;
    }
    await option.click();
  };

  const getOperationOptions = (record: any): ComposeOperationOption[] => {
    const manageOptions: ComposeOperationOption[] = [
      {
        key: 'edit',
        text: t('app.docker.compose.list.operation.edit'),
        click: () => {
          editRef.value?.setParams({ name: record.name });
          editRef.value?.load();
          editRef.value?.show();
        },
      },
      {
        key: 'log',
        text: t('app.docker.compose.list.operation.log'),
        click: () => {
          logsRef.value?.connect(record.config_files);
          logsRef.value?.show();
        },
      },
      {
        key: 'openDirectory',
        text: t('app.docker.compose.list.operation.openDirectory'),
        click: () => {
          if (record.path) {
            router.push(createFileRoute(record.path));
          }
        },
      },
    ];

    const lifecycleOptions: ComposeOperationOption[] = [
      {
        key: 'up',
        text: t('app.docker.compose.list.operation.up'),
        visible: record.status !== COMPOSE_STATUS.Removing,
        click: async () => {
          await handleOperate(record.name, 'up');
        },
      },
      {
        key: 'stop',
        text: t('app.docker.compose.list.operation.stop'),
        visible: [
          COMPOSE_STATUS.Running,
          COMPOSE_STATUS.Partial,
          COMPOSE_STATUS.Mixed,
          COMPOSE_STATUS.Restarting,
        ].includes(record.status),
        confirm: t('app.docker.compose.list.operation.stop.confirm'),
        click: async () => {
          await handleOperate(record.name, 'stop');
        },
      },
      {
        key: 'start',
        text: t('app.docker.compose.list.operation.start'),
        visible: [
          COMPOSE_STATUS.Exited,
          COMPOSE_STATUS.Dead,
          COMPOSE_STATUS.Paused,
        ].includes(record.status),
        click: async () => {
          await handleOperate(record.name, 'start');
        },
      },
      {
        key: 'restart',
        text: t('app.docker.compose.list.operation.restart'),
        visible: [
          COMPOSE_STATUS.Running,
          COMPOSE_STATUS.Partial,
          COMPOSE_STATUS.Mixed,
          COMPOSE_STATUS.Restarting,
        ].includes(record.status),
        click: async () => {
          await handleOperate(record.name, 'restart');
        },
      },
      {
        key: 'pull',
        text: t('app.docker.compose.list.operation.pull'),
        visible: record.status !== COMPOSE_STATUS.Removing,
        click: async () => {
          await handleOperate(record.name, 'pull');
        },
      },
      {
        key: 'down',
        text: t('app.docker.compose.list.operation.down'),
        visible: [
          COMPOSE_STATUS.Running,
          COMPOSE_STATUS.Exited,
          COMPOSE_STATUS.Partial,
          COMPOSE_STATUS.Dead,
          COMPOSE_STATUS.Mixed,
          COMPOSE_STATUS.Paused,
          COMPOSE_STATUS.Restarting,
        ].includes(record.status),
        confirm: t('app.docker.compose.list.operation.down.confirm'),
        click: async () => {
          await handleOperate(record.name, 'down', {
            remove_volumes: false,
          });
        },
      },
      {
        key: 'downWithVolumes',
        text: t('app.docker.compose.list.operation.downWithVolumes'),
        status: 'danger',
        visible: [
          COMPOSE_STATUS.Running,
          COMPOSE_STATUS.Exited,
          COMPOSE_STATUS.Partial,
          COMPOSE_STATUS.Dead,
          COMPOSE_STATUS.Mixed,
          COMPOSE_STATUS.Paused,
          COMPOSE_STATUS.Restarting,
        ].includes(record.status),
        confirm: t(
          'app.docker.compose.list.operation.downWithVolumes.confirm1'
        ),
        click: async () => {
          await handleDownWithVolumes(record.name);
        },
      },
    ];

    const dangerOptions: ComposeOperationOption[] = [
      {
        key: 'delete',
        text: t('app.docker.compose.list.operation.delete'),
        status: 'danger',
        confirm: t('app.docker.compose.list.operation.delete.confirm'),
        visible: record.status !== COMPOSE_STATUS.Removing,
        click: async () => {
          loading.value = true;
          try {
            await deleteComposeApi({ name: record.name });
            Message.success(t('common.message.operationSuccess'));
            reload();
          } catch (err: any) {
            await showErrorWithDockerCheck(
              err.message || t('common.message.operationError'),
              err
            );
          } finally {
            loading.value = false;
          }
        },
      },
    ];

    const primaryKeyByStatus: Record<string, string> = {
      [COMPOSE_STATUS.Running]: 'stop',
      [COMPOSE_STATUS.Exited]: 'up',
      [COMPOSE_STATUS.Dead]: 'up',
      [COMPOSE_STATUS.Partial]: 'restart',
      [COMPOSE_STATUS.Mixed]: 'restart',
      [COMPOSE_STATUS.Paused]: 'down',
      [COMPOSE_STATUS.Restarting]: 'down',
    };

    const visibleLifecycle = lifecycleOptions.filter((item) => item.visible);
    const preferredPrimaryKey = primaryKeyByStatus[record.status];
    const primaryLifecycle =
      visibleLifecycle.find((item) => item.key === preferredPrimaryKey) ||
      visibleLifecycle[0];
    const secondaryLifecycle = visibleLifecycle.filter(
      (item) => item.key !== primaryLifecycle?.key
    );

    const visibleManage = manageOptions;
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
      (item) => item.visible !== false && item.status !== 'danger'
    );
    const upOption = options.find((item) => item.key === 'up');
    const ordered = upOption
      ? [upOption, ...options.filter((item) => item.key !== 'up')]
      : options;
    return ordered.slice(0, 3);
  };

  const getMoreOptions = (record: any) => {
    const primaryKeys = new Set(
      getPrimaryOptions(record).map((item) => item.key)
    );
    return getOperationOptions(record).filter(
      (item) => item.visible !== false && !primaryKeys.has(item.key)
    );
  };

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 如果 Docker 状态变化，可以重新加载 Compose 列表
    if (status === 'installed') {
      reload();
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // Docker 安装完成后重新加载 Compose 列表
    reload();
  };
</script>

<style scoped>
  .compose-operation-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
    align-items: center;
  }

  .danger-option {
    color: rgb(var(--danger-6));
  }
</style>
