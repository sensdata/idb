<template>
  <a-drawer
    :visible="visible"
    width="500px"
    :title="$t('category.manage.title')"
    placement="right"
    :footer="false"
    @cancel="handleClose"
  >
    <idb-table
      v-if="visible"
      ref="gridRef"
      :columns="computedColumns"
      :params="config.params"
      :fetch="config.api.getList"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleCreate">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('category.action.create') }}
        </a-button>
      </template>
      <template #operation="{ record }">
        <div class="operation">
          <a-link @click="handleEdit(record)">
            {{ $t('common.edit') }}
          </a-link>
          <a-link status="danger" @click="handleDelete(record)">
            {{ $t('common.delete') }}
          </a-link>
        </div>
      </template>
    </idb-table>
  </a-drawer>
  <category-form-modal ref="formRef" :config="config" @ok="reload" />
</template>

<script setup lang="ts">
  import { GlobalComponents, computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { useConfirm } from '@/composables/confirm';
  import CategoryFormModal from './category-form-modal.vue';
  import type { CategoryManageConfig, Category } from '../types/category';

  interface Props {
    config: CategoryManageConfig;
  }

  const props = defineProps<Props>();
  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref();
  const { confirm } = useConfirm();

  const defaultColumns = [
    {
      dataIndex: 'name',
      title: t('category.column.name'),
      width: 200,
      ellipsis: true,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 140,
      align: 'left' as const,
      slotName: 'operation',
    },
  ];

  const computedColumns = computed(() => {
    return props.config.columns || defaultColumns;
  });

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    formRef.value?.show();
  };

  const handleEdit = (category: Category) => {
    const nameField = props.config.nameField || 'name';
    formRef.value?.setData({ name: category[nameField] });
    formRef.value?.show();
  };

  const handleDelete = async (category: Category) => {
    const nameField = props.config.nameField || 'name';
    const categoryName = category[nameField];

    try {
      await confirm(t('category.delete.confirm', { name: categoryName }));
      await props.config.api.delete({
        category: categoryName,
        ...(props.config.params || {}),
      });
      Message.success(t('category.message.delete_success'));
      reload();
    } catch (err: any) {
      Message.error(err?.message);
    }
  };

  const show = () => {
    visible.value = true;
  };

  const handleClose = () => {
    visible.value = false;
    emit('ok');
  };

  defineExpose({
    show,
  });
</script>

<style scoped>
  .operation {
    display: flex;
    gap: 12px;
  }

  .operation :first-child {
    padding-left: 0;
  }
</style>
