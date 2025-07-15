<template>
  <app-sidebar-layout>
    <template #sidebar>
      <category-tree
        ref="categoryTreeRef"
        v-model:selected-category="params.category"
        :categories="categoryItems"
        :show-title="true"
        :enable-category-management="true"
        :category-manage-config="categoryManageConfig"
        @create="handleCategoryCreate"
        @category-manage-ok="handleCategoryManageOk"
      />
    </template>
    <template #main>
      <idb-table
        ref="gridRef"
        class="service-table"
        :loading="loading"
        :params="params"
        :columns="columns"
        :fetch="fetchServiceList"
        :auto-load="false"
      >
        <template #leftActions>
          <a-button type="primary" @click="handleCreate">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.service.list.action.create') }}
          </a-button>
          <a-button
            v-if="type === SERVICE_TYPE.Global"
            @click="handleSyncGlobal"
          >
            <template #icon>
              <icon-sync />
            </template>
            {{ $t('app.service.list.action.sync') }}
          </a-button>
        </template>
        <template #status="{ record }: { record: ServiceEntity }">
          <div class="status-cell">
            <a-tag :color="record.linked ? 'green' : 'gray'" class="status-tag">
              {{
                record.linked
                  ? $t('app.service.list.status.activated')
                  : $t('app.service.list.status.deactivated')
              }}
            </a-tag>
          </div>
        </template>
        <template #operation="{ record }: { record: ServiceEntity }">
          <idb-table-operation
            type="button"
            :options="getServiceOperationOptions(record)"
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
  <history-drawer ref="historyRef" />
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
  import { SERVICE_TYPE, SERVICE_ACTION } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { ServiceEntity } from '@/entity/Service';
  import {
    syncGlobalServiceApi,
    getServiceCategoryListApi,
  } from '@/api/service';
  import { useConfirm } from '@/composables/confirm';
  import { useLogger } from '@/composables/use-logger';
  import useCurrentHost from '@/composables/current-host';
  import AppSidebarLayout from '@/components/app-sidebar-layout/index.vue';
  import CategoryTree from '@/components/idb-tree/category-tree.vue';
  import { createServiceCategoryManageConfig } from './adapters/category-manage-adapter';
  import { useServiceList } from './composables/use-service-list';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';

  // 定义组件引用类型接口
  interface FormDrawerInstance extends InstanceType<typeof FormDrawer> {
    show: (params?: {
      name?: string;
      type?: SERVICE_TYPE;
      category?: string;
      isEdit?: boolean;
      record?: ServiceEntity;
    }) => Promise<void>;
  }

  const props = defineProps({
    type: {
      type: String as PropType<SERVICE_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { confirm } = useConfirm();
  const { logError, logInfo } = useLogger('ServiceList');
  const { currentHostId } = useCurrentHost();

  // 使用组合式函数
  const {
    params,
    loading,
    fetchServiceList,
    deleteService,
    toggleServiceStatus,
  } = useServiceList(props.type);

  // 分类管理配置
  const categoryManageConfig = computed(() =>
    createServiceCategoryManageConfig(props.type, currentHostId.value || 0)
  );

  // 分类数据状态
  const categoryItems = ref<string[]>([]);
  const categoryLoading = ref(false);

  // 组件引用
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const formRef = ref<FormDrawerInstance>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const historyRef = ref<InstanceType<typeof HistoryDrawer>>();

  // 记录最后一次手动设置的分类，用于防止重置
  const lastManualCategory = ref<string>('');

  // 表格列配置
  const columns = [
    {
      title: t('app.service.list.columns.name'),
      dataIndex: 'name',
      width: 200,
      ellipsis: true,
      tooltip: true,
    },
    {
      title: t('app.service.list.columns.description'),
      dataIndex: 'content',
      width: 300,
      ellipsis: true,
      tooltip: true,
      render: ({ record }: { record: ServiceEntity }) => {
        // 从content字段中提取Description行的内容作为描述
        if (!record.content) return '';
        const lines = record.content.split('\n');
        const descriptionLine = lines.find((line) =>
          line.trim().startsWith('Description=')
        );
        return descriptionLine
          ? descriptionLine.replace('Description=', '').trim()
          : '';
      },
    },
    {
      title: t('app.service.list.columns.status'),
      dataIndex: 'status',
      slotName: 'status',
      width: 120,
    },
    {
      title: t('app.service.list.columns.size'),
      dataIndex: 'size',
      width: 100,
      render: ({ record }: { record: ServiceEntity }) => {
        return `${(record.size / 1024).toFixed(2)} KB`;
      },
    },
    {
      title: t('app.service.list.columns.mod_time'),
      dataIndex: 'mod_time',
      width: 180,
      render: ({ record }: { record: ServiceEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      title: t('app.service.list.columns.operation'),
      slotName: 'operation',
      width: 240,
      align: 'left' as const,
    },
  ];

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
      const ret = await getServiceCategoryListApi({
        type: props.type,
        page: 1,
        page_size: 1000,
        host: hostId,
      });
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

  // 创建服务
  const handleCreate = () => {
    formRef.value?.show({
      type: props.type,
      category: params.value.category,
      isEdit: false,
    });
  };

  // 编辑服务
  const handleEdit = (record: ServiceEntity) => {
    formRef.value?.show({
      type: props.type,
      category: params.value.category, // 使用左侧目录树当前选择的分类
      isEdit: true,
      record,
    });
  };

  // 激活/停用服务
  const handleAction = async (
    record: ServiceEntity,
    action: SERVICE_ACTION
  ) => {
    try {
      const actionText =
        action === SERVICE_ACTION.Activate
          ? t('app.service.list.operation.activate')
          : t('app.service.list.operation.deactivate');

      await confirm(
        t('app.service.list.confirm.action', {
          action: actionText,
          name: record.name,
        })
      );

      const success = await toggleServiceStatus(record, action);
      if (success) {
        gridRef.value?.reload();
      }
    } catch (error) {
      logError('Failed to action service:', error);
    }
  };

  // 查看日志
  const handleViewLogs = (record: ServiceEntity) => {
    logsRef.value?.show({
      type: props.type,
      category: params.value.category,
      name: record.name,
    });
  };

  // 查看历史
  const handleViewHistory = (record: ServiceEntity) => {
    historyRef.value?.show({
      type: props.type,
      category: params.value.category,
      name: record.name,
    });
  };

  // 删除服务
  const handleDelete = async (record: ServiceEntity) => {
    const success = await deleteService(record);
    if (success) {
      gridRef.value?.reload();
    }
  };

  // 获取操作按钮配置
  const getServiceOperationOptions = (record: ServiceEntity) => [
    {
      text: t('common.edit'),
      click: () => handleEdit(record),
    },
    {
      text: record.linked
        ? t('app.service.list.operation.deactivate')
        : t('app.service.list.operation.activate'),
      click: () =>
        handleAction(
          record,
          record.linked ? SERVICE_ACTION.Deactivate : SERVICE_ACTION.Activate
        ),
    },
    {
      text: t('common.delete'),
      status: 'danger' as const,
      click: () => handleDelete(record),
    },
    {
      text: t('app.service.list.operation.logs'),
      click: () => handleViewLogs(record),
    },
    {
      text: t('app.service.list.operation.history'),
      click: () => handleViewHistory(record),
    },
  ];

  // 同步全局仓库
  const handleSyncGlobal = async () => {
    try {
      await confirm(t('app.service.list.confirm.sync'));

      // 调用真实的API接口
      await syncGlobalServiceApi();

      Message.success(t('app.service.list.success.sync'));
      gridRef.value?.reload();
    } catch (error) {
      logError('Failed to sync global:', error);
      Message.error(t('app.service.list.error.sync'));
    }
  };

  /**
   * 处理分类创建
   */
  const handleCategoryCreate = () => {
    // CategoryTree 组件自己会处理分类创建
    // 这里不需要额外的逻辑
  };

  // 刷新并选择分类
  const refreshAndSelectCategory = async (category: string) => {
    lastManualCategory.value = category;
    params.value.category = category;
    await nextTick();
    categoryTreeRef.value?.refresh();
    gridRef.value?.reload();
  };

  // 处理新分类更新
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

  // 表单提交成功回调
  const handleFormOk = async (newCategory?: string) => {
    if (newCategory) {
      await handleNewCategoryUpdate(newCategory);
    } else {
      gridRef.value?.reload();
    }
  };

  // 分类变更回调
  const handleCategoryChange = async (category: string) => {
    if (category) {
      await refreshAndSelectCategory(category);
    }
  };

  // 分类管理成功回调
  const handleCategoryManageOk = async () => {
    await loadCategories();
    gridRef.value?.reload();
  };

  // 重置组件状态
  const resetComponentsState = () => {
    // 先刷新分类树，确保目录树状态正确
    categoryTreeRef.value?.refresh();
    // 延迟重新加载表格数据，等待分类树刷新完成
    nextTick(() => {
      gridRef.value?.reload();
    });
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

  // 监听type变化
  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      // 不清空分类选择，让分类树组件自己决定选择哪个分类
      lastManualCategory.value = '';
      nextTick(() => {
        categoryTreeRef.value?.refresh();
      });
    }
  );

  // 监听主机ID变化
  watch(
    () => currentHostId.value,
    (newHostId) => {
      params.value.host = newHostId;
    }
  );

  onMounted(() => {
    resetComponentsState();
    loadCategories();
  });

  // 暴露方法给父组件
  defineExpose({
    resetComponentsState: () => {
      resetComponentsState();
      loadCategories(); // 重置状态时也要重新加载分类
    },
  });
</script>

<style scoped>
  .service-table {
    height: 100%;
  }

  .status-cell {
    display: flex;
    align-items: center;
  }

  .status-tag {
    margin: 0;
  }
</style>
