import { FileInfoEntity } from '@/entity/FileInfo';
import request from '@/helper/api-helper';
import { ApiListParams, ApiListResult } from '@/types/global';

export interface FileListApiParams extends ApiListParams {
  path?: string;
}
export function getFileListApi(params?: FileListApiParams) {
  return request.get<ApiListResult<FileInfoEntity>>('files', params);
}

// todo: api
// 1. 缺少path参数
// 2. 返回值和FileInfo不一致
export function getFileInfoApi(data: { path: string }) {
  return request.get<FileInfoEntity>('files/info', data);
}

export function getFileSizeApi(data: { source: string }) {
  return request.get('files/size', data);
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
  return request.post('files', data);
}

// todo: api
// 1. 确认force_delete是不是永久删除
export interface DeleteFileParams {
  source: string;
  force_delete: boolean;
  is_dir: boolean;
}
export function deleteFileApi(data: DeleteFileParams) {
  return request.delete('files', data);
}

// todo: api
// 1. is_dir需要和路径一起
// 2. 确认force_delete是不是永久删除
export interface BatchDeleteFileParams {
  force_delete?: boolean;
  sources: Array<{
    path: string;
    is_dir: boolean;
  }>;
}
export function batchDeleteFileApi(data: BatchDeleteFileParams) {
  return request.delete('files/batch', data);
}

// todo: api
// 1. mode类型需要为string
export interface BatchUpdateRoleParams {
  group: string;
  user: string;
  mode: string;
  sources: string[];
  sub: true;
}
export function batchUpdateFileRoleApi(data: BatchUpdateRoleParams) {
  return request.put('files/batch/role', data);
}

// todo: api
// 1. 缺少当前api， files/mode为多余
export interface BatchUpdateModeParams {
  mode: string;
  sources: string[];
  sub: true;
}
export function batchUpdateFileModeApi(data: BatchUpdateModeParams) {
  return request.put('files/batch/mode', data);
}

// todo: api
// 1. 缺少当前api， files/owner为多余
export interface BatchUpdateOwnerParams {
  group: string;
  user: string;
  sources: string[];
  sub: true;
}
export function batchUpdateFileOwnerApi(data: BatchUpdateOwnerParams) {
  return request.put('files/batch/owner', data);
}

export interface CompressionParams {
  dst: string; // 输出目录
  files: string[]; // 文件路径
  name: string; // 压缩包名称
  replace: boolean; // 覆盖已有文件
  type: string; // 压缩类型
}
export function compressFilesApi(data: CompressionParams) {
  return request.post('files/compress', data);
}

export interface DecompressionParams {
  dst: string; // 输出目录
  path: string; // 文件路径
  type: string; // 压缩类型
}
export function decompressFilesApi(data: DecompressionParams) {
  return request.post('files/decompress', data);
}

export function getFileContentApi(data: { path: string }) {
  return request.get('files/content', data);
}

export function updateFileContentApi(data: {
  source: string;
  content: string;
}) {
  return request.put('files/content', data);
}

export function downloadFileApi(data: { source: string }) {
  return request.get('files/download', data);
}

export function uploadFileApi(data: { dest: string; file: File }) {
  const formData = new FormData();
  formData.append('dest', data.dest);
  formData.append('file', data.file);

  return request.post('files/upload', formData);
}

export function getFavoriteFilesApi(data: ApiListParams) {
  return request.get<ApiListResult<FileInfoEntity>>('files/favorites', data);
}

export function favoriteFileApi(data: { source: string }) {
  return request.post('files/favorites', data);
}

export function unFavoriteFileApi(data: { id: number }) {
  return request.delete('files/favorites', data);
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
  return request.put('files/move', data);
}

export interface RenameFileParams {
  name: string;
  source: string;
}
export function renameFileApi(data: RenameFileParams) {
  return request.put('files/rename', data);
}
