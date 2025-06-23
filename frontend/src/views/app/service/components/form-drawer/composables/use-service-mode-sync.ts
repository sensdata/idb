import { Ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { Message } from '@arco-design/web-vue';
import { useLogger } from '@/composables/use-logger';
import { FormDataType } from './use-service-form-state';
import { useServiceParser } from './use-service-parser';
import ServiceForm from '../service-form.vue';
import ServiceRaw from '../service-raw.vue';

export function useServiceModeSync(
  formData: Ref<FormDataType>,
  rawContent: Ref<string>,
  activeTab: Ref<string>,
  originalRawContent: Ref<string>,
  serviceFormRef: Ref<InstanceType<typeof ServiceForm> | undefined>,
  serviceRawRef: Ref<InstanceType<typeof ServiceRaw> | undefined>,
  setFormChanged: (changed: boolean) => void
) {
  const { t } = useI18n();
  const { logDebug, logError } = useLogger('ServiceModeSync');

  // 检查表单是否有变更
  const checkFormChanges = async (originalContent: string) => {
    if (activeTab.value === 'form') {
      try {
        const currentFormData = await serviceFormRef.value?.getFormData();
        if (currentFormData) {
          // 比较当前生成的配置内容与原始内容
          setFormChanged(currentFormData.content !== originalContent);
        } else {
          setFormChanged(false);
        }
      } catch (error) {
        setFormChanged(false);
      }
    } else {
      // 原始模式：比较原始内容
      // 直接检查当前内容与原始内容是否不同
      setFormChanged(rawContent.value !== originalContent);
    }
  };

  // 从表单模式同步到原始模式
  const syncFormToRaw = async () => {
    try {
      const currentFormData = await serviceFormRef.value?.getFormData();
      if (currentFormData && currentFormData.content) {
        rawContent.value = currentFormData.content;

        // 检查变更状态
        await nextTick();
        // 使用存储的原始内容进行比较，而不是依赖formData中的属性
        await checkFormChanges(originalRawContent.value);
        return true;
      }
      return false;
    } catch (error) {
      Message.error(t('app.service.form.error.form_to_raw'));
      return false;
    }
  };

  // 从原始模式同步到表单模式
  const syncRawToForm = async () => {
    if (!rawContent.value) return false;

    try {
      logDebug('开始同步：从文件模式到表单模式');

      // 分析当前内容，逐行打印
      const lines = rawContent.value.split('\n');
      logDebug(`文件共有 ${lines.length} 行`);
      lines.forEach((line, index) => {
        if (line.trim().startsWith('Environment=')) {
          logDebug(`第 ${index + 1} 行 (Environment): "${line.trim()}"`);
        }
      });

      const originalRaw = originalRawContent.value;

      // 导入服务解析器
      const { parseServiceConfig, parseServiceConfigStructured } =
        useServiceParser();

      // 完整解析当前文件内容
      const parsedConfig = parseServiceConfig(rawContent.value);
      const structuredConfig = parseServiceConfigStructured(rawContent.value);

      logDebug('解析结果:', JSON.stringify(parsedConfig, null, 2));
      logDebug('环境变量值:', parsedConfig.environment);

      // 构建新的表单数据对象
      const newFormData = {
        // 保留原有的name和category
        name: formData.value?.name || '',
        category: formData.value?.category || '',
        // 使用新解析的配置
        parsedConfig: {
          ...parsedConfig,
          rawContent: rawContent.value, // 确保包含原始内容
        },
        // 包含结构化配置
        structuredConfig,
        // 保存原始内容用于比较
        originalContent: rawContent.value,
      };

      // 完全替换formData值以触发响应式更新
      formData.value = { ...newFormData };

      // 检查内容是否变化
      const hasChanged = rawContent.value !== originalRaw;
      setFormChanged(hasChanged);
      logDebug('同步完成，内容是否有变化:', hasChanged);

      return true;
    } catch (error) {
      logError('从文件模式同步到表单模式失败:', error);
      Message.error(t('app.service.form.error.parse'));
      return false;
    }
  };

  return {
    checkFormChanges,
    syncFormToRaw,
    syncRawToForm,
  };
}
