import { useI18n } from 'vue-i18n';
import { nextTick, Ref } from 'vue';
import { CRONTAB_TYPE, SCRIPT_TYPE } from '@/config/enum';
import { getScriptDetailApi } from '@/api/script';
import { usePeriodUtils } from './use-period-utils';
import { StateFlags } from './use-form-state';

// 表单状态接口
interface FormState {
  content: string;
  content_mode: 'direct' | 'script';
  period_details: any[];
  mark: string;
  type?: CRONTAB_TYPE;
  selectedScript?: string;
  selectedCategory?: string;
}

export const useContentHandler = () => {
  const { t } = useI18n();
  const { convertPeriodToCronExpression, generateFormattedPeriodComment } =
    usePeriodUtils();

  // 将CRONTAB_TYPE转换为SCRIPT_TYPE
  const getScriptType = (crontabType?: CRONTAB_TYPE): SCRIPT_TYPE => {
    return crontabType === CRONTAB_TYPE.Global
      ? SCRIPT_TYPE.Global
      : SCRIPT_TYPE.Local;
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

    if (
      flags.userEditingContent.value &&
      !forceUpdate &&
      !isScriptModeWithoutScript
    ) {
      return;
    }

    await nextTick();
    if (formState.period_details && formState.period_details.length > 0) {
      const formattedComment = generateFormattedPeriodComment(
        formState.period_details
      );

      const cronExpression = convertPeriodToCronExpression(
        formState.period_details
      );

      let actualCommand = '';

      if (formState.content) {
        const contentLines = formState.content.split('\n');

        if (contentLines.length > 0) {
          const firstLine = contentLines[0] || '';
          if (firstLine.startsWith('#')) {
            contentLines.shift();

            if (
              contentLines.length > 0 &&
              contentLines[0].startsWith('# ') &&
              (contentLines[0].includes(t('app.crontab.form.mark.prefix')) ||
                contentLines[0].includes('备注:') ||
                contentLines[0].includes('Notes:'))
            ) {
              contentLines.shift();
            }
          }

          if (
            contentLines.length > 0 &&
            /^\S+\s+\S+\s+\S+\s+\S+\s+\S+/.test(contentLines[0])
          ) {
            const cronLine = contentLines[0];
            const cronMatch = cronLine.match(
              /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)(.*)$/
            );
            if (cronMatch && cronMatch.length > 2) {
              actualCommand = cronMatch[2].trim();
              contentLines.shift();
            } else {
              contentLines.shift();
            }
          }

          if (!actualCommand && contentLines.length > 0) {
            actualCommand = contentLines.join('\n').trim();
          }
        }
      }

      const markPrefix = t('app.crontab.form.mark.prefix');
      const markLine = formState.mark
        ? `# ${markPrefix} ${formState.mark}`
        : '';

      let newContent = `# ${formattedComment}`;
      if (markLine) {
        newContent += `\n${markLine}`;
      }
      newContent += `\n${cronExpression}${
        actualCommand ? ' ' + actualCommand : ''
      }`;

      formState.content = newContent;
    }

    await nextTick();
    if (flags.isUpdatingFromPeriod.value) {
      flags.isUpdatingFromPeriod.value = false;
    }
  };

  const updateMarkInScriptMode = async (
    formState: FormState,
    markValue: string,
    selectedCategory: Ref<string | undefined>,
    selectedScript: Ref<string | undefined>,
    scriptParams: Ref<string>
  ) => {
    try {
      if (!selectedCategory.value || !selectedScript.value) return;

      const scriptDetail = await getScriptDetailApi({
        name: selectedScript.value,
        category: selectedCategory.value,
        type: getScriptType(formState.type),
      });

      const scriptPath = scriptDetail.source;

      const baseContent = scriptParams.value
        ? `${scriptPath} ${scriptParams.value}`
        : scriptPath;

      const formattedComment = generateFormattedPeriodComment(
        formState.period_details
      );

      const markPrefix = t('app.crontab.form.mark.prefix');
      const markLine = markValue ? `# ${markPrefix} ${markValue}` : '';

      const cronExpression = convertPeriodToCronExpression(
        formState.period_details
      );

      let newContent = `# ${formattedComment}`;
      if (markLine) {
        newContent += `\n${markLine}`;
      }
      newContent += `\n${cronExpression} ${baseContent}`;

      await nextTick();
      formState.content = newContent;
    } catch (error) {
      console.error('Error fetching script details:', error);
    }
  };

  const updateContentWithParams = async (
    formState: FormState,
    selectedCategory: Ref<string | undefined>,
    selectedScript: Ref<string | undefined>,
    scriptParams: Ref<string>
  ) => {
    if (
      formState.content_mode === 'script' &&
      selectedScript.value &&
      selectedCategory.value
    ) {
      try {
        const scriptDetail = await getScriptDetailApi({
          name: selectedScript.value,
          category: selectedCategory.value,
          type: getScriptType(formState.type),
        });

        const scriptPath = scriptDetail.source;

        const baseContent = scriptParams.value
          ? `${scriptPath} ${scriptParams.value}`
          : scriptPath;

        const commandContent = baseContent;

        const formattedComment = generateFormattedPeriodComment(
          formState.period_details
        );

        const markPrefix = t('app.crontab.form.mark.prefix');
        const markLine = formState.mark
          ? `# ${markPrefix} ${formState.mark}`
          : '';

        const cronExpression = convertPeriodToCronExpression(
          formState.period_details
        );

        let newContent = `# ${formattedComment}`;
        if (markLine) {
          newContent += `\n${markLine}`;
        }
        newContent += `\n${cronExpression} ${commandContent}`;

        formState.content = newContent;
      } catch (error) {
        console.error('Error fetching script details:', error);
      }
    }
  };

  return {
    updateContentWithPeriod,
    updateMarkInScriptMode,
    updateContentWithParams,
  };
};
