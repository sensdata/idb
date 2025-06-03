<template>
  <div
    class="diff-content"
    role="region"
    :aria-label="$t('components.diffViewer.diffViewer')"
  >
    <div v-if="isValidDiff" class="diff-side-by-side">
      <section
        class="diff-column left-column"
        aria-labelledby="historical-header"
      >
        <header id="historical-header" class="diff-header">
          {{ $t('components.diffViewer.historical') }}
        </header>
        <pre
          class="diff-text"
          role="textbox"
          :aria-label="$t('components.diffViewer.historicalContent')"
          aria-readonly="true"
          tabindex="0"
          v-html="sanitizedHistorical"
        ></pre>
      </section>

      <div
        class="diff-divider"
        role="separator"
        aria-orientation="vertical"
      ></div>

      <section
        class="diff-column right-column"
        aria-labelledby="current-header"
      >
        <header id="current-header" class="diff-header">
          {{ $t('components.diffViewer.current') }}
        </header>
        <pre
          class="diff-text"
          role="textbox"
          :aria-label="$t('components.diffViewer.currentContent')"
          aria-readonly="true"
          tabindex="0"
          v-html="sanitizedCurrent"
        ></pre>
      </section>
    </div>

    <div v-else class="empty-state" role="status" aria-live="polite">
      <a-empty :description="emptyStateText" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, toRefs } from 'vue';
  import DOMPurify from 'dompurify';
  import type { ParsedDiff } from './types';

  defineOptions({
    name: 'DiffView',
  });

  export interface Props {
    parsedDiff: ParsedDiff | null;
    emptyStateText?: string;
  }

  const props = withDefaults(defineProps<Props>(), {
    parsedDiff: null,
    emptyStateText: 'No differences found',
  });

  // 使用 toRefs 提升性能，避免不必要的响应式解包
  const { parsedDiff, emptyStateText } = toRefs(props);

  // 数据验证 - 添加更严格的验证
  const isValidDiff = computed(() => {
    return Boolean(
      parsedDiff.value &&
        typeof parsedDiff.value.historical === 'string' &&
        typeof parsedDiff.value.current === 'string' &&
        (parsedDiff.value.historical.trim() !== '' ||
          parsedDiff.value.current.trim() !== '')
    );
  });

  // 安全的HTML清理函数 - 使用DOMPurify
  const sanitizeHtml = (html: string): string => {
    if (!html) return '';

    // 配置DOMPurify，只允许diff相关的安全标签和属性
    const config = {
      ALLOWED_TAGS: ['span', 'div', 'br', 'pre', 'del', 'ins'],
      ALLOWED_ATTR: ['class'],
      ALLOWED_URI_REGEXP: /^$/, // 不允许任何URI
      FORBID_SCRIPTS: true,
      FORBID_TAGS: [
        'script',
        'style',
        'iframe',
        'object',
        'embed',
        'link',
        'meta',
      ],
      FORBID_ATTR: ['style', 'onclick', 'onload', 'onerror'],
    };

    return DOMPurify.sanitize(html, config);
  };

  const sanitizedHistorical = computed(() =>
    sanitizeHtml(parsedDiff.value?.historical || '')
  );

  const sanitizedCurrent = computed(() =>
    sanitizeHtml(parsedDiff.value?.current || '')
  );
</script>

<style scoped>
  .diff-content {
    flex: 1;
    margin-bottom: 16px;
    overflow: auto;
    background: var(--color-bg-1);
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .diff-side-by-side {
    display: flex;
    height: 100%;
    min-height: 400px;
  }

  .diff-column {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .left-column {
    border-right: 1px solid var(--color-border-2);
  }

  .right-column {
    border-left: 1px solid var(--color-border-2);
  }

  .diff-divider {
    flex-shrink: 0;
    width: 2px;
    background: var(--color-border-2);
  }

  .diff-header {
    padding: 8px 16px;
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-2);
    background: var(--color-bg-2);
    border-bottom: 1px solid var(--color-border-2);
  }

  .diff-text {
    flex: 1;
    padding: 16px;
    margin: 0;
    overflow: auto;
    font-family: Monaco, Menlo, 'Ubuntu Mono', monospace;
    font-size: 13px;
    line-height: 1.6;
    color: var(--color-text-1);
    overflow-wrap: break-word;
    white-space: pre-line;
    outline: none;
    background: var(--color-bg-1);
    border: none;
  }

  .diff-text:focus {
    border-radius: 2px;
    box-shadow: 0 0 0 2px var(--color-primary-light-3);
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 300px;
  }

  /* Diff 样式 - 使用更高特异性选择器保持兼容性 */
  .diff-content .diff-text :deep(.diff-added) {
    display: block;
    padding: 4px 8px;
    margin: 2px 0;
    color: var(--color-success-dark-1, #24292f);
    background-color: var(--color-success-light-5, #e6ffed);
    border-left: 4px solid var(--color-success, #28a745);
    border-radius: 4px;
  }

  .diff-content .diff-text :deep(.diff-removed) {
    display: block;
    padding: 4px 8px;
    margin: 2px 0;
    color: var(--color-danger-dark-1, #24292f);
    background-color: var(--color-danger-light-5, #ffebe9);
    border-left: 4px solid var(--color-danger, #d73a49);
    border-radius: 4px;
  }

  .diff-content .diff-text :deep(.diff-normal) {
    display: block;
    color: var(--color-text-1);
  }

  /* 内联 diff 样式 */
  .diff-content .diff-text :deep(.diff-inline-added) {
    display: inline;
    padding: 2px 4px;
    color: var(--color-success-dark-1, #24292f);
    background-color: var(--color-success-light-5, #e6ffed);
    border: 1px solid var(--color-success, #28a745);
    border-radius: 2px;
  }

  .diff-content .diff-text :deep(.diff-inline-deleted) {
    display: inline;
    padding: 2px 4px;
    color: var(--color-danger, #d73a49);
    text-decoration: line-through;
    background-color: var(--color-danger-light-5, #ffebe9);
    border: 1px solid var(--color-danger, #d73a49);
    border-radius: 2px;
  }

  .diff-content .diff-text :deep(.diff-inline-normal) {
    display: inline;
    color: var(--color-text-1);
  }

  /* 媒体查询 - 响应式设计 */
  @media (width <= 768px) {
    .diff-side-by-side {
      flex-direction: column;
    }
    .left-column {
      border-right: none;
      border-bottom: 1px solid var(--color-border-2);
    }
    .right-column {
      border-top: 1px solid var(--color-border-2);
      border-left: none;
    }
    .diff-divider {
      display: none;
    }
  }

  /* 高对比度模式支持 */
  @media (prefers-contrast: high) {
    .diff-text {
      border: 2px solid var(--color-border-3);
    }
  }

  /* 减少动画模式支持 */
  @media (prefers-reduced-motion: reduce) {
    .diff-text {
      transition: none;
    }
  }
</style>
