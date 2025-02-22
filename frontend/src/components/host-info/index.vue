<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <a-button class="back" @click="gotoManage">
        <template #icon>
          <icon-arrow-left />
        </template>
      </a-button>
      <div class="host-name truncate">Hostname</div>
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
            <div class="info-item-label">内存: </div>
            <div class="info-item-content"> {{ state.memory_usage }} </div>
          </div>
          <div class="info-item">
            <div class="info-item-label">网络: </div>
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
      <div class="app-list-title">应用列表</div>
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
  import { inject, reactive, ref, onMounted } from 'vue';
  import UpStreamIcon from '@/assets/icons/upstream.svg';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import router from '@/router';
  import { SELECT_HOST } from '@/router/constants';
  import { getHostStatusApi } from '@/api/host';
  import { Message } from '@arco-design/web-vue';
  import useCurrentHost from '@/hooks/current-host';
  import { formatTransferSpeed } from '@/utils/format';

  defineProps<{
    collapsed: boolean;
  }>();

  const { currentHostId } = useCurrentHost();

  const state = reactive({
    cpu_usage: '0%',
    memory_usage: '0MB/0MB',
    network_up: '0KB/s',
    network_down: '0KB/s',
  });

  const isLoading = ref(false);

  const refreshStatus = async () => {
    if (!currentHostId.value || isLoading.value) return;

    isLoading.value = true;
    try {
      const result = await getHostStatusApi(currentHostId.value);
      state.cpu_usage = result.cpu + '%';
      state.memory_usage = result.mem_used + '/' + result.mem_total;
      state.network_up = formatTransferSpeed(result.tx, 0);
      state.network_down = formatTransferSpeed(result.rx, 0);
    } catch (error) {
      Message.error('获取状态失败');
    } finally {
      isLoading.value = false;
    }
  };

  const gotoManage = () => {
    router.push(SELECT_HOST);
  };

  const openTerminal = inject<() => void>('openTerminal');

  const gotoSysInfo = () => {
    router.push('/app/sysinfo');
  };

  onMounted(() => {
    if (currentHostId.value) {
      refreshStatus();
    }
  });
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
      &:hover {
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
