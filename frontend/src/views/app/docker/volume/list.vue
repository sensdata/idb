<template>
  <div>
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />

    <a-row class="mb-4" :gutter="12">
      <a-col :xs="24" :sm="8">
        <a-card size="small">
          <a-statistic
            :title="t('app.docker.volume.stats.filteredTotal')"
            :value="stats.filteredTotal"
          />
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-card size="small">
          <a-statistic
            :title="t('app.docker.volume.stats.named')"
            :value="stats.namedCount"
          />
        </a-card>
      </a-col>
      <a-col :xs="24" :sm="8">
        <a-card size="small">
          <a-statistic
            :title="t('app.docker.volume.stats.anonymous')"
            :value="stats.anonymousCount"
          />
        </a-card>
      </a-col>
    </a-row>

    <idb-table
      ref="tableRef"
      :columns="columns"
      :has-search="true"
      :fetch="fetchVolumes"
      :beforeFetchHook="beforeFetchHook"
    >
      <template #leftActions>
        <a-space size="small">
          <a-radio-group
            v-model="volumeTypeFilter"
            type="button"
            size="small"
            @change="handleTypeFilterChange"
          >
            <a-radio
              v-for="option in volumeTypeOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-radio>
          </a-radio-group>
          <a-button type="primary" @click="onCreateVolumeClick">
            {{ t('app.docker.volume.list.action.create') }}
          </a-button>
        </a-space>
      </template>
      <template #name="{ record }">
        <div class="volume-name-cell">
          <a-link :hoverable="false" @click="handleInspect(record)">
            <a-tooltip :content="record.name">
              <span>{{ formatVolumeName(record.name) }}</span>
            </a-tooltip>
          </a-link>
          <a-tag
            size="small"
            :color="isAnonymousVolume(record.name) ? 'orange' : 'arcoblue'"
          >
            {{
              isAnonymousVolume(record.name)
                ? t('app.docker.volume.list.type.anonymous')
                : t('app.docker.volume.list.type.named')
            }}
          </a-tag>
          <a-button type="text" size="mini" @click="copyText(record.name)">
            <icon-copy />
          </a-button>
        </div>
      </template>
      <template #mountPoint="{ record }">
        <div class="mount-point-cell">
          <a-tooltip :content="record.mount_point">
            <span>{{ shortenMiddle(record.mount_point, 72) }}</span>
          </a-tooltip>
          <a-button
            type="text"
            size="mini"
            @click="copyText(record.mount_point)"
          >
            <icon-copy />
          </a-button>
        </div>
      </template>
      <template #operation="{ record }">
        <idb-table-operation
          type="dropdown"
          :options="getOperationOptions(record)"
        />
      </template>
    </idb-table>
    <create-volume-drawer ref="createVolumeRef" @success="reload" />
    <inspect-drawer ref="inspectRef" />
  </div>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { formatTime } from '@/utils/format';
  import { ApiListResult } from '@/types/global';
  import {
    getVolumesApi,
    batchDeleteVolumeApi,
    inspectApi,
  } from '@/api/docker';
  import IdbTableOperation from '@/components/idb-table-operation/index.vue';
  import CreateVolumeDrawer from './components/create-volume-drawer.vue';
  import InspectDrawer from './components/inspect-drawer.vue';

  const { t } = useI18n();
  const tableRef = ref();
  const inspectRef = ref<InstanceType<typeof InspectDrawer>>();
  const createVolumeRef = ref();
  const volumeTypeFilter = ref<'all' | 'named' | 'anonymous'>('all');
  const stats = ref({
    filteredTotal: 0,
    namedCount: 0,
    anonymousCount: 0,
  });
  const reload = () => tableRef.value?.reload();
  const MAX_FILTER_SCAN_SIZE = 5000;

  const volumeTypeOptions = computed(() => [
    { value: 'all', label: t('app.docker.volume.list.filter.type.all') },
    { value: 'named', label: t('app.docker.volume.list.filter.type.named') },
    {
      value: 'anonymous',
      label: t('app.docker.volume.list.filter.type.anonymous'),
    },
  ]);

  const isAnonymousVolume = (name: string) =>
    /^[a-f0-9]{64}$/i.test(name || '');
  const formatVolumeName = (name: string) => {
    if (!isAnonymousVolume(name)) {
      return name;
    }
    return `${name.slice(0, 12)}...${name.slice(-8)}`;
  };
  const shortenMiddle = (text: string, maxLength = 72) => {
    if (!text || text.length <= maxLength) {
      return text || '-';
    }
    const head = Math.floor((maxLength - 3) / 2);
    const tail = maxLength - 3 - head;
    return `${text.slice(0, head)}...${text.slice(-tail)}`;
  };

  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text || '');
      Message.success(t('common.message.copy_success'));
    } catch {
      Message.error(t('common.message.copy_failed'));
    }
  };

  const handleTypeFilterChange = () => {
    tableRef.value?.load?.({ page: 1 });
  };

  const fetchVolumes = async (params: any) => {
    if (volumeTypeFilter.value === 'all') {
      const result = await getVolumesApi(params);
      const items = Array.isArray(result.items) ? result.items : [];
      const anonymousCount = items.filter((item: any) =>
        isAnonymousVolume(item.name)
      ).length;
      stats.value = {
        filteredTotal: result.total || 0,
        namedCount: items.length - anonymousCount,
        anonymousCount,
      };
      return result;
    }

    const fullResult = await getVolumesApi({
      ...params,
      page: 1,
      page_size: MAX_FILTER_SCAN_SIZE,
    });
    const fullItems = Array.isArray(fullResult.items) ? fullResult.items : [];
    const filteredItems = fullItems.filter((item: any) => {
      const anonymous = isAnonymousVolume(item.name);
      return volumeTypeFilter.value === 'anonymous' ? anonymous : !anonymous;
    });

    const currentPage = Number(params.page) || 1;
    const pageSize = Number(params.page_size) || 20;
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    const pageItems = filteredItems.slice(start, end);

    const anonymousCount = filteredItems.filter((item: any) =>
      isAnonymousVolume(item.name)
    ).length;
    stats.value = {
      filteredTotal: filteredItems.length,
      namedCount: filteredItems.length - anonymousCount,
      anonymousCount,
    };

    return {
      items: pageItems,
      total: filteredItems.length,
      page: currentPage,
      page_size: pageSize,
    } as ApiListResult<any>;
  };

  const beforeFetchHook = (fetchParams: any) => {
    const nextParams = { ...fetchParams };
    if (typeof nextParams.search === 'string') {
      const keyword = nextParams.search.trim();
      nextParams.info = keyword || undefined;
    }
    delete nextParams.search;
    return nextParams;
  };

  async function handleInspect(record: any) {
    try {
      const data = await inspectApi({
        type: 'volume',
        id: record.name!,
      });
      inspectRef.value?.show(data.info);
    } catch (err: any) {
      await showErrorWithDockerCheck(err?.message, err);
    }
  }

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.docker.volume.list.column.name'),
      width: 180,
      slotName: 'name',
    },
    {
      dataIndex: 'driver',
      title: t('app.docker.volume.list.column.driver'),
      width: 100,
    },
    {
      dataIndex: 'mount_point',
      title: t('app.docker.volume.list.column.mount_point'),
      width: 320,
      slotName: 'mountPoint',
    },
    {
      dataIndex: 'created_at',
      title: t('app.docker.volume.list.column.created'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return formatTime(record.created_at);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'left' as const,
      width: 100,
      slotName: 'operation',
    },
  ];

  const onCreateVolumeClick = () => createVolumeRef.value?.show();

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.volume.list.operation.inspect'),
      click: () => handleInspect(record),
    },
    {
      text: t('app.docker.volume.list.operation.delete'),
      status: 'danger' as const,
      confirm: t('app.docker.volume.list.operation.delete.confirm'),
      click: async () => {
        try {
          await batchDeleteVolumeApi({ force: false, sources: record.name });
          Message.success(t('app.docker.volume.list.operation.delete.success'));
          reload();
        } catch (e: any) {
          await showErrorWithDockerCheck(
            e.message || t('app.docker.volume.list.operation.delete.failed'),
            e
          );
        }
      },
    },
  ];

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 如果 Docker 状态变化，可以重新加载存储卷列表
    if (status === 'installed') {
      reload();
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // Docker 安装完成后重新加载存储卷列表
    reload();
  };
</script>

<style scoped>
  .volume-name-cell {
    display: flex;
    gap: 0.375rem;
    align-items: center;
  }

  .mount-point-cell {
    display: flex;
    gap: 0.25rem;
    align-items: center;
  }
</style>
