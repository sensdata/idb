import { ref, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useLogger } from '@/hooks/use-logger';
import type { SendMsgDo } from '../type';

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
  sessionName?: string;
  termRef?: TerminalRef;
  isRenaming?: boolean;
  renameValue?: string;
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
  const { logError, logWarn } = useLogger('TerminalTabs');

  // 生成唯一的key
  const generateKey = (): string => {
    return `term_${Date.now()}_${Math.random().toString(36).slice(2)}`;
  };

  // 添加新的终端会话
  const addItem = (options: AddItemOptions): string => {
    try {
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
        logWarn(`Terminal item with key ${key} not found`);
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
      const activeItem = terms.value.find(
        (item) => item.key === activeKey.value
      );
      if (activeItem?.termRef?.focus) {
        activeItem.termRef.focus();
      }
    } catch (error) {
      logError('Failed to focus terminal:', error);
    }
  };

  // 设置终端引用
  const setTerminalRef = (el: any, item: TermSessionItem): void => {
    if (
      el &&
      typeof el.focus === 'function' &&
      typeof el.dispose === 'function' &&
      typeof el.sendWsMsg === 'function'
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

  // 获取活跃的终端项
  const getActiveItem = (): TermSessionItem | undefined => {
    return terms.value.find((item) => item.key === activeKey.value);
  };

  // 检查是否有指定key的终端
  const hasItem = (key: string): boolean => {
    return terms.value.some((item) => item.key === key);
  };

  // 组件卸载时清理资源
  onUnmounted(() => {
    clearAll();
  });

  return {
    terms,
    activeKey,
    addItem,
    removeItem,
    focus,
    setTerminalRef,
    clearAll,
    getActiveItem,
    hasItem,
  };
}
