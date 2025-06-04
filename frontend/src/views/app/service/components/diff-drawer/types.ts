import { SERVICE_TYPE } from '@/config/enum';

export interface DiffParams {
  type: SERVICE_TYPE;
  category: string;
  name: string;
  commit: string;
}

export type RestoreSuccessCallback = () => void;

export interface DiffDrawerExpose {
  show: (params: DiffParams, onRestoreSuccess?: RestoreSuccessCallback) => void;
}
