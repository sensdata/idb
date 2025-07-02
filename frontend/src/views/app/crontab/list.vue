<template>
  <div class="crontab-main">
    <div class="crontab-content">
      <div class="crontab-left-panel">
        <category-tree
          ref="categoryTreeRef"
          v-model:selected="params.category"
          :type="type"
        />
      </div>
      <div class="crontab-right-panel">
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
            <a-button @click="handleCategoryManage">
              <template #icon>
                <icon-settings />
              </template>
              {{ $t('app.crontab.category.manage.title') }}
            </a-button>
          </template>
          <template #status="{ record }: { record: CrontabEntity }">
            <div class="status-cell">
              <a-tag
                :color="record.linked ? 'green' : 'gray'"
                class="status-tag"
              >
                {{
                  record.linked
                    ? $t('app.crontab.list.status.running')
                    : $t('app.crontab.list.status.not_running')
                }}
              </a-tag>
            </div>
          </template>
          <template #operation="{ record }: { record: CrontabEntity }">
            <div class="operation">
              <a-button type="text" size="small" @click="handleEdit(record)">
                {{ $t('common.edit') }}
              </a-button>
              <a-button
                type="text"
                size="small"
                @click="
                  handleAction(
                    record,
                    record.linked ? 'deactivate' : 'activate'
                  )
                "
              >
                {{
                  record.linked
                    ? $t('app.crontab.list.operation.deactivate')
                    : $t('app.crontab.list.operation.activate')
                }}
              </a-button>
              <a-button
                type="text"
                size="small"
                status="danger"
                @click="handleDelete(record)"
              >
                {{ $t('common.delete') }}
              </a-button>
            </div>
          </template>
        </idb-table>
      </div>
    </div>
    <form-drawer
      ref="formRef"
      :type="type"
      @ok="handleFormOk"
      @category-change="handleCategoryChange"
    />
    <logs-drawer ref="logsRef" />
    <category-manage
      ref="categoryManageRef"
      :type="type"
      @ok="handleCategoryManageOk"
    />
  </div>
</template>

<script setup lang="ts">
  import {
    GlobalComponents,
    PropType,
    ref,
    watch,
    onMounted,
    nextTick,
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
    CrontabListApiParams,
  } from '@/api/crontab';
  import useLoading from '@/composables/loading';
  import { useConfirm } from '@/composables/confirm';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import { usePeriodUtils } from './components/form-drawer/composables/use-period-utils';
  import CategoryTree from './components/category-tree/index.vue';
  import CategoryManage from './components/category-manage/index.vue';

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
  const { generateFormattedPeriodComment, parseCronExpression } =
    usePeriodUtils();

  // 组件引用
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const categoryManageRef = ref<InstanceType<typeof CategoryManage>>();
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

  // 从cron表达式中提取周期信息
  function extractPeriodFromCronExpression(lines: string[]): string | null {
    const cronLine = lines.find((line) => !line.startsWith('#'));
    if (!cronLine) return null;

    const cronParts = cronLine.trim().split(/\s+/);
    if (cronParts.length < 5) return null;

    const cronExpression = cronParts.slice(0, 5).join(' ');
    const periodDetail = parseCronExpression(cronExpression);
    if (!periodDetail) return null;

    const formattedComment = generateFormattedPeriodComment([periodDetail]);
    const parts = formattedComment.split(':');
    if (parts.length <= 1) return null;

    return parts.slice(1).join(':').trim();
  }

  // 从记录中提取周期信息的工具函数
  function extractPeriodFromRecord(record: CrontabEntity): string {
    if (!record.content) {
      return record.period_expression || '';
    }

    const lines = record.content.split('\n');

    // 只从cron表达式中提取周期信息
    const periodFromCron = extractPeriodFromCronExpression(lines);
    if (periodFromCron) return periodFromCron;

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
      align: 'center' as const,
      slotName: 'status',
    },
    {
      dataIndex: 'period',
      title: t('app.crontab.list.column.period'),
      width: 150,
      render: ({ record }: { record: CrontabEntity }) =>
        extractPeriodFromRecord(record),
    },
    {
      dataIndex: 'mod_time',
      title: t('app.crontab.list.column.mod_time'),
      width: 160,
      render: ({ record }: { record: CrontabEntity }) => {
        return formatTimeWithoutSeconds(record.mod_time);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 210,
      align: 'center' as const,
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

    // 创建参数的副本，避免修改原始对象
    const requestParams = {
      ...fetchParams,
      type: params.value.type,
      category: params.value.category,
    };

    try {
      // 始终使用v-model绑定的category值，确保与左侧目录树同步
      return await getCrontabListApi(requestParams as CrontabListApiParams);
    } catch (error) {
      // 如果是因为category参数问题导致的错误，返回空数据而不是抛出异常
      if (error instanceof Error && error.message === 'OK') {
        console.warn('Category parameter issue detected, returning empty data');
        return Promise.resolve({
          items: [],
          total: 0,
          page: 1,
          page_size: params.value.page_size,
        });
      }
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
        await categoryTreeRef.value.reload();
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
    setLoading(true);
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

  // 处理分类管理
  const handleCategoryManage = () => {
    categoryManageRef.value?.show();
  };

  // 处理分类管理确认
  const handleCategoryManageOk = () => {
    // 刷新左侧分类树
    categoryTreeRef.value?.reload();
    // 刷新表格数据
    reload();
  };

  // 从分类树中选择第一个可用分类
  const selectFirstAvailableCategory = () => {
    if (!categoryTreeRef.value) return;

    // 使用类型断言访问items
    const categoryTree = categoryTreeRef.value as any;
    if (!categoryTree.items?.value?.length) return;

    const categories = categoryTree.items.value;
    if (categories.length > 0) {
      refreshAndSelectCategory(categories[0]);
    }
  };

  // 选择合适的分类：优先使用当前分类，否则使用第一个可用分类
  const selectAppropriateCategory = async (currentCategory: string) => {
    if (currentCategory) {
      // 恢复原始分类
      lastManualCategory.value = currentCategory;
      await refreshAndSelectCategory(currentCategory);
      return;
    }

    if (!categoryTreeRef.value) return;

    // 如果分类树有内容，使用第一个分类
    const categoryTree = categoryTreeRef.value as any;
    if (!categoryTree.items?.value?.length) return;

    const categories = categoryTree.items.value;
    if (categories.length > 0) {
      const firstCategory = categories[0];
      lastManualCategory.value = firstCategory;
      await refreshAndSelectCategory(firstCategory);
    }
  };

  // 重置所有组件状态
  const resetComponentsState = async () => {
    try {
      // 保存当前分类
      const currentCategory =
        params.value.category || lastManualCategory.value || '';

      // 暂时清空所有状态
      params.value.category = '';
      lastManualCategory.value = '';

      // 强制刷新分类树
      if (categoryTreeRef.value) {
        await categoryTreeRef.value.reload();
      }

      // 等待DOM更新
      await nextTick();

      // 尝试选择分类
      await selectAppropriateCategory(currentCategory);

      // 确保表格数据更新
      await nextTick();
      reload();
    } catch (e) {
      // 如果重置失败，尝试回到默认分类
      selectFirstAvailableCategory();
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
      categoryTreeRef.value?.reload();
    }
  };

  // 组件挂载完成后，确保分类初始化不会自动重置
  onMounted(async () => {
    // 挂载时重置所有状态，确保新创建的页面状态正常
    await resetComponentsState();
  });
</script>

<style scoped>
  .crontab-main {
    display: flex;
    flex-direction: column;
    height: calc(100vh - 170px);
  }

  .crontab-content {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .crontab-left-panel {
    width: 200px;
    padding: 16px 0;
    overflow-y: auto;
    border-right: 1px solid var(--color-border);
  }

  .crontab-right-panel {
    flex: 1;
    padding: 16px;
    overflow: auto;
  }

  .crontab-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    justify-content: center;
  }

  .status-tag {
    padding: 0 8px;
  }

  .operation {
    display: flex;
    justify-content: center;
  }

  .operation :deep(.arco-btn-size-small) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
