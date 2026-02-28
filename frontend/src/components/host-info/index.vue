<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <a-button class="back" @click="gotoManage">
        <template #icon>
          <icon-arrow-left />
        </template>
      </a-button>
      <div class="host-name truncate"
        >{{ hostStore.current?.addr || hostStore.current?.name }}
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
  // 记录当前已连接的 hostId，避免重复连接
  const connectedHostId = ref<number | null>(null);

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

    // 如果已经连接到相同的 hostId，不需要重新连接
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

  onUnmounted(() => {
    stopSSE();
  });

  // 记录上次请求的 hostId，避免重复请求
  const lastRequestedHostId = ref<number | null>(null);

  watch(
    () => hostStore.currentId,
    (v?: number, oldV?: number) => {
      if (v) {
        // 只有当 hostId 真正变化时才重新请求
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
    {
      immediate: true,
    }
  );

  const gotoManage = () => {
    console.log('gotoManage clicked, navigating to /manage/host');
    router.push('/manage/host');
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
    height: 4rem;
    padding: 0 0.571rem;
    border-bottom: 1px solid var(--color-border);
    .back {
      position: relative;
      z-index: 10;
      pointer-events: auto;
      cursor: pointer;
    }
    .host-name {
      flex: 1;
      min-width: 100px;
      margin-left: 0.571rem;
      font-size: 1.143rem;
      font-weight: 500;
      color: var(--color-text-1);
    }
  }

  .info {
    padding: 1.143rem 0.571rem;
    .info-content {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.714rem 0.357rem 0.714rem 0.714rem;
      background-color: var(--color-fill-2);
      border-radius: 0.286rem;
      .info-content-left {
        flex: 1;
        min-width: 0;
        .info-item {
          display: flex;
          align-items: center;
          justify-content: flex-start;
          margin-bottom: 0.714rem;
          font-size: 1rem;
          line-height: 1.571rem;
          &:last-child {
            margin-bottom: 0;
          }
          .info-item-label {
            margin-right: 0.357rem;
            font-weight: 500;
            color: var(--color-text-1);
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
            width: 0.857rem;
            height: 0.857rem;
            margin-right: 0.143rem;
          }
          .downstream {
            margin-left: 0.571rem;
          }
        }
      }
    }
  }

  .app-list {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 4rem;
    padding: 0 0.571rem;
    border-top: 1px solid var(--color-border-2);
    .app-list-title {
      font-size: 1.143rem;
      font-weight: 500;
      line-height: 1.714rem;
      color: var(--color-text-1);
      text-indent: 0.714rem;
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
      width: 2.286rem;
      height: 2.286rem;
      color: var(--color-text-2);
      cursor: pointer;
      border-radius: 0.143rem;
      &:hover {
        color: var(--color-text-1);
        background-color: var(--color-fill-2);
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
