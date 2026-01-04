<template>
  <div class="container">
    <Breadcrumb :items="['menu.app.rsync', 'app.rsync.client.title']" />
    <a-card class="general-card" :title="$t('app.rsync.client.title')">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="handleCreate">
            <template #icon><icon-plus /></template>
            {{ $t('app.rsync.client.action.create') }}
          </a-button>
          <a-button @click="handleRefresh">
            <template #icon><icon-refresh /></template>
            {{ $t('common.refresh') }}
          </a-button>
        </a-space>
      </template>

      <a-table
        :loading="loading"
        :data="tableData"
        :pagination="pagination"
        row-key="id"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column
            :title="$t('app.rsync.client.columns.name')"
            data-index="name"
            :width="180"
          />
          <a-table-column
            :title="$t('app.rsync.client.columns.direction')"
            data-index="direction"
            :width="120"
          >
            <template #cell="{ record }">
              <a-tag
                :color="
                  record.direction === 'local_to_remote' ? 'blue' : 'green'
                "
              >
                {{ getDirectionLabel(record.direction) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.rsync.client.columns.localPath')"
            data-index="local_path"
          />
          <a-table-column
            :title="$t('app.rsync.client.columns.remoteType')"
            data-index="remote_type"
            :width="120"
          >
            <template #cell="{ record }">
              {{ getRemoteTypeLabel(record.remote_type) }}
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.rsync.client.columns.remoteHost')"
            :width="180"
          >
            <template #cell="{ record }">
              {{ record.remote_host }}:{{ record.remote_port }}
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.rsync.client.columns.state')"
            data-index="state"
            :width="100"
          >
            <template #cell="{ record }">
              <a-tag :color="getStateColor(record.state)">
                {{ getStateLabel(record.state) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('common.table.operation')"
            align="center"
            :width="200"
          >
            <template #cell="{ record }">
              <a-space>
                <a-button type="text" size="small" @click="handleEdit(record)">
                  {{ $t('common.edit') }}
                </a-button>
                <a-button
                  v-if="record.state !== 'running'"
                  type="text"
                  size="small"
                  status="success"
                  @click="handleRetry(record)"
                >
                  {{ $t('app.rsync.client.action.run') }}
                </a-button>
                <a-popconfirm
                  v-if="record.state === 'running'"
                  :content="$t('app.rsync.client.message.confirmCancel')"
                  @ok="handleCancel(record)"
                >
                  <a-button type="text" size="small" status="warning">
                    {{ $t('app.rsync.client.action.cancel') }}
                  </a-button>
                </a-popconfirm>
                <a-button type="text" size="small" @click="handleTest(record)">
                  {{ $t('common.test') }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  @click="handleViewLogs(record)"
                >
                  {{ $t('app.rsync.client.logs.button') }}
                </a-button>
                <a-popconfirm
                  :content="$t('app.rsync.client.message.confirmDelete')"
                  @ok="handleDelete(record)"
                >
                  <a-button type="text" size="small" status="danger">
                    {{ $t('common.delete') }}
                  </a-button>
                </a-popconfirm>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <!-- 创建/编辑抽屉 -->
    <rsync-client-drawer
      v-model:visible="drawerVisible"
      :edit-data="currentRecord"
      @ok="handleDrawerOk"
    />

    <!-- 日志查看抽屉 -->
    <rsync-client-logs-drawer
      v-model:visible="logsDrawerVisible"
      :task-id="currentLogTaskId"
      :task-state="currentLogTaskState"
    />
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, onUnmounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import Breadcrumb from '@/components/breadcrumb/index.vue';
  import {
    getRsyncClientTaskListApi,
    deleteRsyncClientTaskApi,
    testRsyncClientTaskApi,
    retryRsyncClientTaskApi,
    cancelRsyncClientTaskApi,
  } from '@/api/database';
  import { RsyncClientTask } from '@/entity/Database';
  import RsyncClientDrawer from './components/rsync-client-drawer.vue';
  import RsyncClientLogsDrawer from './components/rsync-client-logs-drawer.vue';

  const { t } = useI18n();
  const loading = ref(false);
  const tableData = ref<RsyncClientTask[]>([]);
  const drawerVisible = ref(false);
  const logsDrawerVisible = ref(false);
  const currentLogTaskId = ref('');
  const currentLogTaskState = ref('');
  const currentRecord = ref<RsyncClientTask | null>(null);
  const pagination = ref({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // 轮询定时器：当存在运行中的任务时，定期刷新列表
  let pollingTimer: number | null = null;

  const getDirectionLabel = (direction: string) => {
    return direction === 'local_to_remote'
      ? t('app.rsync.client.direction.localToRemote')
      : t('app.rsync.client.direction.remoteToLocal');
  };

  const getRemoteTypeLabel = (type: string) => {
    return type === 'rsync'
      ? t('app.rsync.client.remoteType.rsync')
      : t('app.rsync.client.remoteType.ssh');
  };

  const getStateColor = (state: string) => {
    const colors: Record<string, string> = {
      pending: 'gray',
      running: 'blue',
      success: 'green',
      succeeded: 'green',
      failed: 'red',
      canceled: 'orange',
    };
    return colors[state] || 'gray';
  };

  const getStateLabel = (state: string) => {
    const stateKey = `app.rsync.client.state.${state}`;
    return t(stateKey);
  };

  const fetchData = async (withLoading = true) => {
    if (withLoading) {
      loading.value = true;
    }
    try {
      const res = await getRsyncClientTaskListApi({
        page: pagination.value.current,
        page_size: pagination.value.pageSize,
      });
      tableData.value = res.tasks || [];
      pagination.value.total = res.total || 0;

      // 检查是否存在运行中的任务，决定是否开启/关闭轮询
      const hasRunning = tableData.value.some(
        (item) => item.state === 'running'
      );
      if (hasRunning) {
        if (pollingTimer === null) {
          pollingTimer = window.setInterval(() => {
            // 轮询时不展示全局 loading，避免表格闪烁
            fetchData(false);
          }, 3000);
        }
      } else if (pollingTimer !== null) {
        clearInterval(pollingTimer);
        pollingTimer = null;
      }
    } catch (error) {
      Message.error(t('app.rsync.client.message.fetchListFailed'));
    } finally {
      if (withLoading) {
        loading.value = false;
      }
    }
  };

  const handleRefresh = () => {
    fetchData();
  };

  const handlePageChange = (page: number) => {
    pagination.value.current = page;
    fetchData();
  };

  const handlePageSizeChange = (pageSize: number) => {
    pagination.value.pageSize = pageSize;
    fetchData();
  };

  const handleCreate = () => {
    currentRecord.value = null;
    drawerVisible.value = true;
  };

  const handleEdit = (record: RsyncClientTask) => {
    currentRecord.value = record;
    drawerVisible.value = true;
  };

  const handleRetry = async (record: RsyncClientTask) => {
    try {
      await retryRsyncClientTaskApi({ id: record.id });
      Message.success(t('app.rsync.client.message.runSuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.client.message.runFailed'));
    }
  };

  const handleTest = async (record: RsyncClientTask) => {
    try {
      await testRsyncClientTaskApi({ id: record.id });
      Message.success(t('app.rsync.client.message.testSuccess'));
    } catch (error) {
      Message.error(t('app.rsync.client.message.testFailed'));
    }
  };

  const handleCancel = async (record: RsyncClientTask) => {
    try {
      await cancelRsyncClientTaskApi({ id: record.id });
      Message.success(t('app.rsync.client.message.cancelSuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.client.message.cancelFailed'));
    }
  };

  const handleDelete = async (record: RsyncClientTask) => {
    try {
      await deleteRsyncClientTaskApi({ id: record.id });
      Message.success(t('app.rsync.client.message.deleteSuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.client.message.deleteFailed'));
    }
  };

  const handleDrawerOk = () => {
    fetchData();
  };

  const handleViewLogs = (record: RsyncClientTask) => {
    currentLogTaskId.value = record.id;
    currentLogTaskState.value = record.state;
    logsDrawerVisible.value = true;
  };

  onMounted(() => {
    fetchData();
  });

  onUnmounted(() => {
    if (pollingTimer !== null) {
      clearInterval(pollingTimer);
      pollingTimer = null;
    }
  });
</script>

<style scoped lang="less">
  .container {
    padding: 1.43rem;
  }
</style>
