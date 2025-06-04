import { reactive, toRefs } from 'vue';
import { SERVICE_TYPE } from '@/config/enum';
import { ServiceEntity } from '@/entity/Service';

interface DrawerParams {
  type: SERVICE_TYPE;
  category: string;
  name?: string;
  isEdit: boolean;
  record?: ServiceEntity;
}

// 表单数据类型
export interface FormDataType {
  name: string;
  category: string;
  parsedConfig: any;
  structuredConfig?: any;
  originalContent?: string;
  content?: string;
  rawContent?: string;
}

export function useServiceFormState() {
  // 使用reactive统一管理状态
  const state = reactive({
    visible: false,
    activeTab: 'form' as string,
    params: {
      type: SERVICE_TYPE.Local,
      category: '',
      isEdit: false,
    } as DrawerParams,
    formData: {
      name: '',
      category: '',
      parsedConfig: {},
    } as FormDataType,
    rawContent: '',
    hasChanges: false,
    originalRawContent: '',
  });

  // 重置所有状态
  const resetState = () => {
    state.visible = false;
    state.formData = {
      name: '',
      category: '',
      parsedConfig: {},
    };
    state.rawContent = '';
    state.hasChanges = false;
    state.originalRawContent = '';
    state.activeTab = 'form';
  };

  // 设置表单是否已变更
  const setFormChanged = (changed: boolean) => {
    state.hasChanges = changed;
  };

  return {
    ...toRefs(state),
    resetState,
    setFormChanged,
  };
}
