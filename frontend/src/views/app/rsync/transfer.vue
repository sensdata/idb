<template>
  <div class="container">
    <Breadcrumb :items="['menu.app.rsync']" />
    <a-card class="general-card" :title="$t('menu.app.rsync')">
      <template #extra>
        <a-space>
          <a-button type="primary" @click="handleCreateTask">
            <template #icon><icon-plus /></template>
            {{ $t('app.rsync.action.createTask') }}
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
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column
            :title="$t('app.rsync.columns.id')"
            data-index="id"
            :width="200"
          />
          <a-table-column
            :title="$t('app.rsync.columns.src')"
            data-index="src"
          />
          <a-table-column
            :title="$t('app.rsync.columns.dst')"
            data-index="dst"
          />
          <a-table-column
            :title="$t('app.rsync.columns.mode')"
            data-index="mode"
            :width="120"
          >
            <template #cell="{ record }">
              <a-tag :color="modeTagColor(record.mode)">
                {{
                  record.mode === 'copy'
                    ? $t('app.rsync.mode.copy')
                    : $t('app.rsync.mode.incremental')
                }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.rsync.columns.status')"
            data-index="status"
            :width="120"
          >
            <template #cell="{ record }">
              <a-tag :color="statusTagColor(record.status)">
                {{ $t(`app.rsync.status.${record.status || 'unknown'}`) }}
              </a-tag>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.rsync.columns.progress')"
            data-index="progress"
            :width="150"
          >
            <template #cell="{ record }">
              <a-progress :percent="record.progress" />
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('common.table.operation')"
            align="center"
            :width="200"
          >
            <template #cell="{ record }">
              <a-space>
                <a-button
                  type="text"
                  size="small"
                  @click="handleViewDetail(record)"
                >
                  {{ $t('common.detail') }}
                </a-button>
                <a-button
                  v-if="record.status === 'running'"
                  type="text"
                  size="small"
                  status="warning"
                  @click="handleCancelTask(record)"
                >
                  {{ $t('common.cancel') }}
                </a-button>
                <a-button
                  v-if="record.status === 'failed'"
                  type="text"
                  size="small"
                  :loading="retryingTaskId === record.id"
                  @click="handleRetryTask(record)"
                >
                  {{ $t('common.button.retry') }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  status="danger"
                  @click="handleDeleteTask(record)"
                >
                  {{ $t('common.delete') }}
                </a-button>
              </a-space>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </a-card>

    <!-- 创建任务抽屉 -->
    <a-drawer
      v-model:visible="createDrawerVisible"
      :title="$t('app.rsync.create.title')"
      width="600px"
      @ok="handleSubmitTask"
    >
      <a-form :model="taskForm" layout="vertical">
        <a-form-item :label="$t('app.rsync.form.srcHost')">
          <a-select
            v-model="taskForm.src_host_id"
            :placeholder="$t('app.rsync.form.placeholder.srcHost')"
          >
            <a-option
              v-for="host in hostList"
              :key="host.id"
              :value="host.id"
              :label="host.name"
            />
          </a-select>
        </a-form-item>
        <a-form-item :label="$t('app.rsync.form.srcPath')">
          <file-selector
            v-model="taskForm.src"
            :placeholder="$t('app.rsync.form.placeholder.srcPath')"
            :host="taskForm.src_host_id"
            type="all"
          />
        </a-form-item>
        <a-form-item :label="$t('app.rsync.form.dstHost')">
          <a-select
            v-model="taskForm.dst_host_id"
            :placeholder="$t('app.rsync.form.placeholder.dstHost')"
          >
            <a-option
              v-for="host in hostList"
              :key="host.id"
              :value="host.id"
              :label="host.name"
            />
          </a-select>
        </a-form-item>
        <a-form-item :label="$t('app.rsync.form.dstPath')">
          <file-selector
            v-model="taskForm.dst"
            :placeholder="$t('app.rsync.form.placeholder.dstPath')"
            :host="taskForm.dst_host_id"
            type="dir"
          />
        </a-form-item>
        <a-form-item :label="$t('app.rsync.form.mode')">
          <a-radio-group v-model="taskForm.mode">
            <a-radio value="copy">{{ $t('app.rsync.mode.copy') }}</a-radio>
            <a-radio value="incremental">{{
              $t('app.rsync.mode.incremental')
            }}</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-drawer>

    <!-- 任务详情抽屉 -->
    <a-drawer
      v-model:visible="detailDrawerVisible"
      :title="$t('app.rsync.detail.title', { id: currentTask?.id })"
      width="600px"
    >
      <a-descriptions v-if="currentTask" :column="1" bordered>
        <a-descriptions-item :label="$t('app.rsync.field.taskId')">
          {{ currentTask.id }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.src')">
          {{ currentTask.src }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.dst')">
          {{ currentTask.dst }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.cacheDir')">
          {{ currentTask.cache_dir }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.mode')">
          {{
            currentTask.mode === 'copy'
              ? $t('app.rsync.mode.copy')
              : $t('app.rsync.mode.incremental')
          }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.status')">
          <a-tag :color="statusTagColor(currentTask.status)">
            {{ $t(`app.rsync.status.${currentTask.status || 'unknown'}`) }}
          </a-tag>
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.progress')">
          <a-progress :percent="currentTask.progress" />
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.step')">
          {{ currentTask.step }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.startTime')">
          {{ currentTask.start_time }}
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.endTime')">
          {{ currentTask.end_time }}
        </a-descriptions-item>
        <a-descriptions-item
          v-if="currentTask.error"
          :label="$t('app.rsync.field.error')"
        >
          <pre style="color: var(--color-danger-6); white-space: pre-wrap">{{
            currentTask.error
          }}</pre>
        </a-descriptions-item>
        <a-descriptions-item :label="$t('app.rsync.field.lastLog')">
          <pre style="white-space: pre-wrap">{{ currentTask.last_log }}</pre>
        </a-descriptions-item>
      </a-descriptions>
    </a-drawer>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted, onUnmounted, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import {
    getRsyncTaskListApi,
    getRsyncTaskDetailApi,
    createRsyncTaskApi,
    deleteRsyncTaskApi,
    cancelRsyncTaskApi,
    retryRsyncTaskApi,
  } from '@/api/database';
  import { getHostListApi } from '@/api/host';
  import { RsyncTaskInfo } from '@/entity/Database';
  import FileSelector from '@/components/file/file-selector/index.vue';

  const { t } = useI18n();
  const loading = ref(false);
  const tableData = ref<RsyncTaskInfo[]>([]);
  const pagination = ref({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const createDrawerVisible = ref(false);
  const detailDrawerVisible = ref(false);
  const currentTask = ref<RsyncTaskInfo | null>(null);
  const hostList = ref<any[]>([]);
  const retryingTaskId = ref<string | null>(null);
  const DETAIL_REFRESH_INTERVAL = 3000; // ms
  let detailRefreshTimer: number | undefined;

  const taskForm = ref({
    src_host_id: undefined as number | undefined,
    src: '',
    dst_host_id: undefined as number | undefined,
    dst: '',
    mode: 'copy' as 'copy' | 'incremental',
  });

  const statusTagColor = (status?: string): string => {
    if (status === 'running') return 'var(--color-primary-6)';
    if (status === 'success') return 'var(--color-success-6)';
    if (status === 'failed') return 'var(--color-danger-6)';
    return 'var(--color-fill-3)';
  };

  const modeTagColor = (mode: 'copy' | 'incremental'): string => {
    return mode === 'copy'
      ? 'var(--color-primary-6)'
      : 'var(--color-success-6)';
  };

  const clearDetailAutoRefresh = () => {
    if (detailRefreshTimer !== undefined) {
      window.clearInterval(detailRefreshTimer);
      detailRefreshTimer = undefined;
    }
  };

  const startDetailAutoRefresh = (taskId: string) => {
    clearDetailAutoRefresh();
    detailRefreshTimer = window.setInterval(async () => {
      if (!detailDrawerVisible.value) {
        clearDetailAutoRefresh();
        return;
      }
      try {
        const res = await getRsyncTaskDetailApi({ id: taskId });
        currentTask.value = res;
        if (res.status !== 'running') {
          clearDetailAutoRefresh();
        }
      } catch (error) {
        // stop auto refresh on error to avoid spamming backend
        clearDetailAutoRefresh();
      }
    }, DETAIL_REFRESH_INTERVAL);
  };

  const fetchData = async () => {
    loading.value = true;
    try {
      const res = await getRsyncTaskListApi({
        page: pagination.value.current,
        page_size: pagination.value.pageSize,
      });
      tableData.value = res.tasks || [];

      pagination.value.total = res.total || 0;
    } catch (error) {
      Message.error(t('app.rsync.message.fetchListFailed'));
    } finally {
      loading.value = false;
    }
  };

  const fetchHosts = async () => {
    try {
      const res = await getHostListApi({
        page: 1,
        page_size: 1000,
      });
      hostList.value = res.items || [];
    } catch (error) {
      Message.error(t('app.rsync.message.fetchHostFailed'));
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

  const handleCreateTask = async () => {
    taskForm.value = {
      src_host_id: undefined,
      src: '',
      dst_host_id: undefined,
      dst: '',
      mode: 'copy',
    };
    createDrawerVisible.value = true;
    // 打开抽屉时才获取主机列表
    await fetchHosts();
  };

  const handleSubmitTask = async () => {
    if (
      !taskForm.value.src_host_id ||
      !taskForm.value.dst_host_id ||
      !taskForm.value.src ||
      !taskForm.value.dst
    ) {
      Message.warning(t('app.rsync.message.formIncomplete'));
      return;
    }

    try {
      await createRsyncTaskApi(taskForm.value as any);
      Message.success(t('app.rsync.message.createSuccess'));
      createDrawerVisible.value = false;
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.message.createFailed'));
    }
  };

  const handleViewDetail = async (record: RsyncTaskInfo) => {
    try {
      const res = await getRsyncTaskDetailApi({ id: record.id });
      currentTask.value = res;
      detailDrawerVisible.value = true;
      if (res.status === 'running') {
        startDetailAutoRefresh(res.id);
      } else {
        clearDetailAutoRefresh();
      }
    } catch (error) {
      Message.error(t('app.rsync.message.fetchDetailFailed'));
    }
  };

  const handleCancelTask = async (record: RsyncTaskInfo) => {
    try {
      await cancelRsyncTaskApi({ id: record.id });
      Message.success(t('app.rsync.message.cancelSuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.message.cancelFailed'));
    }
  };

  const handleRetryTask = async (record: RsyncTaskInfo) => {
    retryingTaskId.value = record.id;
    try {
      await retryRsyncTaskApi({ id: record.id });
      Message.success(t('app.rsync.message.retrySuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.message.retryFailed'));
    } finally {
      retryingTaskId.value = null;
    }
  };

  const handleDeleteTask = async (record: RsyncTaskInfo) => {
    try {
      await deleteRsyncTaskApi({ id: record.id });
      Message.success(t('app.rsync.message.deleteSuccess'));
      fetchData();
    } catch (error) {
      Message.error(t('app.rsync.message.deleteFailed'));
    }
  };

  watch(detailDrawerVisible, (visible) => {
    if (!visible) {
      clearDetailAutoRefresh();
    } else if (currentTask.value && currentTask.value.status === 'running') {
      startDetailAutoRefresh(currentTask.value.id);
    }
  });

  onMounted(() => {
    fetchData();
  });

  onUnmounted(() => {
    clearDetailAutoRefresh();
  });
</script>

<style scoped lang="less">
  .container {
    padding: 20px;
  }
</style>
