import { FileInfoEntity } from '@/entity/FileInfo';

export type FileItem = FileInfoEntity & {
  open?: boolean;
  selected?: boolean;
};
