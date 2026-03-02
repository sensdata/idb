<template>
  <div class="host-info" :class="{ collapsed: collapsed }">
    <div class="ctrl">
      <router-link v-slot="{ navigate }" to="/manage/host" custom>
        <a-button class="back" @click="navigate">
          <template #icon>
            <icon-arrow-left />
          </template>
        </a-button>
      </router-link>
      <div class="host-name truncate"
        >{{ hostStore.current?.addr || hostStore.current?.name }}
      </div>
      <a-button class="btn" @click="openTerminal?.()">
        <template #icon>
          <icon-code-square />
        </template>
      </a-button>
    </div>
    <div class="app-list">
      <div class="app-list-title">{{ $t('host.info.applist') }}</div>
      <div class="actions">
        <span class="setting"><icon-settings /></span>
        <span class="home"><icon-home /></span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { inject } from 'vue';
  import { useHostStore } from '@/store';

  defineProps<{
    collapsed: boolean;
  }>();

  const hostStore = useHostStore();

  const openTerminal = inject<() => void>('openTerminal');
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
