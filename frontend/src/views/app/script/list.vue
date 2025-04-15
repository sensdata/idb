<template>
  <div class="script-layout">
    <div class="script-sidebar">
      <category-tree
        ref="categoryTreeRef"
        v-model:selected="selectedCat"
        :type="type"
      />
    </div>
    <div class="script-main">
      <idb-table
        ref="gridRef"
        class="script-table"
        :loading="loading"
        :params="params"
        :columns="columns"
        :fetch="getScriptListApi"
        :auto-load="false"
      >
        <template #leftActions>
          <a-button type="primary" @click="handleCreate">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.script.list.action.create') }}
          </a-button>
          <a-button type="outline" @click="handleGroupManage">
            <template #icon>
              <icon-list />
            </template>
            {{ $t('app.script.list.action.group_manage') }}
          </a-button>
        </template>
        <template #history_version="{ record }: { record: ScriptEntity }">
          <a-button
            type="text"
            size="small"
            @click="handleHistoryVersion(record)"
          >
            {{ $t('app.script.list.operation.view_history') }}
          </a-button>
        </template>
        <template #operation="{ record }: { record: ScriptEntity }">
          <div class="operation">
            <a-button type="text" size="small" @click="handleEdit(record)">
              {{ $t('common.edit') }}
            </a-button>
            <a-button type="text" size="small" @click="handleRun(record)">
              {{ $t('app.script.list.operation.run') }}
            </a-button>
            <a-button type="text" size="small" @click="handleLog(record)">
              {{ $t('app.script.list.operation.log') }}
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
    <form-drawer ref="formRef" :type="type" @ok="reload" />
    <logs-drawer ref="logsRef" />
    <logs-view-modal
      ref="runResultModalRef"
      :title="$t('app.script.run.result.title')"
    />
    <category-manage
      ref="categoryManageRef"
      :type="type"
      @ok="handleCategoryManageOk"
    />
    <history-version ref="historyRef" :type="type" @ok="reload" />
  </div>
</template>

<script setup lang="ts">
  import { GlobalComponents, PropType, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { ScriptEntity } from '@/entity/Script';
  import {
    deleteScriptApi,
    getScriptListApi,
    runScriptApi,
  } from '@/api/script';
  import useLoading from '@/hooks/loading';
  import { useConfirm } from '@/hooks/confirm';
  import usetCurrentHost from '@/hooks/current-host';
  import LogsViewModal from '@/components/logs-view/modal.vue';
  import CategoryTree from './components/category-tree/index.vue';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import CategoryManage from './components/category-manage/index.vue';
  import HistoryVersion from './components/history-version/index.vue';

  const props = defineProps({
    type: {
      type: String as PropType<SCRIPT_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { currentHostId } = usetCurrentHost();

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<InstanceType<typeof FormDrawer>>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const runResultModalRef = ref<InstanceType<typeof LogsViewModal>>();
  const categoryManageRef = ref<InstanceType<typeof CategoryManage>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const historyRef = ref<InstanceType<typeof HistoryVersion>>();
  const selectedCat = ref();
  watch(selectedCat, (val) => {
    if (val) {
      gridRef.value?.load({
        category: val,
      });
    }
  });
  const { loading, setLoading } = useLoading();
  const { confirm } = useConfirm();
  const params = reactive({
    type: props.type,
    category: selectedCat.value,
  });

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.script.list.column.name'),
      width: 150,
      slotName: 'name',
    },
    {
      dataIndex: 'mod_time',
      title: t('app.script.list.column.mod_time'),
      width: 160,
      render: ({ record }: { record: ScriptEntity }) => {
        return formatTime(record.mod_time);
      },
    },
    {
      dataIndex: 'create_time',
      title: t('app.script.list.column.create_time'),
      width: 125,
      render: ({ record }: { record: ScriptEntity }) => {
        return formatTime(record.create_time);
      },
    },
    {
      dataIndex: 'history_version',
      title: t('app.script.list.column.history_version'),
      width: 120,
      slotName: 'history_version',
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 210,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    formRef.value?.show();
  };

  const handleGroupManage = () => {
    categoryManageRef.value?.show();
  };

  const handleCategoryManageOk = () => {
    categoryTreeRef.value?.reload();
  };

  const handleHistoryVersion = (record: ScriptEntity) => {
    historyRef.value?.setParams({
      category: selectedCat.value,
      name: record.name,
    });
    historyRef.value?.show();
  };

  const handleEdit = (record: ScriptEntity) => {
    formRef.value?.setParams({
      name: record.name,
      category: selectedCat.value,
    });
    formRef.value?.load();
    formRef.value?.show();
  };
  const handleRun = async (record: ScriptEntity) => {
    setLoading(true);
    try {
      const result = await runScriptApi({
        host_id: currentHostId.value!,
        script_path: record.source,
      });
      if (!result.err) {
        runResultModalRef.value?.setContent(result.out);
        runResultModalRef.value?.show();
        // Message.success(t('app.script.list.message.run_success'));
      } else {
        Message.error(result.err);
      }
    } finally {
      setLoading(false);
    }
  };
  const handleLog = (record: ScriptEntity) => {
    logsRef.value?.show({
      path: record.source,
    });
  };
  const handleDelete = async (record: ScriptEntity) => {
    if (
      await confirm({
        content: t('app.script.list.delete.confirm', { name: record.name }),
      })
    ) {
      setLoading(true);
      try {
        await deleteScriptApi(record);
        Message.success(t('app.script.list.message.delete_success'));
      } finally {
        setLoading(false);
        reload();
      }
    }
  };
</script>

<style scoped>
  .script-layout {
    position: relative;
    min-height: calc(100vh - 240px);
    margin-top: 20px;
    padding-left: 240px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .script-sidebar {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    width: 240px;
    height: 100%;
    padding: 4px 8px;
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
  }

  .script-main {
    min-width: 0;
    height: 100%;
    padding: 20px;
  }

  .operation :deep(.arco-btn-size-small) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
