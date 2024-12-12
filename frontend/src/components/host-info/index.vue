<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <a-button class="back">
        <template #icon>
          <icon-arrow-left />
        </template>
      </a-button>
      <div class="host-name truncate">Hostname</div>
      <a-button class="cmd">
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
  import { reactive } from 'vue';
  import UpStreamIcon from '@/assets/icons/upstream.svg';
  import DownStreamIcon from '@/assets/icons/downstream.svg';

  defineProps<{
    collapsed: boolean;
  }>();

  const state = reactive({
    cpu_usage: '',
    cpu_total: '',
    memory_usage: '',
    memory_total: '',
  });

</script>

<style scoped lang="less">
  .ctrl {
    height: 56px;
    border-bottom: 1px solid var(--color-border);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 8px;
    .host-name {
      margin-left: 8px;
      flex: 1;
      min-width: 100px;
      font-size: 16px;
      font-weight: 500;
      color: var(--color-text-1);
    }
  }
  .info {
    padding: 16px 8px;
    .info-content {
      background-color: var(--color-fill-2);
      padding: 10px 5px 10px 10px;
      border-radius: 4px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      .info-content-left {
        flex: 1;
        min-width: 0;
        .info-item {
          font-size: 14px;
          line-height: 22px;
          margin-bottom: 10px;
          display: flex;
          align-items: center;
          justify-content: flex-start;
          &:last-child {
            margin-bottom: 0;
          }
          .info-item-label {
            color: var(--color-text-1);
            font-weight: 500;
            margin-right: 5px;
          }
          .info-item-content {
            flex: 1;
            min-width: 0;
            color: var(--color-text-3);
            display: flex;
            align-items: center;
            justify-content: flex-start;
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
    border-top: 1px solid var(--color-border-2);
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 8px;
    .app-list-title {
      font-size: 16px;
      font-weight: 500;
      line-height: 24px;
      text-indent: 10px;
      color: var(--color-text-1);
    }
    .actions {
      display: flex;
      align-items: center;
      justify-content: center;
    }
    .actions > span {
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      cursor: pointer;
      border-radius: 2px;
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
