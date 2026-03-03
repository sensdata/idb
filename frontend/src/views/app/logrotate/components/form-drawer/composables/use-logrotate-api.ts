import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useApiWithLoading } from '@/composables/use-api-with-loading';
import { LOGROTATE_TYPE } from '@/config/enum';

// API imports
import {
  createLogrotateRawApi,
  updateLogrotateContentApi,
  getLogrotateContentApi,
} from '@/api/logrotate';
import { generateLogrotateContentFromForm } from '../../../utils/content';

import type { FormData } from '../types';

export function useLogrotateApi(setLoading: (loading: boolean) => void) {
  const { t } = useI18n();
  const { executeApi } = useApiWithLoading(setLoading);

  // 加载日志轮转内容
  const loadContent = async (
    type: LOGROTATE_TYPE,
    category: string,
    name: string,
    hostId: number
  ) => {
    return executeApi(
      async () => {
        return getLogrotateContentApi(type, category, name, hostId);
      },
      {
        onError: (error) => {
          const errorMessage =
            error instanceof Error ? error.message : 'Unknown error';
          Message.error(
            `${t('app.logrotate.form.load_content_failed')}: ${errorMessage}`
          );
          return '';
        },
      }
    );
  };

  // 提交表单数据
  const submitLogrotate = async (
    activeMode: 'form' | 'raw',
    formData: FormData,
    rawContent: string,
    isEdit: boolean,
    currentType: LOGROTATE_TYPE,
    originalCategory: string,
    originalName: string,
    hostId: number
  ) => {
    return executeApi(
      async () => {
        if (activeMode === 'form') {
          // 即使是表单模式，也通过raw接口提交
          // 根据表单数据生成原始配置内容
          const generatedContent = generateLogrotateContentFromForm(formData);

          if (isEdit) {
            // 编辑现有配置
            await updateLogrotateContentApi(
              currentType,
              originalCategory,
              originalName,
              generatedContent,
              hostId
            );
          } else {
            // 创建新配置
            await createLogrotateRawApi(
              currentType,
              formData.category,
              formData.name,
              generatedContent,
              hostId
            );
          }
        } else if (isEdit) {
          // 文件模式编辑
          await updateLogrotateContentApi(
            currentType,
            originalCategory,
            originalName,
            rawContent,
            hostId
          );
        } else {
          // 文件模式创建
          await createLogrotateRawApi(
            currentType,
            formData.category,
            formData.name,
            rawContent,
            hostId
          );
        }

        return isEdit
          ? t('app.logrotate.form.update_success')
          : t('app.logrotate.form.create_success');
      },
      {
        onError: (error: any) => {
          const errorMessage =
            error instanceof Error ? error.message : String(error);
          Message.error(
            `${
              isEdit
                ? t('app.logrotate.form.update_failed')
                : t('app.logrotate.form.create_failed')
            }: ${errorMessage}`
          );
          throw error;
        },
      }
    );
  };

  return {
    loadContent,
    submitLogrotate,
  };
}
