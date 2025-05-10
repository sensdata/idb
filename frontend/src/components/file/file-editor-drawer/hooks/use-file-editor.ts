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
  const lineCount = ref(100); // Default line count
  const viewMode = ref<ContentViewMode>('full');
  const batchSize = ref(100); // 每次加载的行数, 从1000改为100
  const editorInstance = ref<any>(null);
  const eventSource = shallowRef<EventSource | null>(null);
  const isFollowMode = ref(false);

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
        setTimeout(() => {
          instance.scrollDOM.scrollTop = 0;
        }, 100);
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

    // 清空原有日志内容
    content.value = '';
    originalContent.value = '';

    try {
      // 创建 SSE 连接
      eventSource.value = connectFileTailFollowApi(hostId, filePath);

      // 处理日志事件
      eventSource.value.addEventListener('log', (event: Event) => {
        if (event instanceof MessageEvent) {
          // 将新内容添加到编辑器的顶部
          content.value += event.data;
          originalContent.value = content.value;

          // 自动保持在顶部位置
          if (editorInstance.value && editorInstance.value.scrollDOM) {
            setTimeout(() => {
              editorInstance.value.scrollDOM.scrollTop = 0;
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

      // 处理连接关闭事件
      eventSource.value.addEventListener('close', () => {
        // Empty listener
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
    lineCount.value = fileItem.line_count || 100;

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
            setTimeout(() => {
              editorInstance.value.scrollDOM.scrollTop = 0;
            }, 100);
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
        startFollowMode(file.value.path);
      }
      return;
    }

    try {
      // 即使是相同模式下改变行数，也始终显示加载状态
      setLoading(true);

      // 清空编辑器内容，以便显示加载状态
      content.value = '';

      // 如果提供了行数，更新行数
      if (lines !== undefined) {
        lineCount.value = lines;
      }

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
    isFollowMode,
    batchSize,
    changeViewMode,
    setFile,
    saveFile,
    setEditorInstance,
    startFollowMode,
    stopFollowMode,
    cleanup,
  };
}
