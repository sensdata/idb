<template>
  <div>
    <!-- 加载更多按钮 -->
    <div v-if="canLoadMore" class="load-more-button">
      <a-button
        type="outline"
        status="normal"
        long
        :loading="isLoadingMore"
        @click="$emit('loadMore')"
      >
        {{
          isLoadingMore
            ? t('app.file.editor.loadingMore')
            : t('app.file.editor.clickToLoadMore')
        }}
      </a-button>
    </div>

    <!-- 文件内容结束提示 -->
    <div v-if="!canLoadMore" class="end-of-content">
      <a-alert type="info" banner>
        {{
          viewMode === 'tail'
            ? t('app.file.editor.reachedStart')
            : t('app.file.editor.reachedEnd')
        }}
      </a-alert>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import { ContentViewMode } from '@/components/file/file-editor-drawer/types';

  const { t } = useI18n();

  defineProps({
    viewMode: {
      type: String as () => ContentViewMode,
      required: true,
    },
    isLoadingMore: {
      type: Boolean,
      default: false,
    },
    canLoadMore: {
      type: Boolean,
      default: true,
    },
  });

  defineEmits(['loadMore']);
</script>

<style scoped>
  .load-more-button {
    position: sticky;
    bottom: 0;
    z-index: 10;
    width: 100%;
    padding: 8px 16px;
    background-color: var(--color-bg-1);
  }

  .end-of-content {
    position: sticky;
    bottom: 0;
    z-index: 10;
    width: 100%;
    padding: 8px 0;
    background-color: var(--color-bg-1);
    opacity: 0.9;
  }
</style>
