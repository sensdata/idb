import { Ref, nextTick } from 'vue';
import { CRONTAB_PERIOD_TYPE } from '@/config/enum';
import type { FormState, StateFlags } from './use-form-state';
import { usePeriodUtils } from './use-period-utils';
import { useContentHandler } from './use-content-handler';

interface ScriptSelections {
  selectedCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

export const useEventHandlers = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections
) => {
  const { parseCronExpression } = usePeriodUtils();
  const {
    updateContentWithPeriod,
    updateMarkInScriptMode,
    updateContentWithParams,
  } = useContentHandler();

  const handleContentChange = (content: string): void => {
    if (
      flags.isUpdatingFromPeriod.value ||
      flags.isInitialLoad.value ||
      formState.content_mode !== 'direct'
    ) {
      return;
    }

    flags.userEditingContent.value = true;
    formState.content = content;

    nextTick(() => {
      flags.userEditingContent.value = false;
    });
  };

  const handleMarkInput = (): void => {
    flags.userEditingContent.value = true;

    nextTick(() => {
      flags.userEditingContent.value = false;
    });
  };

  const handleMarkBlur = (): void => {
    flags.isInitialLoad.value = false;
    flags.isUpdatingFromPeriod.value = false;
    flags.userEditingContent.value = false;

    nextTick(() => {
      updateContentWithPeriod(formState, flags, true);
    });
  };

  const handleMarkValueChange = (value: string): void => {
    flags.isInitialLoad.value = false;
    const wasUpdating = flags.isUpdatingFromPeriod.value;
    flags.isUpdatingFromPeriod.value = true;

    nextTick(() => {
      try {
        if (
          formState.content_mode === 'script' &&
          selections.selectedScript.value &&
          selections.selectedCategory.value
        ) {
          updateMarkInScriptMode(
            formState,
            value,
            selections.selectedCategory,
            selections.selectedScript,
            selections.scriptParams
          );
        } else {
          updateContentWithPeriod(formState, flags, true);
        }
      } catch (error) {
        console.error('Error updating mark value:', error);
      } finally {
        if (!wasUpdating) {
          nextTick(() => {
            flags.isUpdatingFromPeriod.value = false;
          });
        }
      }
    });
  };

  const handleContentBlur = (content: string): void => {
    if (
      formState.content_mode !== 'direct' ||
      flags.isUpdatingFromPeriod.value
    ) {
      return;
    }

    flags.isUpdatingFromPeriod.value = true;

    try {
      const contentLines = content.split('\n');
      let cronExpression = '';

      if (
        contentLines.length > 1 &&
        contentLines[0]?.startsWith('# ') &&
        /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)/.test(contentLines[1])
      ) {
        const matches = contentLines[1].match(
          /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s*(.*)/
        );
        if (matches && matches.length > 1) {
          cronExpression = matches[1];
        }
      }
      if (/^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)/.test(contentLines[0])) {
        const matches = contentLines[0].match(
          /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s*(.*)/
        );
        if (matches && matches.length > 1) {
          cronExpression = matches[1];
        }
      }

      if (cronExpression) {
        const periodDetail = parseCronExpression(cronExpression);
        if (periodDetail) {
          flags.userEditingContent.value = true;

          nextTick(() => {
            formState.period_details = [{ ...periodDetail }];

            nextTick(() => {
              flags.userEditingContent.value = false;
            });
          });
        }
      }
    } catch (error) {
      console.error('Error processing content blur:', error);
    } finally {
      nextTick(() => {
        flags.isUpdatingFromPeriod.value = false;
      });
    }
  };

  const handlePeriodChange = (): void => {
    flags.userEditingContent.value = false;
    formState.period_details = [...formState.period_details];

    if (formState.content_mode === 'script') {
      if (
        selections.selectedScript.value &&
        selections.selectedCategory.value
      ) {
        flags.isUpdatingFromPeriod.value = true;
        try {
          updateContentWithParams(
            formState,
            selections.selectedCategory,
            selections.selectedScript,
            selections.scriptParams
          );
        } catch (error) {
          console.error('Error updating content with params:', error);
        } finally {
          nextTick(() => {
            flags.isUpdatingFromPeriod.value = false;
          });
        }
      } else {
        flags.isUpdatingFromPeriod.value = true;
        try {
          updateContentWithPeriod(formState, flags, true);
        } catch (error) {
          console.error('Error updating content with period:', error);
        } finally {
          nextTick(() => {
            flags.isUpdatingFromPeriod.value = false;
          });
        }
      }
      return;
    }

    if (flags.isUpdatingFromPeriod.value) {
      nextTick(() => {
        flags.isUpdatingFromPeriod.value = true;
        try {
          updateContentWithPeriod(formState, flags, true);
        } catch (error) {
          console.error('Error updating content with period:', error);
        } finally {
          nextTick(() => {
            flags.isUpdatingFromPeriod.value = false;
          });
        }
      });
      return;
    }

    flags.isUpdatingFromPeriod.value = true;
    try {
      updateContentWithPeriod(formState, flags, true);
    } catch (error) {
      console.error(
        'Error updating content with period in normal path:',
        error
      );
    } finally {
      nextTick(() => {
        flags.isUpdatingFromPeriod.value = false;
      });
    }
  };

  const handleContentModeChange = (fetchCategories: () => void): void => {
    if (formState.content_mode === 'script') {
      flags.userEditingContent.value = true;

      nextTick(() => {
        try {
          selections.selectedCategory.value = undefined;
          selections.selectedScript.value = undefined;
          selections.scriptParams.value = '';

          formState.content = '';

          fetchCategories();

          updateContentWithPeriod(formState, flags, true);
        } catch (error) {
          console.error('Error handling content mode change:', error);
        } finally {
          nextTick(() => {
            flags.userEditingContent.value = false;
          });
        }
      });
    } else {
      nextTick(() => {
        try {
          flags.isUpdatingFromPeriod.value = true;

          try {
            updateContentWithPeriod(formState, flags, true);
          } finally {
            nextTick(() => {
              flags.isUpdatingFromPeriod.value = false;
            });
          }
        } catch (error) {
          console.error('Error switching to direct mode:', error);
        }
      });
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
    handleContentChange,
    handleMarkInput,
    handleMarkBlur,
    handleMarkValueChange,
    handleContentBlur,
    handlePeriodChange,
    handleContentModeChange,
    initializePeriodDetails,
  };
};
