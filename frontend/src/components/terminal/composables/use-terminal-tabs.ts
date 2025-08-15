import { ref, onUnmounted, watch, computed, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { debounce } from 'lodash';
import { useLogger } from '@/composables/use-logger';
import type { SendMsgDo } from '../type';
import { useTerminalPersistence } from './use-terminal-persistence';

// 定义终端引用的接口，避免使用any
export interface TerminalRef {
  focus: () => void;
  fit: () => void;
  forceRefit: () => void;
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
  // 保存所有主机的终端会话
  const allTerms = ref<TermSessionItem[]>([]);

  // 当前主机ID，用于过滤显示的会话
  const currentHostId = ref<number>();

  // 缓存每个主机的会话，避免重复计算
  const hostTermsCache = ref<Map<number, TermSessionItem[]>>(new Map());

  // 当前主机的终端会话（计算属性，基于allTerms过滤）
  const terms = computed(() => {
    if (!currentHostId.value) return [];

    const currentTerms = allTerms.value.filter(
      (item) => item.hostId === currentHostId.value
    );

    // 更新缓存
    hostTermsCache.value.set(currentHostId.value, currentTerms);

    return currentTerms;
  });

  const activeKey = ref<string>();
  const { t } = useI18n();
  const { logError, logWarn } = useLogger('TerminalTabs');
  const {
    saveHostSessionState,
    loadHostSessionState,
    deleteHostSessionState,
    hasHostSessionState,
    convertToRuntimeSessions,
  } = useTerminalPersistence();

  // 终端标签页数量限制
  const MAX_TABS = 5;

  // 获取活跃的终端项（使用computed优化性能）
  const activeItem = computed(() => {
    return terms.value.find((item) => item.key === activeKey.value);
  });

  // 生成唯一的key
  const generateKey = (): string => {
    return `term_${Date.now()}_${Math.random().toString(36).slice(2)}`;
  };

  // 保存当前状态到持久化存储
  const saveCurrentState = (hostId?: number): void => {
    const targetHostId = hostId || currentHostId.value;
    if (!targetHostId) return;

    // 获取指定主机的会话
    const hostTerms = allTerms.value.filter(
      (item) => item.hostId === targetHostId
    );
    if (hostTerms.length === 0) return;

    // 从会话中获取主机信息
    const firstTerm = hostTerms[0];
    if (!firstTerm) return;

    const hostName = firstTerm.hostName;

    try {
      // 保存指定主机的会话
      saveHostSessionState(
        targetHostId,
        hostName,
        hostTerms,
        targetHostId === currentHostId.value ? activeKey.value : undefined
      );
      logWarn(`Saved ${hostTerms.length} sessions for host ${targetHostId}`);
    } catch (error) {
      logError('Failed to save terminal state:', error);
    }
  };

  // 添加新的终端会话
  const addItem = (options: AddItemOptions): string => {
    try {
      // 检查指定主机的标签页数量限制
      const hostTabs = allTerms.value.filter(
        (item) => item.hostId === options.hostId
      );
      if (hostTabs.length >= MAX_TABS) {
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

      // 添加到所有会话列表
      allTerms.value.push(newItem);

      // 如果是当前主机的会话，设置为活跃
      if (options.hostId === currentHostId.value) {
        activeKey.value = key;
      }

      // 立即保存状态（对于有sessionId的attach类型会话）
      if (
        options.type === 'attach' &&
        options.sessionId &&
        options.hostId === currentHostId.value
      ) {
        saveCurrentState();
      }

      logWarn(
        `Added session ${key} for host ${options.hostId} (current: ${currentHostId.value})`
      );
      return key;
    } catch (error) {
      logError('Failed to add terminal item:', error);
      throw error;
    }
  };

  // 移除终端会话
  const removeItem = (key: string | number): void => {
    try {
      const index = allTerms.value.findIndex((item) => item.key === key);
      if (index === -1) {
        return;
      }

      // 安全地清理终端引用
      const item = allTerms.value[index];
      if (item.termRef?.dispose) {
        try {
          item.termRef.dispose();
        } catch (error) {
          logError('Error disposing terminal:', error);
        }
      }

      // 从所有会话列表中移除
      allTerms.value.splice(index, 1);

      // 清理缓存
      hostTermsCache.value.clear();

      // 更新活跃的key（只考虑当前主机的会话）
      if (key === activeKey.value) {
        const currentHostTerms = terms.value;
        const newActiveIndex = Math.min(index, currentHostTerms.length - 1);
        activeKey.value = currentHostTerms[newActiveIndex]?.key;
      }

      logWarn(`Removed session ${key}`);
    } catch (error) {
      logError('Failed to remove terminal item:', error);
    }
  };

  // 聚焦当前活跃的终端并适配尺寸
  const focus = (): void => {
    const terminal = activeItem.value?.termRef;
    if (!terminal?.focus) return;

    try {
      terminal.focus();

      // 切换标签页时重新适配终端尺寸，解决显示问题
      if (terminal.fit) {
        requestAnimationFrame(() => {
          terminal.fit?.();
        });
      }
    } catch (error) {
      logError('Failed to focus terminal:', error);
    }
  };

  // 验证并设置终端引用
  const isValidTerminalRef = (el: unknown): el is TerminalRef => {
    return (
      el !== null &&
      typeof el === 'object' &&
      'focus' in el &&
      'fit' in el &&
      'forceRefit' in el &&
      'dispose' in el &&
      'sendWsMsg' in el &&
      typeof (el as any).focus === 'function' &&
      typeof (el as any).fit === 'function' &&
      typeof (el as any).forceRefit === 'function' &&
      typeof (el as any).dispose === 'function' &&
      typeof (el as any).sendWsMsg === 'function'
    );
  };

  // 设置终端引用
  const setTerminalRef = (el: unknown, item: TermSessionItem): void => {
    if (isValidTerminalRef(el)) {
      item.termRef = el;
    } else if (el === null) {
      // 清理引用
      item.termRef = undefined;
    }
  };

  // 清空终端会话
  const clearAll = (options?: { hostId?: number; dispose?: boolean }): void => {
    try {
      const { hostId, dispose = true } = options || {};

      if (hostId) {
        // 清空指定主机的会话
        const hostTerms = allTerms.value.filter(
          (item) => item.hostId === hostId
        );

        if (dispose) {
          // 安全地清理指定主机的终端引用
          hostTerms.forEach((item) => {
            if (item.termRef?.dispose) {
              try {
                item.termRef.dispose();
              } catch (error) {
                logError('Error disposing terminal during clear host:', error);
              }
            }
          });
        }

        // 从所有会话中移除指定主机的会话
        allTerms.value = allTerms.value.filter(
          (item) => item.hostId !== hostId
        );

        // 如果清空的是当前主机，重置activeKey
        if (hostId === currentHostId.value) {
          activeKey.value = undefined;
        }

        logWarn(
          `Cleared ${hostTerms.length} sessions for host ${hostId} (dispose: ${dispose})`
        );
      } else {
        // 清空所有会话
        if (dispose) {
          // 安全地清理所有终端引用
          allTerms.value.forEach((item) => {
            if (item.termRef?.dispose) {
              try {
                item.termRef.dispose();
              } catch (error) {
                logError('Error disposing terminal during clear all:', error);
              }
            }
          });
        }

        // 清空数组
        allTerms.value.splice(0);
        activeKey.value = undefined;

        // 清理缓存
        hostTermsCache.value.clear();

        logWarn(`Cleared all sessions (dispose: ${dispose})`);
      }
    } catch (error) {
      logError('Failed to clear terminals:', error);
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
    return allTerms.value.filter((item) => item.hostId === hostId).length;
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

      // 清空指定主机的现有会话（但不断开连接）
      clearAll({ hostId, dispose: false });

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
          allTerms.value.push(newItem);
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

      // 恢复活跃的标签页（只考虑当前主机的会话）
      const currentHostTerms = terms.value;
      if (
        savedState.activeKey &&
        currentHostTerms.some((item) => item.key === savedState.activeKey)
      ) {
        activeKey.value = savedState.activeKey;
      } else if (currentHostTerms.length > 0) {
        activeKey.value = currentHostTerms[0].key;
      }

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
      if (currentHostId.value) {
        saveCurrentState(currentHostId.value);
      }
    },
    500, // 减少防抖时间，提高响应性
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

  // 切换主机功能 - 保持后台会话连接
  const switchToHost = (hostId: number): void => {
    const previousHostId = currentHostId.value;

    // 异步保存当前主机状态，不阻塞切换
    if (previousHostId && terms.value.length > 0) {
      // 使用 nextTick 在视图更新后保存上一个主机的状态，确保保存目标正确
      nextTick(() => {
        saveCurrentState(previousHostId);
      });
    }

    // 立即切换到新主机
    currentHostId.value = hostId;

    // 立即更新activeKey到新主机的第一个会话
    const newHostTerms = allTerms.value.filter(
      (item) => item.hostId === hostId
    );
    if (newHostTerms.length > 0) {
      activeKey.value = newHostTerms[0].key;
    } else {
      activeKey.value = undefined;
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

    // 最后清理终端（真正断开连接）
    clearAll({ dispose: true });
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
    // 新增的主机切换功能
    switchToHost,
    allTerms, // 暴露所有会话用于调试
  };
}
