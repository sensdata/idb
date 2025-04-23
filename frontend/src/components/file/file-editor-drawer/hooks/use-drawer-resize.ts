import { ref, onMounted, onUnmounted } from 'vue';

export default function useDrawerResize() {
  const drawerWidth = ref(1200);
  const isFullScreen = ref(false);
  const initialWidth = ref(0);
  const initialX = ref(0);

  const handleMouseMove = (e: MouseEvent) => {
    const offsetX = initialX.value - e.clientX;
    drawerWidth.value = Math.max(
      500,
      Math.min(2000, initialWidth.value + offsetX)
    );
  };

  const stopResize = () => {
    document.removeEventListener('mousemove', handleMouseMove);
    document.removeEventListener('mouseup', stopResize);
  };

  const startResize = (e: MouseEvent) => {
    initialWidth.value = drawerWidth.value;
    initialX.value = e.clientX;

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', stopResize);
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

  const handleWindowResize = () => {
    if (isFullScreen.value) {
      drawerWidth.value = window.innerWidth;
    }
  };

  onMounted(() => {
    window.addEventListener('resize', handleWindowResize);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', handleWindowResize);
  });

  return {
    drawerWidth,
    isFullScreen,
    setDrawerWidth,
    startResize,
    toggleFullScreen,
  };
}
