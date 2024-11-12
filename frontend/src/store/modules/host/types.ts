import { HostEntity } from '@/entity/Host';

export interface HostState {
  currentId?: number;
  current?: HostEntity;
  items: HostEntity[];
}
