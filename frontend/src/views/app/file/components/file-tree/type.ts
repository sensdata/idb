import { SimpleFileInfoEntity } from '@/entity/FileInfo';

export type FileTreeItem = SimpleFileInfoEntity & {
  open?: boolean;
  selected?: boolean;
  loading?: boolean;
  items?: FileTreeItem[];
};
