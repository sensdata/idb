<template>
  <a-drawer
    v-model:visible="visible"
    :title="$t('app.service.history.title')"
    :width="DRAWER_WIDTH"
    :footer="false"
    unmount-on-close
    :aria-labelledby="titleId"
    role="dialog"
  >
    <div class="history-drawer">
      <a-table
        :data="historyList"
        :loading="loading"
        :pagination="pagination"
        :row-class="getRowClass"
        :aria-label="$t('app.service.history.table.aria_label')"
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
      >
        <template #columns>
          <a-table-column
            :title="$t('app.service.history.column.commit')"
            data-index="commit"
            :width="COLUMN_WIDTHS.commit"
          >
            <template #cell="{ record, rowIndex }">
              <div class="commit-cell">
                <a-tag
                  size="small"
                  :aria-label="
                    $t('app.service.history.commit.aria_label', {
                      commit: formatCommitHash(record.commit),
                    })
                  "
                >
                  {{ formatCommitHash(record.commit) }}
                </a-tag>
                <a-tag
                  v-if="isCurrentVersion(rowIndex)"
                  :color="'rgb(var(--success-6))'"
                  size="small"
                  class="current-tag"
                  :aria-label="$t('app.service.history.current.aria_label')"
                >
                  {{ $t('app.service.history.current') }}
                </a-tag>
              </div>
            </template>
          </a-table-column>
          <a-table-column
            :title="$t('app.service.history.column.message')"
            data-index="message"
            ellipsis
            tooltip
            :min-width="COLUMN_WIDTHS.message"
          />
          <a-table-column
            :title="$t('app.service.history.column.author')"
            data-index="author"
            :width="COLUMN_WIDTHS.author"
          />
          <a-table-column
            :title="$t('app.service.history.column.date')"
            data-index="date"
            :width="COLUMN_WIDTHS.date"
          />
          <a-table-column
            :title="$t('common.operation')"
            :width="COLUMN_WIDTHS.operation"
            fixed="right"
          >
            <template #cell="{ record, rowIndex }">
              <div class="operation-buttons">
                <a-button
                  type="text"
                  size="small"
                  :disabled="isCurrentVersion(rowIndex)"
                  :aria-label="
                    $t('app.service.history.operation.diff.aria_label', {
                      commit: formatCommitHash(record.commit),
                    })
                  "
                  @click="onDiff(record)"
                >
                  {{ $t('app.service.history.operation.diff') }}
                </a-button>
                <a-button
                  type="text"
                  size="small"
                  :disabled="isCurrentVersion(rowIndex)"
                  :aria-label="
                    $t('app.service.history.operation.restore.aria_label', {
                      commit: formatCommitHash(record.commit),
                    })
                  "
                  @click="onRestore(record)"
                >
                  {{ $t('app.service.history.operation.restore') }}
                </a-button>
              </div>
            </template>
          </a-table-column>
        </template>

        <!-- Empty state -->
        <template #empty>
          <div class="empty-state">
            <div class="empty-state__icon">📋</div>
            <div class="empty-state__text">
              {{ $t('app.service.history.empty.message') }}
            </div>
          </div>
        </template>
      </a-table>
    </div>

    <!-- 文件对比抽屉 -->
    <service-diff-drawer ref="diffDrawerRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, computed, nextTick, onMounted } from 'vue';
  import { formatCommitHash } from '@/utils/format';
  import { useLogger } from '@/composables/use-logger';
  import type { ServiceHistoryEntity } from '@/entity/Service';
  import type { HistoryParams, HistoryDrawerExpose } from './types';
  import { useHistoryData } from './composables/use-history-data';
  import ServiceDiffDrawer from '../diff-drawer/index.vue';
  import type { DiffDrawerExpose } from '../diff-drawer/types';

  // 常量定义
  const DRAWER_WIDTH = 800;
  const COLUMN_WIDTHS = {
    commit: 140,
    message: 200, // 最小宽度
    author: 100,
    date: 160,
    operation: 180,
  } as const;

  // 响应式状态
  const visible = ref(false);
  const diffDrawerRef = ref<DiffDrawerExpose | null>(null);
  const titleId = ref('');

  // 日志工具
  const { logWarn, logError } = useLogger('HistoryDrawer');

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

  // 计算属性
  const isCurrentVersion = computed(() => {
    return (rowIndex: number): boolean => rowIndex === 0;
  });

  /**
   * 获取表格行的CSS类名 (使用计算属性优化性能)
   */
  const getRowClass = computed(() => {
    return (record: ServiceHistoryEntity, rowIndex: number): string => {
      return isCurrentVersion.value(rowIndex) ? 'current-version-row' : '';
    };
  });

  // 生命周期
  onMounted(() => {
    titleId.value = `history-drawer-${Date.now()}-${Math.random()
      .toString(36)
      .substr(2, 9)}`;
  });

  /**
   * 显示历史记录抽屉
   * @param params - 历史记录查询参数
   */
  const show = async (params: HistoryParams): Promise<void> => {
    visible.value = true;
    await nextTick(); // 确保DOM更新完成
    initializeHistory(params);
  };

  /**
   * 处理文件对比操作
   * @param record - 要对比的历史记录
   */
  const onDiff = (record: ServiceHistoryEntity): void => {
    if (!diffDrawerRef.value || !currentParams.value) {
      logWarn('DiffDrawer ref or currentParams is not available');
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
  const onRestore = async (record: ServiceHistoryEntity): Promise<void> => {
    try {
      const success = await handleRestore(record);
      if (success) {
        visible.value = false;
      }
    } catch (error) {
      logError('Restore operation failed:', error);
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
    min-height: 400px;
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

  .operation-buttons {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 48px 16px;
    color: var(--color-text-3);
  }

  .empty-state__icon {
    margin-bottom: 16px;
    font-size: 48px;
    opacity: 0.5;
  }

  .empty-state__text {
    font-size: 14px;
  }

  /* 深度选择器样式 */
  :deep(.current-version-row) {
    background-color: var(--color-fill-1);
  }

  :deep(.current-version-row:hover) {
    background-color: var(--color-fill-2) !important;
  }

  /* 响应式设计 */
  @media (width <= 768px) {
    .history-drawer {
      min-height: 300px;
    }
    .operation-buttons {
      flex-direction: column;
      gap: 2px;
    }
  }

  /* 无障碍访问增强 */
  @media (prefers-reduced-motion: reduce) {
    :deep(.arco-drawer-body) {
      transition: none;
    }
  }
</style>
