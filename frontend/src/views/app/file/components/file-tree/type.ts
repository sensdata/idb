import { SimpleFileInfoEntity } from '@/entity/FileInfo';

/**
 * 文件树项类型定义
 * 扩展基础文件信息实体，添加文件树特有的属性
 */
export type FileTreeItem = SimpleFileInfoEntity & {
  /**
   * 是否展开状态（仅对文件夹有效）
   */
  open?: boolean;

  /**
   * 是否选中状态
   */
  selected?: boolean;

  /**
   * 是否加载中状态（仅对文件夹有效）
   */
  loading?: boolean;

  /**
   * 子项列表（仅对文件夹有效）
   */
  items?: FileTreeItem[];
};
