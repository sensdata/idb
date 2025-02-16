import { FileInfoEntity } from '@/entity/FileInfo';

export enum FileSelectType {
  FILE = 'file',
  DIR = 'dir',
  ALL = 'all',
}

export type FileItem = Pick<FileInfoEntity, 'name' | 'path' | 'is_dir'>;

export interface FileSelectProps {
  modelValue?: string;
  placeholder?: string;
  disabled?: boolean;
  readonly?: boolean;
  error?: boolean;
  allowCreate?: boolean;
  initialPath?: string;
  type?: FileSelectType | string;
}

export interface FileSelectEmits {
  (e: 'update:modelValue', value: string): void;
  (e: 'change', value: string): void;
}
