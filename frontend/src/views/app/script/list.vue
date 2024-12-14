<template>
  <div class="script-layout">
    <div class="script-sidebar">
      <category-tree v-model:selected="selectedCat" :items="cats" />
    </div>
    <div class="script-main">
      <idb-table
        ref="gridRef"
        :params="params"
        :columns="columns"
        :fetch="getScriptListApi"
      >
        <template #leftActions>
          <a-button type="primary" @click="handleCreate">
            <template #icon>
              <icon-plus />
            </template>
            {{ $t('app.script.list.action.create') }}
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
          <a-button type="text" size="small" @click="handleEdit(record)">
            {{ $t('common.edit') }}
          </a-button>
          <a-button type="text" size="small" @click="handleRun(record)">
            {{ $t('app.script.list.operation.run') }}
          </a-button>
          <a-button type="text" size="small" @click="handleLog(record)">
            {{ $t('app.script.list.operation.log') }}
          </a-button>
          <a-button type="text" size="small" @click="handleDelete(record)">
            {{ $t('common.delete') }}
          </a-button>
        </template>
      </idb-table>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { PropType, ref } from 'vue';
  import { ScriptType } from '@/config/enum';
  import { useI18n } from 'vue-i18n';
  import { formatTime } from '@/utils/format';
  import { ScriptEntity } from '@/entity/Script';
  import { getScriptListApi } from '@/api/script';
  import CategoryTree from './components/category-tree.vue';

  const props = defineProps({
    type: {
      type: String as PropType<ScriptType>,
      required: true,
    },
  });

  const { t } = useI18n();

  const selectedCat = ref(null);
  const params = ref({
    type: props.type,
    category: selectedCat.value,
  });
  const cats = [null, '分类1', '分类2', '分类3'];

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
      width: 125,
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
      dataIndex: 'category',
      title: t('app.script.list.column.category'),
      width: 125,
      render: ({ record }: { record: ScriptEntity }) => {
        return record.category ? record.category : t('app.script.category.all');
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
      width: 200,
      slotName: 'operation',
    },
  ];

  const handleCreate = () => {
    console.log('handleCreate');
  };
  const handleHistoryVersion = (record: ScriptEntity) => {
    console.log('handleHistoryVersion', record);
  };
  const handleEdit = (record: ScriptEntity) => {
    console.log('handleEdit', record);
  };
  const handleRun = (record: ScriptEntity) => {
    console.log('handleRun', record);
  };
  const handleLog = (record: ScriptEntity) => {
    console.log('handleLog', record);
  };
  const handleDelete = (record: ScriptEntity) => {
    console.log('handleDelete', record);
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
</style>
