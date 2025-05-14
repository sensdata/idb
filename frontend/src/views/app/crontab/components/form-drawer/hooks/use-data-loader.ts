import { toRaw, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getCrontabDetailApi } from '@/api/crontab';
import { CRONTAB_TYPE } from '@/config/enum';
import { FormState, StateFlags } from './use-form-state';

// 定义常量
const SCRIPT_PATHS = {
  LOCAL: '/var/lib/idb/data/scripts/local/',
  GLOBAL: '/var/lib/idb/data/scripts/global/',
};

// 定义类型接口
interface ScriptSelections {
  selectedCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

export const useDataLoader = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections,
  fetchScripts: () => Promise<void>
) => {
  const { t } = useI18n();

  const extractMarkFromContent = (contentLines: string[]): string => {
    const hasMarkComment =
      contentLines.length > 1 &&
      contentLines[1]?.startsWith('# ') &&
      (contentLines[1].includes(t('app.crontab.form.mark.prefix')) ||
        contentLines[1].includes('备注:') ||
        contentLines[1].includes('Notes:'));

    if (hasMarkComment) {
      const line = contentLines[1];
      const markPrefix = t('app.crontab.form.mark.prefix');

      if (line.includes(markPrefix)) {
        return line
          .substring(line.indexOf(markPrefix) + markPrefix.length)
          .trim();
      }
      if (line.includes('备注:')) {
        return line.substring(line.indexOf('备注:') + '备注:'.length).trim();
      }
      if (line.includes('Notes:')) {
        return line.substring(line.indexOf('Notes:') + 'Notes:'.length).trim();
      }
    }

    return '';
  };

  // 处理脚本路径和相关设置
  const processScriptPath = (
    path: string
  ): { relativePath: string; type: CRONTAB_TYPE } | null => {
    if (path.startsWith(SCRIPT_PATHS.LOCAL)) {
      return {
        relativePath: path.substring(SCRIPT_PATHS.LOCAL.length),
        type: CRONTAB_TYPE.Local,
      };
    }
    if (path.startsWith(SCRIPT_PATHS.GLOBAL)) {
      return {
        relativePath: path.substring(SCRIPT_PATHS.GLOBAL.length),
        type: CRONTAB_TYPE.Global,
      };
    }
    return null;
  };

  const parseScriptFromContent = async (content: string): Promise<boolean> => {
    if (
      !content ||
      !(
        content.includes(SCRIPT_PATHS.LOCAL) ||
        content.includes(SCRIPT_PATHS.GLOBAL)
      )
    ) {
      return false;
    }

    formState.content_mode = 'script';
    const contentParts = content.split(' ');

    // 确定脚本路径
    const fullPath =
      contentParts[0] === 'sh' && contentParts.length >= 2
        ? contentParts[1]
        : contentParts[0];

    const pathInfo = processScriptPath(fullPath);
    if (!pathInfo) {
      return false;
    }

    formState.type = pathInfo.type;
    const relativePath = pathInfo.relativePath;

    const pathParts = relativePath.split('/');
    if (pathParts.length >= 2) {
      selections.selectedCategory.value = pathParts[0];
      await fetchScripts();
      selections.selectedScript.value = pathParts[1];

      // 处理脚本参数
      if (contentParts[0] === 'sh' && contentParts.length > 2) {
        selections.scriptParams.value = contentParts.slice(2).join(' ');
      } else if (contentParts[0] !== 'sh' && contentParts.length > 1) {
        selections.scriptParams.value = contentParts.slice(1).join(' ');
      }
    }

    return true;
  };

  const extractCommandContent = (data: any): string => {
    if (!data.content) {
      return data.content;
    }

    const contentLines = data.content.split('\n');
    const hasPeriodComment = contentLines[0]?.startsWith('# ');

    const mark = extractMarkFromContent(contentLines);
    if (mark) {
      formState.mark = mark;
    }

    const cronLineIndex = mark ? 2 : 1;
    const cronLine =
      contentLines.length > cronLineIndex ? contentLines[cronLineIndex] : '';
    const isCronLine = /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)/.test(cronLine);

    if (hasPeriodComment && isCronLine) {
      const cronParts = cronLine.match(/^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.*)/);
      if (cronParts && cronParts.length > 2) {
        return cronParts[2];
      }

      const startIndex = mark ? 3 : 2;
      return contentLines.slice(startIndex).join('\n');
    }

    const firstLineIsCron = /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)/.test(
      contentLines[0]
    );
    if (firstLineIsCron) {
      const cronParts = contentLines[0].match(
        /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.*)/
      );
      if (cronParts && cronParts.length > 2) {
        return cronParts[2];
      }
    }

    return data.content;
  };

  interface CrontabDetailParams {
    value?: { id: number };
  }

  const loadData = async (paramsRef: CrontabDetailParams): Promise<void> => {
    flags.isInitialLoad.value = true;

    try {
      if (!paramsRef.value) {
        return;
      }

      const data = await getCrontabDetailApi(toRaw(paramsRef.value));
      Object.assign(formState, data);

      const extractedContent = extractCommandContent(data);
      const isScript = await parseScriptFromContent(extractedContent);

      if (!isScript) {
        formState.content = extractedContent;
      }
    } catch (error) {
      console.error('Error loading crontab data:', error);
    } finally {
      flags.isInitialLoad.value = false;
    }
  };

  return {
    loadData,
  };
};
