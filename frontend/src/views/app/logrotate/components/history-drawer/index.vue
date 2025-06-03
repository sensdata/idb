<template>
  <a-drawer
    v-model:visible="visible"
    :title="$t('app.logrotate.history.title')"
    :width="800"
    :footer="false"
    unmount-on-close
  >
    <div class="history-drawer">
      <a-table
        :data="historyList"
        :loading="loading"
        :pagination="pagination"
        :row-class="getRowClass"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column
            :title="$t('app.logrotate.history.column.commit')"
            data-index="commit"
            :width="140"
          >
            <template #cell="{ record, rowIndex }">
              <div class="commit-cell">
                <a-tag size="small">{{ formatCommit(record.commit) }}</a-tag>
                <a-tag
                  v-if="rowIndex === 0"
                  color="green"
                  size="small"
                  class="current-tag"
                >
                  {{ $t('app.logrotate.history.current') }}
                </a-tag>
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.logrotate.history.column.message')"
            data-index="message"
            ellipsis
            tooltip
          />
          <a-table-column
            :title="$t('app.logrotate.history.column.author')"
            data-index="author"
            :width="100"
          />
          <a-table-column
            :title="$t('app.logrotate.history.column.date')"
            data-index="date"
            :width="160"
          />
          <a-table-column
            :title="$t('common.operation')"
            :width="180"
            fixed="right"
          >
            <template #cell="{ record, rowIndex }">
              <a-button
                type="text"
                size="small"
                :disabled="rowIndex === 0"
                @click="onDiff(record)"
              >
                {{ $t('app.logrotate.history.operation.diff') }}
              </a-button>
              <a-button
                type="text"
                size="small"
                :disabled="rowIndex === 0"
                @click="onRestore(record)"
              >
                {{ $t('app.logrotate.history.operation.restore') }}
              </a-button>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>

    <!-- 文件对比抽屉 -->
    <logrotate-diff-drawer ref="diffDrawerRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { formatCommitHash } from '@/utils/format';
  import type { LogrotateHistory } from '@/entity/Logrotate';
  import type { HistoryParams, HistoryDrawerExpose } from './types';
  import { useHistoryData } from './hooks/use-history-data';
  import LogrotateDiffDrawer from '../diff-drawer/index.vue';
  import type { DiffDrawerExpose } from '../diff-drawer/types';

  const visible = ref(false);
  const diffDrawerRef = ref<DiffDrawerExpose | null>(null);

  // 使用自定义 hook 管理历史数据
  const {
    historyList,
    pagination,
    loading,
    initializeHistory,
    handlePageChange,
    handlePageSizeChange,
    handleRestore,
    currentParams,
    loadHistory,
  } = useHistoryData();

  /**
   * 格式化提交哈希，只显示前8位
   * @param commit - 完整的提交哈希
   * @returns 格式化后的提交哈希
   */
  const formatCommit = formatCommitHash;

  /**
   * 获取表格行的CSS类名
   * @param record - 记录数据
   * @param rowIndex - 行索引
   * @returns CSS类名
   */
  const getRowClass = (record: LogrotateHistory, rowIndex: number): string => {
    return rowIndex === 0 ? 'current-version-row' : '';
  };

  /**
   * 显示历史记录抽屉
   * @param params - 历史记录查询参数
   */
  const show = (params: HistoryParams): void => {
    visible.value = true;
    initializeHistory(params);
  };

  /**
   * 处理文件对比操作
   * @param record - 要对比的历史记录
   */
  const onDiff = (record: LogrotateHistory): void => {
    if (!diffDrawerRef.value || !currentParams.value) {
      return;
    }

    diffDrawerRef.value.show(
      {
        type: currentParams.value.type,
        category: currentParams.value.category,
        name: currentParams.value.name,
        commit: record.commit,
      },
      () => {
        // 恢复成功后刷新历史数据
        loadHistory();
      }
    );
  };

  /**
   * 处理恢复操作
   * @param record - 要恢复的历史记录
   */
  const onRestore = async (record: LogrotateHistory): Promise<void> => {
    const success = await handleRestore(record);
    if (success) {
      visible.value = false;
    }
  };

  // 暴露方法给父组件，使用明确的类型定义
  defineExpose<HistoryDrawerExpose>({
    show,
  });
</script>

<style scoped>
  .history-drawer {
    height: 100%;
  }

  .commit-cell {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .current-tag {
    height: 18px;
    padding: 0 4px;
    font-size: 10px;
    line-height: 16px;
  }

  :deep(.current-version-row) {
    background-color: var(--color-fill-1);
  }

  :deep(.current-version-row:hover) {
    background-color: var(--color-fill-2) !important;
  }
</style>
