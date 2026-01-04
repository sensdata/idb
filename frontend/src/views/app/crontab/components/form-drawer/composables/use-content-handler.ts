import { useI18n } from 'vue-i18n';
import { nextTick, Ref } from 'vue';
import { CRONTAB_TYPE, SCRIPT_TYPE } from '@/config/enum';
import { getScriptDetailApi } from '@/api/script';
import { PeriodDetailDo } from '@/entity/Crontab';
import { usePeriodUtils } from './use-period-utils';
import { useCronDescription } from './use-cron-description';
import { StateFlags } from './use-form-state';

interface ScriptDetailApiParams {
  name: string;
  category: string;
  type: SCRIPT_TYPE;
  host?: number;
}

// 表单状态接口
interface FormState {
  content: string;
  content_mode: 'direct' | 'script';
  period_details: PeriodDetailDo[];
  mark: string;
  type?: CRONTAB_TYPE;
  selectedScript?: string;
  selectedCategory?: string;
  command: string;
}

export const useContentHandler = () => {
  const { t } = useI18n();
  const { convertPeriodToCronExpression } = usePeriodUtils();
  const { getCronDescription } = useCronDescription();

  // 提取构建内容字符串的通用方法，减少代码重复
  const buildContentString = (
    formState: FormState,
    command: string
  ): string => {
    const user = (formState as any).user || 'root';
    if (!formState.period_details || formState.period_details.length === 0) {
      return '';
    }

    const cronExpression = convertPeriodToCronExpression(
      formState.period_details
    );
    // 使用 cronstrue 生成人类可读的周期描述
    const periodDescription = getCronDescription(cronExpression);
    const markPrefix = t('app.crontab.form.mark.prefix');
    const markLine = formState.mark ? `# ${markPrefix} ${formState.mark}` : '';

    let newContent = `# ${t(
      'app.crontab.period.execution_period'
    )}: ${periodDescription}`;
    if (markLine) {
      newContent += `\n${markLine}`;
    }
    newContent += `\n${cronExpression} ${user} ${command}`;

    return newContent;
  };

  // 提取从脚本获取命令的逻辑，减少代码重复
  const getScriptCommand = async (
    scriptName: string,
    scriptCategory: string,
    scriptType: SCRIPT_TYPE,
    scriptParams: string,
    currentHostId?: number
  ): Promise<string> => {
    try {
      // 获取脚本详细信息，以获取正确的source路径
      const scriptDetail = await getScriptDetailApi({
        name: scriptName,
        category: scriptCategory,
        type: scriptType,
        host: currentHostId,
      } as ScriptDetailApiParams);

      if (!scriptDetail || !scriptDetail.source) {
        console.error(
          `Failed to get script source path for "${scriptName}" in category "${scriptCategory}"`
        );
        return '';
      }

      // 使用脚本的source字段作为路径
      const scriptSourcePath = scriptDetail.source;

      return `${scriptSourcePath}${scriptParams ? ' ' + scriptParams : ''}`;
    } catch (error) {
      console.error(
        `Error getting script command for "${scriptName}" in category "${scriptCategory}":`,
        error
      );
      return '';
    }
  };

  // 更新表单内容的通用方法，减少重复代码
  const updateFormContent = (formState: FormState, command: string): void => {
    formState.command = command;
    formState.content = buildContentString(formState, command);
  };

  const updateContentWithParams = async (
    formState: FormState,
    selectedScriptSourceCategory: Ref<string | undefined>,
    selectedScript: Ref<string | undefined>,
    scriptParams: Ref<string>,
    currentHostId?: number
  ) => {
    if (!selectedScriptSourceCategory.value || !selectedScript.value) {
      return;
    }

    try {
      const command = await getScriptCommand(
        selectedScript.value,
        selectedScriptSourceCategory.value,
        formState.type === CRONTAB_TYPE.Global
          ? SCRIPT_TYPE.Global
          : SCRIPT_TYPE.Local,
        scriptParams.value,
        currentHostId
      );

      if (command) {
        updateFormContent(formState, command);
      }
    } catch (error) {
      console.error('Error updating content with script source:', error);
    }
  };

  const updateContentWithPeriod = async (
    formState: FormState,
    flags: StateFlags,
    forceUpdate = false
  ) => {
    if (flags.isInitialLoad.value) {
      flags.isInitialLoad.value = false;
    }

    const isScriptModeWithoutScript =
      formState.content_mode === 'script' &&
      (!formState.selectedScript || !formState.selectedCategory);

    if (!forceUpdate && !isScriptModeWithoutScript) {
      return;
    }

    await nextTick();

    const actualCommand =
      formState.content_mode === 'direct' && formState.command
        ? formState.command
        : '';

    if (actualCommand) {
      updateFormContent(formState, actualCommand);
    }

    await nextTick();
    if (flags.isUpdatingFromPeriod.value) {
      flags.isUpdatingFromPeriod.value = false;
    }
  };

  // 从表单内容提取命令的通用方法
  const extractCommandFromContent = (content: string): string => {
    if (!content) return '';

    const contentLines = content.split('\n');
    const lastLine = contentLines[contentLines.length - 1] || '';
    const cronParts = lastLine.trim().split(/\s+/);

    // 新格式：time user command
    if (cronParts.length >= 7) {
      // 从第7个部分开始是命令，索引5是用户
      return cronParts.slice(6).join(' ');
    }

    // 兼容旧格式：time command
    if (cronParts.length >= 6) {
      return cronParts.slice(5).join(' ');
    }

    return '';
  };

  /**
   * 更新备注信息的内容处理函数
   * 仅在输入框失去焦点时触发，确保稳定可靠的更新
   */
  const updateContentFromMark = async (formState: FormState) => {
    let actualCommand = '';

    if (formState.content_mode === 'direct') {
      actualCommand = formState.command || '';
    } else if (formState.content_mode === 'script') {
      // 脚本模式下使用当前已经选择的脚本路径
      actualCommand = formState.content
        ? extractCommandFromContent(formState.content)
        : '';
    }

    if (actualCommand) {
      updateFormContent(formState, actualCommand);
    }
  };

  const updateMarkInScriptMode = async (
    formState: FormState,
    mark: string,
    selectedScriptSourceCategory: Ref<string | undefined>,
    selectedScript: Ref<string | undefined>,
    scriptParams: Ref<string>,
    currentHostId?: number
  ) => {
    if (!selectedScriptSourceCategory.value || !selectedScript.value) {
      return;
    }

    // 更新标记后重新生成内容
    formState.mark = mark;
    await updateContentWithParams(
      formState,
      selectedScriptSourceCategory,
      selectedScript,
      scriptParams,
      currentHostId
    );
  };

  return {
    updateContentWithPeriod,
    updateMarkInScriptMode,
    updateContentWithParams,
    updateContentFromMark,
  };
};
