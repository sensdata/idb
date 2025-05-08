<template>
  <a-drawer
    ref="drawerRef"
    :visible="visible"
    :width="drawerWidth"
    :closable="true"
    :unmount-on-close="false"
    class="file-editor-drawer"
    @cancel="handleCancel"
    @close="handleClose"
  >
    <template #title>
      <div class="drawer-title">
        <span>{{ file ? file.name : t('app.file.editor.title') }}</span>
      </div>
    </template>

    <div
      ref="resizeHandleRef"
      class="resize-handle"
      :style="{
        left: `${drawerLeft}px`,
        ...resizeHandleStyle,
        opacity: loading || !content ? 0 : 1,
      }"
      @mousedown="handleResizeStart"
    ></div>

    <a-spin
      ref="drawerBodyRef"
      :spinning="loading === true"
      class="file-editor-container"
      tip="加载文件中..."
    >
      <template #icon>
        <icon-loading />
      </template>

      <!-- 工具栏 -->
      <EditorToolbar
        :drawer-width="drawerWidth"
        :is-full-screen="isFullScreen"
        :read-only="readOnly"
        @update:drawer-width="setDrawerWidth"
        @toggle-full-screen="toggleFullScreen"
        @toggle-edit-mode="toggleEditMode"
      />

      <!-- 编辑器内容 -->
      <EditorContent
        v-model="content"
        :loading="loading"
        :extensions="extensions"
        :is-partial-view="isPartialView"
        :read-only="readOnly"
        @editor-ready="handleEditorReady"
      />

      <!-- 视图模式控制 -->
      <FileViewMode
        v-if="viewMode !== 'loading'"
        ref="viewModeControlRef"
        :view-mode="viewMode"
        :line-count="lineCount"
        :batch-size="batchSize"
        :is-large-file="isLargeFile === true"
        @change-view-mode="handleViewModeChange"
      />

      <!-- 加载更多控制 -->
      <LoadMoreControls
        v-if="isLargeFile && ['head', 'tail'].includes(viewMode as string)"
        :view-mode="viewMode"
        :is-loading-more="isLoadingMore"
        :can-load-more="canLoadMore"
        @load-more="handleLoadMore"
      />
    </a-spin>

    <template #footer>
      <EditorFooter
        :file="file"
        :is-edited="isEdited"
        :is-partial-view="isPartialView"
        :read-only="readOnly"
        @cancel="handleCancel"
        @save="handleSave"
      />
    </template>
  </a-drawer>
</template>

<script lang="ts" setup>
  import {
    ref,
    shallowRef,
    watch,
    computed,
    nextTick,
    onUnmounted,
    onMounted,
    type Ref,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useConfirm } from '@/hooks/confirm';
  import { IconLoading } from '@arco-design/web-vue/es/icon';
  import { ContentViewMode } from '@/components/file/file-editor-drawer/types';
  import { EditorView } from '@codemirror/view';

  import useFileEditor from './hooks/use-file-editor';
  import useEditorConfig from './hooks/use-editor-config';
  import useDrawerResize from './hooks/use-drawer-resize';

  import EditorToolbar from './editor-toolbar.vue';
  import EditorFooter from './editor-footer.vue';
  import EditorContent from './editor-content.vue';
  import FileViewMode from './file-view-mode.vue';
  import LoadMoreControls from './load-more-controls.vue';

  // 事件监听器组合式函数
  const useEventListener = (
    target: Window | Document | HTMLElement,
    event: string,
    callback: (...args: any[]) => void
  ) => {
    onMounted(() => {
      target.addEventListener(event, callback);
    });

    onUnmounted(() => {
      target.removeEventListener(event, callback);
    });
  };

  // 滚动监听组合式函数
  const useScrollListener = (
    elRef: Ref<HTMLElement | null>,
    callback: (scrollTop: number) => void
  ) => {
    let scrollableElement: HTMLElement | null = null;

    // 查找可滚动元素的辅助函数
    const findScrollableElement = (el: HTMLElement): HTMLElement => {
      // 针对Arco Design组件的特殊处理
      const arcoBody = el.querySelector('.arco-drawer-body');
      return (arcoBody as HTMLElement) || el;
    };

    // 滚动事件处理函数
    const handleScroll = (e: Event) => {
      if (e.target instanceof HTMLElement) {
        callback(e.target.scrollTop);
      }
    };

    // 观察引用元素变化
    watch(
      () => elRef.value,
      (newEl) => {
        // 清理旧的事件监听
        if (scrollableElement) {
          scrollableElement.removeEventListener('scroll', handleScroll);
          scrollableElement = null;
        }

        // 设置新的事件监听
        if (newEl) {
          scrollableElement = findScrollableElement(newEl);
          scrollableElement.addEventListener('scroll', handleScroll);
        }
      },
      { immediate: true }
    );

    // 在组件卸载时清理事件监听
    onUnmounted(() => {
      if (scrollableElement) {
        scrollableElement.removeEventListener('scroll', handleScroll);
      }
    });
  };

  const { t } = useI18n();
  const visible = ref(false);
  const editorView = shallowRef();
  const viewModeControlRef = ref();
  const drawerRef = ref<HTMLElement | null>(null);
  const resizeHandleRef = ref<HTMLElement | null>(null);
  const drawerBodyRef = ref<HTMLElement | null>(null);
  const readOnly = ref(true); // 默认为只读模式

  // 拖拽状态变量
  const isResizing = ref(false);
  const initialWidth = ref(0);
  const initialX = ref(0);

  const {
    file,
    content,
    loading,
    isEdited,
    viewMode,
    lineCount,
    isPartialView,
    isLoadingMore,
    canLoadMore,
    loadMoreContent,
    setFile: loadFile,
    saveFile,
    changeViewMode,
    setEditorInstance,
    cleanup,
    batchSize,
  } = useFileEditor();

  const {
    drawerWidth,
    isFullScreen,
    drawerLeft,
    resizeHandleStyle,
    setDrawerWidth,
    toggleFullScreen,
    updateScrollPosition,
  } = useDrawerResize();
  const { extensions } = useEditorConfig(file);

  const { confirm } = useConfirm();

  const emit = defineEmits<{
    (e: 'ok'): void;
  }>();

  // 切换编辑模式
  const toggleEditMode = () => {
    readOnly.value = !readOnly.value;
  };

  // ----- 拖拽调整大小相关函数 -----
  const handleResizeStart = (e: MouseEvent) => {
    initialWidth.value = drawerWidth.value;
    initialX.value = e.clientX;
    isResizing.value = true;
  };

  const handleMouseMove = (e: MouseEvent) => {
    if (isResizing.value) {
      const offsetX = initialX.value - e.clientX;
      drawerWidth.value = Math.max(
        500,
        Math.min(2000, initialWidth.value + offsetX)
      );
    }
  };

  const handleMouseUp = () => {
    isResizing.value = false;
  };

  // 窗口大小变化事件处理
  const handleWindowResize = () => {
    if (isFullScreen.value) {
      drawerWidth.value = window.innerWidth;
    }
  };

  // 使用组合式函数处理全局事件
  useEventListener(document, 'mousemove', handleMouseMove);
  useEventListener(document, 'mouseup', handleMouseUp);
  useEventListener(window, 'resize', handleWindowResize);

  // 使用组合式函数处理滚动监听
  useScrollListener(drawerBodyRef, updateScrollPosition);

  // 在组件卸载时清理文件编辑器资源
  onUnmounted(() => {
    cleanup();
  });

  // ----- 编辑器相关函数 -----
  const handleEditorReady = (editor: EditorView) => {
    editorView.value = editor;
    setEditorInstance(editor);

    // 使用nextTick确保DOM更新完成
    nextTick(() => {
      if (editorView.value) {
        setEditorInstance(editorView.value);
      }
    });
  };

  // ----- 内容管理相关函数 -----
  const handleLoadMore = () => {
    loadMoreContent();
  };

  const handleViewModeChange = (mode: ContentViewMode, lines?: number) => {
    changeViewMode(mode, lines);
  };

  // ----- 弹窗控制函数 -----
  const handleCancel = async () => {
    // 如果文件已编辑但未保存，提示用户
    if (isEdited.value) {
      const result = await confirm({
        title: t('app.file.editor.unsavedChanges'),
        content: t('app.file.editor.confirmClose'),
        okText: t('common.form.okText'),
        cancelText: t('common.form.cancelText'),
      });
      if (!result) {
        return;
      }
    }

    // 确保停止实时追踪连接
    cleanup();

    // 重置编辑模式为只读
    readOnly.value = true;

    // 关闭抽屉，但不触发ok事件，从而不会导致页面刷新
    visible.value = false;
  };

  // 处理点击抽屉外部关闭的事件
  const handleClose = () => {
    // 确保停止实时追踪连接
    cleanup();
    // 重置编辑模式为只读
    readOnly.value = true;
  };

  const handleSave = async () => {
    if (readOnly.value) return; // 只读模式不允许保存

    try {
      await saveFile();
      emit('ok');
    } catch (error) {
      console.error('保存文件失败:', error);
    }
  };

  // ----- 计算属性 -----
  // 是否是大文件（用于决定显示哪些按钮）
  const isLargeFile = computed(() => file.value && file.value.size > 100000);

  // ----- 监听器 -----
  // 监听视图模式变化
  watch(
    () => viewMode.value,
    () => {
      // 当视图模式变化时，确保滚动到顶部
      nextTick(() => {
        if (editorView.value) {
          // 使用EditorView的scrollIntoView方法滚动到顶部
          editorView.value.dispatch({
            effects: EditorView.scrollIntoView(0),
          });
        }
      });
    }
  );

  // 监听加载状态变化，在加载完成后清除按钮loading状态
  watch(
    () => loading.value,
    (newLoading, oldLoading) => {
      // 当加载完成时（从true变为false），清除loading按钮状态
      if (
        oldLoading === true &&
        newLoading === false &&
        viewModeControlRef.value
      ) {
        viewModeControlRef.value.clearLoadingState();
      }
    }
  );

  // 当打开小文件时，检查视图模式以确保不在head或tail模式
  watch(
    () => file.value?.size,
    (newSize) => {
      if (newSize !== undefined && newSize <= 100000) {
        // 如果是小文件，而且当前是head或tail模式，切换回full模式
        if (viewMode.value === 'head' || viewMode.value === 'tail') {
          changeViewMode('full');
        }
      }
    }
  );

  // 文件编辑器 API
  defineExpose({
    setFile: loadFile,
    show: () => {
      visible.value = true;
    },
    hide: () => {
      // 确保停止实时追踪连接
      cleanup();
      visible.value = false;
    },
  });
</script>

<style scoped>
  .file-editor-drawer :deep(.arco-drawer-body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 0;
  }

  .file-editor-container {
    position: relative;
    display: flex;
    flex: 1;
    flex-direction: column;
    width: 100%;
    overflow: hidden;
  }

  /* 加强loading效果 */
  :deep(.arco-spin) {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :deep(.arco-spin-loading) {
    font-size: 24px;
  }

  :deep(.arco-spin-tip) {
    margin-top: 12px;
    font-size: 14px;
  }

  .drawer-title {
    display: flex;
    gap: 8px;
    align-items: center;
    font-weight: 500;
  }

  .resize-handle {
    position: fixed;
    top: 0;
    z-index: 1001;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 5px;
    height: 100vh;
    background: linear-gradient(to right, transparent, rgb(0 0 0 / 5%));
    cursor: col-resize;
    transition: opacity 0.3s ease;
    pointer-events: auto;
  }

  .resize-handle[style*='opacity: 0'] {
    pointer-events: none;
  }

  .resize-handle::after {
    width: 2px;
    height: 30px;
    background-color: rgb(0 0 0 / 20%);
    border-radius: 2px;
    content: '';
  }

  .resize-handle:hover {
    background: linear-gradient(to right, transparent, rgb(0 0 0 / 12%));
  }

  .resize-handle:hover::after {
    width: 3px;
    background-color: rgb(0 0 0 / 40%);
  }

  .file-editor-drawer :deep(.arco-drawer-footer) {
    padding: 12px 20px;
    border-top: 1px solid var(--color-border);
  }
</style>
