<template>
  <a-tabs
    v-model:active-key="activeKey"
    class="terminal-tabs"
    type="card-gutter"
    auto-switch
    lazy-load
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
              {{ $t('components.terminal.session.addTitle') }}
            </div>
            <span class="arco-icon-hover popover-close" @click="handlePopClose">
              <icon-close />
            </span>
          </div>
          <div class="popover-body">
            <a-form :model="formState">
              <a-form-item
                field="host_id"
                :label="$t('components.terminal.session.host')"
              >
                <a-select
                  v-model="formState.host_id"
                  :placeholder="
                    $t('components.terminal.session.hostPlaceHolder')
                  "
                  :options="hostOptions"
                  allow-clear
                  allow-create
                  allow-search
                />
              </a-form-item>
              <a-form-item
                field="type"
                :label="$t('components.terminal.session.type')"
              >
                <a-radio-group v-model="formState.type" type="button">
                  <a-radio value="start">{{
                    $t('components.terminal.session.start')
                  }}</a-radio>
                  <a-radio value="attach">{{
                    $t('components.terminal.session.attach')
                  }}</a-radio>
                </a-radio-group>
              </a-form-item>
              <a-form-item
                field="session"
                :label="$t('components.terminal.session.session')"
              >
                <a-select
                  v-if="formState.type === 'attach'"
                  v-model="formState.session"
                  :placeholder="
                    $t('components.terminal.session.attachSession.placeholder')
                  "
                  :loading="sessionLoading"
                  :options="sessionOptions"
                  allow-clear
                  allow-create
                  allow-search
                />
                <a-input
                  v-else
                  v-model="formState.session_name"
                  :placeholder="
                    $t('components.terminal.session.startSession.placeholder')
                  "
                />
              </a-form-item>
              <a-form-item>
                <a-button type="primary" @click="handleAddSession">
                  {{ $t('components.terminal.session.add') }}
                </a-button>
              </a-form-item>
            </a-form>
          </div>
        </template>
      </a-popover>
    </template>
    <a-tab-pane v-for="item of terms" :key="item.key" :title="item.title">
      <terminal
        ref="item.termRef"
        :hostId="item.hostId"
        :path="item.path"
        @session="handleServerName(item as any, $event)"
      />
      <template #title>
        {{ item.title }}
        <a-dropdown
          position="bottom"
          @select="handleClose(item as any, $event as any)"
        >
          <span class="arco-icon-hover arco-tabs-tab-close-btn">
            <a-icon-close />
          </span>
          <template #content>
            <a-doption value="quit">{{
              $t('components.terminal.session.quit')
            }}</a-doption>
            <a-doption value="detach">{{
              $t('components.terminal.session.detach')
            }}</a-doption>
          </template>
        </a-dropdown>
      </template>
    </a-tab-pane>
  </a-tabs>
</template>

<script setup lang="ts">
  import { reactive, ref, Ref, watch } from 'vue';
  import { useHostStore } from '@/store';
  import { HostEntity } from '@/entity/Host';
  import {
    detachTerminalSessionApi,
    getTerminalSessionsApi,
    quitTerminalSessionApi,
  } from '@/api/terminal';
  import { serializeQueryParams } from '@/utils';
  import Terminal from './terminal.vue';

  type TerminalInstance = InstanceType<typeof Terminal> | undefined;

  interface TermSessionItem {
    key: string;
    type: 'attach' | 'start';
    hostId: number;
    title: string;
    termRef: Ref<TerminalInstance>;
    session: string;
    path: string;
  }

  const activeKey = ref<string>();
  const hostStore = useHostStore();
  const terms = ref<TermSessionItem[]>([]);

  function addSession(options: {
    type: 'attach' | 'start';
    host: HostEntity;
    session?: string;
  }) {
    const termRef: any = ref<TerminalInstance>();
    const key = Math.random().toString(36).slice(2);
    terms.value.push({
      key,
      type: options.type,
      hostId: options.host.id,
      title: options.host.name + (options.session ? '-' + options.session : ''),
      path:
        `terminals/{host}/start?` +
        serializeQueryParams({
          type: options.type,
          host: options.host.id,
          session: options.session || '',
        }),
      termRef,
      session: options.session || '',
    });
    activeKey.value = key;
  }

  function removeSession(key: string | number) {
    const index = terms.value.findIndex((item) => item.key === key);
    if (index !== -1) {
      terms.value.splice(index, 1);
    }
    if (key === activeKey.value) {
      activeKey.value =
        terms.value[index]?.key ||
        terms.value[index - 1]?.key ||
        terms.value[0]?.key;
    }
  }

  function handleClose(item: TermSessionItem, action: 'quit' | 'detach') {
    if (action === 'quit') {
      removeSession(item.key);
      quitTerminalSessionApi(item.hostId, {
        session: item.session,
      });
    } else if (action === 'detach') {
      removeSession(item.key);
      detachTerminalSessionApi(item.hostId, {
        session: item.session,
      });
    }
  }

  const popoverVisible = ref(false);
  function handlePopClose() {
    popoverVisible.value = false;
  }

  const formState = reactive({
    host_id: hostStore.current?.id || undefined,
    type: 'start' as 'attach' | 'start',
    session: '',
    session_name: '',
  });
  const hostOptions = ref(
    hostStore.items.map((item) => ({
      label: item.addr + '　　' + item.name,
      value: item.id,
    }))
  );
  const sessionOptions = ref<{ label: string; value: string }[]>([]);
  const sessionLoading = ref(false);
  async function loadSessionOptions(hostId: number) {
    sessionLoading.value = true;
    try {
      const res = await getTerminalSessionsApi(hostId);
      sessionOptions.value = (res.items || []).map((item: string) => ({
        label: item,
        value: item,
      }));
    } finally {
      sessionLoading.value = false;
    }
  }
  function handleAddSession() {
    if (!formState.host_id) {
      return;
    }
    if (formState.type === 'attach' && !formState.session) {
      return;
    }
    addSession({
      type: formState.type,
      host: hostStore.items.find((item) => item.id === formState.host_id)!,
      session:
        formState.type === 'attach'
          ? formState.session
          : formState.session_name,
    });
    popoverVisible.value = false;
  }

  // receive server name from terminal component
  function handleServerName(item: TermSessionItem, session: string) {
    item.title = item.hostId + '-' + session;
    item.session = session;
  }

  watch(popoverVisible, (val) => {
    if (val) {
      formState.host_id = hostStore.current?.id || undefined;
      formState.type = 'start';
      formState.session = '';
      formState.session_name = '';
    }
  });

  watch(
    () => [formState.host_id, formState.type],
    () => {
      if (formState.type === 'attach' && formState.host_id) {
        loadSessionOptions(formState.host_id);
      }
    }
  );

  defineExpose({
    addSession,
  });
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
    margin-left: 6px;
    padding: 0 6px;
  }

  .arco-tabs-nav-add-btn .arco-icon-hover::before {
    width: 24px;
    height: 24px;
  }

  .popover-head {
    position: relative;
    padding: 12px 16px;
    border-bottom: 1px solid var(--color-border-2);
  }

  .popover-close {
    position: absolute;
    top: 50%;
    right: 16px;
    color: var(--color-text-1);
    font-weight: normal;
    font-size: 12px;
    transform: translateY(-50%);
    cursor: pointer;
  }

  .popover-body {
    padding: 12px 16px;
  }

  .host-list {
    margin-top: 16px;
  }

  .host-item {
    height: 32px;
    padding: 6px 12px;
    overflow: hidden;
    font-size: 14px;
    line-height: 20px;
    white-space: nowrap;
    text-overflow: ellipsis;
    border: 1px solid var(--color-border-2);
    cursor: pointer;
  }

  .host-item:hover {
    background-color: var(--color-fill-2);
  }

  .host-addr {
    margin-right: 24px;
  }
</style>
