<template>
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
      <a-button
        v-if="!isSystemType"
        :type="BUTTON_TYPE_PRIMARY"
        @click="handleCreate"
      >
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('app.logrotate.list.action.create') }}
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
      <idb-table-operation
        type="button"
        :options="getLogrotateOperationOptions(record)"
      />
    </template>
  </idb-table>

  <!-- 子组件 -->
  <form-drawer ref="formRef" :type="type" @ok="handleFormOk" />
  <history-drawer ref="historyRef" />

  <a-modal
    v-model:visible="operateResultVisible"
    :title="operateResultTitle"
    :footer="false"
    :width="760"
    :mask-closable="true"
  >
    <a-typography-paragraph :copyable="true" class="operate-result-content">
      {{ operateResultText }}
    </a-typography-paragraph>
  </a-modal>
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

  import { PropType, ref, watch, onMounted, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { LOGROTATE_TYPE, LOGROTATE_FREQUENCY } from '@/config/enum';
  import { LogrotateEntity } from '@/entity/Logrotate';
  import { useLogger } from '@/composables/use-logger';
  import { useConfirm } from '@/composables/confirm';
  import useCurrentHost from '@/composables/current-host';
  import { operateLogrotateApi } from '@/api/logrotate';
  import FormDrawer from './components/form-drawer/index.vue';
  import HistoryDrawer from './components/history-drawer/index.vue';

  // 使用组合式函数
  import { useLogrotateList } from './composables/use-logrotate-list';
  import { useLogrotateColumns } from './composables/use-logrotate-columns';
  import { useLogrotateActions } from './composables/use-logrotate-actions';
  import { DEFAULT_LOGROTATE_CATEGORY, TABLE_CONFIG } from './constants';

  // 常量定义
  const BUTTON_TYPE_PRIMARY = 'primary' as const;

  // 国际化
  const { t } = useI18n();
  const { confirm } = useConfirm();
  const { currentHostId } = useCurrentHost();

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
  const isSystemType = computed(() => props.type === LOGROTATE_TYPE.System);

  // 组件引用
  const gridRef = ref<GridInstance>();
  const formRef = ref<FormDrawerInstance>();
  const historyRef = ref<InstanceType<typeof HistoryDrawer>>();
  const operateResultVisible = ref(false);
  const operateResultTitle = ref('');
  const operateResultText = ref('');

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

  // 事件处理函数
  const handleCreate = () => {
    formRef.value?.show({
      type: props.type,
      category:
        props.type === LOGROTATE_TYPE.System ? '' : DEFAULT_LOGROTATE_CATEGORY,
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

  const showOperateResult = (
    operation: 'test' | 'execute',
    result: string
  ): void => {
    operateResultTitle.value = t(
      operation === 'test'
        ? 'app.logrotate.system.operate.result.test_title'
        : 'app.logrotate.system.operate.result.execute_title'
    );
    operateResultText.value =
      result?.trim() || t('app.logrotate.system.operate.result.empty');
    operateResultVisible.value = true;
  };

  const handleSystemOperate = async (
    record: LogrotateEntity,
    operation: 'test' | 'execute'
  ): Promise<void> => {
    try {
      if (currentHostId.value === undefined) {
        Message.error(t('app.logrotate.list.message.no_host_selected'));
        return;
      }

      if (operation === 'execute') {
        const confirmed = await confirm({
          title: t('app.logrotate.system.operate.execute_confirm.title'),
          content: t('app.logrotate.system.operate.execute_confirm.content', {
            name: record.name,
          }),
        });
        if (!confirmed) {
          return;
        }
      }

      const response = await operateLogrotateApi({
        type: record.type,
        category: record.category || '',
        name: record.name,
        operation,
        host: currentHostId.value,
      });

      Message.success(
        operation === 'test'
          ? t('app.logrotate.system.operate.test_success')
          : t('app.logrotate.system.operate.execute_success')
      );
      showOperateResult(operation, response.result);
    } catch (error: any) {
      const detail = error?.response?.data?.data;
      if (typeof detail === 'string' && detail.trim()) {
        showOperateResult(operation, detail);
      }
      Message.error(
        error?.message ||
          (operation === 'test'
            ? t('app.logrotate.system.operate.test_failed')
            : t('app.logrotate.system.operate.execute_failed'))
      );
    }
  };

  // 获取操作按钮配置
  const getLogrotateOperationOptions = (record: LogrotateEntity) => {
    if (isSystemType.value) {
      return [
        {
          text: t('app.logrotate.system.operate.test'),
          click: () => handleSystemOperate(record, 'test'),
        },
        {
          text: t('app.logrotate.system.operate.execute'),
          click: () => handleSystemOperate(record, 'execute'),
        },
        {
          text: t('common.view'),
          click: () => handleEdit(record),
        },
      ];
    }

    return [
      {
        text: t('common.edit'),
        click: () => handleEdit(record),
      },
      {
        text: getActionButtonInfo(record).text,
        click: () => handleActionClick(record),
      },
      {
        text: t('app.logrotate.list.operation.history'),
        click: () => handleHistory(record),
      },
      {
        text: t('common.delete'),
        status: 'danger' as const,
        click: () => handleDeleteClick(record),
      },
    ];
  };

  const handleFormOk = async () => {
    gridRef.value?.reload();
  };

  watch(
    () => props.type,
    (newType) => {
      params.value.type = newType;
      params.value.category =
        newType === LOGROTATE_TYPE.System ? '' : DEFAULT_LOGROTATE_CATEGORY;
      params.value.page = TABLE_CONFIG.DEFAULT_PAGE;
      params.value.page_size = TABLE_CONFIG.DEFAULT_PAGE_SIZE;
      logInfo(`类型切换: ${newType}`);
      gridRef.value?.reload();
    }
  );

  // 暴露方法给父组件
  defineExpose({
    resetComponentsState: () => {
      params.value.page = TABLE_CONFIG.DEFAULT_PAGE;
      params.value.page_size = TABLE_CONFIG.DEFAULT_PAGE_SIZE;
      params.value.category = isSystemType.value
        ? ''
        : DEFAULT_LOGROTATE_CATEGORY;
      gridRef.value?.reload();
    },
  });

  onMounted(() => {
    params.value.category = isSystemType.value
      ? ''
      : DEFAULT_LOGROTATE_CATEGORY;
    gridRef.value?.reload();
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

  .operate-result-content {
    max-height: 420px;
    overflow: auto;
    word-break: break-word;
    white-space: pre-wrap;
  }
</style>
