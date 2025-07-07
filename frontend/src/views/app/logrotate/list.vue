<template>
  <app-sidebar-layout>
    <template #sidebar>
      <category-tree
        ref="categoryTreeRef"
        v-model:selected-category="params.category"
        :category-config="categoryConfig"
        :enable-category-management="true"
        :host-id="currentHostId"
        :categories="categoryItems"
        :show-title="false"
        @create="handleCategoryCreate"
        @category-created="handleCategoryCreated"
        @category-updated="handleCategoryUpdated"
        @category-deleted="handleCategoryDeleted"
      />
    </template>
    <template #main>
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
          <category-manage-button
            :config="categoryManageConfig"
            @ok="handleCategoryManageOk"
          />
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
    </template>
  </app-sidebar-layout>

  <!-- 子组件 -->
  <form-drawer
    ref="formRef"
    :type="type"
    @ok="handleFormOk"
    @category-change="handleCategoryChange"
  />
  <history-drawer ref="historyRef" />
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

  import { PropType, ref, watch, onMounted, nextTick, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';
  import { LogrotateEntity } from '@/entity/Logrotate';
  import { useLogger } from '@/composables/use-logger';
  import { getLogrotateCategoriesApi } from '@/api/logrotate';
  import { Message } from '@arco-design/web-vue';
  import useCurrentHost from '@/composables/current-host';
  import AppSidebarLayout from '@/components/app-sidebar-layout/index.vue';
  import CategoryTree from '@/components/idb-tree/category-tree.vue';
  import CategoryManageButton from '@/components/idb-tree/components/category-manage-button/index.vue';
  import { createLogrotateCategoryConfig } from './adapters/category-adapter';
  import { createLogrotateCategoryManageConfig } from './adapters/category-manage-adapter';
  import FormDrawer from './components/form-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';

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

  // Composables
  const { currentHostId } = useCurrentHost();

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

  // 分类管理配置
  const categoryConfig = computed(() =>
    createLogrotateCategoryConfig(props.type)
  );

  // 分类管理配置
  const categoryManageConfig = computed(() =>
    createLogrotateCategoryManageConfig(props.type, currentHostId.value || 0)
  );

  // 分类数据状态
  const categoryItems = ref<string[]>([]);
  const categoryLoading = ref(false);

  // 组件引用
  const gridRef = ref<GridInstance>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
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

  /**
   * 加载分类列表
   */
  const loadCategories = async () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error('Host ID is required');
      return;
    }

    // 防止重复加载
    if (categoryLoading.value) {
      logInfo('分类正在加载中，跳过重复请求');
      return;
    }

    logInfo('开始加载分类列表');
    categoryLoading.value = true;
    try {
      const ret = await getLogrotateCategoriesApi(props.type, 1, 1000, hostId);
      logInfo(`分类 API 返回数据:`, ret);

      const newItems = [...ret.items.map((item) => item.name)];
      logInfo(`处理后的分类列表:`, newItems);

      // 如果当前选中的分类不在列表中，添加到列表中
      if (params.value.category && !newItems.includes(params.value.category)) {
        newItems.push(params.value.category);
        logInfo(`添加当前选中分类到列表: ${params.value.category}`);
      }

      categoryItems.value = newItems;
      logInfo(`分类列表已更新，当前选中: ${params.value.category}`);

      // 如果没有选择任何分类且列表不为空，选择第一个分类
      if (!params.value.category && newItems.length > 0) {
        logInfo(`自动选择第一个分类: ${newItems[0]}`);
        params.value.category = newItems[0];
      }
    } catch (err: any) {
      logError('加载分类失败', err);
      Message.error(err?.message || 'Failed to load categories');
    } finally {
      categoryLoading.value = false;
      logInfo('分类加载完成');
    }
  };

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

  /**
   * 处理分类创建
   */
  const handleCategoryCreate = () => {
    // CategoryTree 组件自己会处理分类创建
    // 这里不需要额外的逻辑
  };

  /**
   * 处理分类创建成功
   */
  const handleCategoryCreated = async (categoryName: string) => {
    logInfo(`分类创建成功: ${categoryName}`);
    await loadCategories();
    gridRef.value?.reload();
  };

  /**
   * 处理分类更新成功
   */
  const handleCategoryUpdated = async (oldName: string, newName: string) => {
    logInfo(`分类更新成功: ${oldName} -> ${newName}`);
    await loadCategories();
    gridRef.value?.reload();
  };

  /**
   * 处理分类删除成功
   */
  const handleCategoryDeleted = async (categoryName: string) => {
    logInfo(`分类删除成功: ${categoryName}`);
    await loadCategories();
    gridRef.value?.reload();
  };

  const handleCategoryChange = async (category: string) => {
    if (category) {
      await refreshAndSelectCategory(category);
    }
  };

  const handleNewCategoryUpdate = async (newCategory: string) => {
    try {
      lastManualCategory.value = newCategory;
      await nextTick();

      // 重新加载分类列表以确保新分类在列表中
      await loadCategories();

      await refreshAndSelectCategory(newCategory);

      await nextTick();
      if (params.value.category !== newCategory) {
        params.value.category = newCategory;
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

  // 处理分类管理确认
  const handleCategoryManageOk = async () => {
    await loadCategories();
    gridRef.value?.reload();
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
      // 类型变化时重置组件状态并重新加载分类
      resetComponentsState();
      loadCategories();
    }
  );

  // 暴露方法给父组件
  defineExpose({
    resetComponentsState: () => {
      resetComponentsState();
      loadCategories(); // 重置状态时也要重新加载分类
    },
  });

  onMounted(() => {
    resetComponentsState();
    loadCategories(); // 组件挂载时加载分类列表
  });
</script>

<style scoped>
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
