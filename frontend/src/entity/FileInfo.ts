export interface FileInfoEntity {
  content: string;
  extension: string;
  favorite_id: number;
  gid: string;
  group: string;
  user: string;
  is_dir: boolean;
  is_hidden: boolean;
  is_symlink: boolean;
  item_total: number;
  items?: FileInfoEntity[];
  link_path: string;
  mime_type: string;
  mod_time: string;
  mode: string;
  name: string;
  path: string;
  size: number;
  type: string;
  uid: string;
  update_time: string;
}
