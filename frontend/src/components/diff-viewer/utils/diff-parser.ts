import { ParsedDiff } from '../types';

// 常量定义
const REGEX_PATTERNS = {
  INS_TAG: /<ins[^>]*>([\s\S]*?)<\/ins>/g,
  DEL_TAG: /<del[^>]*>([\s\S]*?)<\/del>/g,
  SPAN_TAG: /<span(?![^>]*class="diff-)[^>]*>/g,
} as const;

const DIFF_CLASSES = {
  NORMAL: 'diff-normal',
  INLINE_NORMAL: 'diff-inline-normal',
  ADDED: 'diff-added',
  INLINE_ADDED: 'diff-inline-added',
  REMOVED: 'diff-removed',
  INLINE_DELETED: 'diff-inline-deleted',
} as const;

/**
 * 检查内容是否为块级内容（包含换行符或段落标记）
 */
const isBlockContent = (content: string): boolean => {
  return (
    content.includes('\n') ||
    content.includes('&para;') ||
    content.includes('<br')
  );
};

/**
 * 创建带CSS类的span标签
 */
const createSpanWithClass = (content: string, className: string): string => {
  return `<span class="${className}">${content}</span>`;
};

/**
 * 处理删除标签（del）
 */
const processDelTags = (content: string, isHistorical: boolean): string => {
  return content.replace(REGEX_PATTERNS.DEL_TAG, (match, delContent) => {
    if (isHistorical) {
      // 历史版本：显示为普通内容
      const className = isBlockContent(delContent)
        ? DIFF_CLASSES.NORMAL
        : DIFF_CLASSES.INLINE_NORMAL;
      return createSpanWithClass(delContent, className);
    }
    // 当前版本：显示为删除状态
    const className = isBlockContent(delContent)
      ? DIFF_CLASSES.REMOVED
      : DIFF_CLASSES.INLINE_DELETED;
    return createSpanWithClass(delContent, className);
  });
};

/**
 * 处理插入标签（ins）
 */
const processInsTags = (content: string, isHistorical: boolean): string => {
  if (isHistorical) {
    // 历史版本：移除ins标签内容
    return content.replace(REGEX_PATTERNS.INS_TAG, '');
  }
  // 当前版本：显示为新增内容
  return content.replace(REGEX_PATTERNS.INS_TAG, (match, insContent) => {
    const className = isBlockContent(insContent)
      ? DIFF_CLASSES.ADDED
      : DIFF_CLASSES.INLINE_ADDED;
    return createSpanWithClass(insContent, className);
  });
};

/**
 * 处理普通span标签
 */
const processSpanTags = (content: string): string => {
  return content.replace(
    REGEX_PATTERNS.SPAN_TAG,
    `<span class="${DIFF_CLASSES.INLINE_NORMAL}">`
  );
};

/**
 * 处理内容的核心逻辑
 */
const processContent = (content: string, isHistorical: boolean): string => {
  let result = content;

  // 按顺序处理各种标签
  result = processInsTags(result, isHistorical);
  result = processDelTags(result, isHistorical);
  result = processSpanTags(result);

  return result;
};

/**
 * 解析diff内容为侧边对比格式
 */
export const parseDiffToSideBySide = (diffHtml: string): ParsedDiff => {
  if (!diffHtml?.trim()) {
    return { historical: '', current: '' };
  }

  return {
    historical: processContent(diffHtml, true),
    current: processContent(diffHtml, false),
  };
};
