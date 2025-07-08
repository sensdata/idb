<template>
  <app-sidebar-layout>
    <template #sidebar>
      <category-tree
        ref="categoryTreeRef"
        v-model:selected-category="selectedCat"
        :categories="categoryNames"
        :show-title="false"
        @create="handleCategoryCreate"
      />
    </template>
    <template #main>
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
          <category-manage-button
            :config="categoryManageConfig"
            @ok="handleCategoryManageOk"
          />
        </template>
        <template #history_version="{ record }: { record: ScriptEntity }">
          <a-link @click="handleHistoryVersion(record)">
            {{ $t('app.script.list.operation.view_history') }}
          </a-link>
        </template>
        <template #operation="{ record }: { record: ScriptEntity }">
          <idb-table-operation
            type="button"
            :options="getScriptOperationOptions(record)"
          />
        </template>
      </idb-table>
    </template>
  </app-sidebar-layout>
  <form-drawer
    ref="formRef"
    :type="type"
    @ok="reload"
    @category-change="handleCategoryChange"
  />
  <logs-drawer ref="logsRef" />
  <logs-view-modal
    ref="runResultModalRef"
    :title="$t('app.script.run.result.title')"
  />
  <history-version ref="historyRef" :type="type" @ok="reload" />
</template>

<script setup lang="ts">
  import {
    GlobalComponents,
    PropType,
    reactive,
    ref,
    watch,
    onMounted,
    computed,
  } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { formatTime } from '@/utils/format';
  import { ScriptEntity } from '@/entity/Script';
  import {
    deleteScriptApi,
    getScriptListApi,
    runScriptApi,
    getScriptCategoryListApi,
  } from '@/api/script';
  import useLoading from '@/composables/loading';
  import { useConfirm } from '@/composables/confirm';
  import usetCurrentHost from '@/composables/current-host';
  import AppSidebarLayout from '@/components/app-sidebar-layout/index.vue';
  import LogsViewModal from '@/components/logs-view/modal.vue';
  import CategoryTree from '@/components/idb-tree/category-tree.vue';
  import CategoryManageButton from '@/components/idb-tree/components/category-manage-button/index.vue';
  import FormDrawer from './components/form-drawer/index.vue';
  import LogsDrawer from './components/logs-drawer/index.vue';
  import HistoryVersion from './components/history-version/index.vue';
  import { createScriptCategoryManageConfig } from './adapters/category-manage-adapter';

  const props = defineProps({
    type: {
      type: String as PropType<SCRIPT_TYPE>,
      required: true,
    },
  });

  const { t } = useI18n();
  const { currentHostId } = usetCurrentHost();

  // Category data state
  const categoryNames = ref<string[]>([]);
  const categoryLoading = ref(false);

  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<InstanceType<typeof FormDrawer>>();
  const logsRef = ref<InstanceType<typeof LogsDrawer>>();
  const runResultModalRef = ref<InstanceType<typeof LogsViewModal>>();
  const categoryTreeRef = ref<InstanceType<typeof CategoryTree>>();
  const historyRef = ref<InstanceType<typeof HistoryVersion>>();
  const selectedCat = ref();

  // Category management configuration
  const categoryManageConfig = computed(() =>
    createScriptCategoryManageConfig(props.type)
  );

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
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  /**
   * Load categories list
   */
  const loadCategories = async () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error('Host ID is required');
      return;
    }

    if (categoryLoading.value) {
      return;
    }

    categoryLoading.value = true;
    try {
      const ret = await getScriptCategoryListApi({
        page: 1,
        page_size: 1000,
        type: props.type,
      });

      const newItems = [...ret.items.map((item) => item.name)];

      // If current selected category is not in the list, add it
      if (selectedCat.value && !newItems.includes(selectedCat.value)) {
        newItems.push(selectedCat.value);
      }

      categoryNames.value = newItems;

      // Auto-select first category if none selected and list is not empty
      if (!selectedCat.value && newItems.length > 0) {
        selectedCat.value = newItems[0];
      }
    } catch (err: any) {
      Message.error(err?.message || 'Failed to load categories');
    } finally {
      categoryLoading.value = false;
    }
  };

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    formRef.value?.clearParams();
    formRef.value?.show();
  };

  const handleCategoryManageOk = () => {
    loadCategories();
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
        await deleteScriptApi({
          host: currentHostId.value!,
          type: props.type,
          category: selectedCat.value || '',
          name: record.name,
        });
        Message.success(t('app.script.list.message.delete_success'));
      } finally {
        setLoading(false);
        reload();
      }
    }
  };

  // 获取操作按钮配置
  const getScriptOperationOptions = (record: ScriptEntity) => [
    {
      text: t('common.edit'),
      click: () => handleEdit(record),
    },
    {
      text: t('app.script.list.operation.run'),
      click: () => handleRun(record),
    },
    {
      text: t('app.script.list.operation.log'),
      click: () => handleLog(record),
    },
    {
      text: t('common.delete'),
      status: 'danger' as const,
      confirm: t('app.script.list.delete.confirm', { name: record.name }),
      click: () => handleDelete(record),
    },
  ];

  const handleCategoryChange = () => {
    loadCategories();
    gridRef.value?.reload();
  };

  // Category management event handlers
  const handleCategoryCreate = () => {
    // 在分类为空时，可以触发创建分类的操作
    // 现在使用 category-manage-button 组件内部处理
  };

  // Load categories on mount
  onMounted(() => {
    loadCategories();
  });
</script>

<style scoped></style>
