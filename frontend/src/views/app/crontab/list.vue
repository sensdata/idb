<template>
  <app-sidebar-layout>
    <template #sidebar>
      <category-tree
        ref="categoryTreeRef"
        v-model:selected-category="params.category"
        :category-config="categoryConfig"
        :category-manage-config="categoryManageConfig"
        :enable-category-management="true"
        :host-id="currentHostId"
        :categories="categoryItems"
        :show-title="true"
        @create="handleCategoryCreate"
        @category-created="handleCategoryCreated"
        @category-updated="handleCategoryUpdated"
        @category-deleted="handleCategoryDeleted"
        @category-manage-ok="handleCategoryManageOk"
      />
    </template>
    <template #main>
      <idb-table
        ref="gridRef"
        class="crontab-table"
        :loading="loading"
        :params="params"
        :columns="columns"
        :fetch="fetchCrontabList"
      >
        <template #leftActions>
          <a-button type="primary" @click="handleCreate">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.crontab.list.action.create') }}
          </a-button>
        </template>
        <template #status="{ record }: { record: CrontabEntity }">
          <div class="status-cell">
            <span
              v-if="record.linked === true"
              class="status-tag"
              style="color: rgb(var(--success-6))"
            >
              生效中
            </span>
            <span
              v-else-if="record.linked === false"
              class="status-tag"
              style="color: rgb(var(--color-text-4))"
            >
              未激活
            </span>
            <span v-else class="status-tag">
              {{ record.linked }}
            </span>
          </div>
        </template>
        <template #operation="{ record }: { record: CrontabEntity }">
          <idb-table-operation
            type="button"
            :options="getCrontabOperationOptions(record)"
          />
        </template>
      </idb-table>
    </template>
  </app-sidebar-layout>
  <form-drawer
    ref="formRef"
    :type="type"
    @ok="handleFormOk"
    @category-change="handleCategoryChange"
  />
  <logs-drawer ref="logsRef" />
</template>

<script setup lang="ts">
  import {
    GlobalComponents,
    PropType,
    ref,
    watch,
    onMounted,
    nextTick,
    computed,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { formatTimeWithoutSeconds } from '@/utils/format';
  import { CrontabEntity } from '@/entity/Crontab';
  import {
    deleteCrontabApi,
    getCrontabListApi,
    actionCrontabApi,
    getCrontabCategoryListApi,
  } from '@/api/crontab';
  import useLoading from '@/composables/loading';
  import { useConfirm } from '@/composables/confirm';
  import usetCurrentHost from '@/composables/current-host';
  import AppSidebarLayout from '@/components/app-sidebar-layout/index.vue';
  import CategoryTree from '@/components/idb-tree/category-tree.vue';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import { useCronDescription } from './components/form-drawer/composables/use-cron-description';
  import { createCrontabCategoryConfig } from './adapters/category-adapter';
  import { createCrontabCategoryManageConfig } from './adapters/category-manage-adapter';

  // 定义组件引用类型接口，提高类型安全性
  interface FormDrawerInstance extends InstanceType<typeof FormDrawer> {
    setParams: (params: {
      name?: string;
      type?: CRONTAB_TYPE;
      category?: string;
      isEdit?: boolean;
      record?: CrontabEntity;
    }) => void;
    loadData: () => Promise<void>;
    show: (params?: {
      name?: string;
      type?: CRONTAB_TYPE;
      category?: string;
      isEdit?: boolean;
      record?: CrontabEntity;
    }) => Promise<void>;
    hide: () => void;
  }

  const props = defineProps({
    type: {
      type: String as PropType<CRONTAB_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { getCronDescriptionFromContent } = useCronDescription();

  // 获取当前主机ID
  const { currentHostId } = usetCurrentHost();

  // 创建分类管理配置
  const categoryConfig = computed(() =>
    createCrontabCategoryConfig(props.type)
  );

  // 创建分类管理抽屉配置
  const categoryManageConfig = computed(() =>
    createCrontabCategoryManageConfig(props.type)
  );

  // 分类数据状态
  const categoryItems = ref<string[]>([]);

  // 加载分类列表
  const loadCategories = async () => {
    try {
      const response = await getCrontabCategoryListApi({
        type: props.type,
        page: 1,
        page_size: 1000,
      });
      categoryItems.value = response.items.map((item: any) => item.name);
    } catch (error) {
      console.error('Failed to load categories:', error);
      categoryItems.value = [];
    }
  };

  // 组件引用
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const formRef = ref<FormDrawerInstance>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();

  // 查询参数
  const params = ref<{
    type: CRONTAB_TYPE;
    category: string;
    page: number;
    page_size: number;
  }>({
    type: props.type,
    category: '',
    page: 1,
    page_size: 20,
  });

  // 记录最后一次手动设置的分类，用于防止重置
  const lastManualCategory = ref<string>('');

  // 从记录中提取周期信息的工具函数
  function extractPeriodFromRecord(record: CrontabEntity): string {
    if (!record.content) {
      return record.period_expression || '';
    }

    // 使用 cronstrue 从内容中提取周期描述
    const description = getCronDescriptionFromContent(record.content);
    if (description) return description;

    // 最后回退到period_expression
    return record.period_expression || '';
  }

  // 表格列定义
  const columns = [
    {
      dataIndex: 'name',
      title: t('app.crontab.list.column.name'),
      width: 150,
    },
    {
      dataIndex: 'status',
      title: t('app.crontab.list.column.status'),
      width: 120,
      align: 'left' as const,
      slotName: 'status',
    },
    {
      dataIndex: 'period',
      title: t('app.crontab.list.column.period'),
      width: 150,
      align: 'left' as const,
      render: ({ record }: { record: CrontabEntity }) =>
        extractPeriodFromRecord(record),
    },
    {
      dataIndex: 'mod_time',
      title: t('app.crontab.list.column.mod_time'),
      width: 160,
      align: 'left' as const,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTimeWithoutSeconds(record.mod_time);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 210,
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  // 获取定时任务列表
  const fetchCrontabList = async (fetchParams: Record<string, unknown>) => {
    // 如果category为空或未设置，不发送请求
    if (!params.value.category || params.value.category.trim() === '') {
      // 返回空数据，避免API请求
      return Promise.resolve({
        items: [],
        total: 0,
        page: 1,
        page_size: params.value.page_size,
      });
    }

    // 检查hostId是否存在
    if (!currentHostId.value) {
      console.warn('hostId is undefined, skipping API request');
      return Promise.resolve({
        items: [],
        total: 0,
        page: 1,
        page_size: params.value.page_size,
      });
    }

    // 创建参数的副本，避免修改原始对象
    const requestParams = {
      ...fetchParams,
      type: params.value.type,
      category: params.value.category,
    };

    try {
      // 始终使用v-model绑定的category值，确保与左侧目录树同步
      return await getCrontabListApi(requestParams);
    } catch (error) {
      console.error('getCrontabListApi error:', error);
      // 其他错误继续抛出
      throw error;
    }
  };

  // 重新加载表格数据
  const reload = () => {
    gridRef.value?.reload();
  };

  // 强制刷新分类树并选择指定分类
  const refreshAndSelectCategory = async (category: string) => {
    if (!category) return;

    // 记录这是手动设置的分类
    lastManualCategory.value = category;

    // 设置参数，确保选择正确的分类
    params.value.category = category;

    // 刷新分类树并选择类别
    if (categoryTreeRef.value) {
      try {
        await categoryTreeRef.value.refresh();
        await nextTick();
        categoryTreeRef.value.selectCategory(category);
        reload();
      } catch (e) {
        // 即使刷新失败，也尝试选择分类并刷新表格
        categoryTreeRef.value.selectCategory(category);
        reload();
      }
    } else {
      // 如果分类树不可用，直接设置分类并刷新
      params.value.category = category;
      reload();
    }
  };

  // 监听分类变化，重新加载列表
  watch(
    () => params.value.category,
    (newCategory) => {
      // 确保每次类别变化都会记录最新值
      if (newCategory) {
        lastManualCategory.value = newCategory;
      }

      // 如果有最近手动设置的分类，并且新分类为空或与手动设置的不同，再次尝试选择手动设置的分类
      if (
        lastManualCategory.value &&
        (newCategory === '' || newCategory !== lastManualCategory.value)
      ) {
        nextTick(() => {
          if (params.value.category !== lastManualCategory.value) {
            params.value.category = lastManualCategory.value;
            categoryTreeRef.value?.selectCategory(lastManualCategory.value);
          }
        });
        return;
      }

      // 只有当分类值存在且不为空字符串时才刷新
      if (newCategory && newCategory.trim() !== '') {
        reload();
      }
    }
  );

  // 处理创建定时任务
  const handleCreate = () => {
    const currentCategory = params.value.category || '';
    formRef.value?.show({
      type: params.value.type,
      category: currentCategory,
    });
  };

  // 处理表单确认
  const handleFormOk = async (newCategory?: string) => {
    // 如果有新的分类返回
    if (newCategory) {
      // 记录为手动设置的分类
      lastManualCategory.value = newCategory;

      // 使用nextTick替代setTimeout
      await nextTick();

      // 尝试直接选择分类
      if (categoryTreeRef.value) {
        categoryTreeRef.value.selectCategory(newCategory);
      }

      // 刷新并选择分类
      await refreshAndSelectCategory(newCategory);

      // 再次确认选中状态
      await nextTick();
      if (params.value.category !== newCategory) {
        params.value.category = newCategory;
        if (categoryTreeRef.value) {
          categoryTreeRef.value.selectCategory(newCategory);
        }
        reload();
      }
    } else {
      // 仅重新加载表格数据
      reload();
    }
  };

  // 处理编辑定时任务
  const handleEdit = async (record: CrontabEntity) => {
    if (!formRef.value) return;
    try {
      formRef.value.show({
        record,
        isEdit: true,
        category: params.value.category, // Pass the current category to ensure it's selected in the form
      });
    } catch (error) {
      Message.error(t('app.crontab.list.message.edit_error'));
    }
  };

  // 处理定时任务操作（激活/停用）
  const handleAction = async (
    record: CrontabEntity,
    action: 'activate' | 'deactivate'
  ) => {
    try {
      await actionCrontabApi({
        type: params.value.type,
        category: params.value.category,
        name: record.name,
        action,
      });
      // 根据操作类型显示不同的成功消息
      const messageKey =
        action === 'activate'
          ? 'app.crontab.list.message.activate_success'
          : 'app.crontab.list.message.deactivate_success';
      Message.success(t(messageKey));
      // 等待一小段时间后再刷新，确保后端状态已更新
      // eslint-disable-next-line no-promise-executor-return
      await new Promise((resolve) => setTimeout(resolve, 1000));
      // 刷新表格数据
      if (gridRef.value) {
        await gridRef.value.load();
      }
    } catch (err) {
      if (err instanceof Error) {
        Message.error(err.message);
      } else {
        Message.error(String(err));
      }
    }
  };

  // 处理删除定时任务
  const handleDelete = async (record: CrontabEntity) => {
    if (
      await confirm({
        content: t('app.crontab.list.delete.confirm', { name: record.name }),
      })
    ) {
      setLoading(true);
      try {
        await deleteCrontabApi({
          name: record.name,
          type: params.value.type,
          category: params.value.category,
        });
        Message.success(t('app.crontab.list.message.delete_success'));
        reload();
      } catch (err) {
        if (err instanceof Error) {
          Message.error(err.message);
        } else {
          Message.error(String(err));
        }
      } finally {
        setLoading(false);
      }
    }
  };

  // 获取操作按钮配置
  const getCrontabOperationOptions = (record: CrontabEntity) => [
    {
      text: t('common.edit'),
      click: () => handleEdit(record),
    },
    {
      text: record.linked
        ? t('app.crontab.list.operation.deactivate')
        : t('app.crontab.list.operation.activate'),
      click: () =>
        handleAction(record, record.linked ? 'deactivate' : 'activate'),
    },
    {
      text: t('common.delete'),
      status: 'danger' as const,
      confirm: t('app.crontab.list.delete.confirm', { name: record.name }),
      click: () => handleDelete(record),
    },
  ];

  // 处理分类管理功能已集成到 CategoryManageButton 组件中

  // 处理分类管理确认
  const handleCategoryManageOk = () => {
    // 重新加载分类列表
    loadCategories();
    // 刷新左侧分类树
    categoryTreeRef.value?.refresh();
    // 刷新表格数据
    reload();
  };

  // 从分类树中选择第一个可用分类
  const selectFirstAvailableCategory = () => {
    // 直接使用categoryItems.value而不是依赖组件内部状态
    if (categoryItems.value && categoryItems.value.length > 0) {
      const firstCategory = categoryItems.value[0];
      refreshAndSelectCategory(firstCategory);
    }
  };

  // 处理任务创建时分类变更
  const handleCategoryChange = (category: string) => {
    if (category) {
      // 如果创建了新分类，强制刷新并选择
      lastManualCategory.value = category;
      refreshAndSelectCategory(category);
    } else {
      // 仅刷新分类树
      categoryTreeRef.value?.refresh();
    }
  };

  // 处理分类创建
  const handleCategoryCreate = () => {
    // 触发分类管理功能
    // 由于管理按钮现在在分类树头部，这里可以通过分类树组件访问
    if (
      categoryTreeRef.value &&
      categoryTreeRef.value.categoryManageButtonRef
    ) {
      categoryTreeRef.value.categoryManageButtonRef.show();
    }
  };

  // 处理分类创建成功
  const handleCategoryCreated = (categoryName: string) => {
    // 重新加载分类列表
    loadCategories();
    // 选择新创建的分类
    refreshAndSelectCategory(categoryName);
  };

  // 处理分类更新成功
  const handleCategoryUpdated = (oldName: string, newName: string) => {
    // 重新加载分类列表
    loadCategories();
    // 如果当前选中的是被更新的分类，则选择新名称
    if (params.value.category === oldName) {
      refreshAndSelectCategory(newName);
    }
  };

  // 处理分类删除成功
  const handleCategoryDeleted = (categoryName: string) => {
    // 重新加载分类列表
    loadCategories();
    // 如果当前选中的是被删除的分类，则清空选择
    if (params.value.category === categoryName) {
      params.value.category = '';
      selectFirstAvailableCategory();
    }
  };

  // 组件挂载完成后，确保分类初始化不会自动重置
  onMounted(async () => {
    // 先加载分类列表
    await loadCategories();

    // 等待分类数据加载完成后再进行初始化
    await nextTick();

    // 如果有分类数据，直接选择第一个分类，避免复杂的重置逻辑
    if (categoryItems.value && categoryItems.value.length > 0) {
      const firstCategory = categoryItems.value[0];
      lastManualCategory.value = firstCategory;
      await refreshAndSelectCategory(firstCategory);
    }
  });
</script>

<style scoped>
  .crontab-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }

  .status-tag {
    padding: 4px 8px;
    font-size: 12px;
    border-radius: 4px;
  }
</style>
