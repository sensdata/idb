import { BaseEntity } from '@/types/global';
import { HostGroupEntity } from './Group';

export interface HostStatusVo {
  cpu: number;
  disk: number;
  mem: number;
  mem_total: string;
  mem_used: string;
  rx: number;
  tx: number;
}

export interface HostEntity extends BaseEntity, HostStatusVo {
  created_at: number;

  group: HostGroupEntity;
  name: string;
  addr: string;
  port: number;
  user: string;
  auth_mode: string;
  password: string;
  private_key: string;
  pass_phrase: string;

  agent_addr: string;
  agent_port: number;
  agent_mode: string;
  agent_key: string;

  // todo: 需要新增
  is_default?: boolean;
}
