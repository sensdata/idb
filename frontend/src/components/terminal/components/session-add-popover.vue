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
    <a-button type="outline" size="mini" class="add-session-btn">
      <template #icon>
        <icon-plus />
      </template>
      {{ $t('components.terminal.workspace.addSession') }}
    </a-button>
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
              ref="attachSelectRef"
              v-model="formState.sessionId"
              :placeholder="
                $t('components.terminal.session.attachSession.placeholder')
              "
              :loading="sessionLoading"
              :options="sessionOptions"
              allow-clear
              allow-search
            />
            <div v-else>
              <a-input
                v-model="formState.sessionName"
                :placeholder="
                  $t('components.terminal.session.startSession.placeholder')
                "
                :max-length="50"
                show-word-limit
              />
              <div class="start-session-hint">
                {{
                  $t('components.terminal.session.startSession.autoGenerate')
                }}
              </div>
            </div>
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
  import { ref, reactive, watch, computed, onUnmounted, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getTerminalSessionsApi } from '@/api/terminal';
  import { useLogger } from '@/composables/use-logger';

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
  const preferredOpenType = ref<'attach' | 'start'>('start');
  const attachSelectRef = ref<any>();

  const formState = reactive({
    type: 'start' as 'attach' | 'start',
    sessionId: '',
    sessionName: '',
  });

  const sessionOptions = ref<SessionOption[]>([]);
  const sessionLoading = ref(false);
  const isSubmitting = ref(false);

  // 已移除会话名称验证和生成函数，改为后端自动生成

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

      // 在 attach 模式下，自动选择第一个可用（detached）会话
      if (formState.type === 'attach' && !formState.sessionId) {
        const firstAvailable = sessionOptions.value.find(
          (item) => !item.disabled
        );
        if (firstAvailable) {
          formState.sessionId = firstAvailable.value;
        }
      }

      // 打开 attach 弹窗时，自动聚焦到会话选择器
      if (formState.type === 'attach') {
        await nextTick();
        attachSelectRef.value?.focus?.();
      }
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
    formState.type = preferredOpenType.value;
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
    }
    // 对于start类型，不需要验证会话名称，后端会自动生成

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
    } else {
      preferredOpenType.value = 'start';
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

  function openAttachMode(): void {
    preferredOpenType.value = 'attach';
    visible.value = true;
  }

  defineExpose({
    openAttachMode,
  });
</script>

<style scoped>
  .add-session-btn {
    height: 30px;
    padding: 0 10px;
    font-size: 12px;
    font-weight: 500;
    line-height: 1;
    border-radius: 0;
  }

  .add-session-btn :deep(.arco-btn-icon) {
    margin-right: 4px;
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

  .start-session-hint {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.5;
    color: var(--color-text-3);
  }
</style>
