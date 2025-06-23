<template>
  <div class="logrotate-layout">
    <div class="logrotate-sidebar">
      <category-tree
        ref="categoryTreeRef"
        v-model:selected="params.category"
        :type="type"
      />
    </div>
    <div class="logrotate-main">
      <idb-table
        ref="gridRef"
        class="logrotate-table"
        :loading="loading"
        :params="params"
        :columns="columns"
        :fetch="fetchLogrotateList"
        :auto-load="false"
      >
        <template #leftActions>
          <a-button :type="BUTTON_TYPE_PRIMARY" @click="handleCreate">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.logrotate.list.action.create') }}
          </a-button>
          <a-button @click="handleCategoryManage">
            <template #icon>
              <icon-settings />
            </template>
            {{ $t('app.logrotate.category.manage.title') }}
          </a-button>
        </template>

        <template #status="{ record }: { record: LogrotateEntity }">
          <div class="status-cell">
            <a-tag :color="getStatusInfo(record).color" class="status-tag">
              {{ getStatusInfo(record).text }}
            </a-tag>
          </div>
        </template>

        <template #count="{ record }: { record: LogrotateEntity }">
          {{ formatCount(record) }}
        </template>

        <template #frequency="{ record }: { record: LogrotateEntity }">
          {{ formatFrequency(record) }}
        </template>

        <template #operation="{ record }: { record: LogrotateEntity }">
          <div class="operation">
            <a-button
              :type="BUTTON_TYPE_TEXT"
              :size="BUTTON_SIZE"
              @click="handleEdit(record)"
            >
              {{ $t('common.edit') }}
            </a-button>
            <a-button
              :type="BUTTON_TYPE_TEXT"
              :size="BUTTON_SIZE"
              @click="handleActionClick(record)"
            >
              {{ getActionButtonInfo(record).text }}
            </a-button>
            <a-button
              :type="BUTTON_TYPE_TEXT"
              :size="BUTTON_SIZE"
              @click="handleHistory(record)"
            >
              {{ $t('app.logrotate.list.operation.history') }}
            </a-button>
            <a-button
              :type="BUTTON_TYPE_TEXT"
              :size="BUTTON_SIZE"
              :status="BUTTON_STATUS_DANGER"
              @click="handleDeleteClick(record)"
            >
              {{ $t('common.delete') }}
            </a-button>
          </div>
        </template>
      </idb-table>
    </div>
  </div>

  <!-- 子组件 -->
  <form-drawer
    ref="formRef"
    :type="type"
    @ok="handleFormOk"
    @category-change="handleCategoryChange"
  />
  <history-drawer ref="historyRef" />
  <category-manage
    ref="categoryManageRef"
    :type="type"
    @ok="handleCategoryManageOk"
  />
</template>

<script setup lang="ts">
  /**
   * Logrotate 列表组件
   *
   * 功能：
   * - 显示日志轮转配置列表
   * - 支持分类管理和筛选
   * - 提供创建、编辑、删除、启用/禁用等操作
   * - 支持查看历史记录
   */

  import { PropType, ref, watch, onMounted, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';
  import { LogrotateEntity } from '@/entity/Logrotate';
  import { useLogger } from '@/composables/use-logger';
  import FormDrawer from './components/form-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';
  import CategoryTree from './components/category-tree/index.vue';
  import CategoryManage from './components/category-manage/index.vue';

  // 使用组合式函数
  import { useLogrotateList } from './composables/use-logrotate-list';
  import { useLogrotateColumns } from './composables/use-logrotate-columns';
  import { useLogrotateActions } from './composables/use-logrotate-actions';
  import { useCategoryManagement } from './composables/use-category-management';
  import { LAYOUT_CONFIG } from './constants';

  // 常量定义
  const BUTTON_SIZE = 'small' as const;
  const BUTTON_TYPE_TEXT = 'text' as const;
  const BUTTON_TYPE_PRIMARY = 'primary' as const;
  const BUTTON_STATUS_DANGER = 'danger' as const;

  // 国际化
  const { t } = useI18n();

  // 日志记录
  const { logError, logInfo } = useLogger('LogrotateList');

  // 定义组件引用类型接口
  interface FormDrawerInstance extends InstanceType<typeof FormDrawer> {
    show: (params?: {
      name?: string;
      type?: LOGROTATE_TYPE;
      category?: string;
      isEdit?: boolean;
      record?: LogrotateEntity;
    }) => Promise<void>;
  }

  // 定义表格组件引用类型
  interface GridInstance {
    reload: () => void;
    load: (params?: any) => void;
  }

  const props = defineProps({
    type: {
      type: String as PropType<LOGROTATE_TYPE>,
      required: true,
    },
  });

  // 组件引用
  const gridRef = ref<GridInstance>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const categoryManageRef = ref<InstanceType<typeof CategoryManage>>();
  const formRef = ref<FormDrawerInstance>();
  const historyRef = ref<InstanceType<typeof HistoryDrawer>>();

  // 使用组合式函数
  const {
    params,
    loading,
    fetchLogrotateList,
    deleteLogrotate,
    toggleLogrotateStatus,
  } = useLogrotateList(props.type);

  const { columns } = useLogrotateColumns();
  const { getStatusInfo, getActionButtonInfo } = useLogrotateActions();

  // 格式化轮转数量，从 content 中解析实际值
  const formatCount = (record: LogrotateEntity): string => {
    if (!record) return '';

    // 从配置内容中解析轮转次数
    const parseRotateCount = (configContent: string): string => {
      const rotateMatch = configContent.match(/^\s*rotate\s+(\d+)\s*$/m);
      return rotateMatch && rotateMatch[1] ? rotateMatch[1] : '7';
    };

    // 优先从 content 中解析
    if (record.content) {
      return parseRotateCount(record.content);
    }

    // 回退到 count 字段
    if (record.count) {
      if (record.count.startsWith('rotate ')) {
        return record.count.replace(/^rotate\s+/, '');
      }
      return record.count;
    }

    return '7'; // 默认值
  };

  // 格式化轮转频率，从 content 中解析实际值并翻译成更友好的文本
  const formatFrequency = (record: LogrotateEntity): string => {
    if (!record) return '';

    // 从配置内容中解析轮转频率
    const parseFrequency = (configContent: string): LOGROTATE_FREQUENCY => {
      const frequencyMatch = configContent.match(
        /^\s*(daily|weekly|monthly|yearly)\s*$/m
      );
      if (frequencyMatch && frequencyMatch[1]) {
        const value = frequencyMatch[1].toLowerCase();
        switch (value) {
          case 'daily':
            return LOGROTATE_FREQUENCY.Daily;
          case 'weekly':
            return LOGROTATE_FREQUENCY.Weekly;
          case 'monthly':
            return LOGROTATE_FREQUENCY.Monthly;
          case 'yearly':
            return LOGROTATE_FREQUENCY.Yearly;
          default:
            return LOGROTATE_FREQUENCY.Daily;
        }
      }
      return LOGROTATE_FREQUENCY.Daily;
    };

    // 将频率转换为用户友好的文本
    const getFrequencyText = (frequency: LOGROTATE_FREQUENCY): string => {
      return t(`app.logrotate.frequency.${frequency}`);
    };

    // 优先从 content 中解析
    if (record.content) {
      const frequency = parseFrequency(record.content);
      return getFrequencyText(frequency);
    }

    // 回退到 frequency 字段
    if (record.frequency) {
      return getFrequencyText(record.frequency);
    }

    return getFrequencyText(LOGROTATE_FREQUENCY.Daily); // 默认值
  };

  const { lastManualCategory, refreshAndSelectCategory, resetComponentsState } =
    useCategoryManagement(categoryTreeRef, gridRef, params);

  // 事件处理函数
  const handleCreate = () => {
    formRef.value?.show({
      type: props.type,
      category: params.value.category,
      isEdit: false,
    });
  };

  const handleEdit = (record: LogrotateEntity) => {
    formRef.value?.show({
      type: props.type,
      category: record.category,
      name: record.name,
      isEdit: true,
      record,
    });
  };

  const handleDeleteClick = async (record: LogrotateEntity) => {
    try {
      const success = await deleteLogrotate(record);
      if (success) {
        logInfo(`删除配置成功: ${record.name}`);
        gridRef.value?.reload();
      }
    } catch (error) {
      logError('删除失败', error as Error);
      // 这里可以添加用户友好的错误提示
    }
  };

  const handleActionClick = async (record: LogrotateEntity) => {
    try {
      const actionInfo = getActionButtonInfo(record);
      const success = await toggleLogrotateStatus(record, actionInfo.action);
      if (success) {
        logInfo(`${actionInfo.action}配置成功: ${record.name}`);
        gridRef.value?.reload();
      }
    } catch (error) {
      logError('操作失败', error as Error);
    }
  };

  const handleHistory = (record: LogrotateEntity) => {
    historyRef.value?.show({
      type: record.type,
      category: record.category,
      name: record.name,
    });
  };

  const handleCategoryManage = () => {
    categoryManageRef.value?.show();
  };

  const handleCategoryChange = async (category: string) => {
    if (category) {
      await refreshAndSelectCategory(category);
    }
  };

  const handleCategoryManageOk = async () => {
    try {
      // First, clear the current category selection to prevent immediate invalid API calls
      const previousCategory = params.value.category;
      params.value.category = '';

      // Next, refresh the category tree
      if (categoryTreeRef.value) {
        await categoryTreeRef.value.refresh();
        logInfo('Category tree refreshed after category management');

        // Only select the previous category if it still exists
        if (
          previousCategory &&
          categoryTreeRef.value.items &&
          categoryTreeRef.value.items.includes(previousCategory)
        ) {
          params.value.category = previousCategory;
          logInfo(`Restored previous category selection: ${previousCategory}`);
        } else if (previousCategory) {
          logInfo(
            `Previous category '${previousCategory}' no longer exists, selection cleared`
          );
        }
      }

      // Only after the category tree is refreshed and category selection is updated, reload the grid
      // This prevents errors from trying to load a deleted category
      gridRef.value?.reload();
    } catch (error) {
      logError('Error refreshing category tree', error as Error);
    }
  };

  // 处理新分类的选择和更新
  const handleNewCategoryUpdate = async (newCategory: string) => {
    try {
      lastManualCategory.value = newCategory;
      await nextTick();

      if (categoryTreeRef.value) {
        categoryTreeRef.value.selectCategory(newCategory);
      }

      await refreshAndSelectCategory(newCategory);

      await nextTick();
      if (params.value.category !== newCategory) {
        params.value.category = newCategory;
        if (categoryTreeRef.value) {
          categoryTreeRef.value.selectCategory(newCategory);
        }
        gridRef.value?.reload();
      }

      logInfo(`分类更新成功: ${newCategory}`);
    } catch (error) {
      logError('分类更新失败', error as Error);
    }
  };

  const handleFormOk = async (newCategory?: string) => {
    if (newCategory) {
      await handleNewCategoryUpdate(newCategory);
    } else {
      gridRef.value?.reload();
    }
  };

  // 监听器
  watch(
    () => params.value.category,
    (newCategory, oldCategory) => {
      logInfo(`监听器触发 - 分类变化: ${oldCategory} -> ${newCategory}`);
      // 当分类变化时触发重新加载
      // 包括从空变为有值，或从一个值变为另一个值
      if (newCategory && newCategory !== oldCategory) {
        logInfo(`触发表格重新加载，分类: ${newCategory}`);
        gridRef.value?.reload();
      } else {
        logInfo(
          `跳过表格重新加载，原因: newCategory=${newCategory}, oldCategory=${oldCategory}`
        );
      }
    }
  );

  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      params.value.category = '';
      lastManualCategory.value = '';
      // 类型变化时重置组件状态
      resetComponentsState();
    }
  );

  // 暴露方法给父组件
  defineExpose({
    resetComponentsState,
  });

  onMounted(() => {
    resetComponentsState();
  });
</script>

<style scoped>
  .logrotate-layout {
    position: relative;
    min-height: calc(100vh - 240px);
    padding-left: 240px;
    margin-top: 20px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .logrotate-sidebar {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 240px;
    height: 100%;
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
  }

  .logrotate-main {
    min-width: 0;
    height: 100%;
    padding: 20px;
  }

  .logrotate-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    align-items: center;
  }

  .status-tag {
    margin: 0;
  }

  .operation {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    justify-content: center;
    min-width: v-bind('LAYOUT_CONFIG.OPERATION_MIN_WIDTH + "px"');
  }

  .operation :deep(.arco-btn-size-small) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
