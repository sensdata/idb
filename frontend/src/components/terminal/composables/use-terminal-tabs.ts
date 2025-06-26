import { ref, onUnmounted, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { debounce } from 'lodash';
import { useLogger } from '@/composables/use-logger';
import type { SendMsgDo } from '../type';
import { useTerminalPersistence } from './use-terminal-persistence';

// 定义终端引用的接口，避免使用any
export interface TerminalRef {
  focus: () => void;
  dispose: () => void;
  sendWsMsg: (payload: Partial<SendMsgDo>) => void;
}

export interface TermSessionItem {
  key: string;
  type: 'attach' | 'start';
  hostId: number;
  hostName: string;
  title: string;
  sessionId?: string;
  sessionName?: string; // 服务器端的会话名称，不应该被用户重命名影响
  originalSessionName?: string; // 保存原始的服务器会话名称，用于恢复会话
  termRef?: TerminalRef;
  isRenaming?: boolean;
  renameValue?: string;
  isCustomTitle?: boolean; // 标记标题是否为用户自定义
}

export interface AddItemOptions {
  type: 'attach' | 'start';
  hostId: number;
  hostName: string;
  sessionId?: string;
  sessionName?: string;
}

export function useTerminalTabs() {
  const terms = ref<TermSessionItem[]>([]);
  const activeKey = ref<string>();
  const { t } = useI18n();
  const { logError } = useLogger('TerminalTabs');
  const {
    saveHostSessionState,
    loadHostSessionState,
    deleteHostSessionState,
    hasHostSessionState,
    convertToRuntimeSessions,
  } = useTerminalPersistence();

  // 终端标签页数量限制
  const MAX_TABS = 10;

  // 获取活跃的终端项（使用computed优化性能）
  const activeItem = computed(() => {
    return terms.value.find((item) => item.key === activeKey.value);
  });

  // 生成唯一的key
  const generateKey = (): string => {
    return `term_${Date.now()}_${Math.random().toString(36).slice(2)}`;
  };

  // 当前主机ID，用于持久化
  const currentHostId = ref<number>();

  // 保存当前状态到持久化存储
  const saveCurrentState = (): void => {
    // 如果没有任何会话，则不需要保存
    if (terms.value.length === 0) return;

    // 从当前会话中获取主机信息（所有会话应该属于同一个主机）
    const firstTerm = terms.value[0];
    if (!firstTerm) return;

    const hostId = firstTerm.hostId;
    const hostName = firstTerm.hostName;

    // 验证所有会话都属于同一个主机
    const allSameHost = terms.value.every((term) => term.hostId === hostId);
    if (!allSameHost) {
      logError('Found terms from different hosts, cannot save state safely');
      return;
    }

    try {
      saveHostSessionState(hostId, hostName, terms.value, activeKey.value);

      // 更新内部的 currentHostId 以保持同步
      currentHostId.value = hostId;
    } catch (error) {
      logError('Failed to save terminal state:', error);
    }
  };

  // 添加新的终端会话
  const addItem = (options: AddItemOptions): string => {
    try {
      // 检查当前主机的标签页数量限制
      const currentHostTabs = terms.value.filter(
        (item) => item.hostId === options.hostId
      );
      if (currentHostTabs.length >= MAX_TABS) {
        throw new Error(
          t('components.terminal.session.tabLimitReached', { max: MAX_TABS })
        );
      }

      const key = generateKey();

      const newItem: TermSessionItem = {
        key,
        type: options.type,
        hostId: options.hostId,
        hostName: options.hostName,
        title: t('components.terminal.session.connecting'),
        termRef: undefined,
        sessionId: options.sessionId || '',
        sessionName: options.sessionName,
      };

      terms.value.push(newItem);
      activeKey.value = key;

      // 立即保存状态（对于有sessionId的attach类型会话）
      if (options.type === 'attach' && options.sessionId) {
        saveCurrentState();
      }

      return key;
    } catch (error) {
      logError('Failed to add terminal item:', error);
      throw error;
    }
  };

  // 移除终端会话
  const removeItem = (key: string | number): void => {
    try {
      const index = terms.value.findIndex((item) => item.key === key);
      if (index === -1) {
        return;
      }

      // 安全地清理终端引用
      const item = terms.value[index];
      if (item.termRef?.dispose) {
        try {
          item.termRef.dispose();
        } catch (error) {
          logError('Error disposing terminal:', error);
        }
      }

      // 使用splice而不是直接修改length
      terms.value.splice(index, 1);

      // 更新活跃的key
      if (key === activeKey.value) {
        const newActiveIndex = Math.min(index, terms.value.length - 1);
        activeKey.value = terms.value[newActiveIndex]?.key;
      }
    } catch (error) {
      logError('Failed to remove terminal item:', error);
    }
  };

  // 聚焦当前活跃的终端
  const focus = (): void => {
    try {
      if (activeItem.value?.termRef?.focus) {
        activeItem.value.termRef.focus();
      }
    } catch (error) {
      logError('Failed to focus terminal:', error);
    }
  };

  // 设置终端引用
  const setTerminalRef = (el: unknown, item: TermSessionItem): void => {
    if (
      el &&
      typeof el === 'object' &&
      'focus' in el &&
      'dispose' in el &&
      'sendWsMsg' in el &&
      typeof (el as any).focus === 'function' &&
      typeof (el as any).dispose === 'function' &&
      typeof (el as any).sendWsMsg === 'function'
    ) {
      item.termRef = el as TerminalRef;
    } else if (el === null) {
      // 清理引用
      item.termRef = undefined;
    }
  };

  // 清空所有终端
  const clearAll = (): void => {
    try {
      // 安全地清理所有终端引用
      terms.value.forEach((item) => {
        if (item.termRef?.dispose) {
          try {
            item.termRef.dispose();
          } catch (error) {
            logError('Error disposing terminal during clear all:', error);
          }
        }
      });

      // 清空数组
      terms.value.splice(0);
      activeKey.value = undefined;
    } catch (error) {
      logError('Failed to clear all terminals:', error);
    }
  };

  // 获取活跃的终端项（保持向后兼容）
  const getActiveItem = (): TermSessionItem | undefined => {
    return activeItem.value;
  };

  // 检查是否有指定key的终端
  const hasItem = (key: string): boolean => {
    return terms.value.some((item) => item.key === key);
  };

  // 获取指定主机的标签页数量
  const getHostTabsCount = (hostId: number): number => {
    return terms.value.filter((item) => item.hostId === hostId).length;
  };

  // 检查指定主机是否可以添加新标签页
  const canAddTab = (hostId: number): boolean => {
    return getHostTabsCount(hostId) < MAX_TABS;
  };

  // 恢复指定主机的会话状态
  const restoreHostSessions = async (
    hostId: number,
    hostName: string,
    getAllHostSessions?: (hostId: number) => Promise<any[]>
  ): Promise<boolean> => {
    try {
      const savedState = loadHostSessionState(hostId);
      if (!savedState || !savedState.sessions.length) {
        return false;
      }

      // 验证保存的会话确实属于指定的主机
      const validSessions = savedState.sessions.filter(
        (session) => session.hostId === hostId
      );
      if (validSessions.length === 0) {
        deleteHostSessionState(hostId);
        return false;
      }

      // 查询服务器上实际存在的会话
      let serverSessions: any[] = [];
      try {
        if (getAllHostSessions) {
          serverSessions = await getAllHostSessions(hostId);
        }
      } catch (error) {
        logError('Failed to query server sessions:', error);
        // 如果无法查询服务器会话，仍然尝试恢复，但标记为不确定状态
      }

      // 清空当前会话
      clearAll();

      // 智能恢复会话：只恢复可以真正attach到原会话的标签页
      let restoredCount = 0;
      let removedCount = 0;
      const runtimeSessions = convertToRuntimeSessions(validSessions);
      const validSessionsToKeep: any[] = []; // 用于更新缓存

      runtimeSessions.forEach((session) => {
        let shouldRestore = false;
        let shouldKeepInCache = true;

        // 恢复会话的逻辑：只有当可以真正attach到原始会话时才恢复
        // 所有保存的会话都应该是为了恢复到原来的会话，所以都应该是attach类型

        if (session.type === 'start') {
          // start类型的会话不应该被保存用于恢复，这是一个错误状态
          // 从缓存中删除这种无效的会话记录
          shouldRestore = false;
          shouldKeepInCache = false;
          removedCount++;
        } else if (session.type === 'attach') {
          // 检查是否有有效的sessionId
          if (!session.sessionId) {
            // attach类型但没有sessionId，无效的会话记录，从缓存中删除
            shouldRestore = false;
            shouldKeepInCache = false;
            removedCount++;
          } else if (serverSessions.length === 0) {
            // 无法查询服务器会话，保留在缓存中但不恢复
            shouldRestore = false;
            shouldKeepInCache = true;
          } else {
            // 在服务器会话列表中查找对应的会话
            const serverSession = serverSessions.find(
              (s) => s.session === session.sessionId
            );

            if (
              serverSession &&
              serverSession.status?.toLowerCase() === 'detached'
            ) {
              // 服务器上存在且为detached状态，可以真正attach恢复
              shouldRestore = true;
              shouldKeepInCache = true;
            } else if (serverSession) {
              // 服务器上存在但是attached状态，不能恢复，从缓存中删除
              shouldRestore = false;
              shouldKeepInCache = false;
              removedCount++;
            } else {
              // 服务器上不存在此会话，会话已经不存在，从缓存中删除
              shouldRestore = false;
              shouldKeepInCache = false;
              removedCount++;
            }
          }
        } else {
          // 未知的会话类型，从缓存中删除
          shouldRestore = false;
          shouldKeepInCache = false;
          removedCount++;
        }

        // 决定是否保留在缓存中
        if (shouldKeepInCache) {
          // 找到对应的原始持久化会话项
          const originalSession = validSessions.find(
            (s) => s.key === session.key
          );
          if (originalSession) {
            validSessionsToKeep.push(originalSession);
          }
        }

        // 决定是否恢复到运行时
        if (shouldRestore) {
          const newItem: TermSessionItem = {
            ...session,
            hostId, // 确保使用正确的hostId
            hostName, // 确保使用正确的hostName
            termRef: undefined, // 这个会在组件中设置
          };
          terms.value.push(newItem);
          restoredCount++;
        }
      });

      // 如果有会话被移除，更新缓存
      if (removedCount > 0) {
        if (validSessionsToKeep.length > 0) {
          // 重新保存缓存，只保留有效的会话
          // 将持久化的会话项转换为运行时会话项用于保存
          const validRuntimeSessions =
            convertToRuntimeSessions(validSessionsToKeep);

          // 保存更新后的状态
          saveHostSessionState(
            hostId,
            hostName,
            validRuntimeSessions as TermSessionItem[],
            savedState.activeKey
          );
        } else {
          // 没有有效会话了，删除整个主机的缓存
          deleteHostSessionState(hostId);
        }
      }

      // 恢复活跃的标签页
      if (
        savedState.activeKey &&
        terms.value.some((item) => item.key === savedState.activeKey)
      ) {
        activeKey.value = savedState.activeKey;
      } else if (terms.value.length > 0) {
        activeKey.value = terms.value[0].key;
      }

      currentHostId.value = hostId;

      return restoredCount > 0;
    } catch (error) {
      logError('Failed to restore host sessions:', error);
      return false;
    }
  };

  // 检查主机是否有可恢复的会话
  const hasRestoredSessions = (hostId: number): boolean => {
    return hasHostSessionState(hostId);
  };

  // 设置当前主机并保存状态
  const setCurrentHost = (hostId: number): void => {
    currentHostId.value = hostId;
  };

  // 监听会话变化，自动保存状态
  const debouncedSaveState = debounce(
    () => {
      saveCurrentState();
    },
    1000,
    {
      leading: false, // 不在开始时执行
      trailing: true, // 在结束时执行
    }
  );

  // 监听terms和activeKey变化
  const stopWatchingTerms = watch(
    () => ({ terms: terms.value, activeKey: activeKey.value }),
    () => {
      if (currentHostId.value && terms.value.length > 0) {
        debouncedSaveState();
      }
    },
    { deep: true }
  );

  // 清空指定主机的会话（当主机切换时）
  const clearHostSessions = (hostId: number): void => {
    try {
      deleteHostSessionState(hostId);
    } catch (error) {
      logError('Failed to clear host sessions:', error);
    }
  };

  // 组件卸载时清理资源
  onUnmounted(() => {
    // 先保存当前状态（在清理之前）
    if (terms.value.length > 0) {
      saveCurrentState();
    }

    // 取消防抖函数
    debouncedSaveState.cancel();

    // 停止监听
    stopWatchingTerms();

    // 最后清理终端
    clearAll();
  });

  return {
    terms,
    activeKey,
    activeItem,
    addItem,
    removeItem,
    focus,
    setTerminalRef,
    clearAll,
    getActiveItem,
    hasItem,
    getHostTabsCount,
    canAddTab,
    MAX_TABS,
    // 新增的持久化功能
    setCurrentHost,
    restoreHostSessions,
    hasRestoredSessions,
    saveCurrentState,
    clearHostSessions,
  };
}
