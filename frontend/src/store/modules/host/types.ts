import { HostEntity } from '@/entity/host';

export interface HostState {
  currentId?: number;
  current?: HostEntity;
  items: HostEntity[];
}
