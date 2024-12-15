import { ScriptType } from '@/config/enum';
import { BaseEntity } from '@/types/global';

export interface ScriptEntity extends BaseEntity {
  id: number;
  type: ScriptType;
  name: string;
  category?: string;
  content: string;
  create_time: string;
  mod_time: string;
}
