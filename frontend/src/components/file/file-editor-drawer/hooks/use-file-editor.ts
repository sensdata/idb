import { ref, computed } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { updateFileContentApi } from '@/api/file';
import { resolveApiUrl } from '@/helper/api-helper';
import { FileItem } from '@/views/app/file/types/file-item';
import useLoading from '@/hooks/loading';

export default function useFileEditor() {
  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);
  const file = ref<FileItem | null>(null);
  const content = ref('');
  const originalContent = ref('');

  const isEdited = computed(() => {
    return content.value !== originalContent.value;
  });

  const setFile = async (fileItem: FileItem) => {
    file.value = fileItem;

    try {
      setLoading(true);

      // 如果文件大小为0，则直接设置为空内容，不下载文件
      if (fileItem.size === 0) {
        content.value = '';
        originalContent.value = '';
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

  const saveFile = async () => {
    if (!file.value) {
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

  return {
    file,
    content,
    loading,
    isEdited,
    setFile,
    saveFile,
  };
}
