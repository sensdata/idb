import { useLogger } from '@/composables/use-logger';

interface ApiOptions<T> {
  errorMessage?: string;
  defaultValue?: T;
  onSuccess?: (result: T) => void;
  onError?: (error: Error) => void;
}

/**
 * API调用加载状态管理Hook
 */
export function useApiWithLoading(setLoadingState: (loading: boolean) => void) {
  const { logError } = useLogger('ApiWithLoading');

  const executeApi = async <T>(
    apiCall: () => Promise<T>,
    options: ApiOptions<T> = {}
  ): Promise<T> => {
    setLoadingState(true);
    try {
      const result = await apiCall();
      options.onSuccess?.(result);
      return result;
    } catch (error) {
      const err = error instanceof Error ? error : new Error(String(error));

      if (options.errorMessage) {
        logError(options.errorMessage, err);
      }

      options.onError?.(err);

      // 如有默认值则返回，否则抛出错误
      if (options.defaultValue !== undefined) {
        return options.defaultValue;
      }

      throw err;
    } finally {
      setLoadingState(false);
    }
  };

  return {
    executeApi,
  };
}
