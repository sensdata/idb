import { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { getCrontabDetailApi } from '@/api/crontab';
import { getScriptListApi } from '@/api/script';
import { CRONTAB_TYPE, SCRIPT_TYPE } from '@/config/enum';
import { CrontabEntity } from '@/entity/Crontab';
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

  // 获取脚本分类列表
  const getScriptCategoriesWithSource = async (
    scriptType: SCRIPT_TYPE
  ): Promise<any[]> => {
    try {
      const { getScriptCategoryListApi } = await import('@/api/script');
      const result = await getScriptCategoryListApi({
        type: scriptType,
        page: 1,
        page_size: 1000,
        host: currentHostId,
      } as any);

      if (result?.items && Array.isArray(result.items)) {
        return result.items;
      }
    } catch (error) {
      console.error('Error getting script categories:', error);
    }
    return [];
  };

  // 根据脚本路径匹配分类
  const findMatchingCategory = (categories: any[], scriptPath: string): any => {
    // 首先尝试找到路径包含在脚本路径中的分类
    for (const category of categories) {
      if (category.source && scriptPath.includes(category.source)) {
        return category;
      }
    }
    return null;
  };

  // 获取特定分类下的所有脚本
  const getScriptsForCategory = async (
    scriptType: SCRIPT_TYPE,
    categoryName: string
  ): Promise<any[]> => {
    try {
      // 使用导入的API而不是重新导入
      const result = await getScriptListApi({
        type: scriptType,
        category: categoryName,
        page: 1,
        page_size: 1000,
        host: currentHostId,
      } as any);

      if (result?.items && Array.isArray(result.items)) {
        return result.items;
      }
    } catch (error) {
      console.error('Error getting scripts for category:', error);
    }
    return [];
  };

  // 根据脚本路径匹配脚本
  const findMatchingScriptBySource = (
    scripts: any[],
    scriptPath: string
  ): any => {
    return scripts.find((script) => script.source === scriptPath);
  };

  const parseScriptFromContent = async (content: string): Promise<boolean> => {
    if (!content) {
      return false;
    }

    // 设置默认内容模式为脚本
    formState.content_mode = 'script';

    // 解析脚本信息（路径和参数）
    const { fullPath, params } = getScriptInfoFromContent(content);
    if (!fullPath) {
      return false;
    }

    // 根据路径判断脚本类型（全局或本地）
    const isGlobal = fullPath.includes('/global/');
    formState.type = isGlobal ? CRONTAB_TYPE.Global : CRONTAB_TYPE.Local;
    const scriptType =
      formState.type === CRONTAB_TYPE.Global
        ? SCRIPT_TYPE.Global
        : SCRIPT_TYPE.Local;

    // 1. 获取当前类型下所有分类
    const categories = await getScriptCategoriesWithSource(scriptType);
    if (!categories || categories.length === 0) {
      return false;
    }

    // 2. 查找脚本路径匹配的分类
    const matchingCategory = findMatchingCategory(categories, fullPath);
    if (!matchingCategory) {
      return false;
    }

    // 设置找到的分类
    selections.selectedScriptSourceCategory.value = matchingCategory.name;
    selections.scriptParams.value = params;

    // 3. 获取该分类下的所有脚本
    const categoryScripts = await getScriptsForCategory(
      scriptType,
      matchingCategory.name
    );

    // 4. 查找匹配的脚本
    const matchingScript = findMatchingScriptBySource(
      categoryScripts,
      fullPath
    );

    if (matchingScript) {
      // 如果找到完全匹配的脚本，使用它
      selections.selectedScript.value = matchingScript.name;
      return true;
    }

    // 如果没找到完全匹配的脚本，使用路径中的文件名作为脚本名
    const scriptName = getScriptNameFromPath(fullPath);
    selections.selectedScript.value = scriptName;
    // 继续处理，但是可能会因为找不到完全匹配而导致显示问题
    console.warn(
      `No exact script match found for ${fullPath}, using filename: ${scriptName}`
    );
    return true;
  };

  // 从内容中提取命令和cron表达式
  const extractCommandContent = (
    data: any
  ): {
    commandContent: string;
    cronExpression: string | null;
    user: string;
  } => {
    if (!data.content) {
      return { commandContent: '', cronExpression: null, user: 'root' };
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
      return { commandContent: '', cronExpression: null, user: 'root' };
    }

    const cronLine = contentLines[cronLineIndex];

    // 优先匹配 Time-User-Command 格式
    const cronPatternWithUser =
      /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(\S+)\s+(.+)$/;
    const cronPatternLegacy = /^(\S+\s+\S+\s+\S+\s+\S+\s+\S+)\s+(.+)$/;

    let user = 'root';
    let cronExpression: string | null = null;
    let commandContent = '';

    const withUserMatches = cronLine.match(cronPatternWithUser);
    if (withUserMatches && withUserMatches.length >= 4) {
      cronExpression = withUserMatches[1].trim();
      user = withUserMatches[2].trim();
      commandContent = withUserMatches[3].trim();
    } else {
      const legacyMatches = cronLine.match(cronPatternLegacy);
      if (legacyMatches && legacyMatches.length >= 3) {
        cronExpression = legacyMatches[1].trim();
        commandContent = legacyMatches[2].trim();
      } else {
        return {
          commandContent: data.content,
          cronExpression: null,
          user: 'root',
        };
      }
    }

    return {
      commandContent,
      cronExpression,
      user,
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
    const { commandContent, cronExpression, user } =
      extractCommandContent(record);

    // 解析cron表达式
    if (cronExpression) {
      const periodDetail = parseCronExpression(cronExpression);
      if (periodDetail) {
        formState.period_details = [periodDetail];
      }
    }

    // 设置命令字段和用户
    formState.command = commandContent;
    formState.user = user || 'root';

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
