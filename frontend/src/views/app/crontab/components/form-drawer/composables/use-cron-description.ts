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
  const envVarLinePattern = /^[A-Za-z_][A-Za-z0-9_]*\s*=/;

  const tryGetCronDescription = (
    cronExpression: string
  ): { ok: boolean; text: string } => {
    if (!cronExpression) return { ok: false, text: '' };

    try {
      const cronstrueLocale = convertToCronstrueLocale(locale.value);
      const text = cronstrue.toString(cronExpression, {
        locale: cronstrueLocale,
        use24HourTimeFormat: true,
      });
      return { ok: true, text };
    } catch (error) {
      return { ok: false, text: '' };
    }
  };

  /**
   * 将 cron 表达式转换为人类可读的描述
   * @param cronExpression cron 表达式，如 "0 *\/3 * * *"
   * @returns 人类可读的描述，如 "每 3 小时"
   */
  const getCronDescription = (cronExpression: string): string => {
    const result = tryGetCronDescription(cronExpression);
    return result.ok ? result.text : cronExpression;
  };

  /**
   * 从 crontab 文件内容中提取 cron 表达式并转换为描述
   * @param content crontab 文件内容
   * @returns 人类可读的描述
   */
  const getCronDescriptionFromContent = (content: string): string => {
    if (!content) return '';

    const lines = content.split('\n');
    for (const line of lines) {
      const trimmed = line.trim();
      if (
        !trimmed ||
        trimmed.startsWith('#') ||
        envVarLinePattern.test(trimmed)
      )
        continue;

      const parts = trimmed.split(/\s+/);
      if (parts.length < 6) continue;

      const cronExpression = parts.slice(0, 5).join(' ');
      const result = tryGetCronDescription(cronExpression);
      if (result.ok) return result.text;
    }
    return '';
  };

  return {
    getCronDescription,
    getCronDescriptionFromContent,
  };
};
