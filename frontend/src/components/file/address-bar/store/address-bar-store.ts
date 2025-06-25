import { defineStore } from 'pinia';
import { ref } from 'vue';

export const useAddressBarStore = defineStore('address-bar', () => {
  // 用户当前输入的值
  const userInputValue = ref('');

  // 用户是否正在输入（导航状态）
  const isUserNavigating = ref(false);

  // 预期的目标路径
  const expectedPath = ref('');

  // 导航开始时间戳
  const navigationStartTime = ref(0);

  // 设置用户输入状态
  const setUserInput = (value: string) => {
    userInputValue.value = value;
  };

  // 开始导航
  const startNavigation = (targetPath: string) => {
    isUserNavigating.value = true;
    expectedPath.value = targetPath;
    navigationStartTime.value = Date.now();
    userInputValue.value = targetPath.replace(/^\//, ''); // 移除开头的斜杠
  };

  // 完成导航
  const completeNavigation = (actualPath: string) => {
    // 检查是否是我们预期的路径
    if (expectedPath.value === actualPath) {
      isUserNavigating.value = false;
      expectedPath.value = '';
      navigationStartTime.value = 0;
      return true; // 导航成功完成
    }
    return false; // 不是预期的路径
  };

  // 取消导航（超时或其他原因）
  const cancelNavigation = () => {
    isUserNavigating.value = false;
    expectedPath.value = '';
    navigationStartTime.value = 0;
    userInputValue.value = '';
  };

  // 检查导航是否超时（5秒）
  const isNavigationExpired = () => {
    if (!isUserNavigating.value) return false;
    return Date.now() - navigationStartTime.value > 5000;
  };

  // 应该忽略路径变化吗？
  const shouldIgnorePathChange = (newPath: string) => {
    // 如果导航已过期，自动取消
    if (isNavigationExpired()) {
      cancelNavigation();
      return false;
    }

    // 如果正在导航且新路径不是预期路径，则忽略
    return isUserNavigating.value && expectedPath.value !== newPath;
  };

  // 获取当前应该显示的值
  const getDisplayValue = (propsPath: string) => {
    if (isUserNavigating.value && userInputValue.value) {
      return userInputValue.value;
    }
    return propsPath.replace(/^\//, ''); // 移除开头的斜杠
  };

  // 重置所有状态
  const reset = () => {
    userInputValue.value = '';
    isUserNavigating.value = false;
    expectedPath.value = '';
    navigationStartTime.value = 0;
  };

  return {
    // 状态
    userInputValue,
    isUserNavigating,
    expectedPath,
    navigationStartTime,

    // 方法
    setUserInput,
    startNavigation,
    completeNavigation,
    cancelNavigation,
    isNavigationExpired,
    shouldIgnorePathChange,
    getDisplayValue,
    reset,
  };
});
