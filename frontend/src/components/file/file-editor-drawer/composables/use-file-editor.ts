import { ref, computed, shallowRef, nextTick } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import {
  updateFileContentApi,
  getFileTailApi,
  getFileHeadApi,
  connectFileTailFollowApi,
} from '@/api/file';
import { resolveApiUrl } from '@/helper/api-helper';
import useLoading from '@/composables/loading';
import { useLogger } from '@/composables/use-logger';
import { useHostStore } from '@/store';
import {
  ContentViewMode,
  FileItem,
} from '@/components/file/file-editor-drawer/types';

export default function useFileEditor() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const { logError } = useLogger('FileEditor');
  const hostStore = useHostStore();
  const file = ref<FileItem | null>(null);
  const content = ref('');
  const originalContent = ref('');
  const lineCount = ref(100); // Default line count
  const viewMode = ref<ContentViewMode>('full');
  const batchSize = ref(100); // 每次加载的行数, 从1000改为100
  const editorInstance = ref<any>(null);
  const eventSource = shallowRef<EventSource | null>(null);
  const isFollowMode = ref(false);
  const isFollowPaused = ref(false);

  const isEdited = computed(() => {
    return content.value !== originalContent.value;
  });

  const isPartialView = computed(() => {
    return viewMode.value !== 'full';
  });

  // 设置编辑器实例
  const setEditorInstance = (instance: any) => {
    editorInstance.value = instance;

    // 只在特定情况下进行初始化操作
    if (instance && isPartialView.value) {
      // 如果是tail模式，确保滚动到顶部
      if (viewMode.value === 'tail' && instance.scrollDOM) {
        nextTick(() => {
          instance.scrollDOM.scrollTop = 0;
        });
      }
    }
  };

  // 新增: 停止文件跟踪模式
  const stopFollowMode = () => {
    if (eventSource.value) {
      eventSource.value.close();
      eventSource.value = null;
    }
    isFollowMode.value = false;
    isFollowPaused.value = false;

    // 如果正在跟踪模式，切换回尾部模式
    if (viewMode.value === 'follow') {
      viewMode.value = 'tail';
    }
  };

  // 暂停追踪
  const pauseFollowMode = () => {
    isFollowPaused.value = true;
  };

  // 恢复追踪
  const resumeFollowMode = () => {
    isFollowPaused.value = false;

    // 恢复时自动滚动到底部显示最新内容
    if (editorInstance.value && editorInstance.value.scrollDOM) {
      nextTick(() => {
        const scrollDOM = editorInstance.value.scrollDOM;
        scrollDOM.scrollTop = scrollDOM.scrollHeight;
      });
    }
  };

  // 新增: 启动文件跟踪模式
  const startFollowMode = async (filePath: string) => {
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

    // 设置follow模式的默认行数
    lineCount.value = 10;

    try {
      // 先获取文件尾部10行作为初始内容
      const tailData = await getFileTailApi({
        path: filePath,
        numbers: 10,
      });

      // 设置初始内容
      content.value = tailData.content;
      originalContent.value = tailData.content;

      // 创建 SSE 连接
      eventSource.value = connectFileTailFollowApi(hostId, filePath);

      // 处理日志事件
      eventSource.value.addEventListener('log', (event: Event) => {
        if (event instanceof MessageEvent && !isFollowPaused.value) {
          // 只有在未暂停时才更新内容
          content.value += event.data;
          originalContent.value = content.value;

          // 自动滚动到底部显示最新内容
          if (editorInstance.value && editorInstance.value.scrollDOM) {
            nextTick(() => {
              const scrollDOM = editorInstance.value.scrollDOM;
              scrollDOM.scrollTop = scrollDOM.scrollHeight;
            });
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

      // 处理连接关闭事件
      eventSource.value.addEventListener('close', () => {
        // Empty listener
      });

      // 处理错误
      eventSource.value.addEventListener('error', (event) => {
        logError('File follow error:', event);
        // 显示更详细的错误信息
        let errorMessage = t('app.file.editor.followError');
        if (event instanceof ErrorEvent && event.message) {
          errorMessage = `${t('app.file.editor.followError')}: ${
            event.message
          }`;
        }
        Message.error(errorMessage);
        stopFollowMode();
        setLoading(false); // 确保停止加载状态
      });

      // 连接成功后
      eventSource.value.onopen = () => {
        setLoading(false);
        Message.success(t('app.file.editor.followStarted'));
      };
    } catch (error) {
      logError('Failed to start follow mode:', error);
      // 显示更详细的错误信息
      let errorMessage = t('app.file.editor.followFailed');
      if (error instanceof Error) {
        errorMessage = `${t('app.file.editor.followFailed')}: ${error.message}`;
      }
      Message.error(errorMessage);
      stopFollowMode();
      setLoading(false); // 确保在错误时也停止加载状态
    }
  };

  const setFile = async (fileItem: FileItem) => {
    file.value = fileItem;

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

    // Set the line count from the file item or use default based on view mode
    if (fileItem.content_view_mode === 'tail') {
      lineCount.value = fileItem.line_count || 30; // tail模式默认30行
    } else {
      lineCount.value = fileItem.line_count || 100; // head模式默认100行
    }

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
          // 不再需要反转行顺序
          content.value = fileItem.content;
          originalContent.value = fileItem.content;

          // 确保重新加载文件后立即重置滚动位置到顶部
          if (editorInstance.value && editorInstance.value.scrollDOM) {
            nextTick(() => {
              editorInstance.value.scrollDOM.scrollTop = 0;
            });
          }
        } else {
          content.value = fileItem.content;
          originalContent.value = fileItem.content;
        }
        return;
      }

      // 使用fetch直接调用下载API获取文件内容
      const apiUrl = resolveApiUrl('/files/{host}/download', {
        source: fileItem.path,
      });
      const response = await fetch(apiUrl);

      if (!response.ok) {
        // 尝试解析错误响应
        let errorMessage = t('app.file.editor.loadFailed');
        try {
          const errorData = await response.json();
          if (errorData.message) {
            errorMessage = `${t('app.file.editor.loadFailed')}: ${
              errorData.message
            }`;
          }
        } catch {
          // 如果无法解析JSON，使用HTTP状态信息
          errorMessage = `${t('app.file.editor.loadFailed')}: HTTP ${
            response.status
          } ${response.statusText}`;
        }
        throw new Error(errorMessage);
      }

      const fileContent = await response.text();
      content.value = fileContent;
      originalContent.value = fileContent;
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : t('app.file.editor.loadFailed');
      Message.error(errorMessage);
      logError('Failed to load file content:', error);
      content.value = '';
      originalContent.value = '';
    } finally {
      setLoading(false);
    }
  };

  // 更新视图模式 - 修改为支持单独切换模式和单独设置行数
  const changeViewMode = async (mode: ContentViewMode, lines?: number) => {
    if (!file.value) return;

    // 如果之前在跟踪模式，停止跟踪
    if (viewMode.value === 'follow') {
      stopFollowMode();
    }

    // 如果切换到跟踪模式
    if (mode === 'follow') {
      if (file.value) {
        await startFollowMode(file.value.path);
      }
      return;
    }

    try {
      // 即使是相同模式下改变行数，也始终显示加载状态
      setLoading(true);

      // 清空编辑器内容，以便显示加载状态
      content.value = '';

      // 如果提供了行数，更新行数；否则使用对应模式的默认行数
      if (lines !== undefined) {
        lineCount.value = lines;
      } else if (mode === 'tail') {
        lineCount.value = 30; // tail模式默认30行
      } else if (mode === 'head') {
        lineCount.value = 100; // head模式默认100行
      }

      if (mode === 'full') {
        // Load full file content
        const apiUrl = resolveApiUrl('/files/{host}/download', {
          source: file.value.path,
        });
        const response = await fetch(apiUrl);

        if (!response.ok) {
          // 尝试解析错误响应
          let errorMessage = t('app.file.editor.loadFailed');
          try {
            const errorData = await response.json();
            if (errorData.message) {
              errorMessage = `${t('app.file.editor.loadFailed')}: ${
                errorData.message
              }`;
            }
          } catch {
            errorMessage = `${t('app.file.editor.loadFailed')}: HTTP ${
              response.status
            } ${response.statusText}`;
          }
          throw new Error(errorMessage);
        }

        const fileContent = await response.text();
        content.value = fileContent;
        originalContent.value = fileContent;
        viewMode.value = 'full';
      } else if (mode === 'head') {
        // Load head content with specified line count
        const headData = await getFileHeadApi({
          path: file.value.path,
          numbers: lineCount.value,
        });

        content.value = headData.content;
        originalContent.value = headData.content;
        viewMode.value = 'head';
      } else if (mode === 'tail') {
        // Load tail content with specified line count
        const tailData = await getFileTailApi({
          path: file.value.path,
          numbers: lineCount.value,
        });

        content.value = tailData.content;
        originalContent.value = tailData.content;
        viewMode.value = 'tail';
      }

      // 确保滚动到顶部
      if (editorInstance.value && mode !== 'full') {
        nextTick(() => {
          if (editorInstance.value.scrollDOM) {
            editorInstance.value.scrollDOM.scrollTop = 0;
          }
        });
      }
    } catch (error) {
      const errorMessage =
        error instanceof Error
          ? error.message
          : t('app.file.editor.loadFailed');
      Message.error(errorMessage);
      logError('Failed to change view mode:', error);
      // 发生错误时，恢复之前的内容和视图模式
      // 这样用户可以继续使用之前的内容，而不是看到空白
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
      // 显示更详细的错误信息
      let errorMessage = t('app.file.editor.saveFailed');
      if (error instanceof Error) {
        errorMessage = `${t('app.file.editor.saveFailed')}: ${error.message}`;
      }
      Message.error(errorMessage);
      logError('Failed to save file:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  // 清理函数
  const cleanup = () => {
    // 关闭 SSE 连接
    stopFollowMode();
  };

  return {
    file,
    content,
    loading,
    isEdited,
    viewMode,
    lineCount,
    isPartialView,
    isFollowMode,
    isFollowPaused,
    batchSize,
    changeViewMode,
    setFile,
    saveFile,
    setEditorInstance,
    startFollowMode,
    stopFollowMode,
    pauseFollowMode,
    resumeFollowMode,
    cleanup,
  };
}
