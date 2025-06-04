<template>
  <div class="service-main">
    <div class="service-content">
      <div class="service-left-panel">
        <category-tree
          ref="categoryTreeRef"
          v-model:selected="params.category"
          :type="type"
        />
      </div>
      <div class="service-right-panel">
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
            <a-button @click="handleCategoryManage">
              <template #icon>
                <icon-settings />
              </template>
              {{ $t('app.service.category.manage.title') }}
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
              <a-tag
                :color="record.linked ? 'green' : 'gray'"
                class="status-tag"
              >
                {{
                  record.linked
                    ? $t('app.service.list.status.activated')
                    : $t('app.service.list.status.deactivated')
                }}
              </a-tag>
            </div>
          </template>
          <template #operation="{ record }: { record: ServiceEntity }">
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
                    record.linked
                      ? SERVICE_ACTION.Deactivate
                      : SERVICE_ACTION.Activate
                  )
                "
              >
                {{
                  record.linked
                    ? $t('app.service.list.operation.deactivate')
                    : $t('app.service.list.operation.activate')
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
              <a-dropdown>
                <a-button type="text" size="small">
                  {{ $t('common.more') }}
                  <icon-down />
                </a-button>
                <template #content>
                  <a-doption @click="handleViewLogs(record)">
                    {{ $t('app.service.list.operation.logs') }}
                  </a-doption>
                  <a-doption @click="handleViewHistory(record)">
                    {{ $t('app.service.list.operation.history') }}
                  </a-doption>
                </template>
              </a-dropdown>
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
      @category-created="handleCategoryCreated"
    />
    <logs-drawer ref="logsRef" />
    <history-drawer ref="historyRef" />
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
  import { SERVICE_TYPE, SERVICE_ACTION } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { ServiceEntity } from '@/entity/Service';
  import { syncGlobalServiceApi } from '@/api/service';
  import { useConfirm } from '@/hooks/confirm';
  import { useLogger } from '@/hooks/use-logger';
  import { useServiceList } from './hooks/use-service-list';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';
  import CategoryTree from './components/category-tree/index.vue';
  import CategoryManage from './components/category-manage/index.vue';

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
  const { logError } = useLogger('ServiceList');

  // 使用组合式函数
  const {
    params,
    loading,
    fetchServiceList,
    deleteService,
    toggleServiceStatus,
  } = useServiceList(props.type);

  // 组件引用
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const categoryManageRef = ref<InstanceType<typeof CategoryManage>>();
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
      align: 'center' as const,
      fixed: 'right' as const,
    },
  ];

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

  // 分类管理
  const handleCategoryManage = () => {
    categoryManageRef.value?.show();
  };

  // 表单提交成功回调
  const handleFormOk = () => {
    gridRef.value?.reload();
  };

  // 分类变更回调
  const handleCategoryChange = (category: string) => {
    lastManualCategory.value = category;
    params.value.category = category;
    gridRef.value?.reload();
  };

  // 分类创建成功回调
  const handleCategoryCreated = () => {
    // 刷新分类树以显示新创建的分类
    categoryTreeRef.value?.refresh();
  };

  // 分类管理成功回调
  const handleCategoryManageOk = () => {
    categoryTreeRef.value?.refresh();
    gridRef.value?.reload();
  };

  // 重置组件状态
  const resetComponentsState = () => {
    // 如果没有手动选择分类，则重置为空
    if (!lastManualCategory.value) {
      params.value.category = '';
    }
    gridRef.value?.reload();
  };

  // 当分类选择后，自动加载数据
  const handleCategorySelected = () => {
    if (params.value.category) {
      gridRef.value?.load();
    }
  };

  // 监听type变化
  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      params.value.category = '';
      lastManualCategory.value = '';
      nextTick(() => {
        categoryTreeRef.value?.refresh();
        gridRef.value?.reload();
      });
    }
  );

  // 监听分类变化
  watch(
    () => params.value.category,
    () => {
      params.value.page = 1;
      handleCategorySelected();
    }
  );

  onMounted(() => {
    categoryTreeRef.value?.refresh();
  });

  // 暴露方法给父组件
  defineExpose({
    resetComponentsState,
  });
</script>

<style scoped>
  .service-main {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .service-content {
    display: flex;
    flex: 1;
    gap: 16px;
    min-height: 0;
  }

  .service-left-panel {
    flex-shrink: 0;
    width: 200px;
  }

  .service-right-panel {
    flex: 1;
    min-width: 0;
  }

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

  .operation {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    align-items: center;
    justify-content: center;
  }

  .operation :deep(.arco-btn-size-small) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
