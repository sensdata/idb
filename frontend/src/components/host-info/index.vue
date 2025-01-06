<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <a-button class="back" @click="gotoManage">
        <template #icon>
          <icon-arrow-left />
        </template>
      </a-button>
      <div class="host-name truncate">Hostname</div>
      <a-button class="cmd" @click="openTerminal?.()">
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
            <div class="info-item-content"> 17.82% </div>
          </div>
          <div class="info-item">
            <div class="info-item-label">内存: </div>
            <div class="info-item-content"> 1.2G/3.7G </div>
          </div>
          <div class="info-item">
            <div class="info-item-label">网络: </div>
            <div class="info-item-content">
              <up-stream-icon class="info-item-content-icon" />
              <span>128.2K/s</span>
              <down-stream-icon class="info-item-content-icon downstream" />
              <span>128.3K/s</span>
            </div>
          </div>
        </div>
        <div class="info-content-right">
          <a-button class="cmd">
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
        <span class="refresh"><icon-refresh /></span>
        <span class="setting"><icon-settings /></span>
        <span class="home"><icon-home /></span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { inject, reactive } from 'vue';
  import UpStreamIcon from '@/assets/icons/upstream.svg';
  import DownStreamIcon from '@/assets/icons/downstream.svg';
  import router from '@/router';
  import { SELECT_HOST } from '@/router/constants';

  defineProps<{
    collapsed: boolean;
  }>();

  const state = reactive({
    cpu_usage: '',
    cpu_total: '',
    memory_usage: '',
    memory_total: '',
  });

  const gotoManage = () => {
    router.push(SELECT_HOST);
  };

  const openTerminal = inject<() => void>('openTerminal');
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
    .cmd,
    .info,
    .app-list {
      display: none;
    }
  }
</style>
