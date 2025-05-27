<template>
  <a-popover
    v-model:popup-visible="visible"
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
          {{ $t('components.terminal.workspace.addSession') }}
        </div>
        <span class="arco-icon-hover popover-close" @click="handleClose">
          <icon-close />
        </span>
      </div>
      <div class="popover-body">
        <a-form :model="formState" @submit="handleAddSession">
          <a-form-item
            field="type"
            :label="$t('components.terminal.session.type')"
          >
            <a-radio-group v-model="formState.type" type="button">
              <a-radio value="start">
                {{ $t('components.terminal.session.start') }}
              </a-radio>
              <a-radio value="attach">
                {{ $t('components.terminal.session.attach') }}
              </a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item
            field="session"
            :label="$t('components.terminal.session.session')"
          >
            <a-select
              v-if="formState.type === 'attach'"
              v-model="formState.sessionId"
              :placeholder="
                $t('components.terminal.session.attachSession.placeholder')
              "
              :loading="sessionLoading"
              :options="sessionOptions"
              allow-clear
              allow-search
            />
            <a-input
              v-else
              v-model="formState.sessionName"
              :placeholder="
                $t('components.terminal.session.startSession.placeholder')
              "
              :max-length="50"
              show-word-limit
            />
          </a-form-item>
          <a-form-item>
            <a-button
              type="primary"
              :loading="isSubmitting"
              @click="handleAddSession"
            >
              {{ $t('components.terminal.session.add') }}
            </a-button>
          </a-form-item>
        </a-form>
      </div>
    </template>
  </a-popover>
</template>

<script setup lang="ts">
  import { ref, reactive, watch, computed, onUnmounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getTerminalSessionsApi } from '@/api/terminal';
  import { useLogger } from '@/hooks/use-logger';

  interface SessionData {
    type: 'attach' | 'start';
    sessionId?: string;
    sessionName?: string;
  }

  interface Props {
    visible: boolean;
    currentHostId?: number;
  }

  interface Emits {
    (e: 'update:visible', value: boolean): void;
    (e: 'addSession', data: SessionData): void;
  }

  interface SessionOption {
    label: string;
    value: string;
    status: string;
    disabled?: boolean;
  }

  interface SessionItem {
    name: string;
    status: string;
    session: string;
  }

  interface SessionResponse {
    items?: SessionItem[];
  }

  const props = defineProps<Props>();
  const emit = defineEmits<Emits>();

  const { t } = useI18n();
  const { logWarn, logError } = useLogger('SessionAddPopover');

  const visible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value),
  });

  const formState = reactive({
    type: 'start' as 'attach' | 'start',
    sessionId: '',
    sessionName: '',
  });

  const sessionOptions = ref<SessionOption[]>([]);
  const sessionLoading = ref(false);
  const isSubmitting = ref(false);

  // 验证会话名称
  function validateSessionName(name: string): boolean {
    if (!name.trim()) return false;
    // 检查是否包含特殊字符 (避免使用控制字符)
    const invalidChars = /[<>:"/\\|?*]/;
    const hasControlChars = name.split('').some((char) => {
      const code = char.charCodeAt(0);
      return code >= 0 && code <= 31;
    });
    return !invalidChars.test(name) && !hasControlChars;
  }

  // 生成默认会话名称
  function generateDefaultSessionName(): string {
    const timestamp = Date.now().toString().slice(-6);
    return `session-${timestamp}`;
  }

  // 加载会话选项
  async function loadSessionOptions(hostId: number): Promise<void> {
    if (!hostId || hostId <= 0) {
      logWarn('Invalid host ID provided:', hostId);
      return;
    }

    sessionLoading.value = true;
    try {
      const res: SessionResponse = await getTerminalSessionsApi(hostId);

      if (!res.items || !Array.isArray(res.items)) {
        logWarn('Invalid session response format:', res);
        sessionOptions.value = [];
        return;
      }

      // 分别处理 detached 和 attached 会话
      const detachedSessions = res.items
        .filter((item: SessionItem) => {
          return (
            item &&
            typeof item.status === 'string' &&
            item.status.toLowerCase() === 'detached' &&
            item.name &&
            item.session
          );
        })
        .map((item: SessionItem) => ({
          label: `${item.name} (${item.status})`,
          value: item.session,
          status: item.status,
        }));

      const attachedSessions = res.items
        .filter((item: SessionItem) => {
          return (
            item &&
            typeof item.status === 'string' &&
            item.status.toLowerCase() === 'attached' &&
            item.name &&
            item.session
          );
        })
        .map((item: SessionItem) => ({
          label: `${item.name} (${item.status}) - ${t(
            'components.terminal.session.alreadyAttached'
          )}`,
          value: item.session,
          status: item.status,
          disabled: true,
        }));

      // 合并选项，detached 在前，attached 在后
      sessionOptions.value = [...detachedSessions, ...attachedSessions];
    } catch (error) {
      logError('Failed to load terminal sessions:', error);
      Message.error(t('components.terminal.session.loadFailed'));
      sessionOptions.value = [];
    } finally {
      sessionLoading.value = false;
    }
  }

  // 关闭弹窗
  function handleClose(): void {
    visible.value = false;
  }

  // 重置表单
  function resetForm(): void {
    formState.type = 'start';
    formState.sessionId = '';
    formState.sessionName = '';
  }

  // 添加会话
  async function handleAddSession(): Promise<void> {
    if (isSubmitting.value) return;

    // 验证主机ID
    if (!props.currentHostId || props.currentHostId <= 0) {
      Message.error(t('components.terminal.workspace.selectHost'));
      return;
    }

    // 验证表单数据
    if (formState.type === 'attach') {
      if (!formState.sessionId?.trim()) {
        Message.error(t('components.terminal.session.selectSession'));
        return;
      }
    } else if (formState.type === 'start') {
      // 对于start类型，如果没有输入会话名称，生成默认名称
      if (!formState.sessionName.trim()) {
        formState.sessionName = generateDefaultSessionName();
      } else if (!validateSessionName(formState.sessionName)) {
        Message.error(t('components.terminal.session.invalidSessionName'));
        return;
      }
    }

    isSubmitting.value = true;
    try {
      const sessionData: SessionData = {
        type: formState.type,
        sessionId:
          formState.type === 'attach' ? formState.sessionId : undefined,
        sessionName:
          formState.type === 'start' ? formState.sessionName : undefined,
      };

      emit('addSession', sessionData);

      // 重置表单并关闭弹窗
      resetForm();
      visible.value = false;

      Message.success(t('components.terminal.session.addSuccess'));
    } catch (error) {
      logError('Failed to add session:', error);
      Message.error(t('components.terminal.session.addFailed'));
    } finally {
      isSubmitting.value = false;
    }
  }

  // 监听弹窗显示状态
  const stopWatchingVisible = watch(visible, (val: boolean) => {
    if (val) {
      resetForm();

      // 加载可附加的会话
      if (props.currentHostId && props.currentHostId > 0) {
        loadSessionOptions(props.currentHostId);
      }
    }
  });

  // 监听表单类型变化
  const stopWatchingType = watch(
    () => formState.type,
    (newType: 'attach' | 'start') => {
      if (
        newType === 'attach' &&
        props.currentHostId &&
        props.currentHostId > 0
      ) {
        loadSessionOptions(props.currentHostId);
      }
      // 清空相关字段
      formState.sessionId = '';
      formState.sessionName = '';
    }
  );

  // 监听当前主机ID变化
  const stopWatchingHostId = watch(
    () => props.currentHostId,
    (newHostId: number | undefined) => {
      if (
        newHostId &&
        newHostId > 0 &&
        formState.type === 'attach' &&
        visible.value
      ) {
        loadSessionOptions(newHostId);
      }
    }
  );

  // 组件卸载时清理监听器
  onUnmounted(() => {
    stopWatchingVisible();
    stopWatchingType();
    stopWatchingHostId();
  });
</script>

<style scoped>
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
    font-size: 12px;
    font-weight: normal;
    color: var(--color-text-1);
    cursor: pointer;
    transform: translateY(-50%);
  }

  .popover-body {
    padding: 12px 16px;
  }
</style>
