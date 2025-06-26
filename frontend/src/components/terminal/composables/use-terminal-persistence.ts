import { ref, computed, nextTick } from 'vue';
import { useLogger } from '@/composables/use-logger';
import type { TermSessionItem } from './use-terminal-tabs';

// 定义持久化的终端会话项接口
interface PersistedTermSessionItem {
  key: string;
  type: 'attach' | 'start'; // 实际上应该总是 'attach'，因为只有attach类型的会话需要恢复
  hostId: number;
  hostName: string;
  title: string;
  sessionId?: string; // 对于持久化的会话，这个应该总是存在
  sessionName?: string;
  originalSessionName?: string; // 保存原始的服务器会话名称，用于恢复会话
  isCustomTitle?: boolean; // 标记标题是否为用户自定义
  // 不持久化 termRef 和 isRenaming 等运行时状态
}

// 定义主机会话状态
interface HostSessionState {
  hostId: number;
  hostName: string;
  sessions: PersistedTermSessionItem[];
  activeKey?: string;
  lastUpdated: number;
}

// 本地存储的key
const STORAGE_KEY = 'terminal-workspace-state';

export function useTerminalPersistence() {
  const { logError, logWarn } = useLogger('TerminalPersistence');

  // 响应式状态
  const isStorageAvailable = ref(false);
  const cachedStates = ref<Record<string, HostSessionState>>({});
  const lastSyncTime = ref<number>(0);

  // 清理过期的状态（超过7天）
  const MAX_AGE_DAYS = 7;
  const MAX_AGE_MS = MAX_AGE_DAYS * 24 * 60 * 60 * 1000;

  // 检查localStorage是否可用
  const checkStorageAvailability = (): boolean => {
    try {
      const testKey = '__test__';
      localStorage.setItem(testKey, 'test');
      localStorage.removeItem(testKey);
      isStorageAvailable.value = true;
      return true;
    } catch (error) {
      logWarn('localStorage is not available:', error);
      isStorageAvailable.value = false;
      return false;
    }
  };

  // 初始化时检查存储可用性
  checkStorageAvailability();

  // 计算属性：存储状态
  const storageStatus = computed(() => ({
    available: isStorageAvailable.value,
    cachedCount: Object.keys(cachedStates.value).length,
    lastSync: lastSyncTime.value,
  }));

  // 安全地获取存储数据
  const safeGetStorage = (key: string): string | null => {
    if (!isStorageAvailable.value) return null;
    try {
      return localStorage.getItem(key);
    } catch (error) {
      logError('Failed to get storage item:', error);
      return null;
    }
  };

  // 安全地设置存储数据
  const safeSetStorage = (key: string, value: string): boolean => {
    if (!isStorageAvailable.value) return false;
    try {
      localStorage.setItem(key, value);
      return true;
    } catch (error) {
      logError('Failed to set storage item:', error);
      return false;
    }
  };

  // 安全地移除存储数据
  const safeRemoveStorage = (key: string): boolean => {
    if (!isStorageAvailable.value) return false;
    try {
      localStorage.removeItem(key);
      return true;
    } catch (error) {
      logError('Failed to remove storage item:', error);
      return false;
    }
  };

  // 解析存储的状态数据
  const parseStoredStates = (
    stored: string
  ): Record<string, HostSessionState> => {
    try {
      const parsed = JSON.parse(stored) as Record<string, HostSessionState>;
      return parsed || {};
    } catch (error) {
      logError('Failed to parse stored states:', error);
      return {};
    }
  };

  // 同步缓存和存储
  const syncCache = (): void => {
    const stored = safeGetStorage(STORAGE_KEY);
    if (stored) {
      cachedStates.value = parseStoredStates(stored);
      lastSyncTime.value = Date.now();
    }
  };

  // 清理过期状态
  const cleanupExpiredStates = (
    allStates: Record<string, HostSessionState>
  ): Record<string, HostSessionState> => {
    const now = Date.now();
    const cleanedStates: Record<string, HostSessionState> = {};

    Object.entries(allStates).forEach(([hostId, state]) => {
      if (now - state.lastUpdated <= MAX_AGE_MS) {
        cleanedStates[hostId] = state;
      }
    });

    return cleanedStates;
  };

  // 创建持久化会话项
  const createPersistedSessions = (
    sessions: TermSessionItem[],
    hostId: number,
    hostName: string
  ): PersistedTermSessionItem[] => {
    // 保存所有有sessionId的会话，无论是attach还是start类型
    // 一旦会话获得了sessionId，说明它在服务器上存在，就可以被恢复
    const validSessions = sessions.filter((session) => {
      // 有sessionId的会话都可以保存（无论是attach还是start类型）
      if (session.sessionId) {
        return true;
      }

      // 没有sessionId的会话不保存（无法恢复）
      logWarn(
        `Skipping session ${session.key} without sessionId - cannot be restored`
      );
      return false;
    });

    logWarn(
      `Filtering sessions for persistence: ${sessions.length} total, ${validSessions.length} valid (with sessionId)`
    );

    return validSessions.map((session) => ({
      key: session.key,
      type: 'attach', // 保存时统一转换为attach类型，因为恢复时都是attach操作
      hostId, // 使用参数传入的hostId，确保一致性
      hostName, // 使用参数传入的hostName，确保一致性
      title: session.title,
      sessionId: session.sessionId, // 应该总是存在
      sessionName: session.sessionName,
      originalSessionName: session.originalSessionName,
      isCustomTitle: session.isCustomTitle || false,
    }));
  };

  // 删除主机会话状态
  const deleteHostSessionState = (hostId: number): void => {
    try {
      const stored = safeGetStorage(STORAGE_KEY);
      if (!stored) return;

      const allStates = parseStoredStates(stored);
      delete allStates[hostId.toString()];

      const success = safeSetStorage(STORAGE_KEY, JSON.stringify(allStates));
      if (success) {
        // 更新缓存
        delete cachedStates.value[hostId.toString()];
      }
    } catch (error) {
      logError('Failed to delete host session state:', error);
    }
  };

  // 从本地存储加载状态
  const loadHostSessionState = (hostId: number): HostSessionState | null => {
    try {
      // 首先检查缓存
      const hostKey = hostId.toString();
      const cachedState = cachedStates.value[hostKey];

      if (cachedState) {
        // 检查缓存是否过期
        const now = Date.now();
        if (now - cachedState.lastUpdated <= MAX_AGE_MS) {
          return cachedState;
        }
      }

      // 缓存未命中或过期，从存储加载
      const stored = safeGetStorage(STORAGE_KEY);
      if (!stored) return null;

      const allStates = parseStoredStates(stored);
      const hostState = allStates[hostKey];

      if (!hostState) return null;

      // 检查状态是否过期
      const now = Date.now();
      if (now - hostState.lastUpdated > MAX_AGE_MS) {
        logWarn(`Host ${hostId} session state expired, removing`);
        deleteHostSessionState(hostId);
        return null;
      }

      // 更新缓存
      cachedStates.value[hostKey] = hostState;

      return hostState;
    } catch (error) {
      logError('Failed to load host session state:', error);
      return null;
    }
  };

  // 保存主机会话状态到本地存储
  const saveHostSessionState = (
    hostId: number,
    hostName: string,
    sessions: TermSessionItem[],
    activeKey?: string
  ): void => {
    try {
      logWarn(
        `Saving host session state: hostId=${hostId}, hostName=${hostName}, sessions=${sessions.length}, activeKey=${activeKey}`
      );

      // 转换为持久化格式
      const persistedSessions = createPersistedSessions(
        sessions,
        hostId,
        hostName
      );

      logWarn(
        `Persisted sessions: ${JSON.stringify(
          persistedSessions.map((s) => ({
            key: s.key,
            title: s.title,
            sessionId: s.sessionId,
          }))
        )}`
      );

      const hostState: HostSessionState = {
        hostId,
        hostName,
        sessions: persistedSessions,
        activeKey,
        lastUpdated: Date.now(),
      };

      // 获取现有状态
      const stored = safeGetStorage(STORAGE_KEY);
      const allStates = stored ? parseStoredStates(stored) : {};

      logWarn(
        `Existing states before update: ${Object.keys(allStates).join(', ')}`
      );

      // 更新状态
      allStates[hostId.toString()] = hostState;

      logWarn(`States after update: ${Object.keys(allStates).join(', ')}`);

      // 清理过期状态
      const cleanedStates = cleanupExpiredStates(allStates);

      // 保存到本地存储
      const success = safeSetStorage(
        STORAGE_KEY,
        JSON.stringify(cleanedStates)
      );

      if (success) {
        // 更新缓存
        cachedStates.value = cleanedStates;
        lastSyncTime.value = Date.now();
        logWarn(
          `Successfully saved host ${hostId} session state to localStorage`
        );
      } else {
        logError(`Failed to save host ${hostId} session state to localStorage`);
      }
    } catch (error) {
      logError('Failed to save host session state:', error);
    }
  };

  // 获取所有可恢复的主机会话
  const getAllHostSessionStates = (): HostSessionState[] => {
    try {
      // 先同步缓存
      syncCache();

      const stored = safeGetStorage(STORAGE_KEY);
      if (!stored) return [];

      const allStates = parseStoredStates(stored);

      // 清理过期状态
      const cleanedStates = cleanupExpiredStates(allStates);

      // 如果清理后状态有变化，更新存储
      if (Object.keys(cleanedStates).length !== Object.keys(allStates).length) {
        safeSetStorage(STORAGE_KEY, JSON.stringify(cleanedStates));
        cachedStates.value = cleanedStates;
      }

      return Object.values(cleanedStates);
    } catch (error) {
      logError('Failed to load all host session states:', error);
      return [];
    }
  };

  // 清理所有状态
  const clearAllStates = (): void => {
    try {
      const success = safeRemoveStorage(STORAGE_KEY);
      if (success) {
        cachedStates.value = {};
        lastSyncTime.value = 0;
      }
    } catch (error) {
      logError('Failed to clear all states:', error);
    }
  };

  // 检查主机是否有保存的会话状态
  const hasHostSessionState = (hostId: number): boolean => {
    const state = loadHostSessionState(hostId);
    return state !== null && state.sessions.length > 0;
  };

  // 转换持久化的会话项为运行时会话项
  const convertToRuntimeSessions = (
    persistedSessions: PersistedTermSessionItem[]
  ): Omit<TermSessionItem, 'termRef'>[] => {
    return persistedSessions.map((session) => ({
      key: session.key,
      type: session.type, // 应该总是 'attach'
      hostId: session.hostId,
      hostName: session.hostName,
      title: session.title,
      sessionId: session.sessionId, // 应该总是存在
      sessionName: session.sessionName,
      originalSessionName: session.originalSessionName,
      isRenaming: false,
      isCustomTitle: session.isCustomTitle || false,
    }));
  };

  // 强制刷新缓存
  const refreshCache = async (): Promise<void> => {
    await nextTick();
    syncCache();
  };

  return {
    // 响应式状态
    storageStatus,

    // 状态管理
    loadHostSessionState,
    saveHostSessionState,
    deleteHostSessionState,
    hasHostSessionState,

    // 批量操作
    getAllHostSessionStates,
    clearAllStates,

    // 转换工具
    convertToRuntimeSessions,

    // 工具方法
    refreshCache,
    checkStorageAvailability,
  };
}
