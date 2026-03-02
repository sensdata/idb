<template>
  <div class="box">
    <div class="host-switch">
      <a-tooltip position="bottom" :content="$t('components.switchHost.tips')">
        <button class="btn color-primary" @click="handleClick">
          <IconHome />
          <span class="host-name truncate">{{ currentHost?.name }}</span>
        </button>
      </a-tooltip>
    </div>
    <div v-if="currentModuleName" class="current-module truncate">
      {{ currentModuleName }}
    </div>
    <div class="host-status">
      <div class="status-item cpu-item">
        <span class="status-label">CPU:</span>
        <span class="status-value cpu-value">{{ state.cpu_usage }}</span>
      </div>
      <div class="status-item memory-item">
        <span class="status-label">{{ $t('host.info.memory') }}:</span>
        <span class="status-value memory-value">{{ state.memory_usage }}</span>
      </div>
      <div class="status-item network-item">
        <span class="status-label">{{ $t('host.info.network') }}:</span>
        <span class="status-value network">
          <up-stream-icon class="status-icon" />
          <span class="network-value">{{ state.network_up }}</span>
          <down-stream-icon class="status-icon downstream" />
          <span class="network-value">{{ state.network_down }}</span>
        </span>
      </div>
    </div>
  </div>
  <a-drawer
    :width="640"
    :visible="drawerVisible"
    :footer="false"
    unmountOnClose
    @cancel="handleCancel"
  >
    <template #title>
      {{ $t('components.switchHost.currentlabel') }}
      <strong> {{ currentHost?.name }}({{ currentHost?.addr }}) </strong>
    </template>
    <a-input-search
      v-model="searchValue"
      class="mt-2"
      :placeholder="$t('components.switchHost.searchPlaceholder')"
      search-button
      allow-clear
      @clear="() => handleSearch('')"
      @search="handleSearch"
      @press-enter="handleSearchEnter"
    />
    <div class="mt-5">
      <idb-table
        ref="gridRef"
        :columns="columns"
        :fetch="getHostListApi"
        :hasToolbar="false"
      >
        <template #operation="{ record }: { record: HostEntity }">
          <a-button
            v-if="record.id === currentHost?.id"
            type="primary"
            disabled
            size="mini"
          >
            {{ $t('components.switchHost.operation.selected') }}
          </a-button>
          <a-button
            v-else
            type="primary"
            size="mini"
            @click="handleSelect(record)"
          >
            {{ $t('components.switchHost.operation.select') }}
          </a-button>
        </template>
      </idb-table>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch, onUnmounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { HostEntity } from '@/entity/Host';
  import {
    getHostListApi,
    connectHostStatusFollowApi,
    getHostStatusApi,
  } from '@/api/host';
  import { useHostStore } from '@/store';
  import { formatTransferSpeed } from '@/utils/format';
  import usetCurrentHost from '@/composables/current-host';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import UpStreamIcon from '@/assets/icons/upstream.svg';

  const { t } = useI18n();

  const gridRef = ref();
  const route = useRoute();
  const hostStore = useHostStore();
  const { switchHost } = usetCurrentHost();
  const currentHost = computed(() => hostStore.current);
  const currentModuleName = computed(() => {
    const localeKey = route.meta?.locale;
    if (typeof localeKey === 'string' && localeKey.length > 0) {
      return t(localeKey);
    }
    return '';
  });

  const drawerVisible = ref(false);
  const handleClick = () => {
    drawerVisible.value = true;
  };
  const handleCancel = () => {
    drawerVisible.value = false;
  };

  // 搜索
  const searchValue = ref('');
  const handleSearch = (value: string) => {
    searchValue.value = value;
    gridRef.value?.load({
      search: value,
      page: 1,
    });
  };
  const handleSearchEnter = () => {
    handleSearch(searchValue.value);
  };
  const columns = [
    {
      dataIndex: 'addr',
      title: t('components.switchHost.column.addr'),
      width: 150,
    },
    {
      dataIndex: 'name',
      title: t('components.switchHost.column.name'),
      width: 150,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      slotName: 'operation',
    },
  ];

  const handleSelect = (record: HostEntity) => {
    switchHost(record.id, true);
    handleCancel();
  };

  const state = reactive({
    cpu_usage: '0%',
    memory_usage: '0MB/0MB',
    network_up: '0KB/s',
    network_down: '0KB/s',
  });

  const isLoading = ref(false);
  const esRef = ref<EventSource>();
  const connectedHostId = ref<number | null>(null);
  const lastRequestedHostId = ref<number | null>(null);

  const refreshStatus = async () => {
    if (!hostStore.currentId || isLoading.value) {
      return;
    }

    isLoading.value = true;
    try {
      const result = await getHostStatusApi(hostStore.currentId);
      state.cpu_usage = result.cpu + '%';
      state.memory_usage = result.mem_used + '/' + result.mem_total;
      state.network_up = formatTransferSpeed(result.tx);
      state.network_down = formatTransferSpeed(result.rx);
    } catch (error) {
      Message.error(t('host.info.status.failed'));
    } finally {
      isLoading.value = false;
    }
  };

  const stopSSE = () => {
    if (esRef.value) {
      esRef.value.close();
      esRef.value = undefined;
    }
    connectedHostId.value = null;
  };

  const startSSE = () => {
    if (!hostStore.currentId) {
      return;
    }

    if (connectedHostId.value === hostStore.currentId && esRef.value) {
      return;
    }

    stopSSE();
    const es = connectHostStatusFollowApi(hostStore.currentId);
    esRef.value = es;
    connectedHostId.value = hostStore.currentId;

    es.addEventListener('status', (event) => {
      const data = JSON.parse(event.data);
      state.cpu_usage = data.cpu + '%';
      state.memory_usage = data.mem_used + '/' + data.mem_total;
      state.network_up = formatTransferSpeed(data.tx);
      state.network_down = formatTransferSpeed(data.rx);
    });

    es.addEventListener('error', (err) => {
      console.error('SSE error', err);
    });
  };

  watch(
    () => hostStore.currentId,
    (v?: number) => {
      if (v) {
        if (v !== lastRequestedHostId.value) {
          lastRequestedHostId.value = v;
          refreshStatus();
          startSSE();
        }
      } else {
        lastRequestedHostId.value = null;
        stopSSE();
      }
    },
    { immediate: true }
  );

  onUnmounted(() => {
    stopSSE();
  });
</script>

<style scoped>
  .box {
    display: flex;
    gap: 12px;
    align-items: center;
    justify-content: flex-start;
    padding: 10px 16px;
    background-color: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  .host-switch {
    display: flex;
    align-items: center;
    max-width: 60%;
  }

  .current-module {
    box-sizing: border-box;
    display: inline-flex;
    align-items: center;
    max-width: 36%;
    height: 36px;
    padding: 0 12px;
    overflow: hidden;
    font-size: 14px;
    font-weight: 500;
    line-height: 36px;
    color: var(--color-text-2);
    text-align: left;
    background-color: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 18px;
    transition: all 0.2s ease;
  }

  .host-status {
    box-sizing: border-box;
    display: inline-flex;
    flex: 0 1 auto;
    flex-wrap: nowrap;
    gap: 0 10px;
    align-items: center;
    width: fit-content;
    min-width: 500px;
    max-width: 100%;
    height: 36px;
    padding: 0 12px;
    margin-left: auto;
    font-size: 12px;
    line-height: 1.4;
    color: var(--color-text-2);
    background-color: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 18px;
  }

  .status-item {
    display: inline-flex;
    flex: 0 0 auto;
    gap: 4px;
    align-items: center;
    white-space: nowrap;
  }

  .status-label {
    color: var(--color-text-3);
  }

  .status-value {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: flex-end;
    font-feature-settings: 'tnum';
    font-variant-numeric: tabular-nums;
    color: var(--color-text-1);
    white-space: nowrap;
  }

  .cpu-value {
    min-width: 5.5ch;
  }

  .memory-value {
    min-width: 16ch;
  }

  .network {
    gap: 4px;
    min-width: 23ch;
  }

  .network-value {
    display: inline-block;
    width: auto;
    min-width: 8.5ch;
    font-feature-settings: 'tnum';
    font-variant-numeric: tabular-nums;
    text-align: right;
  }

  .status-icon {
    width: 12px;
    height: 12px;
  }

  .downstream {
    margin-left: 4px;
  }

  .current-module::before {
    width: 6px;
    height: 6px;
    margin-right: 8px;
    content: '';
    background-color: var(--color-primary-light-4);
    border-radius: 999px;
  }

  .btn {
    display: inline-flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
    max-width: 100%;
    height: 36px;
    padding: 0 14px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    background-color: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 18px;
    transition: all 0.2s ease;
  }

  .btn:hover {
    color: var(--color-primary-6);
    background-color: var(--color-fill-3);
    border-color: var(--color-border-3);
  }

  .btn:active {
    transform: translateY(1px);
  }

  .host-name {
    max-width: 220px;
  }

  @media (width <= 1280px) {
    .host-switch {
      max-width: 45%;
    }
    .current-module {
      max-width: 28%;
    }
    .host-status {
      min-width: 460px;
      font-size: 11px;
    }
  }

  @media (width <= 992px) {
    .host-switch {
      max-width: 100%;
    }
    .host-status {
      flex-basis: auto;
      width: 100%;
      max-width: 100%;
      margin-left: 0;
      overflow: auto hidden;
    }
    .current-module {
      max-width: 50%;
    }
  }
</style>
