<template>
  <a-drawer
    v-model:visible="visible"
    :title="$t('app.pma.manager.title', { name: composeName })"
    width="800px"
    :footer="false"
    unmount-on-close
  >
    <div v-if="loading && !composeInfo" class="loading-container">
      <a-spin :loading="true" />
    </div>

    <div v-else>
      <a-tabs v-if="composeInfo" default-active-key="info">
        <a-tab-pane key="info" :title="$t('app.pma.tab.info')">
          <a-descriptions :column="2" bordered>
            <a-descriptions-item :label="$t('app.pma.info.name')">
              {{ composeInfo.name }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.pma.info.version')">
              {{ composeInfo.version }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.pma.info.port')">
              {{ composeInfo.port }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.pma.info.status')">
              <a-space>
                <a-tag
                  :color="
                    composeInfo.status === 'running'
                      ? 'green'
                      : composeInfo.status === 'exited'
                      ? 'red'
                      : 'gray'
                  "
                >
                  {{ composeInfo.status }}
                </a-tag>
                <a-button
                  size="mini"
                  :loading="refreshLoading"
                  @click="handleRefreshStatus"
                >
                  {{ $t('app.pma.button.refreshStatus') }}
                </a-button>
              </a-space>
            </a-descriptions-item>
          </a-descriptions>

          <a-divider />

          <a-space>
            <a-button
              v-if="composeInfo.status !== 'running'"
              type="primary"
              :loading="startLoading"
              @click="handleOperation('start')"
            >
              {{ $t('app.pma.button.start') }}
            </a-button>
            <a-button
              v-if="composeInfo.status === 'running'"
              status="warning"
              :loading="stopLoading"
              @click="handleOperation('stop')"
            >
              {{ $t('app.pma.button.stop') }}
            </a-button>
            <a-button
              v-if="composeInfo.status === 'running'"
              :loading="restartLoading"
              @click="handleOperation('restart')"
            >
              {{ $t('app.pma.button.restart') }}
            </a-button>
          </a-space>
        </a-tab-pane>

        <a-tab-pane key="port" :title="$t('app.pma.tab.port')">
          <a-form :model="portForm" layout="vertical">
            <a-form-item :label="$t('app.pma.port.label')">
              <a-input-number
                v-model="portForm.port"
                :min="1024"
                :max="65535"
                :placeholder="$t('app.pma.port.placeholder')"
              />
            </a-form-item>
            <a-form-item>
              <a-button
                type="primary"
                :loading="changePortLoading"
                @click="handleChangePort"
              >
                {{ $t('app.pma.button.changePort') }}
              </a-button>
            </a-form-item>
          </a-form>
        </a-tab-pane>

        <a-tab-pane key="servers" :title="$t('app.pma.tab.servers')">
          <div class="mb-3 flex justify-between items-center">
            <a-input
              v-model="serverQuery.name"
              :placeholder="$t('app.pma.servers.searchPlaceholder')"
              class="w-[240px]"
              @press-enter="loadServers"
            />
            <a-space>
              <a-button @click="loadServers">
                {{ $t('common.refresh') }}
              </a-button>
              <a-button type="primary" @click="handleAddServer">
                {{ $t('common.add') }}
              </a-button>
            </a-space>
          </div>
          <a-table
            :data="serverList"
            :pagination="serverPagination"
            @page-change="handleServerPageChange"
            @page-size-change="handleServerPageSizeChange"
          >
            <template #columns>
              <a-table-column
                :title="$t('app.pma.servers.verbose')"
                data-index="verbose"
              />
              <a-table-column
                :title="$t('app.pma.servers.host')"
                data-index="host"
              />
              <a-table-column
                :title="$t('app.pma.servers.port')"
                data-index="port"
              />
              <a-table-column
                :title="$t('common.table.operation')"
                align="center"
                :width="160"
              >
                <template #cell="{ record }">
                  <a-space>
                    <a-button
                      type="text"
                      size="small"
                      @click="handleEditServer(record)"
                    >
                      {{ $t('common.edit') }}
                    </a-button>
                    <a-button
                      type="text"
                      size="small"
                      status="danger"
                      @click="handleDeleteServer(record)"
                    >
                      {{ $t('common.delete') }}
                    </a-button>
                  </a-space>
                </template>
              </a-table-column>
            </template>
          </a-table>
        </a-tab-pane>
      </a-tabs>
    </div>

    <a-drawer
      v-model:visible="serverFormVisible"
      :title="
        serverFormMode === 'add'
          ? $t('app.pma.servers.addTitle')
          : $t('app.pma.servers.editTitle')
      "
      width="400px"
      :mask-closable="false"
    >
      <a-form :model="serverForm" layout="vertical">
        <a-form-item :label="$t('app.pma.servers.form.verbose')">
          <a-input v-model="serverForm.verbose" />
        </a-form-item>
        <a-form-item :label="$t('app.pma.servers.form.host')">
          <a-input v-model="serverForm.host" />
        </a-form-item>
        <a-form-item :label="$t('app.pma.servers.form.port')">
          <a-input v-model="serverForm.port" />
        </a-form-item>
      </a-form>
      <template #footer>
        <a-space>
          <a-button @click="serverFormVisible = false">
            {{ $t('common.cancel') }}
          </a-button>
          <a-button
            type="primary"
            :loading="serverFormLoading"
            @click="handleSubmitServer"
          >
            {{ $t('common.save') }}
          </a-button>
        </a-space>
      </template>
    </a-drawer>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import {
    getPmaComposesApi,
    pmaOperationApi,
    pmaSetPortApi,
    getPmaServersApi,
    pmaAddServerApi,
    pmaUpdateServerApi,
    pmaRemoveServerApi,
  } from '@/api/pma';
  import type { PmaComposeInfo, PmaServerInfo } from '@/entity/Pma';
  import useCurrentHost from '@/composables/current-host';

  const { t } = useI18n();
  const { currentHostId } = useCurrentHost();

  const visible = ref(false);
  const loading = ref(false);
  const composeName = ref('');
  const composeInfo = ref<PmaComposeInfo | null>(null);

  const startLoading = ref(false);
  const stopLoading = ref(false);
  const restartLoading = ref(false);
  const changePortLoading = ref(false);
  const refreshLoading = ref(false);

  const portForm = ref({
    port: 80,
  });

  const serverQuery = ref({
    name: '',
    page: 1,
    page_size: 20,
  });

  const serverList = ref<PmaServerInfo[]>([]);
  const serverPagination = ref({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  const serverFormVisible = ref(false);
  const serverFormMode = ref<'add' | 'edit'>('add');
  const serverFormLoading = ref(false);
  const serverForm = ref({
    verbose: '',
    host: '',
    port: '',
  });
  const editingServer = ref<PmaServerInfo | null>(null);

  const loadCompose = async () => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    loading.value = true;
    try {
      const res = await getPmaComposesApi({
        page: 1,
        page_size: 100,
      });
      composeInfo.value =
        res.composes.find((c) => c.name === composeName.value) || null;
      if (composeInfo.value) {
        portForm.value.port = Number(composeInfo.value.port) || 80;
      }
    } catch (error) {
      Message.error(t('app.pma.message.loadComposeFailed'));
    } finally {
      loading.value = false;
    }
  };

  const loadServers = async () => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    try {
      const res = await getPmaServersApi({
        name: serverQuery.value.name,
        page: serverPagination.value.current,
        page_size: serverPagination.value.pageSize,
      });
      serverList.value = res.servers || [];
      serverPagination.value.total = res.total || 0;
    } catch (error) {
      Message.error(t('app.pma.message.loadServersFailed'));
    }
  };

  const handleServerPageChange = (page: number) => {
    serverPagination.value.current = page;
    loadServers();
  };

  const handleServerPageSizeChange = (pageSize: number) => {
    serverPagination.value.pageSize = pageSize;
    loadServers();
  };

  const handleOperation = async (operation: 'start' | 'stop' | 'restart') => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    let loadingRef;
    if (operation === 'start') loadingRef = startLoading;
    else if (operation === 'stop') loadingRef = stopLoading;
    else loadingRef = restartLoading;

    loadingRef.value = true;
    try {
      await pmaOperationApi({
        name: composeName.value,
        operation,
      });
      Message.success(t('app.pma.message.operationSuccess', { operation }));
      await loadCompose();
    } catch (error) {
      Message.error(t('app.pma.message.operationFailed', { operation }));
    } finally {
      loadingRef.value = false;
    }
  };

  const handleChangePort = async () => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    changePortLoading.value = true;
    try {
      await pmaSetPortApi({
        name: composeName.value,
        port: String(portForm.value.port),
      });
      Message.success(t('app.pma.message.portChangeSuccess'));
      await loadCompose();
    } catch (error) {
      Message.error(t('app.pma.message.portChangeFailed'));
    } finally {
      changePortLoading.value = false;
    }
  };

  const handleRefreshStatus = async () => {
    refreshLoading.value = true;
    try {
      await loadCompose();
      Message.success(t('app.pma.message.refreshSuccess'));
    } catch (error) {
      Message.error(t('app.pma.message.refreshFailed'));
    } finally {
      refreshLoading.value = false;
    }
  };

  const handleAddServer = () => {
    serverFormMode.value = 'add';
    serverForm.value = {
      verbose: '',
      host: '',
      port: '',
    };
    editingServer.value = null;
    serverFormVisible.value = true;
  };

  const handleEditServer = (record: PmaServerInfo) => {
    serverFormMode.value = 'edit';
    serverForm.value = {
      verbose: record.verbose,
      host: record.host,
      port: record.port,
    };
    editingServer.value = record;
    serverFormVisible.value = true;
  };

  const handleDeleteServer = async (record: PmaServerInfo) => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    try {
      await pmaRemoveServerApi({
        name: composeName.value,
        host: record.host,
        port: record.port,
      });
      Message.success(t('common.message.deleteSuccess'));
      await loadServers();
    } catch (error) {
      Message.error(t('common.message.deleteError'));
    }
  };

  const handleSubmitServer = async () => {
    if (!currentHostId.value) {
      Message.error(t('common.error.noHostSelected'));
      return;
    }
    if (!serverForm.value.host || !serverForm.value.port) {
      Message.warning(t('common.form.required'));
      return;
    }
    serverFormLoading.value = true;
    try {
      const payload = {
        name: composeName.value,
        verbose: serverForm.value.verbose,
        host: serverForm.value.host,
        port: serverForm.value.port,
      };
      if (serverFormMode.value === 'add') {
        await pmaAddServerApi(payload);
        Message.success(t('common.message.saveSuccess'));
      } else {
        await pmaUpdateServerApi(payload);
        Message.success(t('common.message.saveSuccess'));
      }
      serverFormVisible.value = false;
      await loadServers();
    } catch (error) {
      Message.error(t('common.message.saveError'));
    } finally {
      serverFormLoading.value = false;
    }
  };

  const show = (name: string) => {
    composeName.value = name;
    visible.value = true;
    loadCompose();
    loadServers();
  };

  defineExpose({
    show,
  });
</script>

<style scoped>
  .loading-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
  }
</style>
