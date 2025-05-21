import { reactive, computed } from 'vue';
import type { LoadingStates } from '../types';

export function useLoadingState() {
  // 创建响应式的 loading 状态对象
  const loadingStates = reactive<LoadingStates>({
    port: false,
    listen: false,
    root: false,
    passwordAuth: false,
    keyAuth: false,
    reverseLookup: false,
    sourceConfig: false,
  });

  // 是否有任一 loading 状态激活
  const anyLoading = computed<boolean>(() => {
    return Object.values(loadingStates).some((state) => state === true);
  });

  return {
    loadingStates,
    anyLoading,
  };
}
