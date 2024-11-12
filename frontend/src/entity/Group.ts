import { BaseEntity } from '@/types/global';

export interface GroupEntity extends BaseEntity {
  created_at: number;
  group_name: string;
}
