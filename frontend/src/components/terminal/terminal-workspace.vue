<template>
  <div class="terminal-workspace">
    <!-- 左侧主机列表 -->
    <host-sidebar
      :current-host-id="currentHostId"
      @host-select="handleHostSelect"
    />

    <!-- 右侧主体区域 -->
    <div class="workspace-main">
      <div v-if="!currentHostId" class="welcome-state">
        <a-empty :description="$t('components.terminal.workspace.selectHost')">
          <template #image>
            <icon-desktop />
          </template>
        </a-empty>
      </div>

      <template v-else>
        <!-- 终端标签页 -->
        <a-tabs
          v-model:active-key="activeKey"
          class="terminal-tabs"
          type="card-gutter"
          auto-switch
          lazy-load
        >
          <template #extra>
            <session-add-popover
              v-model:visible="popoverVisible"
              :current-host-id="currentHostId"
              @add-session="handleAddSession"
            />
          </template>

          <!-- 终端标签页内容 -->
          <a-tab-pane v-for="item of terms" :key="item.key">
            <terminal
              :ref="(el) => setTerminalRef(el, item)"
              :host-id="item.hostId"
              type="session"
              path="terminals/{host}/start"
              send-heartbeat
              @wsopen="() => handleWsOpen(item)"
              @session="(data) => handleSessionName(item, data)"
            />
            <template #title>
              <terminal-tab-title
                :item="item"
                @action="handleTabAction"
                @rename="handleTabRename"
              />
            </template>
          </a-tab-pane>
        </a-tabs>

        <!-- 加载状态 -->
        <loading-state v-if="terms.length === 0 && isCreatingSession" />

        <!-- 空状态 -->
        <empty-state
          v-else-if="terms.length === 0"
          @create-session="handleCreateFirstSession"
        />
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick, onMounted, onUnmounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useHostStore } from '@/store';
  import { useLogger } from '@/composables/use-logger';
  import { HostEntity } from '@/entity/Host';
  import {
    detachTerminalSessionApi,
    quitTerminalSessionApi,
    renameTerminalSessionApi,
  } from '@/api/terminal';
  import HostSidebar from './host-sidebar.vue';
  import Terminal from './terminal.vue';
  import SessionAddPopover from './components/session-add-popover.vue';
  import TerminalTabTitle from './components/terminal-tab-title.vue';
  import LoadingState from './components/loading-state.vue';
  import EmptyState from './components/empty-state.vue';
  import { MsgType } from './type';
  import {
    useTerminalTabs,
    type TermSessionItem,
  } from './composables/use-terminal-tabs';
  import { useTerminalSessions } from './composables/use-terminal-sessions';

  const { t } = useI18n();
  const hostStore = useHostStore();
  const { logError, logWarn } = useLogger('TerminalWorkspace');

  // 使用组合式函数
  const {
    terms,
    activeKey,
    addItem,
    removeItem,
    focus,
    setTerminalRef,
    clearAll,
  } = useTerminalTabs();

  const { createFirstSession } = useTerminalSessions();

  const currentHostId = ref<number>();
  const popoverVisible = ref(false);
  const isCreatingSession = ref(false);

  // 计算属性
  const currentHost = computed(() =>
    hostStore.items.find((item) => item.id === currentHostId.value)
  );

  // 处理标签页操作
  async function handleTabAction(
    item: TermSessionItem,
    action: 'rename' | 'quit' | 'detach'
  ): Promise<void> {
    if (!item.sessionId && (action === 'quit' || action === 'detach')) {
      Message.warning(t('components.terminal.session.noSessionId'));
      return;
    }

    try {
      if (action === 'rename') {
        // 开始重命名模式
        const index = terms.value.findIndex(
          (term: TermSessionItem) => term.key === item.key
        );
        if (index !== -1) {
          terms.value[index].isRenaming = true;
        }
      } else if (action === 'quit') {
        removeItem(item.key);
        if (item.sessionId) {
          await quitTerminalSessionApi(item.hostId, {
            session: item.sessionId,
          });
        }
      } else if (action === 'detach') {
        removeItem(item.key);
        if (item.sessionId) {
          await detachTerminalSessionApi(item.hostId, {
            session: item.sessionId,
          });
        }
      }
    } catch (error) {
      logError(`Failed to ${action} terminal session:`, error);
      const errorMessages = {
        quit: 'components.terminal.session.quitFailed',
        detach: 'components.terminal.session.detachFailed',
        rename: 'components.terminal.session.renameFailed',
      } as const;
      Message.error(t(errorMessages[action]));
    }
  }

  // 处理标签页重命名
  async function handleTabRename(updatedItem: TermSessionItem): Promise<void> {
    const index = terms.value.findIndex(
      (term: TermSessionItem) => term.key === updatedItem.key
    );

    if (index === -1) {
      logWarn('Terminal item not found for rename:', updatedItem.key);
      return;
    }

    const originalItem = terms.value[index];

    // 如果没有真正改变名称，或者没有sessionId，则只更新前端状态
    if (!updatedItem.sessionId || updatedItem.title === originalItem.title) {
      terms.value[index] = { ...originalItem, isRenaming: false };
      return;
    }

    // 先退出重命名模式，显示原来的标题，等API完成后再更新
    terms.value[index] = { ...originalItem, isRenaming: false };

    try {
      // 调用后端API重命名session
      await renameTerminalSessionApi(updatedItem.hostId, {
        session: updatedItem.sessionId,
        data: updatedItem.title,
      });

      // 成功后更新为新标题
      terms.value[index] = {
        ...originalItem,
        title: updatedItem.title,
        isRenaming: false,
      };
      Message.success(t('components.terminal.session.renameSuccess'));

      logWarn(
        `Successfully renamed session ${updatedItem.sessionId} to "${updatedItem.title}"`
      );
    } catch (error) {
      logError('Failed to rename terminal session:', error);
      Message.error(t('components.terminal.session.renameFailed'));

      // 失败时保持原始状态（已经设置过了，不需要再次设置）
    }
  }

  // 添加会话
  function handleAddSession(sessionData: {
    type: 'attach' | 'start';
    sessionId?: string;
    sessionName?: string;
  }): void {
    if (!currentHost.value) {
      Message.error(t('components.terminal.workspace.hostNotFound'));
      return;
    }

    addItem({
      type: sessionData.type,
      hostId: currentHost.value.id,
      hostName: currentHost.value.name,
      sessionId: sessionData.sessionId,
      sessionName: sessionData.sessionName,
    });

    popoverVisible.value = false;
  }

  // 处理WebSocket连接打开
  function handleWsOpen(item: TermSessionItem): void {
    if (!item.termRef) {
      logWarn('Terminal reference not available');
      return;
    }

    try {
      if (item.type === 'start') {
        item.termRef.sendWsMsg({
          type: MsgType.Start,
          session: item.sessionId || '',
          data: item.sessionName || '', // 传递用户自定义的会话名称，空则让后端自动生成
        });
      } else {
        item.termRef.sendWsMsg({
          type: MsgType.Attach,
          session: item.sessionId,
        });
      }
    } catch (error) {
      logError('Failed to send WebSocket message:', error);
      Message.error(t('components.terminal.session.connectionFailed'));
    }
  }

  // 接收服务器返回的会话名称
  function handleSessionName(
    item: TermSessionItem,
    data: {
      sessionId: string;
      sessionName: string;
    }
  ): void {
    if (!data.sessionId || !data.sessionName) {
      logWarn('Invalid session data received:', data);
      return;
    }

    item.title = data.sessionName;
    item.sessionId = data.sessionId;
    item.sessionName = data.sessionName;
  }

  // 创建第一个会话
  async function handleCreateFirstSession(): Promise<void> {
    if (!currentHost.value) return;

    isCreatingSession.value = true;
    try {
      const sessionData = await createFirstSession(currentHost.value);
      addItem({
        type: sessionData.type,
        hostId: currentHost.value.id,
        hostName: currentHost.value.name,
        sessionId: sessionData.sessionId,
      });
    } catch (error) {
      logError('Failed to create first session:', error);
      // 降级处理：创建新会话
      addItem({
        type: 'start',
        hostId: currentHost.value.id,
        hostName: currentHost.value.name,
      });
      Message.warning(t('components.terminal.session.fallbackToNewSession'));
    } finally {
      isCreatingSession.value = false;
    }
  }

  // 处理主机选择
  async function handleHostSelect(host: HostEntity): Promise<void> {
    try {
      // 清理当前所有终端
      clearAll();

      // 切换主机
      currentHostId.value = host.id;
      hostStore.setCurrentId(host.id);

      // 创建默认会话
      await handleCreateFirstSession();
    } catch (error) {
      logError('Failed to select host:', error);
      Message.error(t('components.terminal.workspace.hostSelectFailed'));
    }
  }

  // 重新初始化工作区（用于弹窗重新打开时）
  const reinitialize = async (): Promise<void> => {
    await nextTick();
    if (currentHostId.value && terms.value.length === 0) {
      await handleCreateFirstSession();
    }
  };

  // 监听活跃标签页变化
  const stopWatchingActiveKey = watch(activeKey, (val) => {
    if (val) {
      nextTick(() => {
        focus();
      });
    }
  });

  // 监听hostStore的当前主机变化
  const stopWatchingHostStore = watch(
    () => hostStore.currentId,
    async (newHostId) => {
      if (newHostId && newHostId !== currentHostId.value) {
        const host = hostStore.items.find((h) => h.id === newHostId);
        if (host) {
          await handleHostSelect(host);
        }
      }
    }
  );

  // 初始化时选择当前主机
  if (hostStore.currentId) {
    currentHostId.value = hostStore.currentId;
  }

  // 组件挂载时初始化
  onMounted(async () => {
    await nextTick();
    if (currentHostId.value && terms.value.length === 0) {
      await handleCreateFirstSession();
    }
  });

  // 组件卸载时清理
  onUnmounted(() => {
    stopWatchingActiveKey();
    stopWatchingHostStore();
    clearAll();
  });

  defineExpose({
    currentHostId,
    addItem,
    focus,
    reinitialize,
  });
</script>

<style scoped>
  .terminal-workspace {
    display: flex;
    height: 100%;
    background: var(--color-bg-1);
  }

  .workspace-main {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .welcome-state {
    display: flex;
    flex: 1;
    align-items: center;
    justify-content: center;
    padding: 32px;
    background: var(--color-bg-1);
  }

  .terminal-tabs {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .terminal-tabs :deep(.arco-tabs-content) {
    flex: 1;
    height: 100%;
    padding-top: 0;
    border: none;
  }

  .terminal-tabs :deep(.arco-tabs-nav-tab) {
    flex: none;
  }

  .terminal-tabs :deep(.arco-tabs-pane) {
    height: calc(100vh - var(--header-height, 140px));
  }

  .terminal-tabs :deep(.arco-tabs-tab) {
    padding-right: 6px;
  }
</style>
