<template>
  <a-spin :loading="loading" class="w-full">
    <!-- Docker 环境检测 -->
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
    />

    <div class="list">
      <div class="list-header">
        <a-input-search
          v-model="searchValue"
          class="w-[240px]"
          :placeholder="$t('common.search')"
          :loading="loading"
          allow-clear
          @clear="() => onSearch('')"
          @search="onSearch"
          @press-enter="onSearchEnter"
        />
        <a-button type="text" size="small" @click="load">
          <template #icon><icon-refresh /></template>
          {{ $t('app.managed.refresh') }}
        </a-button>
      </div>
    </div>
    <div class="list-body">
      <a-empty v-if="items.length === 0 && !loading">
        <template #image>
          <icon-apps />
        </template>
        {{ $t('app.managed.empty') }}
      </a-empty>
      <a-grid
        v-else
        :cols="{ xs: 1, sm: 1, md: 2, lg: 2, xl: 3 }"
        :col-gap="24"
        :row-gap="24"
      >
        <a-grid-item v-for="item of filteredItems" :key="item.id">
          <managed-app-card
            :app="item"
            :container-info="containerInfoMap[item.name]"
            @manage="handleManageDatabase"
            @logs="handleShowLogs"
            @operate="handleOperate"
          />
        </a-grid-item>
      </a-grid>
    </div>
  </a-spin>
  <logs-modal ref="logsRef" />
  <database-manager-drawer ref="databaseManagerRef" />
</template>

<script setup lang="ts">
  import { computed, onMounted, onUnmounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { IconApps, IconRefresh } from '@arco-design/web-vue/es/icon';
  import { Message } from '@arco-design/web-vue';
  import { pick } from 'lodash';
  import useLoading from '@/composables/loading';
  import { AppSimpleEntity } from '@/entity/App';
  import { getManagedAppsApi } from '@/api/store';
  import {
    queryContainersApi,
    operateContainersApi,
    connectContainerUsagesFollowApi,
  } from '@/api/docker';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { useDatabaseManager } from '@/composables/use-database-manager';
  import { useConfirm } from '@/composables/confirm';
  import DockerInstallGuide from '@/components/docker-install-guide/index.vue';
  import DatabaseManagerDrawer from '@/components/database-manager/index.vue';
  import LogsModal from '../docker/container/components/logs-modal.vue';
  import ManagedAppCard from './components/managed-app-card.vue';

  const { t } = useI18n();
  const { getDatabaseType } = useDatabaseManager();
  const { confirm } = useConfirm();

  const items = ref<AppSimpleEntity[]>([]);
  const containerInfoMap = ref<Record<string, any>>({});
  const currentSSE = ref<EventSource | null>(null);

  const { loading, showLoading, hideLoading } = useLoading();

  // 停止 SSE 连接
  const stopSSE = () => {
    if (currentSSE.value) {
      currentSSE.value.close();
      currentSSE.value = null;
    }
  };

  // 启动 SSE 实时监控
  const startSSE = () => {
    stopSSE();

    const es = connectContainerUsagesFollowApi({
      state: 'all',
      page: 1,
      page_size: 100,
    });

    es.addEventListener('status', (event: MessageEvent) => {
      try {
        const usagesData = JSON.parse(event.data);
        if (usagesData?.items) {
          // 更新容器资源使用信息
          for (const usage of usagesData.items) {
            for (const app of items.value) {
              const existing = containerInfoMap.value[app.name];
              if (existing && existing.container_id === usage.container_id) {
                containerInfoMap.value[app.name] = {
                  ...existing,
                  ...pick(usage, [
                    'cpu_percent',
                    'memory_usage',
                    'memory_limit',
                    'memory_percent',
                  ]),
                };
              }
            }
          }
        }
      } catch (e) {
        console.error('解析容器资源使用数据失败', e);
      }
    });

    es.addEventListener('error', () => {
      console.error('SSE 连接错误');
    });

    currentSSE.value = es;
  };

  // 加载容器状态信息
  const loadContainerStatus = async () => {
    if (items.value.length === 0) return;

    try {
      // 查询所有容器状态
      const containerData = await queryContainersApi({
        state: 'all',
        page: 1,
        page_size: 100,
      });

      // 构建容器信息映射（通过容器名称匹配应用名称）
      const map: Record<string, any> = {};
      for (const container of containerData?.items || []) {
        // 容器名称可能包含前缀，尝试匹配
        for (const app of items.value) {
          if (
            container.name === app.name ||
            container.name.includes(app.name) ||
            app.name.includes(container.name.replace(/^\//, ''))
          ) {
            map[app.name] = container;
            break;
          }
        }
      }
      containerInfoMap.value = map;

      // 启动 SSE 实时监控
      startSSE();
    } catch (err: any) {
      console.error('加载容器状态失败', err);
    }
  };

  // 加载管理应用列表
  const load = async () => {
    if (loading.value) {
      return;
    }
    showLoading();
    try {
      const data = await getManagedAppsApi();
      items.value = data?.items || [];
      // 加载容器状态
      await loadContainerStatus();
    } catch (err: any) {
      await showErrorWithDockerCheck(err.message, err);
    } finally {
      hideLoading();
    }
  };

  onMounted(() => {
    load();
  });

  onUnmounted(() => {
    stopSSE();
  });

  const searchValue = ref('');
  const filteredItems = computed(() => {
    if (!searchValue.value) {
      return items.value;
    }
    const keyword = searchValue.value.toLowerCase();
    return items.value.filter(
      (item) =>
        item.name.toLowerCase().includes(keyword) ||
        item.display_name.toLowerCase().includes(keyword)
    );
  });

  const onSearch = (value: string) => {
    searchValue.value = value;
  };
  const onSearchEnter = () => {
    onSearch(searchValue.value);
  };

  // 管理数据库
  const databaseManagerRef = ref<InstanceType<typeof DatabaseManagerDrawer>>();
  const handleManageDatabase = (item: AppSimpleEntity) => {
    const dbType = getDatabaseType(item.name);
    if (!dbType) return;
    databaseManagerRef.value?.show(dbType, item.name);
  };

  // 查看日志
  const logsRef = ref<InstanceType<typeof LogsModal>>();
  const handleShowLogs = (_app: AppSimpleEntity, containerName: string) => {
    logsRef.value?.connect(containerName);
    logsRef.value?.show();
  };

  // 容器操作（启动/停止/重启）
  const handleOperate = async (
    containerName: string,
    operation: 'start' | 'stop' | 'restart'
  ) => {
    const confirmMessages: Record<string, string> = {
      start: t('app.managed.confirm.start'),
      stop: t('app.managed.confirm.stop'),
      restart: t('app.managed.confirm.restart'),
    };

    if (!(await confirm(confirmMessages[operation]))) {
      return;
    }

    try {
      const result = await operateContainersApi({
        names: [containerName],
        operation,
      });
      if (result.success) {
        Message.success(t('app.managed.operation.success'));
        // 刷新容器状态
        await loadContainerStatus();
      } else {
        await showErrorWithDockerCheck(result.message);
      }
    } catch (err: any) {
      await showErrorWithDockerCheck(err.message, err);
    }
  };

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    if (status === 'installed') {
      load();
    }
  };

  defineExpose({
    load,
  });
</script>

<style scoped>
  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.143rem;
  }
</style>
