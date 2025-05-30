import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { useApiWithLoading } from '@/hooks/use-api-with-loading';
import { LOGROTATE_TYPE } from '@/config/enum';

// API imports
import {
  createLogrotateRawApi,
  updateLogrotateContentApi,
  getLogrotateContentApi,
} from '@/api/logrotate';

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

  // 根据表单数据生成logrotate配置内容
  const generateLogrotateContentFromForm = (formData: FormData): string => {
    let content = `${formData.path} {\n`;

    // 添加基本选项
    content += `    ${formData.frequency}\n`;
    content += `    rotate ${formData.count}\n`;

    // 添加布尔选项
    if (formData.compress) {
      content += '    compress\n';

      // 延迟压缩选项
      if (formData.delayCompress) {
        content += '    delaycompress\n';
      }
    }

    // 创建新文件选项
    if (formData.create && formData.create.trim()) {
      const createValue = formData.create.startsWith('create ')
        ? formData.create
        : `create ${formData.create}`;
      content += `    ${createValue}\n`;
    }

    // 缺失日志文件不报错选项
    if (formData.missingOk) {
      content += '    missingok\n';
    }

    // 空日志文件不轮转选项
    if (formData.notIfEmpty) {
      content += '    notifempty\n';
    }

    // 前置命令
    if (formData.preRotate && formData.preRotate.trim()) {
      content += '    prerotate\n';
      content += `        ${formData.preRotate}\n`;
      content += '    endscript\n';
    }

    // 后置命令
    if (formData.postRotate && formData.postRotate.trim()) {
      content += '    postrotate\n';
      content += `        ${formData.postRotate}\n`;
      content += '    endscript\n';
    }

    content += '}';
    return content;
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
