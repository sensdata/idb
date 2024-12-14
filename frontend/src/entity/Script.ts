import { ScriptType } from '@/config/enum';

export interface ScriptEntity {
  id: number;
  type: ScriptType;
  name: string;
  category?: string;
  create_time: string;
  mod_time: string;
}
