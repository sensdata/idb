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
        <idb-table-operation
          type="button"
          :options="getOperationOptions(record)"
        />
      </template>
    </idb-table>
    <logs-modal ref="logsRef" />
    <edit-drawer ref="editRef" />
    <create-drawer ref="createRef" @success="reload" />
    <down-confirm-modal ref="downConfirmRef" @confirm="afterDownConfirm" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, h, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
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
  import DownConfirmModal from './components/down-confirm-modal.vue';

  const { t } = useI18n();
  const router = useRouter();

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
              router.push(`/app/docker/container/${record.container_number}`);
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
              router.push(`/app/docker/container/${record.container_number}`);
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
      width: 180,
      slotName: 'operation',
    },
  ];

  const handleOperate = async (
    name: string,
    operation: 'start' | 'stop' | 'restart' | 'up' | 'down',
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

  const afterDownConfirm = async (params: {
    name: string;
    remove_volumes: boolean;
  }) => {
    await handleOperate(params.name, 'down', {
      remove_volumes: params.remove_volumes,
    });
  };

  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const editRef = ref<InstanceType<typeof EditDrawer>>();
  const createRef = ref<InstanceType<typeof CreateDrawer>>();
  const downConfirmRef = ref<InstanceType<typeof DownConfirmModal>>();

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.compose.list.operation.edit'),
      click: () => {
        editRef.value?.setParams({ name: record.name });
        editRef.value?.load();
        editRef.value?.show();
      },
    },
    {
      text: t('app.docker.compose.list.operation.log'),
      click: () => {
        logsRef.value?.connect(record.config_files);
        logsRef.value?.show();
      },
    },
    {
      text: t('app.docker.compose.list.operation.openDirectory'),
      click: () => {
        if (record.path) {
          router.push(createFileRoute(record.path));
        }
      },
    },
    {
      text: t('app.docker.compose.list.operation.start'),
      visible: [
        COMPOSE_STATUS.Exited,
        COMPOSE_STATUS.Partial,
        COMPOSE_STATUS.Dead,
        COMPOSE_STATUS.Mixed,
      ].includes(record.status),
      click: async () => {
        await handleOperate(record.name, 'start');
      },
    },
    {
      text: t('app.docker.compose.list.operation.stop'),
      visible: [
        COMPOSE_STATUS.Running,
        COMPOSE_STATUS.Partial,
        COMPOSE_STATUS.Dead,
        COMPOSE_STATUS.Mixed,
      ].includes(record.status),
      confirm: t('app.docker.compose.list.operation.stop.confirm'),
      click: async () => {
        await handleOperate(record.name, 'stop');
      },
    },
    {
      text: t('app.docker.compose.list.operation.restart'),
      visible: [
        COMPOSE_STATUS.Running,
        COMPOSE_STATUS.Partial,
        COMPOSE_STATUS.Dead,
        COMPOSE_STATUS.Mixed,
      ].includes(record.status),
      click: async () => {
        await handleOperate(record.name, 'restart');
      },
    },
    {
      text: t('app.docker.compose.list.operation.up'),
      visible: [COMPOSE_STATUS.Exited, COMPOSE_STATUS.Dead].includes(
        record.status
      ),
      click: async () => {
        await handleOperate(record.name, 'up');
      },
    },
    {
      text: t('app.docker.compose.list.operation.down'),
      visible: [
        COMPOSE_STATUS.Running,
        COMPOSE_STATUS.Partial,
        COMPOSE_STATUS.Dead,
        COMPOSE_STATUS.Mixed,
        COMPOSE_STATUS.Paused,
        COMPOSE_STATUS.Restarting,
      ].includes(record.status),
      click: async () => {
        downConfirmRef.value?.show(record.name);
      },
    },
    {
      text: t('app.docker.compose.list.operation.delete'),
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
