import { ref, onMounted, computed, Ref } from 'vue';

export default function useDrawerResize() {
  const drawerWidth = ref(1200);
  const isFullScreen = ref(false);
  const initialWidth = ref(0);
  const initialX = ref(0);
  const isResizing = ref(false);
  const scrollTop = ref(0);

  // 计算抽屉的左侧位置
  const drawerLeft = computed(() => {
    // 由于抽屉是从右侧打开的，左侧位置 = 窗口宽度 - 抽屉宽度
    return window.innerWidth - drawerWidth.value;
  });

  // 计算resize handle的样式
  const resizeHandleStyle = computed(() => ({
    top: `${scrollTop.value}px`,
  }));

  // 处理鼠标移动 - 由父组件在mousemove事件中调用
  const handleMouseMove = (clientX: number) => {
    if (!isResizing.value) return;

    const offsetX = initialX.value - clientX;
    drawerWidth.value = Math.max(
      500,
      Math.min(2000, initialWidth.value + offsetX)
    );
  };

  // 停止调整大小 - 由父组件在mouseup事件中调用
  const stopResize = () => {
    isResizing.value = false;
  };

  // 开始调整大小 - 由父组件在mousedown事件中调用
  const startResize = (clientX: number) => {
    initialWidth.value = drawerWidth.value;
    initialX.value = clientX;
    isResizing.value = true;
  };

  const setDrawerWidth = (width: number) => {
    drawerWidth.value = width;
  };

  const toggleFullScreen = () => {
    isFullScreen.value = !isFullScreen.value;
    if (isFullScreen.value) {
      drawerWidth.value = window.innerWidth;
    } else {
      drawerWidth.value = 1200;
    }
  };

  // 更新窗口大小 - 由父组件在window resize事件中调用
  const handleWindowResize = (windowWidth: number) => {
    if (isFullScreen.value) {
      drawerWidth.value = windowWidth;
    }
  };

  // 更新滚动位置 - 由父组件在scroll事件中调用
  const updateScrollPosition = (newScrollTop: number) => {
    scrollTop.value = newScrollTop;
  };

  return {
    drawerWidth,
    isFullScreen,
    isResizing,
    drawerLeft,
    resizeHandleStyle,
    setDrawerWidth,
    startResize,
    handleMouseMove,
    stopResize,
    toggleFullScreen,
    handleWindowResize,
    updateScrollPosition,
  };
}
