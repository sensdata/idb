import { BaseEntity } from '@/types/global';

export interface HostGroupEntity extends BaseEntity {
  id: number;
  group_name: string;
  created_at: number;
}

export interface HostStatusDo {
  activated: boolean;
  cpu: number;
  disk: number;
  mem: number;
  mem_total: string;
  mem_used: string;
  rx: number;
  tx: number;
}

export interface HostEntity extends BaseEntity, HostStatusDo {
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

  agent_status: {
    status: string;
    connected: string;
  };

  agent_version: string;
  agent_latest: string;

  default?: boolean;
}
