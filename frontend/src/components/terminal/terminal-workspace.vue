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
        <!-- 终端容器 -->
        <div class="terminal-container">
          <!-- 右上角操作按钮 -->
          <div class="terminal-actions">
            <a-button
              type="text"
              size="small"
              :loading="isPruningSessions"
              @click="handlePruneSessions"
            >
              <template #icon>
                <icon-delete />
              </template>
              {{ $t('components.terminal.session.prune') }}
            </a-button>
          </div>

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
                v-if="canAddNewTab"
                v-model:visible="popoverVisible"
                :current-host-id="currentHostId"
                @add-session="handleAddSession"
              />
            </template>

            <!-- 终端标签页内容 - 只显示当前主机的会话标签 -->
            <a-tab-pane v-for="item of terms" :key="item.key">
              <template #title>
                <terminal-tab-title
                  :item="item"
                  @action="handleTabAction"
                  @rename="handleTabRename"
                />
              </template>
              <!-- 标签页内容为空，实际终端在下面渲染 -->
            </a-tab-pane>
          </a-tabs>

          <!-- 所有终端组件 - 保持在DOM中，通过CSS控制显示 -->
          <div class="terminal-content">
            <div
              v-for="item of allTerms"
              :key="item.key"
              class="terminal-wrapper"
              :class="{
                'terminal-active':
                  item.hostId === currentHostId && item.key === activeKey,
                'terminal-hidden':
                  item.hostId !== currentHostId || item.key !== activeKey,
              }"
            >
              <terminal
                :ref="(el) => setTerminalRef(el, item)"
                :host-id="item.hostId"
                type="session"
                path="terminals/{host}/start"
                send-heartbeat
                @wsopen="() => handleWsOpen(item)"
                @session="(data) => handleSessionName(item, data)"
              />
            </div>
          </div>

          <!-- 加载状态 - 绝对定位覆盖在终端内容区域上 -->
          <loading-state
            v-if="
              terms.length === 0 && (isCreatingSession || isRestoringSession)
            "
            class="terminal-loading-overlay"
            :text-key="
              isRestoringSession
                ? 'components.terminal.workspace.restoringSession'
                : 'components.terminal.workspace.creatingSession'
            "
          />

          <!-- 空状态 -->
          <empty-state
            v-else-if="terms.length === 0"
            @create-session="handleCreateFirstSession"
          />
        </div>
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
    pruneTerminalSessionApi,
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
    setCurrentHost,
    restoreHostSessions,
    saveCurrentState,
    switchToHost,
    allTerms,
    canAddTab,
    MAX_TABS,
  } = useTerminalTabs();

  const { createFirstSession, getAllHostSessions } = useTerminalSessions();

  const currentHostId = ref<number>();
  const popoverVisible = ref(false);
  const isCreatingSession = ref(false);
  const isPruningSessions = ref(false);
  const isRestoringSession = ref(false);

  // 计算属性
  const currentHost = computed(() =>
    hostStore.items.find((item) => item.id === currentHostId.value)
  );

  // 计算是否可以添加新标签页
  const canAddNewTab = computed(() => {
    return currentHostId.value ? canAddTab(currentHostId.value) : false;
  });

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
        // 保存移除会话后的状态到缓存
        saveCurrentState();
      } else if (action === 'detach') {
        removeItem(item.key);
        if (item.sessionId) {
          await detachTerminalSessionApi(item.hostId, {
            session: item.sessionId,
          });
        }
        // 保存移除会话后的状态到缓存
        saveCurrentState();
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

      // 成功后更新为新标题，但保持原始的sessionName不变
      terms.value[index] = {
        ...originalItem,
        title: updatedItem.title,
        isCustomTitle: true, // 标记为用户自定义标题
        isRenaming: false,
        // 注意：不修改sessionName，保持服务器端的原始会话名称
      };

      // 保存更新后的状态到缓存
      saveCurrentState();

      Message.success(t('components.terminal.session.renameSuccess'));
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

    try {
      addItem({
        type: sessionData.type,
        hostId: currentHost.value.id,
        hostName: currentHost.value.name,
        sessionId: sessionData.sessionId,
        sessionName: sessionData.sessionName,
      });

      popoverVisible.value = false;
    } catch (error) {
      logError('Failed to add session:', error);
      Message.error(
        error instanceof Error
          ? error.message
          : t('components.terminal.session.addFailed')
      );
    }
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
        // 对于attach类型，使用原始的服务器会话名称，而不是用户自定义的标题
        // 这确保服务器能找到正确的会话进行附加
        const originalSessionName =
          item.originalSessionName || item.sessionName || '';

        // 检查sessionId是否有效（只有在attach类型时才需要sessionId）
        if (!item.sessionId) {
          // 对于没有sessionId的attach类型，可能需要转换为start类型
          // 或者让服务器处理这种情况
        }

        item.termRef.sendWsMsg({
          type: MsgType.Attach,
          session: item.sessionId || '', // 允许空的sessionId，让服务器处理
          data: originalSessionName, // 传递原始的服务器会话名称，确保能正确附加
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

    // 如果当前标题不是默认的"连接中"状态，并且sessionId已经存在，
    // 说明这是一个已经存在的会话（可能是用户重命名过的），不应该被服务器返回的名称覆盖
    const isConnectingTitle =
      item.title === t('components.terminal.session.connecting');
    const hasExistingSession = Boolean(item.sessionId);

    // 检查sessionId是否发生了变化（可能表明attach失败，服务器创建了新session）
    if (
      hasExistingSession &&
      item.sessionId &&
      item.sessionId !== data.sessionId
    ) {
      logError(
        `WARNING: SessionId mismatch! Expected: ${item.sessionId}, Got: ${data.sessionId}. This may indicate attach failed and server created a new session.`
      );
    }

    // 只有在以下情况才更新标题：
    // 1. 当前标题是"连接中"（新建会话）
    // 2. 或者没有自定义标题且当前会话ID为空（新建会话）
    if (isConnectingTitle || (!item.isCustomTitle && !hasExistingSession)) {
      item.title = data.sessionName;
      item.isCustomTitle = false; // 标记为非自定义标题
    } else {
      // 保持现有标题，忽略服务器名称
    }

    // 更新sessionId和sessionName
    item.sessionId = data.sessionId;

    // 保存原始的服务器会话名称，用于会话恢复
    if (!item.originalSessionName) {
      item.originalSessionName = data.sessionName;
    }
    item.sessionName = data.sessionName;

    // 保存更新后的会话信息到缓存
    saveCurrentState();
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
      // 检查是否是标签页数量限制错误
      if (error instanceof Error && error.message.includes('tabLimitReached')) {
        Message.error(error.message);
        return;
      }

      // 降级处理：创建新会话
      try {
        addItem({
          type: 'start',
          hostId: currentHost.value.id,
          hostName: currentHost.value.name,
        });
        Message.warning(t('components.terminal.session.fallbackToNewSession'));
      } catch (fallbackError) {
        logError('Failed to create fallback session:', fallbackError);
        Message.error(
          fallbackError instanceof Error
            ? fallbackError.message
            : t('components.terminal.session.addFailed')
        );
      }
    } finally {
      isCreatingSession.value = false;
    }
  }

  // 处理主机选择
  async function handleHostSelect(host: HostEntity): Promise<void> {
    try {
      // 如果是同一个主机，不需要切换
      if (currentHostId.value === host.id) {
        return;
      }

      // 立即切换UI显示，让用户感觉响应迅速
      switchToHost(host.id);
      currentHostId.value = host.id;
      hostStore.setCurrentId(host.id);
      setCurrentHost(host.id);

      // 检查新主机是否已有运行中的会话
      const existingSessions = terms.value;

      if (existingSessions.length > 0) {
        // 已有运行中的会话，直接显示（连接由子组件自行维护）
        return;
      }

      // 没有运行中的会话，异步处理会话恢复，不阻塞UI切换
      nextTick(async () => {
        isRestoringSession.value = true;
        try {
          const restored = await restoreHostSessions(
            host.id,
            host.name,
            getAllHostSessions
          );

          if (restored) {
            // 成功恢复会话，显示提示信息
            Message.success(
              t('components.terminal.session.sessionsRestored', {
                count: terms.value.length,
              })
            );

            // 依赖子组件的 wsopen 事件来发送 Start/Attach，无需额外延迟处理
            // Terminal 子组件在 WebSocket 打开时会触发 wsopen 事件，父组件监听后调用 handleWsOpen(item)
          } else {
            // 没有保存的会话，创建默认会话
            await handleCreateFirstSession();
          }
        } catch (error) {
          logError('Failed to restore sessions:', error);
          // 恢复失败时，尝试创建默认会话
          try {
            await handleCreateFirstSession();
          } catch (fallbackError) {
            logError('Failed to create fallback session:', fallbackError);
          }
        } finally {
          isRestoringSession.value = false;
        }
      });
    } catch (error) {
      logError('Failed to select host:', error);
      // 如果是标签页数量限制错误，显示具体错误信息
      if (error instanceof Error && error.message.includes('tabLimitReached')) {
        Message.error(error.message);
      } else {
        Message.error(t('components.terminal.workspace.hostSelectFailed'));
      }
    }
  }

  // 重新初始化工作区（用于弹窗重新打开时）
  const reinitialize = async (): Promise<void> => {
    await nextTick();
    if (currentHostId.value && terms.value.length === 0) {
      const currentHostItem = hostStore.items.find(
        (h) => h.id === currentHostId.value
      );
      if (currentHostItem) {
        // 尝试恢复保存的会话
        isRestoringSession.value = true; // 开始恢复会话
        try {
          const restored = await restoreHostSessions(
            currentHostId.value,
            currentHostItem.name,
            getAllHostSessions
          );
          if (!restored) {
            // 如果没有保存的会话，创建默认会话
            await handleCreateFirstSession();
          }
        } finally {
          isRestoringSession.value = false; // 恢复会话结束
        }
      }
    }
  };

  // 监听活跃标签页变化，确保终端正确显示和适配尺寸
  const stopWatchingActiveKey = watch(activeKey, (val) => {
    if (!val) return;

    nextTick(() => {
      focus();
      // 使用 requestAnimationFrame 确保终端完全显示后再次适配尺寸
      requestAnimationFrame(() => {
        const activeItem = terms.value.find((item) => item.key === val);
        activeItem?.termRef?.forceRefit?.();
      });
    });
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
      const currentHostItem = hostStore.items.find(
        (h) => h.id === currentHostId.value
      );
      if (currentHostItem) {
        setCurrentHost(currentHostId.value);
        // 尝试恢复保存的会话
        isRestoringSession.value = true; // 开始恢复会话
        try {
          const restored = await restoreHostSessions(
            currentHostId.value,
            currentHostItem.name,
            getAllHostSessions
          );
          if (!restored) {
            // 如果没有保存的会话，创建默认会话
            await handleCreateFirstSession();
          }
        } finally {
          isRestoringSession.value = false; // 恢复会话结束
        }
      }
    }
  });

  // 组件卸载时清理
  onUnmounted(() => {
    // 保存当前状态
    if (currentHostId.value && terms.value.length > 0) {
      saveCurrentState();
    }

    stopWatchingActiveKey();
    stopWatchingHostStore();

    // 只在真正关闭应用时才断开所有连接
    // 这里可以根据需要决定是否保持后台连接
    clearAll({ dispose: true });
  });

  // 处理终端会话清理
  async function handlePruneSessions(): Promise<void> {
    if (!currentHostId.value) {
      Message.warning(t('components.terminal.workspace.selectHost'));
      return;
    }

    isPruningSessions.value = true;
    try {
      await pruneTerminalSessionApi(currentHostId.value, 'screen');
      Message.success(t('components.terminal.session.pruneSuccess'));
      // 清理成功后，刷新当前的终端列表（如果需要的话）
      // 这里可以选择是否重新创建第一个会话，或者保持当前状态
    } catch (error) {
      logError('Failed to prune terminal sessions:', error);
      console.error('Prune sessions error:', error);
      Message.error(t('components.terminal.session.pruneFailed'));
    } finally {
      isPruningSessions.value = false;
    }
  }

  defineExpose({
    currentHostId,
    terms,
    activeKey,
    popoverVisible,
    isCreatingSession,
    isPruningSessions,
    isRestoringSession,
    currentHost,
    canAddNewTab,
    MAX_TABS,
    handleTabAction,
    handleTabRename,
    handleAddSession,
    handleWsOpen,
    handleSessionName,
    handleCreateFirstSession,
    handleHostSelect,
    reinitialize,
    handlePruneSessions,
  });
</script>

<style scoped>
  .terminal-workspace {
    display: flex;
    gap: 8px;
    height: 100%;
    padding: 0 8px 8px 0;
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
    height: calc(100vh - var(--header-height, 124px));
  }

  .terminal-tabs :deep(.arco-tabs-tab) {
    padding-right: 6px;
  }

  .terminal-tabs :deep(.arco-tabs-nav) {
    height: 32px;
    margin-bottom: 0;
  }

  .terminal-tabs :deep(.arco-tabs-tab) {
    min-width: 84px;
    height: 32px;
    padding: 5px 16px;
    margin-right: 0;
  }

  .terminal-tabs :deep(.arco-tabs-tab-title) {
    font-family: Roboto, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto,
      sans-serif !important;
    font-weight: 400 !important;
    line-height: 22px !important;
    color: var(--color-text-2) !important;

    @apply text-base !important;
  }

  .terminal-tabs
    :deep(.arco-tabs-tab.arco-tabs-tab-active .arco-tabs-tab-title) {
    color: var(--color-primary-6) !important; /* 使用系统primary color变量 */
  }

  .terminal-tabs :deep(.arco-tabs-tab:hover .arco-tabs-tab-title) {
    color: var(--color-primary-5) !important; /* 悬停状态使用primary-5 */
  }

  .terminal-tabs :deep(.arco-tabs-tab-active .arco-tabs-tab-title) {
    color: var(--color-primary-6) !important;
  }

  .terminal-tabs :deep(.arco-tabs-tab:hover .arco-tabs-tab-title) {
    color: var(--color-primary-5) !important;
  }

  .terminal-tabs :deep(.arco-tabs-tab):not(:last-child) {
    margin-right: 2px;
  }

  .terminal-tabs :deep(.arco-tabs-nav-extra) {
    margin-left: 8px;
  }

  .terminal-container {
    position: relative;
    display: flex;
    flex: 1;
    flex-direction: column;
    height: 100%;
    background: var(--color-bg-1);
  }

  .terminal-actions {
    position: absolute;
    top: 0;
    right: 0;
    z-index: 10;
  }

  .terminal-actions .arco-btn {
    color: var(--color-text-2);
    background: rgb(255 255 255 / 80%);
    border: 1px solid var(--color-border-2);
    backdrop-filter: blur(4px);
  }

  .terminal-actions .arco-btn:hover {
    color: var(--color-text-1);
    background: rgb(255 255 255 / 95%);
    border-color: var(--color-border-3);
  }

  /* 隐藏标签页内容，使用独立的终端内容区域 */
  .terminal-tabs :deep(.arco-tabs-content),
  .terminal-tabs :deep(.arco-tabs-content-list),
  .terminal-tabs :deep(.arco-tabs-pane) {
    height: 0;
    padding: 0;
    overflow: hidden;
  }

  .terminal-content {
    position: relative;
    width: 100%;
    height: calc(100% - 40px);
  }

  .terminal-wrapper {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }

  .terminal-active {
    z-index: 1;
    visibility: visible;
    pointer-events: auto;
    opacity: 1;
  }

  .terminal-hidden {
    z-index: 0;
    visibility: hidden;
    pointer-events: none;
    opacity: 0;
  }

  .terminal-loading-overlay {
    position: absolute;
    top: 40px;
    left: 0;
    z-index: 10;
    width: 100%;
    height: calc(100% - 40px);
    background: var(--color-bg-1);
  }

  .terminal-loading-overlay :deep(.loading-state) {
    min-height: 100%;
    padding: 0;
  }
</style>
