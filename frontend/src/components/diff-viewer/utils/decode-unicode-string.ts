import { decode } from 'html-entities';
import { createLogger } from '@/utils/logger';

// 创建模块专用的logger实例
const logger = createLogger('diff-viewer/decode');

/**
 * 解码Unicode转义字符和HTML实体
 * 使用 html-entities 库替代手写实现
 */
export const decodeUnicodeString = (str: string): string => {
  if (!str) return str;

  try {
    let result = str;

    // 先尝试解析JSON字符串（处理双重转义的情况）
    if (result.startsWith('"') && result.endsWith('"')) {
      try {
        result = JSON.parse(result);
      } catch {
        // 如果JSON解析失败，继续使用原字符串
      }
    }

    // 使用 html-entities 库进行解码
    // 它会自动处理：
    // - Unicode转义字符 (\u0026 -> &)
    // - HTML数字实体 (&#38; -> &)
    // - HTML十六进制实体 (&#x26; -> &)
    // - HTML命名实体 (&amp; -> &, &lt; -> <, &para; -> ¶, 等)
    result = decode(result);

    // 处理一些额外的转义字符
    result = result
      .replace(/\\n/g, '\n')
      .replace(/\\r/g, '\r')
      .replace(/\\t/g, '\t')
      .replace(/\\"/g, '"')
      .replace(/\\\\/g, '\\');

    // 处理换行相关的符号和HTML标签
    // 注意：html-entities已经将&para;解码为¶符号，所以这里直接替换¶
    result = result.replace(/¶/g, '\n').replace(/<br\s*\/?>/gi, '\n');

    return result;
  } catch (error) {
    logger.logWarn('Failed to decode content:', error);
    return str;
  }
};
