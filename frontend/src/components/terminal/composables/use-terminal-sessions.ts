import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { getTerminalSessionsApi } from '@/api/terminal';
import { HostEntity } from '@/entity/Host';
import { useLogger } from '@/composables/use-logger';

// 定义会话选项类型
interface SessionOption {
  label: string;
  value: string;
}

// 定义会话创建结果类型
interface SessionCreationResult {
  type: 'attach' | 'start';
  sessionId?: string;
}

// 定义会话项类型
interface SessionItem {
  name: string;
  status: string;
  session: string;
  time?: string;
}

export function useTerminalSessions() {
  const { t } = useI18n();
  const { logWarn, logError } = useLogger('TerminalSessions');
  const sessionOptions = ref<SessionOption[]>([]);
  const sessionLoading = ref<boolean>(false);

  // 加载会话选项
  async function loadSessionOptions(hostId: number): Promise<void> {
    if (!hostId || hostId <= 0) {
      logWarn('Invalid hostId provided to loadSessionOptions:', hostId);
      return;
    }

    sessionLoading.value = true;
    try {
      const res = await getTerminalSessionsApi(hostId);

      if (!res || !Array.isArray(res.items)) {
        logWarn('Invalid response format from getTerminalSessionsApi:', res);
        sessionOptions.value = [];
        return;
      }

      sessionOptions.value = res.items
        .filter((item: SessionItem) => {
          if (!item || typeof item.status !== 'string') {
            logWarn('Invalid session item:', item);
            return false;
          }
          return item.status.toLowerCase() === 'detached';
        })
        .map((item: SessionItem) => ({
          label: `${item.name || 'Unknown'} (${item.status || 'Unknown'})`,
          value: item.session || '',
        }))
        .filter((option) => option.value); // 过滤掉空的session值
    } catch (error) {
      logError('Failed to load terminal sessions:', error);
      Message.error(t('components.terminal.session.loadFailed'));
      sessionOptions.value = [];
    } finally {
      sessionLoading.value = false;
    }
  }

  // 获取主机所有可用的会话（包括attached和detached）
  async function getAllHostSessions(hostId: number): Promise<SessionItem[]> {
    if (!hostId || hostId <= 0) {
      logWarn('Invalid hostId provided to getAllHostSessions:', hostId);
      return [];
    }

    try {
      const res = await getTerminalSessionsApi(hostId);

      if (!res || !Array.isArray(res.items)) {
        logWarn('Invalid response format from getTerminalSessionsApi:', res);
        return [];
      }

      return res.items.filter((item: SessionItem) => {
        return item && item.session && item.name;
      });
    } catch (error) {
      logError('Failed to get all host sessions:', error);
      return [];
    }
  }

  // 创建第一个会话（利用后端自动选择逻辑）
  async function createFirstSession(
    host: HostEntity
  ): Promise<SessionCreationResult> {
    if (!host || !host.id || host.id <= 0) {
      logWarn('Invalid host provided to createFirstSession:', host);
      return {
        type: 'start',
      };
    }

    // 直接返回 attach 类型，不指定 sessionId
    // 后端会自动选择最新的 detached 会话进行 attach
    // 如果没有可用的 detached 会话，后端会自动降级为创建新会话
    return {
      type: 'attach',
      // sessionId 为空，让后端自动处理
    };
  }

  // 清理函数（可选，用于组件卸载时清理状态）
  function clearSessions(): void {
    sessionOptions.value = [];
    sessionLoading.value = false;
  }

  return {
    // 响应式数据
    sessionOptions,
    sessionLoading,
    // 方法
    loadSessionOptions,
    createFirstSession,
    getAllHostSessions,
    clearSessions,
  } as const;
}
