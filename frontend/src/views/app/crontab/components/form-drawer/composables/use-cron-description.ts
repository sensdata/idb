import cronstrue from 'cronstrue/i18n';
import { useI18n } from 'vue-i18n';

/**
 * 将 BCP 47 语言标签转换为 cronstrue 支持的 locale 格式
 * BCP 47 使用连字符（如 zh-CN），cronstrue 使用下划线（如 zh_CN）
 * 对于简单语言代码（如 en-US），cronstrue 只需要语言部分（如 en）
 */
const convertToCronstrueLocale = (bcp47Locale: string): string => {
  // 将连字符替换为下划线
  const underscoreLocale = bcp47Locale.replace('-', '_');

  // 对于英语等只需要语言代码的情况，提取语言部分
  // cronstrue 支持的完整 locale 列表包括：zh_CN, zh_TW, pt_BR 等需要完整代码
  // 其他如 en, es, fr 等只需要语言部分
  const needsFullLocale = ['zh_CN', 'zh_TW', 'pt_BR'];
  if (needsFullLocale.includes(underscoreLocale)) {
    return underscoreLocale;
  }

  // 其他情况只取语言部分（下划线前的部分）
  return underscoreLocale.split('_')[0];
};

/**
 * 将 cron 表达式转换为人类可读的描述
 * 使用 cronstrue 库，支持多语言
 */
export const useCronDescription = () => {
  const { locale } = useI18n();

  /**
   * 将 cron 表达式转换为人类可读的描述
   * @param cronExpression cron 表达式，如 "0 *\/3 * * *"
   * @returns 人类可读的描述，如 "每 3 小时"
   */
  const getCronDescription = (cronExpression: string): string => {
    if (!cronExpression) return '';

    try {
      const cronstrueLocale = convertToCronstrueLocale(locale.value);
      return cronstrue.toString(cronExpression, {
        locale: cronstrueLocale,
        use24HourTimeFormat: true,
      });
    } catch (error) {
      console.error('Failed to parse cron expression:', cronExpression, error);
      return cronExpression;
    }
  };

  /**
   * 从 crontab 文件内容中提取 cron 表达式并转换为描述
   * @param content crontab 文件内容
   * @returns 人类可读的描述
   */
  const getCronDescriptionFromContent = (content: string): string => {
    if (!content) return '';

    const lines = content.split('\n');
    // 找到第一个非注释行
    const cronLine = lines.find((line) => line.trim() && !line.startsWith('#'));
    if (!cronLine) return '';

    // 提取 cron 表达式（前5个字段）
    const parts = cronLine.trim().split(/\s+/);
    if (parts.length < 5) return '';

    const cronExpression = parts.slice(0, 5).join(' ');
    return getCronDescription(cronExpression);
  };

  return {
    getCronDescription,
    getCronDescriptionFromContent,
  };
};
