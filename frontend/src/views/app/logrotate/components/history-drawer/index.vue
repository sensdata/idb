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
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column
            :title="$t('app.logrotate.history.column.commit')"
            data-index="commit"
            :width="120"
          >
            <template #cell="{ record }">
              <a-tag size="small">{{ formatCommit(record.commit) }}</a-tag>
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
            :width="120"
            fixed="right"
          >
            <template #cell="{ record }">
              <a-button type="text" size="small" @click="onRestore(record)">
                {{ $t('app.logrotate.history.operation.restore') }}
              </a-button>
            </template>
          </a-table-column>
        </template>
      </a-table>
    </div>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import type { LogrotateHistory } from '@/entity/Logrotate';
  import type { HistoryParams, HistoryDrawerExpose } from './types';
  import { useHistoryData } from './hooks/use-history-data';

  const visible = ref(false);

  // 使用自定义 hook 管理历史数据
  const {
    historyList,
    pagination,
    loading,
    initializeHistory,
    handlePageChange,
    handlePageSizeChange,
    handleRestore,
  } = useHistoryData();

  /**
   * 格式化提交哈希，只显示前8位
   * @param commit - 完整的提交哈希
   * @returns 格式化后的提交哈希
   */
  const formatCommit = (commit: string): string => {
    return commit.substring(0, 8);
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
</style>
