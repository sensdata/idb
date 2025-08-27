<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <a-button class="back" @click="gotoManage">
        <template #icon>
          <icon-arrow-left />
        </template>
      </a-button>
      <div class="host-name truncate"
        >{{ hostStore.current?.name || hostStore.current?.addr }}
      </div>
      <a-button class="btn" @click="openTerminal?.()">
        <template #icon>
          <icon-code-square />
        </template>
      </a-button>
    </div>
    <div class="info">
      <div class="info-content">
        <div class="info-content-left">
          <div class="info-item">
            <div class="info-item-label">CPU: </div>
            <div class="info-item-content"> {{ state.cpu_usage }} </div>
          </div>
          <div class="info-item">
            <div class="info-item-label">{{ $t('host.info.memory') }}: </div>
            <div class="info-item-content"> {{ state.memory_usage }} </div>
          </div>
          <div class="info-item">
            <div class="info-item-label">{{ $t('host.info.network') }}: </div>
            <div class="info-item-content">
              <up-stream-icon class="info-item-content-icon" />
              <span>{{ state.network_up }}</span>
              <down-stream-icon class="info-item-content-icon downstream" />
              <span>{{ state.network_down }}</span>
            </div>
          </div>
        </div>
        <div class="info-content-right">
          <a-button class="btn" @click="gotoSysInfo()">
            <template #icon>
              <icon-right />
            </template>
          </a-button>
        </div>
      </div>
    </div>
    <div class="app-list">
      <div class="app-list-title">{{ $t('host.info.applist') }}</div>
      <div class="actions">
        <span class="refresh" @click="refreshStatus">
          <icon-refresh :spin="isLoading" />
        </span>
        <span class="setting"><icon-settings /></span>
        <span class="home"><icon-home /></span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { inject, reactive, ref, watch, onUnmounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import router from '@/router';
  import { useHostStore } from '@/store';
  import { SELECT_HOST } from '@/router/constants';
  import { connectHostStatusFollowApi, getHostStatusApi } from '@/api/host';
  import { formatTransferSpeed } from '@/utils/format';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import UpStreamIcon from '@/assets/icons/upstream.svg';

  defineProps<{
    collapsed: boolean;
  }>();

  const hostStore = useHostStore();
  const { t } = useI18n();

  const state = reactive({
    cpu_usage: '0%',
    memory_usage: '0MB/0MB',
    network_up: '0KB/s',
    network_down: '0KB/s',
  });

  const isLoading = ref(false);

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

  const esRef = ref<EventSource>();
  const stopSSE = () => {
    if (esRef.value) {
      esRef.value.close();
    }
  };
  const startSSE = () => {
    if (!hostStore.currentId) {
      return;
    }

    stopSSE();

    const es = connectHostStatusFollowApi(hostStore.currentId);
    esRef.value = es;
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
  onUnmounted(() => {
    stopSSE();
  });

  watch(
    () => hostStore.currentId,
    (v?: number) => {
      if (v) {
        refreshStatus();
        startSSE();
      } else {
        stopSSE();
      }
    },
    {
      immediate: true,
    }
  );

  const gotoManage = () => {
    router.push(SELECT_HOST);
  };

  const openTerminal = inject<() => void>('openTerminal');

  const gotoSysInfo = () => {
    router.push('/app/sysinfo');
  };
</script>

<style scoped lang="less">
  .ctrl {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    padding: 0 8px;
    border-bottom: 1px solid var(--color-border);
    .host-name {
      flex: 1;
      min-width: 100px;
      margin-left: 8px;
      color: var(--color-text-1);
      font-weight: 500;
      font-size: 16px;
    }
  }

  .info {
    padding: 16px 8px;
    .info-content {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 10px 5px 10px 10px;
      background-color: var(--color-fill-2);
      border-radius: 4px;
      .info-content-left {
        flex: 1;
        min-width: 0;
        .info-item {
          display: flex;
          align-items: center;
          justify-content: flex-start;
          margin-bottom: 10px;
          font-size: 14px;
          line-height: 22px;
          &:last-child {
            margin-bottom: 0;
          }
          .info-item-label {
            margin-right: 5px;
            color: var(--color-text-1);
            font-weight: 500;
          }
          .info-item-content {
            display: flex;
            flex: 1;
            align-items: center;
            justify-content: flex-start;
            min-width: 0;
            color: var(--color-text-3);
          }
          .info-item-content-icon {
            width: 12px;
            height: 12px;
            margin-right: 2px;
          }
          .downstream {
            margin-left: 8px;
          }
        }
      }
    }
  }

  .app-list {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 56px;
    padding: 0 8px;
    border-top: 1px solid var(--color-border-2);
    .app-list-title {
      color: var(--color-text-1);
      font-weight: 500;
      font-size: 16px;
      line-height: 24px;
      text-indent: 10px;
    }
    .actions {
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .actions > span {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      border-radius: 2px;
      cursor: pointer;
      color: var(--color-text-2);
      &:hover {
        background-color: var(--color-fill-2);
        color: var(--color-text-1);
      }
    }
  }

  .collapsed {
    .host-name,
    .btn,
    .info,
    .app-list {
      display: none;
    }
  }
</style>
