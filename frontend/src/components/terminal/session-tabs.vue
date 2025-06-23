<template>
  <div class="session-tabs">
    <div class="tabs-header">
      <div class="tabs-nav">
        <div
          v-for="session in sessions"
          :key="session.session"
          class="tab-item"
          :class="{ active: session.session === currentSessionId }"
          @click="handleSessionSelect(session)"
        >
          <span
            class="tab-status"
            :class="getSessionStatusClass(session)"
          ></span>
          <span class="tab-name" :title="session.name">{{ session.name }}</span>
          <a-dropdown
            position="bottom"
            @select="handleSessionAction(session, $event as any)"
          >
            <span class="tab-close" @click.stop>
              <icon-more />
            </span>
            <template #content>
              <a-doption :value="SESSION_ACTIONS.DETACH">
                {{ $t('components.terminal.session.detach') }}
              </a-doption>
              <a-doption :value="SESSION_ACTIONS.QUIT">
                {{ $t('components.terminal.session.quit') }}
              </a-doption>
              <a-doption :value="SESSION_ACTIONS.RENAME">
                {{ $t('components.terminal.session.rename') }}
              </a-doption>
            </template>
          </a-dropdown>
        </div>

        <div class="tab-add" @click="showCreateModal = true">
          <icon-plus />
          <span>{{ $t('components.terminal.session.new') }}</span>
        </div>
      </div>

      <div class="tabs-actions">
        <a-button
          type="text"
          size="small"
          :loading="loading"
          @click="debouncedRefreshSessions"
        >
          <template #icon>
            <icon-refresh />
          </template>
        </a-button>
        <a-button type="text" size="small" @click="handlePruneSessions">
          <template #icon>
            <icon-delete />
          </template>
          {{ $t('components.terminal.session.prune') }}
        </a-button>
      </div>
    </div>

    <div v-if="sessions.length === 0 && !loading" class="empty-sessions">
      <a-empty :description="$t('components.terminal.workspace.noSessions')">
        <template #image>
          <icon-code-square />
        </template>
        <a-button type="primary" @click="showCreateModal = true">
          {{ $t('components.terminal.session.createFirst') }}
        </a-button>
      </a-empty>
    </div>

    <!-- 创建会话弹窗 -->
    <a-modal
      v-model:visible="showCreateModal"
      :title="$t('components.terminal.session.create')"
      @ok="handleCreateSession"
      @cancel="resetCreateForm"
    >
      <a-form :model="createForm" layout="vertical">
        <a-form-item
          field="name"
          :label="$t('components.terminal.session.name')"
          :rules="nameValidationRules"
        >
          <a-input
            v-model="createForm.name"
            :placeholder="$t('components.terminal.session.namePlaceholder')"
          />
        </a-form-item>
        <a-form-item
          field="type"
          :label="$t('components.terminal.session.type')"
        >
          <a-select
            v-model="createForm.type"
            :placeholder="$t('components.terminal.session.typePlaceholder')"
          >
            <a-option :value="SESSION_TYPES.SCREEN">Screen</a-option>
            <a-option :value="SESSION_TYPES.TMUX">Tmux</a-option>
            <a-option :value="SESSION_TYPES.BASH">Bash</a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 重命名弹窗 -->
    <a-modal
      v-model:visible="showRenameModal"
      :title="$t('components.terminal.session.rename')"
      @ok="handleRenameSession"
      @cancel="resetRenameForm"
    >
      <a-form :model="renameForm" layout="vertical">
        <a-form-item
          field="name"
          :label="$t('components.terminal.session.newName')"
          :rules="nameValidationRules"
        >
          <a-input
            v-model="renameForm.name"
            :placeholder="$t('components.terminal.session.namePlaceholder')"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { debounce } from 'lodash';
  import { useLogger } from '@/composables/use-logger';
  import {
    getTerminalSessionsApi,
    detachTerminalSessionApi,
    quitTerminalSessionApi,
    renameTerminalSessionApi,
    pruneTerminalSessionApi,
  } from '@/api/terminal';

  // 常量定义
  const SESSION_TYPES = {
    SCREEN: 'screen',
    TMUX: 'tmux',
    BASH: 'bash',
  } as const;

  const SESSION_ACTIONS = {
    DETACH: 'detach',
    QUIT: 'quit',
    RENAME: 'rename',
  } as const;

  const SESSION_STATUS = {
    ATTACHED: 'attached',
    DETACHED: 'detached',
  } as const;

  // 类型定义
  type SessionAction = (typeof SESSION_ACTIONS)[keyof typeof SESSION_ACTIONS];

  interface SessionInfo {
    session: string;
    name: string;
    status: string;
    time: string;
  }

  // Composables
  const { t } = useI18n();
  const { logError, logInfo } = useLogger('SessionTabs');

  const emit = defineEmits<{
    sessionSelect: [session: SessionInfo];
    sessionCreate: [sessionName: string, sessionType: string];
    sessionsChanged: [];
  }>();

  const props = defineProps<{
    hostId?: number;
    currentSessionId?: string;
  }>();

  const sessions = ref<SessionInfo[]>([]);
  const loading = ref(false);
  const showCreateModal = ref(false);
  const showRenameModal = ref(false);
  const currentRenameSession = ref<SessionInfo | null>(null);

  const createForm = reactive({
    name: '',
    type: SESSION_TYPES.SCREEN,
  });

  const renameForm = reactive({
    name: '',
  });

  // 表单验证规则
  const nameValidationRules = computed(() => [
    {
      required: true,
      message: t('components.terminal.session.nameRequired'),
    },
    {
      minLength: 1,
      message: t('components.terminal.session.nameMinLength'),
    },
    {
      maxLength: 50,
      message: t('components.terminal.session.nameMaxLength'),
    },
  ]);

  const refreshSessions = async () => {
    if (!props.hostId) {
      sessions.value = [];
      emit('sessionsChanged');
      return;
    }

    loading.value = true;
    try {
      const result = await getTerminalSessionsApi(props.hostId);
      sessions.value = result.items || [];
      logInfo(
        `Successfully loaded ${sessions.value.length} sessions for host ${props.hostId}`
      );
      emit('sessionsChanged');
    } catch (error) {
      logError('Failed to load sessions:', error);
      Message.error(t('components.terminal.session.loadFailed'));
      sessions.value = [];
      emit('sessionsChanged');
    } finally {
      loading.value = false;
    }
  };

  // 防抖的刷新函数
  const debouncedRefreshSessions = debounce(refreshSessions, 300);

  const resetCreateForm = () => {
    createForm.name = '';
    createForm.type = SESSION_TYPES.SCREEN;
  };

  const resetRenameForm = () => {
    renameForm.name = '';
    currentRenameSession.value = null;
  };

  const getSessionStatusClass = (session: SessionInfo) => {
    switch (session.status.toLowerCase()) {
      case SESSION_STATUS.ATTACHED:
        return 'status-attached';
      case SESSION_STATUS.DETACHED:
        return 'status-detached';
      default:
        return 'status-unknown';
    }
  };

  const handleSessionSelect = (session: SessionInfo) => {
    emit('sessionSelect', session);
  };

  const handleSessionAction = async (
    session: SessionInfo,
    action: SessionAction
  ) => {
    if (!props.hostId) return;

    try {
      switch (action) {
        case SESSION_ACTIONS.DETACH:
          await detachTerminalSessionApi(props.hostId, {
            session: session.session,
          });
          logInfo(
            `Successfully detached session: ${session.name} (${session.session})`
          );
          Message.success(t('components.terminal.session.detachSuccess'));
          await refreshSessions();
          break;
        case SESSION_ACTIONS.QUIT:
          await quitTerminalSessionApi(props.hostId, {
            session: session.session,
          });
          logInfo(
            `Successfully quit session: ${session.name} (${session.session})`
          );
          Message.success(t('components.terminal.session.quitSuccess'));
          await refreshSessions();
          break;
        case SESSION_ACTIONS.RENAME:
          currentRenameSession.value = session;
          renameForm.name = session.name;
          showRenameModal.value = true;
          logInfo(
            `Opening rename modal for session: ${session.name} (${session.session})`
          );
          break;
        default:
          break;
      }
    } catch (error) {
      logError(`Failed to ${action} session:`, error);
      const errorMessage =
        error instanceof Error ? error.message : 'Unknown error';
      Message.error(
        `${t('components.terminal.session.actionFailed')}: ${errorMessage}`
      );
    }
  };

  const handleCreateSession = () => {
    if (!createForm.name.trim()) {
      Message.error(t('components.terminal.session.nameRequired'));
      return;
    }

    logInfo(`Creating new session: ${createForm.name} (${createForm.type})`);
    emit('sessionCreate', createForm.name, createForm.type);
    showCreateModal.value = false;
    resetCreateForm();
  };

  const handleRenameSession = async () => {
    if (
      !props.hostId ||
      !currentRenameSession.value ||
      !renameForm.name.trim()
    ) {
      return;
    }

    try {
      await renameTerminalSessionApi(props.hostId, {
        session: currentRenameSession.value.session,
        data: renameForm.name,
      });
      logInfo(
        `Successfully renamed session from "${currentRenameSession.value.name}" to "${renameForm.name}"`
      );
      Message.success(t('components.terminal.session.renameSuccess'));
      showRenameModal.value = false;
      resetRenameForm();
      await refreshSessions();
    } catch (error) {
      logError('Failed to rename session:', error);
      Message.error(t('components.terminal.session.renameFailed'));
    }
  };

  const handlePruneSessions = async () => {
    if (!props.hostId) return;

    try {
      await pruneTerminalSessionApi(props.hostId);
      logInfo(`Successfully pruned sessions for host ${props.hostId}`);
      Message.success(t('components.terminal.session.pruneSuccess'));
      await refreshSessions();
    } catch (error) {
      logError('Failed to prune sessions:', error);
      Message.error(t('components.terminal.session.pruneFailed'));
    }
  };

  // 监听hostId变化，自动刷新会话列表
  watch(
    () => props.hostId,
    () => {
      debouncedRefreshSessions();
    },
    { immediate: true }
  );

  onMounted(() => {
    refreshSessions();
  });

  defineExpose({
    refreshSessions,
    sessions,
  });
</script>

<style scoped>
  .session-tabs {
    background: #fafafa;
    border-bottom: 1px solid #e8e8e8;
  }

  .tabs-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    min-height: 48px;
    padding: 12px 16px 12px;
  }

  .tabs-nav {
    display: flex;
    flex: 1;
    gap: 0;
    align-items: center;
    overflow-x: auto;
  }

  .tab-item {
    position: relative;
    display: flex;
    gap: 6px;
    align-items: center;
    min-width: 80px;
    max-width: 160px;
    padding: 6px 12px;
    margin-right: 4px;
    font-size: 13px;
    color: #722ed1;
    white-space: nowrap;
    cursor: pointer;
    background: #fff;
    border: 1px solid #d9d9d9;
    border-radius: 4px;
    transition: all 0.15s ease;
  }

  .tab-item:hover {
    background: #fafafa;
    border-color: #b37feb;
  }

  .tab-item.active {
    font-weight: 500;
    color: #722ed1;
    background: #fff;
    border-color: #722ed1;
    box-shadow: 0 2px 4px rgb(114 46 209 / 10%);
  }

  .tab-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: inherit;
  }

  .tab-status {
    flex-shrink: 0;
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  .status-attached {
    background: #52c41a;
  }

  .status-detached {
    background: #bfbfbf;
  }

  .status-unknown {
    background: #faad14;
  }

  .tab-close {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 16px;
    height: 16px;
    padding: 2px;
    color: #722ed1;
    border-radius: 3px;
    opacity: 0;
    transition: all 0.15s ease;
  }

  .tab-close:hover {
    background: #f0f0f0;
  }

  .tab-item:hover .tab-close {
    opacity: 0.7;
  }

  .tab-item.active .tab-close {
    opacity: 0.8;
  }

  .tab-close:hover {
    opacity: 1 !important;
  }

  .tab-add {
    display: flex;
    gap: 6px;
    align-items: center;
    padding: 6px 12px;
    margin-left: 4px;
    font-size: 12px;
    color: #722ed1;
    white-space: nowrap;
    cursor: pointer;
    background: #fff;
    border: 1px dashed #d9d9d9;
    border-radius: 4px;
    transition: all 0.15s ease;
  }

  .tab-add:hover {
    background: #fafafa;
    border-color: #722ed1;
    border-style: solid;
  }

  .tabs-actions {
    display: flex;
    gap: 6px;
    align-items: center;
    margin-left: 12px;
  }

  .tabs-actions .arco-btn {
    color: #722ed1;
    border-color: #d9d9d9;
  }

  .tabs-actions .arco-btn:hover {
    color: #722ed1;
    border-color: #722ed1;
  }

  .empty-sessions {
    padding: 32px;
    text-align: center;
  }
</style>
