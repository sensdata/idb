import FolderIcon from '@/assets/icons/color-folder.svg';
import FolderOpenIcon from '@/assets/icons/color-folder-open.svg';
import FileIcon from '@/assets/icons/drive-file.svg';
import { FileItem } from '../../types/file-item';

export default function IconRender(props: { item: FileItem }) {
  const { item } = props;
  if (item.is_dir) {
    if (item.open) {
      return <FolderOpenIcon />;
    }
    return <FolderIcon />;
  }

  return <FileIcon />;
}
