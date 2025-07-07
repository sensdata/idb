<template>
  <div class="app-sidebar-layout" :style="layoutStyle">
    <div class="app-sidebar" :style="sidebarStyle">
      <slot name="sidebar" />
    </div>
    <div class="app-main" :style="mainStyle">
      <slot name="main" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, CSSProperties } from 'vue';

  interface Props {
    sidebarWidth?: number;
    sidebarPadding?: string;
    mainPadding?: string;
    containerMarginTop?: number;
    minHeight?: string;
    showBorder?: boolean;
    borderRadius?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    sidebarWidth: 208,
    sidebarPadding: '4px 0',
    mainPadding: '20px',
    containerMarginTop: 0,
    minHeight: 'calc(100vh - 240px)',
    showBorder: true,
    borderRadius: '4px',
  });

  // 计算布局样式
  const layoutStyle = computed((): CSSProperties => {
    const styles: CSSProperties = {
      position: 'relative',
      minHeight: props.minHeight,
      paddingLeft: `${props.sidebarWidth}px`,
      marginTop:
        props.containerMarginTop > 0 ? `${props.containerMarginTop}px` : '0',
    };

    if (props.showBorder) {
      styles.border = '1px solid var(--color-border-2)';
      styles.borderRadius = props.borderRadius;
    }

    return styles;
  });

  // 计算侧边栏样式
  const sidebarStyle = computed(
    (): CSSProperties => ({
      position: 'absolute',
      top: 0,
      bottom: 0,
      left: 0,
      width: `${props.sidebarWidth}px`,
      height: '100%',
      padding: props.sidebarPadding,
      overflow: 'auto',
      borderRight: '1px solid var(--color-border-2)',
      transition: 'width 0.3s ease',
    })
  );

  // 计算主内容区域样式
  const mainStyle = computed(
    (): CSSProperties => ({
      minWidth: 0,
      height: '100%',
      padding: props.mainPadding,
    })
  );
</script>

<style scoped>
  .app-sidebar-layout {
    /* 桌面布局 */
    --sidebar-width-desktop: 208px;
    --sidebar-width-tablet: 180px;
    --sidebar-width-small-tablet: 160px;
    --sidebar-width-mobile: 140px;
    --sidebar-width-small-mobile: 120px;
  }

  /* 响应式设计 */

  /* 桌面布局 */
  @media screen and (width >= 992px) {
    .app-sidebar-layout {
      padding-left: var(--sidebar-width-desktop);
    }
    .app-sidebar {
      width: var(--sidebar-width-desktop);
    }
  }

  /* 平板设备 */
  @media screen and (width <= 991px) {
    .app-sidebar-layout {
      padding-left: var(--sidebar-width-tablet);
    }
    .app-sidebar {
      width: var(--sidebar-width-tablet);
    }
  }

  /* 小型平板 */
  @media screen and (width <= 768px) {
    .app-sidebar-layout {
      padding-left: var(--sidebar-width-small-tablet);
    }
    .app-sidebar {
      width: var(--sidebar-width-small-tablet);
    }
    .app-main {
      padding: 15px;
    }
  }

  /* 手机设备 */
  @media screen and (width <= 576px) {
    .app-sidebar-layout {
      padding-left: var(--sidebar-width-mobile);
    }
    .app-sidebar {
      width: var(--sidebar-width-mobile);
      padding: 4px 4px;
    }
    .app-main {
      padding: 10px;
    }
  }

  /* 小型手机 */
  @media screen and (width <= 480px) {
    .app-sidebar-layout {
      padding-left: var(--sidebar-width-small-mobile);
    }
    .app-sidebar {
      width: var(--sidebar-width-small-mobile);
      padding: 4px 2px;
    }
    .app-main {
      padding: 8px;
    }
  }
</style>
