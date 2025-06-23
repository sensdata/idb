import { nextTick } from 'vue';
import { useLogger } from '@/composables/use-logger';
import type { FormData } from '../types';

export function useModeManager(
  activeMode: any, // Ref<'form' | 'raw'>
  previousMode: any, // Ref<'form' | 'raw'>
  generateRawContent: (formData: FormData) => void,
  parseRawContentToForm: (formData: FormData) => FormData | null,
  updateForm: (values: Partial<FormData>) => void,
  formData: FormData
) {
  const { log } = useLogger('ModeManager');

  // Handle mode switching
  const handleModeChange = async (mode: string | number) => {
    const modeStr = String(mode) as 'form' | 'raw';
    const currentMode = previousMode.value;

    log('🔄 模式切换:', { from: currentMode, to: modeStr });

    // If no actual change, return
    if (currentMode === modeStr) {
      log('⚠️ 没有真正的模式切换，跳过处理');
      return;
    }

    if (modeStr === 'raw' && currentMode === 'form') {
      // Switch from form to raw: generate raw content
      log('📝 从表单模式切换到文件模式');
      generateRawContent(formData);
    } else if (modeStr === 'form' && currentMode === 'raw') {
      // Switch from raw to form: parse raw content to form
      log('🔄 从文件模式切换到表单模式');

      const parsedData = parseRawContentToForm(formData);
      if (parsedData) {
        log('📝 更新表单数据');
        updateForm(parsedData);
        await nextTick();
      }
    }

    // Update current and previous modes
    previousMode.value = modeStr;
    activeMode.value = modeStr;
    log('✅ 模式已更新为:', activeMode.value);
  };

  return {
    handleModeChange,
  };
}
