import { Ref, unref } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useClipboard } from '@/composables/use-clipboard';
import { moveFileApi, getFileDetailApi } from '@/api/file';
import { resolveApiUrl } from '@/helper/api-helper';
import { openWindow } from '@/utils';

// Import component types
import ModeDrawer from '@/components/file/mode-drawer/index.vue';
import OwnerDrawer from '@/components/file/owner-drawer/index.vue';
import CreateFileDrawer from '@/components/file/create-file-drawer/index.vue';
import CreateFolderDrawer from '@/components/file/create-folder-drawer/index.vue';
import CompressDrawer from '@/components/file/compress-drawer/index.vue';
import DecompressDrawer from '@/components/file/decompress-drawer/index.vue';
import DeleteFileModal from '@/components/file/delete-file-modal/index.vue';
import UploadFilesDrawer from '@/components/file/upload-files-drawer/index.vue';
import RenameDrawer from '@/components/file/rename-drawer/index.vue';
import PropertyDrawer from '@/components/file/property-drawer/index.vue';
import FileEditorDrawer from '@/components/file/file-editor-drawer/index.vue';
import { FileItem } from '@/components/file/file-editor-drawer/types';

interface FileStore {
  pwd: string;
  selected: FileItem[];
  clipboardItems: FileItem[];
  cutActive: boolean;
  clearSelected: () => void;
  handleOpen: (record: FileItem) => void;
  handleGoto: (path: string) => void;
  handleCopy: () => void;
  handleCut: () => void;
}

interface FileOperationsParams {
  store: FileStore;
  setLoading: (loading: boolean) => void;
  modeDrawerRef: Ref<InstanceType<typeof ModeDrawer> | undefined>;
  ownerDrawerRef: Ref<InstanceType<typeof OwnerDrawer> | undefined>;
  createFileDrawerRef: Ref<InstanceType<typeof CreateFileDrawer> | undefined>;
  createFolderDrawerRef: Ref<
    InstanceType<typeof CreateFolderDrawer> | undefined
  >;
  compressDrawerRef: Ref<InstanceType<typeof CompressDrawer> | undefined>;
  decompressDrawerRef: Ref<InstanceType<typeof DecompressDrawer> | undefined>;
  deleteFileModalRef: Ref<InstanceType<typeof DeleteFileModal> | undefined>;
  uploadFilesDrawerRef: Ref<InstanceType<typeof UploadFilesDrawer> | undefined>;
  renameDrawerRef?: Ref<InstanceType<typeof RenameDrawer> | undefined>;
  propertyDrawerRef?: Ref<InstanceType<typeof PropertyDrawer> | undefined>;
  fileEditorDrawerRef?: Ref<InstanceType<typeof FileEditorDrawer> | undefined>;
  openTerminal?: () => void;
  selected: Ref<FileItem[]>;
}

export const useFileOperations = (params: FileOperationsParams) => {
  const { t } = useI18n();
  const { copyText } = useClipboard();
  const {
    store,
    setLoading,
    modeDrawerRef,
    ownerDrawerRef,
    createFileDrawerRef,
    createFolderDrawerRef,
    compressDrawerRef,
    decompressDrawerRef,
    deleteFileModalRef,
    uploadFilesDrawerRef,
    renameDrawerRef,
    propertyDrawerRef,
    fileEditorDrawerRef,
    openTerminal,
    selected,
  } = params;

  // 打开文件的函数
  const openFileInEditor = async (fileOrPath: FileItem | string) => {
    try {
      const filePath =
        typeof fileOrPath === 'string' ? fileOrPath : fileOrPath.path;

      // 提取目录路径
      const lastSlashIndex = filePath.lastIndexOf('/');
      const dirPath =
        lastSlashIndex > 0 ? filePath.substring(0, lastSlashIndex) : '/';

      // 如果不在当前目录，先导航
      if (store.pwd !== dirPath) {
        store.handleGoto(dirPath);
      }

      // 获取文件详情
      const fileDetail = await getFileDetailApi({
        path: filePath,
        expand: false,
      });

      if (!fileDetail) return;

      // 根据文件大小决定打开方式
      if (fileDetail.size > 1048576) {
        // 大文件下载
        Message.info(t('app.file.list.message.largeFileDownload'));
        const downloadUrl = resolveApiUrl('/files/{host}/download', {
          source: fileDetail.path,
        });
        openWindow(downloadUrl, { download: fileDetail.name });
      } else if (fileEditorDrawerRef?.value) {
        // 打开编辑器
        fileEditorDrawerRef.value.setFile(fileDetail);
        fileEditorDrawerRef.value.show();
      }
    } catch (error) {
      console.error('File open error:', error);
      Message.error(t('app.file.list.message.fileOpenFailed'));
    }
  };

  const handleModifyMode = (record: FileItem) => {
    modeDrawerRef.value?.setData(record);
    modeDrawerRef.value?.show();
  };

  const handleModifyOwner = (record: FileItem) => {
    ownerDrawerRef.value?.setData(record);
    ownerDrawerRef.value?.show();
  };

  const handleCreate = (key: string) => {
    switch (key) {
      case 'createFolder':
        createFolderDrawerRef.value?.setData({
          pwd: store.pwd,
        });
        createFolderDrawerRef.value?.show();
        break;
      case 'createFile':
        createFileDrawerRef.value?.setData({
          pwd: store.pwd,
        });
        createFileDrawerRef.value?.show();
        break;
      default:
        break;
    }
  };

  const handleCopy = () => {
    if (!selected.value?.length) {
      return;
    }
    store.handleCopy();
  };

  const handleCut = () => {
    if (!selected.value?.length) {
      return;
    }
    store.handleCut();
  };

  const handlePaste = async () => {
    if (!store.clipboardItems.length) {
      return false;
    }

    setLoading(true);
    try {
      await moveFileApi({
        sources: store.clipboardItems.map((item: FileItem) => item.path),
        dest: store.pwd,
        cover: false,
        type: store.cutActive ? 'cut' : 'copy',
      });
      store.clearSelected();
      return true;
    } catch (error) {
      console.error('Paste error:', error);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleUpload = () => {
    uploadFilesDrawerRef.value?.show();
    uploadFilesDrawerRef.value?.setData({
      directory: store.pwd,
    });
  };

  const handleTerminal = () => {
    openTerminal?.();
  };

  const handleBatchCompress = () => {
    if (!selected.value?.length) {
      return;
    }
    compressDrawerRef.value?.setFiles(unref(selected));
    compressDrawerRef.value?.show();
  };

  const handleBatchDecompress = () => {
    if (!selected.value?.length) {
      return;
    }
    decompressDrawerRef.value?.setFiles(unref(selected));
    decompressDrawerRef.value?.show();
  };

  const handleBatchDelete = () => {
    if (!selected.value?.length) {
      return;
    }
    deleteFileModalRef.value?.setData(unref(selected));
    deleteFileModalRef.value?.show();
  };

  const handleRename = (record: FileItem) => {
    if (renameDrawerRef?.value) {
      renameDrawerRef.value.setData({
        path: record.path,
      });
      renameDrawerRef.value.show();
    } else {
      modeDrawerRef.value?.setData(record);
      modeDrawerRef.value?.show();
    }
  };

  const handleCopyPath = async (record: FileItem) => {
    try {
      await copyText(record.path);
      Message.success(t('app.file.list.message.copyPathSuccess'));
    } catch (err) {
      Message.error(t('app.file.list.message.copyPathFailed'));
    }
  };

  const handleProperty = (record: FileItem) => {
    if (propertyDrawerRef?.value) {
      propertyDrawerRef.value.setData(record);
      propertyDrawerRef.value.show();
    } else {
      modeDrawerRef.value?.setData(record);
      modeDrawerRef.value?.show();
    }
  };

  const handleDelete = (record: FileItem) => {
    deleteFileModalRef.value?.setData([record]);
    deleteFileModalRef.value?.show();
  };

  const handleDownload = (record: FileItem) => {
    const a = document.createElement('a');
    a.href = resolveApiUrl('/files/{host}/download', { source: record.path });
    a.download = record.name;
    a.click();
  };

  const handleOperation = (key: string, record: FileItem) => {
    switch (key) {
      case 'open':
        if (record.is_dir) {
          store.handleOpen(record);
        } else {
          openFileInEditor(record);
        }
        break;
      case 'mode':
        handleModifyMode(record);
        break;
      case 'rename':
        handleRename(record);
        break;
      case 'copyPath':
        handleCopyPath(record);
        break;
      case 'download':
        handleDownload(record);
        break;
      case 'property':
        handleProperty(record);
        break;
      case 'delete':
        handleDelete(record);
        break;
      default:
        break;
    }
  };

  return {
    handleModifyMode,
    handleModifyOwner,
    handleCreate,
    handleCopy,
    handleCut,
    handlePaste,
    handleUpload,
    handleTerminal,
    handleBatchCompress,
    handleBatchDecompress,
    handleBatchDelete,
    handleRename,
    handleCopyPath,
    handleProperty,
    handleDelete,
    handleDownload,
    handleOperation,
    openFileInEditor,
  };
};
