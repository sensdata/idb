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

  // 创建第一个会话（智能选择attach或start）
  async function createFirstSession(
    host: HostEntity
  ): Promise<SessionCreationResult> {
    if (!host || !host.id || host.id <= 0) {
      logWarn('Invalid host provided to createFirstSession:', host);
      return {
        type: 'start',
      };
    }

    try {
      // 先查询服务器上的会话
      const serverSessions = await getAllHostSessions(host.id);
      logWarn(
        `Found ${serverSessions.length} sessions on server for host ${host.id}`
      );

      // 查找detached状态的会话
      const detachedSessions = serverSessions.filter(
        (session) => session.status?.toLowerCase() === 'detached'
      );

      if (detachedSessions.length > 0) {
        // 选择最新的detached会话
        const latestSession = detachedSessions.reduce((latest, current) => {
          // 如果有时间信息，选择最新的；否则选择第一个
          if (latest.time && current.time) {
            return new Date(current.time) > new Date(latest.time)
              ? current
              : latest;
          }
          return latest;
        });

        // 确保选择的会话确实有sessionId
        if (latestSession.session) {
          logWarn(
            `Will attach to detached session: ${latestSession.session} (${latestSession.name})`
          );
          return {
            type: 'attach',
            sessionId: latestSession.session,
          };
        }
        logWarn(
          'Selected detached session has no sessionId, will create new session'
        );
        return {
          type: 'start',
        };
      }
      // 没有detached会话，创建新会话
      logWarn('No detached sessions found, will create new session');
      return {
        type: 'start',
      };
    } catch (error) {
      logError(
        'Failed to query server sessions for createFirstSession:',
        error
      );
      // 查询失败，降级为创建新会话
      return {
        type: 'start',
      };
    }
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
