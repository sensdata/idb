import { BaseEntity } from '@/types/global';

export interface HostGroupEntity extends BaseEntity {
  id: number;
  group_name: string;
  created_at: number;
}
