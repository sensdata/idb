import { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getCrontabDetailApi } from '@/api/crontab';
import { getScriptListApi } from '@/api/script';
import { CRONTAB_TYPE, SCRIPT_TYPE } from '@/config/enum';
import { CrontabEntity } from '@/entity/Crontab';
import { ScriptEntity } from '@/entity/Script';
import { FormState, StateFlags } from './use-form-state';
import { usePeriodUtils } from './use-period-utils';

// 定义类型接口
interface ScriptSelections {
  selectedScriptSourceCategory: Ref<string | undefined>;
  selectedScript: Ref<string | undefined>;
  scriptParams: Ref<string>;
}

// 编辑参数接口
interface EditParams {
  name?: string;
  type?: CRONTAB_TYPE;
  category?: string;
  isEdit?: boolean;
}

export const useDataLoader = (
  formState: FormState,
  flags: StateFlags,
  selections: ScriptSelections,
  fetchScripts: () => Promise<void>,
  currentHostId?: any
) => {
  const { t } = useI18n();
  const { parseCronExpression } = usePeriodUtils();

  // 从内容行中提取标记信息
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

  // 查找匹配脚本
  const findMatchingScript = async (
    sourceType: SCRIPT_TYPE,
    category: string,
    fullPath: string
  ): Promise<ScriptEntity | undefined> => {
    try {
      const scriptListResult = await getScriptListApi({
        type: sourceType,
        category,
        page: 1,
        page_size: 100,
        host: currentHostId,
      } as any);

      if (scriptListResult?.items && Array.isArray(scriptListResult.items)) {
        return scriptListResult.items.find(
          (script: ScriptEntity) => fullPath === script.source
        );
      }
    } catch (error) {
      console.error('Error finding script by source:', error);
    }

    return undefined;
  };

  // 从路径获取脚本名称
  const getScriptNameFromPath = (fullPath: string): string => {
    const pathSegments = fullPath.split('/').filter(Boolean);
    return pathSegments.length >= 1
      ? pathSegments[pathSegments.length - 1]
      : '';
  };

  // 从内容解析脚本路径和参数
  const getScriptInfoFromContent = (
    content: string
  ): {
    fullPath: string;
    params: string;
  } => {
    const contentParts = content.split(' ');

    // 确定脚本路径
    const fullPath =
      contentParts[0] === 'sh' && contentParts.length >= 2
        ? contentParts[1]
        : contentParts[0];

    // 处理脚本参数
    let params = '';
    if (contentParts[0] === 'sh' && contentParts.length > 2) {
      params = contentParts.slice(2).join(' ');
    } else if (contentParts[0] !== 'sh' && contentParts.length > 1) {
      params = contentParts.slice(1).join(' ');
    }

    return { fullPath, params };
  };

  // 从脚本路径解析类别
  const getCategoryFromPath = (fullPath: string): string | undefined => {
    const pathSegments = fullPath.split('/').filter(Boolean);
    if (pathSegments.length < 2) {
      return undefined;
    }

    return pathSegments[pathSegments.length - 2];
  };

  const parseScriptFromContent = async (content: string): Promise<boolean> => {
    if (!content) {
      return false;
    }

    formState.content_mode = 'script';

    // 解析脚本信息
    const { fullPath, params } = getScriptInfoFromContent(content);

    // 从路径中提取类型和相对路径信息
    const isGlobal = fullPath.includes('/global/');
    formState.type = isGlobal ? CRONTAB_TYPE.Global : CRONTAB_TYPE.Local;

    // 获取类别
    const category = getCategoryFromPath(fullPath);
    if (!category) {
      return false;
    }

    selections.selectedScriptSourceCategory.value = category;
    selections.scriptParams.value = params;

    // 等待脚本加载完成
    await fetchScripts();

    // 查找匹配的脚本
    const scriptType =
      formState.type === CRONTAB_TYPE.Global
        ? SCRIPT_TYPE.Global
        : SCRIPT_TYPE.Local;

    const matchingScript = await findMatchingScript(
      scriptType,
      selections.selectedScriptSourceCategory.value,
      fullPath
    );

    if (matchingScript) {
      selections.selectedScript.value = matchingScript.name;
    } else {
      const scriptName = getScriptNameFromPath(fullPath);
      selections.selectedScript.value = scriptName;
    }

    return true;
  };

  // 从内容中提取命令和cron表达式
  const extractCommandContent = (
    data: any
  ): { commandContent: string; cronExpression: string | null } => {
    if (!data.content) {
      return { commandContent: '', cronExpression: null };
    }

    const contentLines = data.content.split('\n');

    const mark = extractMarkFromContent(contentLines);
    if (mark) {
      formState.mark = mark;
    }

    let cronLineIndex = 0;

    // 跳过注释行查找crontab表达式
    while (
      cronLineIndex < contentLines.length &&
      contentLines[cronLineIndex]?.startsWith('#')
    ) {
      cronLineIndex++;
    }

    if (cronLineIndex >= contentLines.length) {
      return { commandContent: '', cronExpression: null };
    }

    const cronLine = contentLines[cronLineIndex];

    // 匹配cron表达式(5个字段)和后面的命令
    const cronPattern = /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$/;
    const cronMatches = cronLine.match(cronPattern);

    if (cronMatches && cronMatches.length >= 3) {
      return {
        commandContent: cronMatches[2].trim(),
        cronExpression: cronMatches[1].trim(),
      };
    }

    return {
      commandContent: data.content,
      cronExpression: null,
    };
  };

  // 重置表单基本信息
  const resetBasicFormInfo = (record: CrontabEntity) => {
    formState.name = record.name || '';
    formState.type = record.type || CRONTAB_TYPE.Local;
    formState.mark = record.mark || '';
    formState.content_mode = record.content_mode || 'direct';
    formState.command = '';
  };

  // 处理周期详情
  const handlePeriodDetails = async (record: CrontabEntity) => {
    // 如果有period_details，直接使用
    if (record.period_details && record.period_details.length > 0) {
      formState.period_details = record.period_details;
      return { usedExistingPeriod: true };
    }

    // 否则尝试从content解析
    const { commandContent, cronExpression } = extractCommandContent(record);

    // 解析cron表达式
    if (cronExpression) {
      const periodDetail = parseCronExpression(cronExpression);
      if (periodDetail) {
        formState.period_details = [periodDetail];
      }
    }

    // 设置命令字段
    formState.command = commandContent;

    return {
      usedExistingPeriod: false,
      commandContent,
      cronExpression,
    };
  };

  // 处理内容模式
  const handleContentMode = async (
    record: CrontabEntity,
    commandContent: string,
    hasExplicitMode: boolean
  ) => {
    // 如果未指定content_mode，则尝试解析为脚本
    if (!hasExplicitMode) {
      const isScript = await parseScriptFromContent(commandContent);

      if (!isScript) {
        formState.content_mode = 'direct';
      }
    }
  };

  const loadDataFromRecord = async (record: CrontabEntity): Promise<void> => {
    flags.isInitialLoad.value = true;

    try {
      if (!record) {
        return;
      }

      // 重置表单基本信息
      resetBasicFormInfo(record);

      // 处理周期详情
      const { usedExistingPeriod, commandContent } = await handlePeriodDetails(
        record
      );

      // 处理内容模式
      const hasExplicitMode = !!record.content_mode;
      if (!usedExistingPeriod && commandContent) {
        await handleContentMode(record, commandContent, hasExplicitMode);
      }

      // 最后设置content，确保模式已经确定
      formState.content = record.content || '';
    } catch (error) {
      console.error('Error loading data from record:', error);
    } finally {
      flags.isInitialLoad.value = false;
    }
  };

  const loadData = async (
    paramsRef: Ref<EditParams | undefined>
  ): Promise<void> => {
    flags.isInitialLoad.value = true;

    try {
      if (
        !paramsRef.value ||
        !paramsRef.value.name ||
        !paramsRef.value.type ||
        !paramsRef.value.category
      ) {
        return;
      }

      // 构造API参数
      const params = {
        name: paramsRef.value.name,
        type: paramsRef.value.type,
        category: paramsRef.value.category,
      };

      const data = await getCrontabDetailApi(params);

      if (!data) {
        return;
      }

      // 使用相同的加载逻辑处理返回的数据
      await loadDataFromRecord(data);
    } catch (error) {
      console.error('Error loading data:', error);
    } finally {
      flags.isInitialLoad.value = false;
    }
  };

  return {
    loadData,
    loadDataFromRecord,
  };
};
