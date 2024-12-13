import { FileInfoEntity } from '@/entity/FileInfo';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface FileListApiParams extends ApiListParams {
  path?: string;
  // todo: 接口差参数
  show_hidden?: boolean;
}
export function getFileListApi(params?: FileListApiParams) {
  return request
    .get<ApiListResult<FileInfoEntity>>('files/:host', params)
    .then((res: any) => {
      return {
        total: res?.item_total,
        items: res?.items,
        page: params?.page || 1,
        page_size: params?.page_size || 20,
      };
    });
}

export function getFileInfoApi(data: { path: string }) {
  return request.get<FileInfoEntity>('files/:host', {
    ...data,
    page: 1,
    page_size: 1,
  });
}

export function getFileSizeApi(data: { source: string }) {
  return request.get('files/:host/size', data);
}

export interface CreateFileParams {
  source: string; // 文件路径
  is_dir: boolean; // 是否是目录
  is_link?: boolean; // 是否是链接
  is_symlink?: boolean; // 是否是软链接
  link_path?: string; // 链接路径
  mode?: string; // 文件权限
}
export function createFileApi(data: CreateFileParams) {
  return request.post('files/:host', data);
}

// todo: api
// 1. permanently_delete待实现
export interface DeleteFileParams {
  source: string;
  force_delete: boolean;
  permanently_delete?: boolean;
  is_dir: boolean;
}
export function deleteFileApi(data: DeleteFileParams) {
  return request.delete('files/:host', data);
}

// todo: api
// 1. is_dir需要和路径一起
// 2. permanently_delete待实现
export interface BatchDeleteFileParams {
  force_delete?: boolean;
  permanently_delete?: boolean;
  sources: Array<{
    path: string;
    is_dir: boolean;
  }>;
}
export function batchDeleteFileApi(data: BatchDeleteFileParams) {
  return request.delete('files/:host/batch', data);
}

// todo: api
// 1. mode类型需要为string
export interface BatchUpdateRoleParams {
  group: string;
  user: string;
  mode: string;
  sources: string[];
  sub: boolean;
}
export function batchUpdateFileRoleApi(data: BatchUpdateRoleParams) {
  return request.put('files/:host/batch/role', data);
}

// todo: api
// 1. 缺少当前api， files/mode为多余
export interface BatchUpdateModeParams {
  mode: string;
  sources: string[];
  sub: boolean;
}
export function batchUpdateFileModeApi(data: BatchUpdateModeParams) {
  return request.put('files/:host/batch/mode', data);
}
export interface UpdateOwnerParams {
  group: string;
  user: string;
  source: string;
  sub: boolean;
}
export function updateFileOwnerApi(data: UpdateOwnerParams) {
  return request.put('files/:host/owner', data);
}

export interface CompressionParams {
  dst: string; // 输出目录
  files: string[]; // 文件路径
  name: string; // 压缩包名称
  replace: boolean; // 覆盖已有文件
  type: string; // 压缩类型
}
export function compressFilesApi(data: CompressionParams) {
  return request.post('files/:host/compress', data);
}

export interface DecompressionParams {
  dst: string; // 输出目录
  path: string; // 文件路径
  type: string; // 压缩类型
}
export function decompressFilesApi(data: DecompressionParams) {
  return request.post('files/:host/decompress', data);
}

export function getFileContentApi(data: { path: string }) {
  return request.get('files/:host/content', data);
}

export function updateFileContentApi(data: {
  source: string;
  content: string;
}) {
  return request.put('files/:host/content', data);
}

export function downloadFileApi(data: { source: string }) {
  return request.get('files/:host/download', data);
}

export function uploadFileApi(data: { dest: string; file: File }) {
  const formData = new FormData();
  formData.append('dest', data.dest);
  formData.append('file', data.file);

  return request.post('files/:host/upload', formData);
}

export function getFavoriteFilesApi(data: ApiListParams) {
  return request.get<ApiListResult<FileInfoEntity>>(
    'files/:host/favorites',
    data
  );
}

export function favoriteFileApi(data: { source: string }) {
  return request.post('files/:host/favorites', data);
}

export function unFavoriteFileApi(data: { id: number }) {
  return request.delete('files/:host/favorites', data);
}

// todo: 移动之前需要先检查目标文件是否存在
export interface MoveFileParams {
  dest: string;
  name?: string;
  sources: string[];
  cover: boolean;
  type: 'copy' | 'move';
}
export function moveFileApi(data: MoveFileParams) {
  return request.put('files/:host/move', data);
}

export interface RenameFileParams {
  name: string;
  source: string;
}
export function renameFileApi(data: RenameFileParams) {
  return request.put('files/:host/rename', data);
}
