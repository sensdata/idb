import { FileInfoEntity } from '@/entity/FileInfo';

export type ContentViewMode = 'full' | 'head' | 'tail' | 'loading' | 'follow';

export type FileItem = FileInfoEntity & {
  open?: boolean;
  selected?: boolean;
  loading?: boolean;
  is_tail?: boolean; // Deprecated: use content_view_mode instead
  content_view_mode?: ContentViewMode; // Viewing mode: full, head, or tail
  line_count?: number; // Number of lines to view in head/tail mode
};
