import { ref, computed, shallowRef } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import {
  updateFileContentApi,
  getFileTailApi,
  getFileHeadApi,
  connectFileTailFollowApi,
} from '@/api/file';
import { resolveApiUrl } from '@/helper/api-helper';
import useLoading from '@/hooks/loading';
import { useHostStore } from '@/store';
import {
  ContentViewMode,
  FileItem,
} from '@/components/file/file-editor-drawer/types';

export default function useFileEditor() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const hostStore = useHostStore();
  const file = ref<FileItem | null>(null);
  const content = ref('');
  const originalContent = ref('');
  const lineCount = ref(1000); // Default line count
  const viewMode = ref<ContentViewMode>('full');
  const isLoadingMore = ref(false);
  const canLoadMore = ref(true);
  const currentOffset = ref(0);
  const batchSize = ref(1000); // 每次加载的行数, 从500改为1000
  const editorInstance = ref<any>(null);
  const eventSource = shallowRef<EventSource | null>(null);
  const isFollowMode = ref(false);

  const isEdited = computed(() => {
    return content.value !== originalContent.value;
  });

  const isPartialView = computed(() => {
    return viewMode.value !== 'full';
  });

  // 检查是否需要加载更多内容
  const checkNeedLoadMore = (): void => {
    if (!file.value || !isPartialView.value) return;

    // 根据文件大小和当前已加载行数判断是否还有更多内容可加载
    const totalSize = file.value.size;
    const avgLineSize = 100; // 假设平均每行100字节
    const estimatedTotalLines = Math.ceil(totalSize / avgLineSize);

    // 如果当前行数接近预估总行数，则认为无法加载更多
    if (currentOffset.value >= estimatedTotalLines) {
      canLoadMore.value = false;
    } else {
      canLoadMore.value = true;
    }
  };

  // 加载更多尾部内容
  const loadMoreTail = async (): Promise<void> => {
    if (!file.value || isLoadingMore.value || !canLoadMore.value) return;

    try {
      isLoadingMore.value = true;

      // 计算新的起始偏移量
      const newOffset = currentOffset.value + batchSize.value;

      // 获取更多内容
      const tailData = await getFileTailApi({
        path: file.value.path,
        numbers: newOffset,
      });

      // 更新当前偏移量
      currentOffset.value = newOffset;

      // 从新内容中提取更多的行
      const currentLines = content.value.split('\n');
      const newLines = tailData.content.split('\n');

      // 检查是否还有更多内容可加载
      if (
        newLines.length <= currentLines.length ||
        newLines.length - currentLines.length < 10
      ) {
        canLoadMore.value = false;
        Message.info({
          content: t('app.file.editor.reachedStart'),
          duration: 2000,
        });
      } else {
        canLoadMore.value = true;

        // 如果成功加载了更多内容，显示提示
        Message.success({
          content: t('app.file.editor.loadedMore', {
            count: newLines.length - currentLines.length,
          }),
          duration: 1000,
        });
      }

      // 行反转处理
      const contentLines = tailData.content.split('\n');
      contentLines.reverse();

      // 更新内容
      content.value = contentLines.join('\n');
      originalContent.value = content.value;

      // 更新行数
      lineCount.value = newOffset;
    } catch (error) {
      console.error('Failed to load more content:', error);
      Message.error(t('app.file.editor.loadMoreFailed'));
    } finally {
      // 500ms后再设置isLoadingMore为false，避免太快看不到loading效果
      setTimeout(() => {
        isLoadingMore.value = false;
      }, 500);
    }
  };

  // 加载更多头部内容
  const loadMoreHead = async (): Promise<void> => {
    if (!file.value || isLoadingMore.value || !canLoadMore.value) return;

    try {
      isLoadingMore.value = true;

      // 计算新的起始偏移量
      const newOffset = currentOffset.value + batchSize.value;

      // 获取更多内容
      const headData = await getFileHeadApi({
        path: file.value.path,
        numbers: newOffset,
      });

      // 更新当前偏移量
      currentOffset.value = newOffset;

      // 检查是否还有更多内容可加载
      const currentLines = content.value.split('\n').length;
      const newLines = headData.content.split('\n').length;

      if (newLines <= currentLines || newLines - currentLines < 10) {
        canLoadMore.value = false;
        Message.info({
          content: t('app.file.editor.reachedEnd'),
          duration: 2000,
        });
      } else {
        canLoadMore.value = true;

        // 如果成功加载了更多内容，显示提示
        Message.success({
          content: t('app.file.editor.loadedMore', {
            count: newLines - currentLines,
          }),
          duration: 1000,
        });
      }

      // 更新内容
      content.value = headData.content;
      originalContent.value = headData.content;

      // 更新行数
      lineCount.value = newOffset;
    } catch (error) {
      console.error('Failed to load more content:', error);
      Message.error(t('app.file.editor.loadMoreFailed'));
    } finally {
      // 500ms后再设置isLoadingMore为false，避免太快看不到loading效果
      setTimeout(() => {
        isLoadingMore.value = false;
      }, 500);
    }
  };

  // 统一加载更多内容的函数，根据当前视图模式决定调用哪个具体函数
  const loadMoreContent = () => {
    if (viewMode.value === 'tail') {
      loadMoreTail();
    } else if (viewMode.value === 'head') {
      loadMoreHead();
    }
  };

  // 设置编辑器实例
  const setEditorInstance = (instance: any) => {
    editorInstance.value = instance;

    // 只在特定情况下进行初始化操作
    if (instance && isPartialView.value) {
      // 如果是tail模式，确保滚动到顶部
      if (viewMode.value === 'tail' && instance.scrollDOM) {
        setTimeout(() => {
          instance.scrollDOM.scrollTop = 0;
        }, 100);
      }

      // 检查是否需要展示"加载更多"按钮
      checkNeedLoadMore();
    }
  };

  // 新增: 停止文件跟踪模式
  const stopFollowMode = () => {
    if (eventSource.value) {
      eventSource.value.close();
      eventSource.value = null;
    }
    isFollowMode.value = false;

    // 如果正在跟踪模式，切换回尾部模式
    if (viewMode.value === 'follow') {
      viewMode.value = 'tail';
    }
  };

  // 新增: 启动文件跟踪模式
  const startFollowMode = (filePath: string) => {
    if (eventSource.value) {
      // 如果已存在连接，先关闭
      stopFollowMode();
    }

    // 获取当前主机ID
    const hostId = hostStore.currentId ?? hostStore.defaultId;
    if (!hostId) {
      Message.error(t('app.file.editor.noHostSelected'));
      return;
    }

    setLoading(true);
    isFollowMode.value = true;
    viewMode.value = 'follow' as ContentViewMode;

    try {
      // 创建 SSE 连接
      eventSource.value = connectFileTailFollowApi(hostId, filePath);

      // 处理日志事件
      eventSource.value.addEventListener('log', (event: Event) => {
        if (event instanceof MessageEvent) {
          // 追加新的内容到编辑器
          content.value += event.data;
          originalContent.value = content.value;

          // 自动滚动到底部
          if (editorInstance.value && editorInstance.value.scrollDOM) {
            setTimeout(() => {
              editorInstance.value.scrollDOM.scrollTop =
                editorInstance.value.scrollDOM.scrollHeight;
            }, 100);
          }
        }
      });

      // 处理心跳事件
      eventSource.value.addEventListener('heartbeat', () => {
        // Heartbeat received
      });

      // 处理状态事件
      eventSource.value.addEventListener('status', () => {
        // Status event received
      });

      // 处理错误
      eventSource.value.addEventListener('error', (event) => {
        console.error('File follow error:', event);
        Message.error(t('app.file.editor.followError'));
        stopFollowMode();
      });

      // 连接成功后
      eventSource.value.onopen = () => {
        setLoading(false);
        Message.success(t('app.file.editor.followStarted'));
      };
    } catch (error) {
      console.error('Failed to start follow mode:', error);
      Message.error(t('app.file.editor.followFailed'));
      stopFollowMode();
    }
  };

  const setFile = async (fileItem: FileItem) => {
    file.value = fileItem;

    // 重置状态
    canLoadMore.value = true;
    currentOffset.value = 0;

    // 始终先设置加载状态
    setLoading(true);

    // 处理显式的loading状态
    if (fileItem.loading || fileItem.content_view_mode === 'loading') {
      // 如果是加载状态，清空内容以显示加载效果
      content.value = '';
      originalContent.value = '';
      viewMode.value = 'loading';
      // 这里不返回，让外部调用完成后主动关闭loading
      return;
    }

    // Set the view mode from the file item
    if (fileItem.content_view_mode) {
      viewMode.value = fileItem.content_view_mode;
    } else if (fileItem.is_tail) {
      viewMode.value = 'tail';
    } else {
      viewMode.value = 'full';
    }

    // Set the line count from the file item or use default
    lineCount.value = fileItem.line_count || 1000;
    currentOffset.value = lineCount.value;

    try {
      // 如果文件大小为0，则直接设置为空内容，不下载文件
      if (fileItem.size === 0) {
        content.value = '';
        originalContent.value = '';
        return;
      }

      // 处理部分内容视图（head/tail）
      if (isPartialView.value && fileItem.content) {
        if (viewMode.value === 'tail' && fileItem.content) {
          // 反转行顺序，使最新的行显示在顶部
          const contentLines = fileItem.content.split('\n');
          contentLines.reverse();
          content.value = contentLines.join('\n');
          originalContent.value = content.value;

          // 确保重新加载文件后立即重置滚动位置到顶部
          if (editorInstance.value && editorInstance.value.scrollDOM) {
            setTimeout(() => {
              editorInstance.value.scrollDOM.scrollTop = 0;
            }, 100);
          }
        } else {
          content.value = fileItem.content;
          originalContent.value = fileItem.content;
        }

        // 检查是否需要显示加载更多按钮
        checkNeedLoadMore();

        return;
      }

      // 使用fetch直接调用下载API获取文件内容
      const apiUrl = resolveApiUrl('/files/{host}/download', {
        source: fileItem.path,
      });
      const response = await fetch(apiUrl);

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const fileContent = await response.text();
      content.value = fileContent;
      originalContent.value = fileContent;
    } catch (error) {
      Message.error(t('app.file.editor.loadFailed'));
      console.error('Failed to load file content:', error);
      content.value = '';
      originalContent.value = '';
    } finally {
      setLoading(false);
    }
  };

  // 更新视图模式
  const changeViewMode = async (mode: ContentViewMode, lines?: number) => {
    if (!file.value) return;

    // 如果之前在跟踪模式，停止跟踪
    if (viewMode.value === 'follow') {
      stopFollowMode();
    }

    // If switching to follow mode
    if (mode === 'follow') {
      if (file.value) {
        startFollowMode(file.value.path);
      }
      return;
    }

    // 重置状态
    canLoadMore.value = true;

    const newLineCount = lines || lineCount.value;
    currentOffset.value = newLineCount;

    try {
      setLoading(true);

      if (mode === 'full') {
        // Load full file content
        const apiUrl = resolveApiUrl('/files/{host}/download', {
          source: file.value.path,
        });
        const response = await fetch(apiUrl);

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const fileContent = await response.text();
        content.value = fileContent;
        originalContent.value = fileContent;
        viewMode.value = 'full';
      } else if (mode === 'head') {
        // Load head content
        const headData = await getFileHeadApi({
          path: file.value.path,
          numbers: newLineCount,
        });

        content.value = headData.content;
        originalContent.value = headData.content;
        viewMode.value = 'head';
        lineCount.value = newLineCount;

        // 检查是否需要显示加载更多按钮
        checkNeedLoadMore();
      } else if (mode === 'tail') {
        // Load tail content
        const tailData = await getFileTailApi({
          path: file.value.path,
          numbers: newLineCount,
        });

        // 反转行顺序，使最新的行显示在顶部
        const contentLines = tailData.content.split('\n');
        contentLines.reverse();
        const reversedContent = contentLines.join('\n');

        content.value = reversedContent;
        originalContent.value = reversedContent;
        viewMode.value = 'tail';
        lineCount.value = newLineCount;

        // 检查是否需要显示加载更多按钮
        checkNeedLoadMore();
      }

      // 确保滚动到顶部
      if (editorInstance.value && mode !== 'full') {
        setTimeout(() => {
          if (editorInstance.value.scrollDOM) {
            editorInstance.value.scrollDOM.scrollTop = 0;
          }
        }, 100);
      }
    } catch (error) {
      Message.error(t('app.file.editor.loadFailed'));
      console.error('Failed to change view mode:', error);
    } finally {
      setLoading(false);
    }
  };

  const saveFile = async () => {
    if (!file.value) {
      return false;
    }

    // 如果是部分内容视图，不允许保存
    if (isPartialView.value) {
      Message.warning(t('app.file.editor.partialViewNoSave'));
      return false;
    }

    try {
      setLoading(true);
      await updateFileContentApi({
        source: file.value.path,
        content: content.value,
      });

      originalContent.value = content.value;
      Message.success(t('app.file.editor.saveSuccess'));
      return true;
    } catch (error) {
      Message.error(t('app.file.editor.saveFailed'));
      console.error('Failed to save file:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  // 清理函数
  const cleanup = () => {
    // 关闭 SSE 连接
    stopFollowMode();

    // 其他现有清理逻辑...
  };

  return {
    file,
    content,
    loading,
    isEdited,
    viewMode,
    lineCount,
    isPartialView,
    isLoadingMore,
    canLoadMore,
    isFollowMode,
    loadMoreContent,
    changeViewMode,
    setFile,
    saveFile,
    setEditorInstance,
    startFollowMode,
    stopFollowMode,
    cleanup,
  };
}
