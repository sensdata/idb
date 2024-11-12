<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="100vh"
    placement="bottom"
  >
    <template #title>{{ $t('components.terminalDrawer.title') }}</template>
    <a-tabs
      class="terminal-tabs"
      type="card-gutter"
      :editable="true"
      auto-switch
      lazy-load
      @delete="handleDelete"
    >
      <template #extra>
        <a-popover
          v-model:popup-visible="popoverVisible"
          placement="bottom"
          trigger="click"
          :content-style="{
            padding: 0,
            width: '320px',
          }"
        >
          <div class="arco-tabs-nav-add-btn">
            <span class="arco-icon-hover">
              <icon-plus />
            </span>
          </div>
          <template #content>
            <div class="popover-head">
              <div class="arco-popover-title">
                {{ $t('components.terminalDrawer.add.title') }}
              </div>
              <span
                class="arco-icon-hover popover-close"
                @click="handlePopClose"
              >
                <icon-close />
              </span>
            </div>
            <div class="popover-body">
              <a-input-search
                v-model="searchValue"
                :placeholder="
                  $t('components.terminalDrawer.add.search.placeholder')
                "
                allow-clear
                @clear="handleSearch('')"
                @search="handleSearch"
                @press-enter="handleSearchEnter"
              />
              <div class="host-list">
                <div
                  v-for="item of hostItems"
                  :key="item.id"
                  class="host-item"
                  @click="handleSelectHost(item)"
                >
                  <span class="host-addr">{{ item.addr }}</span>
                  <span class="host-name"> {{ item.name }}</span>
                </div>
              </div>
            </div>
          </template>
        </a-popover>
      </template>
      <a-tab-pane v-for="item of terms" :key="item.key" :title="item.title">
        <x-term ref="item.xtermRef" :hostId="item.hostId" />
      </a-tab-pane>
    </a-tabs>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch, Ref } from 'vue';
  import XTerm from '@/components/xterm/index.vue';
  import { useHostStore } from '@/store';
  import { HostEntity } from '@/entity/Host';

  type XTermInstance = InstanceType<typeof XTerm> | undefined;

  interface TermItem {
    key: string;
    hostId: number;
    title: string;
    xtermRef: Ref<XTermInstance>;
  }

  const visible = defineModel('visible', {
    type: Boolean,
    required: true,
  });
  const hostStore = useHostStore();
  const terms = ref<TermItem[]>([]);
  function addTerm(host: HostEntity) {
    const xtermRef: any = ref<XTermInstance>();
    const key = Math.random().toString(36).slice(2);
    terms.value.push({
      key,
      hostId: host.id,
      title: host.name,
      xtermRef,
    });
  }

  function handleDelete(key: string | number) {
    const index = terms.value.findIndex((item) => item.key === key);
    if (index !== -1) {
      terms.value.splice(index, 1);
    }
  }

  const firstShow = ref(true);
  watch(visible, (val) => {
    if (val && firstShow.value) {
      firstShow.value = false;
      addTerm(hostStore.default);
    }
  });

  const popoverVisible = ref(false);
  function handlePopClose() {
    popoverVisible.value = false;
  }

  function handleSelectHost(host: HostEntity) {
    addTerm(host);
    popoverVisible.value = false;
  }

  const searchValue = ref('');
  const hostItems = ref<HostEntity[]>(hostStore.items.slice(0));
  function handleSearch(value: string) {
    hostItems.value = hostStore.items.filter((item) => {
      return item.addr.includes(value) || item.name.includes(value);
    });
  }
  function handleSearchEnter() {
    handleSearch(searchValue.value);
  }
</script>

<style scoped>
  .terminal-tabs :deep(.arco-tabs-content) {
    padding-top: 0;
    border: none;
  }
  .terminal-tabs :deep(.arco-tabs-nav-tab) {
    flex: none;
  }
  .terminal-tabs :deep(.arco-tabs-pane) {
    height: calc(100vh - 104px);
  }
  .terminal-tabs :deep(.arco-tabs-nav-extra .arco-tabs-nav-add-btn) {
    padding: 0 6px;
    margin-left: 6px;
  }
  .arco-tabs-nav-add-btn .arco-icon-hover::before {
    width: 24px;
    height: 24px;
  }
  .popover-head {
    position: relative;
    border-bottom: 1px solid var(--color-border-2);
    padding: 12px 16px;
  }
  .popover-close {
    position: absolute;
    right: 16px;
    top: 50%;
    transform: translateY(-50%);
    cursor: pointer;
    font-size: 12px;
    font-weight: normal;
    color: var(--color-text-1);
  }
  .popover-body {
    padding: 12px 16px;
  }
  .host-list {
    margin-top: 16px;
  }
  .host-item {
    padding: 6px 12px;
    height: 32px;
    line-height: 20px;
    border: 1px solid var(--color-border-2);
    font-size: 14px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    cursor: pointer;
  }
  .host-item:hover {
    background-color: var(--color-fill-2);
  }
  .host-addr {
    margin-right: 24px;
  }
</style>
