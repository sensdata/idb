<template>
  <a-drawer
    ref="drawerRef"
    :visible="visible"
    :closable="true"
    :unmount-on-close="false"
    :width="drawerWidth"
    class="file-editor-drawer"
    @cancel="handleCancel"
    @close="handleClose"
  >
    <template #title>
      <div class="drawer-title">
        <span class="title-path">{{
          file ? file.path : t('app.file.editor.title')
        }}</span>
        <a-tag v-if="viewMode !== 'loading'" :color="modeTagColor">
          {{ modeTagText }}
        </a-tag>
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

    <div class="file-editor-container" @scroll="handleScroll">
      <a-spin
        :spinning="loading === true"
        tip="加载文件中..."
        style="height: 100%"
      >
        <template #icon>
          <icon-loading />
        </template>

        <!-- 工具栏容器 - 左侧EditorToolbar，右侧FileViewMode -->
        <div class="toolbar-container">
          <EditorToolbar
            :drawer-width="drawerWidth"
            :is-full-screen="isFullScreen"
            :read-only="readOnly"
            :view-mode="viewMode"
            @update:drawer-width="setDrawerWidth"
            @toggle-full-screen="toggleFullScreen"
            @toggle-edit-mode="toggleEditMode"
          />

          <FileViewMode
            v-if="viewMode !== 'loading'"
            ref="viewModeControlRef"
            :view-mode="viewMode"
            :line-count="lineCount"
            :batch-size="batchSize"
            :is-large-file="isLargeFile === true"
            :is-follow-mode="isFollowMode"
            :is-follow-paused="isFollowPaused"
            @change-view-mode="handleViewModeChange"
            @pause-follow="pauseFollowMode"
            @resume-follow="resumeFollowMode"
          />
        </div>

        <div
          v-if="viewMode !== 'loading'"
          :class="[
            'editor-mode-banner',
            isEditing ? 'is-editing' : 'is-readonly',
          ]"
        >
          <icon-edit v-if="isEditing" />
          <icon-lock v-else />
          <span>{{ modeHintText }}</span>
        </div>

        <!-- 编辑器内容 -->
        <CodeEditor
          v-model="content"
          :loading="loading"
          :file="file"
          :is-partial-view="isPartialView"
          :read-only="readOnly"
          :loading-text="t('app.file.editor.loadingContent')"
          @editor-ready="handleEditorReady"
          @content-double-click="handleContentDoubleClick"
        />
      </a-spin>
    </div>

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
    computed,
    shallowRef,
    onUnmounted,
    watch,
    nextTick,
    onMounted,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { EditorView } from '@codemirror/view';
  import {
    IconLoading,
    IconEdit,
    IconLock,
  } from '@arco-design/web-vue/es/icon';
  import { useConfirm } from '@/composables/confirm';
  import CodeEditor from '@/components/code-editor/index.vue';
  import useFileEditor from './composables/use-file-editor';
  import useDrawerResize from './composables/use-drawer-resize';
  import { ContentViewMode, FileItem } from './types';
  import FileViewMode from './file-view-mode.vue';
  import EditorToolbar from './editor-toolbar.vue';
  import EditorFooter from './editor-footer.vue';

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

  const { t } = useI18n();
  const visible = ref(false);
  const editorView = shallowRef();
  const viewModeControlRef = ref();
  const drawerRef = ref<HTMLElement | null>(null);
  const resizeHandleRef = ref<HTMLElement | null>(null);
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
    isFollowMode,
    isFollowPaused,
    setFile: loadFile,
    saveFile,
    changeViewMode,
    setEditorInstance,
    pauseFollowMode,
    resumeFollowMode,
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

  const { confirm } = useConfirm();

  const emit = defineEmits<{
    (e: 'ok'): void;
  }>();

  const isEditing = computed(
    () => readOnly.value === false && viewMode.value === 'full'
  );
  const modeTagText = computed(() =>
    isEditing.value
      ? t('app.file.editor.modeEditing')
      : t('app.file.editor.modeReadOnly')
  );
  const modeTagColor = computed(() =>
    isEditing.value ? 'rgb(var(--success-6))' : 'rgb(var(--warning-6))'
  );
  const modeHintText = computed(() =>
    isEditing.value
      ? t('app.file.editor.modeEditingHint')
      : t('app.file.editor.modeReadOnlyHint')
  );

  // 切换编辑模式
  const toggleEditMode = () => {
    if (viewMode.value !== 'full') {
      readOnly.value = true;
      return;
    }
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

  // 滚动事件处理函数 - 完全使用Vue的事件系统
  const handleScroll = (event: Event) => {
    const target = event.target as HTMLElement;
    if (target) {
      updateScrollPosition(target.scrollTop);
    }
  };

  // 在组件卸载时清理文件编辑器资源
  onUnmounted(() => {
    cleanup();
  });

  // ----- 编辑器相关函数 -----
  const handleEditorReady = (payload: { view: EditorView }) => {
    // 设置编辑器实例，从payload中获取view实例
    editorView.value = payload.view;
    setEditorInstance(payload.view);

    // 使用nextTick确保DOM更新完成
    nextTick(() => {
      if (editorView.value) {
        setEditorInstance(editorView.value);
      }
    });
  };

  const handleContentDoubleClick = () => {
    if (readOnly.value && viewMode.value === 'full') {
      readOnly.value = false;
    }
  };

  // ----- 内容管理相关函数 -----
  const handleViewModeChange = (mode: ContentViewMode, lines?: number) => {
    changeViewMode(mode, lines);
  };

  // ----- 弹窗控制函数 -----
  const handleCancel = async () => {
    // 仅当处于编辑模式且有未保存更改时才进行二次确认
    if (!readOnly.value && isEdited.value) {
      const result = await confirm({
        title: t('app.file.editor.unsavedChanges'),
        content: t('app.file.editor.confirmClose'),
        okText: t('common.confirm.okText'),
        cancelText: t('common.confirm.cancelText'),
      });
      if (!result) {
        return;
      }
    }

    // 确保停止实时追踪连接
    cleanup();

    // 无论如何关闭时都回到只读模式
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
      const success = await saveFile();
      // 只有保存成功时才触发ok事件关闭抽屉
      if (success) {
        emit('ok');
      }
      // 如果保存失败，saveFile内部已经显示了错误信息，这里不需要额外处理
    } catch (error) {
      console.error('保存文件失败:', error);
      // 保存失败时不关闭抽屉，让用户可以重试或查看错误信息
    }
  };

  // ----- 计算属性 -----
  // 是否是大文件（用于决定显示哪些按钮）
  const isLargeFile = computed(() => file.value && file.value.size > 100000);

  // ----- 监听器 -----
  // 监听视图模式变化
  watch(
    () => viewMode.value,
    (newMode) => {
      if (newMode !== 'full') {
        readOnly.value = true;
      }
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
  const setFile = (nextFile: FileItem) => {
    // 切换文件时默认回到只读，避免误改
    readOnly.value = true;
    loadFile(nextFile);
  };

  defineExpose({
    setFile,
    setReadOnly: (value: boolean) => {
      if (viewMode.value === 'full') {
        readOnly.value = value;
      } else {
        readOnly.value = true;
      }
    },
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
    height: 100%;
    padding: 0 16px 16px 16px; /* 进一步减小顶部padding */
    overflow: hidden auto; /* 启用垂直滚动 */
  }

  /* 确保Spin组件不影响布局 */
  .file-editor-container :deep(.arco-spin) {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  .file-editor-container :deep(.arco-spin-container) {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
  }

  /* 工具栏容器样式 */
  .toolbar-container {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    margin-bottom: 12px;
  }

  /* 响应式布局 - 在小屏幕上垂直堆叠 */
  @media (width <= 768px) {
    .toolbar-container {
      flex-direction: column;
      gap: 8px;
      align-items: stretch;
    }
  }

  /* 确保CodeEditor正确显示 */
  .file-editor-container :deep(.code-editor) {
    flex: 1;
    order: 0; /* 在工具栏之后显示 */
    min-height: 0; /* 允许flex收缩 */
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

  .title-path {
    max-width: min(70vw, 860px);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .editor-mode-banner {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 8px 10px;
    margin-bottom: 12px;
    font-size: 12px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .editor-mode-banner.is-readonly {
    color: rgb(var(--warning-7));
    background-color: rgb(var(--warning-1));
    border-color: rgb(var(--warning-3));
  }

  .editor-mode-banner.is-editing {
    color: rgb(var(--success-7));
    background-color: rgb(var(--success-1));
    border-color: rgb(var(--success-3));
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
    pointer-events: auto;
    cursor: col-resize;
    background: linear-gradient(to right, transparent, rgb(0 0 0 / 5%));
    transition: opacity 0.3s ease;
  }

  .resize-handle[style*='opacity: 0'] {
    pointer-events: none;
  }

  .resize-handle::after {
    width: 2px;
    height: 30px;
    content: '';
    background-color: rgb(0 0 0 / 20%);
    border-radius: 2px;
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
