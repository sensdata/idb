import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { getTerminalSessionsApi } from '@/api/terminal';
import { HostEntity } from '@/entity/Host';
import { useLogger } from '@/hooks/use-logger';

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

  // 创建第一个会话（直接尝试attach，让后端自动选择session）
  async function createFirstSession(
    host: HostEntity
  ): Promise<SessionCreationResult> {
    if (!host || !host.id || host.id <= 0) {
      logWarn('Invalid host provided to createFirstSession:', host);
      return {
        type: 'start',
      };
    }

    // 直接尝试attach，后端会自动选择最新的detached session
    // 如果没有可用的session，后端会返回错误，前端再fallback到创建新session
    return {
      type: 'attach',
      sessionId: '', // 空sessionId让后端自动选择
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
    clearSessions,
  } as const;
}
