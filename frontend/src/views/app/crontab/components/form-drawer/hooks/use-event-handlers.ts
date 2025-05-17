import { Ref } from 'vue';
import { CRONTAB_PERIOD_TYPE } from '@/config/enum';
import type { FormState, StateFlags } from './use-form-state';
import { useContentHandler } from './use-content-handler';

interface ScriptSelections {
  selectedScriptSourceCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

const handleOperationError = (operation: string, error: unknown): void => {
  console.error(`Error ${operation}:`, error);
};

export const useEventHandlers = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections
) => {
  const {
    updateContentWithPeriod,
    updateContentWithParams,
    updateContentFromMark,
  } = useContentHandler();

  const noOp = (): void => {
    // Intentionally empty - used as a no-operation function
  };

  const handleCommandChange = (value: string): void => {
    formState.command = value;
    updateContentWithPeriod(formState, flags, true).catch((error) =>
      handleOperationError('updating content with period', error)
    );
  };

  const handleMarkInput = (): void => {
    // 当用户输入时，我们不做任何操作，等到失去焦点时再更新
  };

  const handleMarkBlur = async (): Promise<void> => {
    try {
      await updateContentFromMark(formState);
    } catch (error) {
      handleOperationError('updating content from mark', error);
    }
  };

  const handleMarkValueChange = (): void => {
    // 值变更时不执行任何操作，等待失去焦点时更新
  };

  const handleContentBlur = (): void => {
    // Intentionally empty - content blur handled elsewhere
  };

  const updateScriptContent = async (): Promise<void> => {
    try {
      flags.isUpdatingFromPeriod.value = true;
      await updateContentWithParams(
        formState,
        selections.selectedScriptSourceCategory,
        selections.selectedScript,
        selections.scriptParams
      );
    } catch (error) {
      handleOperationError('updating content with params', error);
    } finally {
      flags.isUpdatingFromPeriod.value = false;
    }
  };

  const updateNormalContent = async (forceUpdate = false): Promise<void> => {
    try {
      flags.isUpdatingFromPeriod.value = true;
      await updateContentWithPeriod(formState, flags, forceUpdate);
    } catch (error) {
      handleOperationError('updating content with period', error);
    } finally {
      flags.isUpdatingFromPeriod.value = false;
    }
  };

  const handlePeriodChange = async (): Promise<void> => {
    flags.userEditingContent.value = false;
    formState.period_details = [...formState.period_details];

    if (formState.content_mode === 'script') {
      if (
        selections.selectedScript.value &&
        selections.selectedScriptSourceCategory.value
      ) {
        await updateScriptContent();
      } else {
        await updateNormalContent(true);
      }
      return;
    }

    await updateNormalContent(true);
  };

  const switchToScriptMode = async (
    fetchCategories: () => void
  ): Promise<void> => {
    flags.userEditingContent.value = true;

    try {
      selections.selectedScriptSourceCategory.value = undefined;
      selections.selectedScript.value = undefined;
      selections.scriptParams.value = '';

      formState.content = '';
      formState.command = '';

      fetchCategories();

      await updateNormalContent(true);
    } catch (error) {
      handleOperationError('switching to script mode', error);
    } finally {
      flags.userEditingContent.value = false;
    }
  };

  const switchToDirectMode = async (): Promise<void> => {
    try {
      await updateNormalContent(true);
    } catch (error) {
      handleOperationError('switching to direct mode', error);
    }
  };

  const handleContentModeChange = (fetchCategories: () => void): void => {
    if (formState.content_mode === 'script') {
      switchToScriptMode(fetchCategories);
    } else {
      switchToDirectMode();
    }
  };

  const initializePeriodDetails = (): void => {
    if (!formState.period_details || formState.period_details.length === 0) {
      formState.period_details = [
        {
          type: CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES,
          week: 1,
          day: 1,
          hour: 0,
          minute: 1,
          second: 0,
        },
      ];
    }
  };

  return {
    handleContentChange: noOp,
    handleCommandChange,
    handleMarkInput,
    handleMarkBlur,
    handleMarkValueChange,
    handleContentBlur,
    handlePeriodChange,
    handleContentModeChange,
    initializePeriodDetails,
  };
};
