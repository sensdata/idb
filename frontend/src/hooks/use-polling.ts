import { ref, onUnmounted } from 'vue';

export interface PollingOptions {
  pollingFunction: () => void | Promise<void>;
  interval: number;
  immediate?: boolean;
  initialDelay?: number;
  onBeforeStart?: (id: number) => void;
  onInvalidId?: (id: number) => void;
  onError?: (error: unknown) => void;
}

/**
 * 用于处理轮询操作的组合式函数
 */
export const usePolling = (options: PollingOptions) => {
  const pollingTimer = ref<number | null>(null);
  const isPolling = ref(false);

  /**
   * 执行轮询任务，并处理可能的错误
   */
  const executePollingTask = async (): Promise<void> => {
    try {
      await options.pollingFunction();
    } catch (error) {
      options.onError?.(error);
    }
  };

  /**
   * 停止轮询
   */
  const stopPolling = (): void => {
    if (pollingTimer.value) {
      clearInterval(pollingTimer.value);
      pollingTimer.value = null;
      isPolling.value = false;
    }
  };

  /**
   * 开始轮询
   */
  const startPolling = (id: number): void => {
    if (id && id > 0) {
      stopPolling();

      if (options.onBeforeStart) {
        options.onBeforeStart(id);
      }

      isPolling.value = true;

      if (options.immediate) {
        executePollingTask();

        if (options.initialDelay) {
          setTimeout(() => {
            executePollingTask();
          }, options.initialDelay);
        }
      }

      pollingTimer.value = window.setInterval(() => {
        executePollingTask();
      }, options.interval);
    } else if (options.onInvalidId) {
      options.onInvalidId(id);
    }
  };

  /**
   * 暂停轮询
   */
  const pausePolling = (): void => {
    if (pollingTimer.value) {
      clearInterval(pollingTimer.value);
      pollingTimer.value = null;
    }
  };

  /**
   * 恢复轮询
   */
  const resumePolling = (id: number): void => {
    if (isPolling.value && !pollingTimer.value) {
      pollingTimer.value = window.setInterval(() => {
        executePollingTask();
      }, options.interval);
    } else {
      startPolling(id);
    }
  };

  // 组件卸载时清理定时器
  onUnmounted(() => {
    stopPolling();
  });

  return {
    startPolling,
    stopPolling,
    pausePolling,
    resumePolling,
    isPolling,
  };
};

export default usePolling;
