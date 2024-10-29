import { GroupEntity } from './group';

export interface HostEntity {
  created_at: number;

  group: GroupEntity;
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

  is_default?: boolean;
  cpu_rate: number;
  memory_rate: number;
  disk_rate: number;
}
