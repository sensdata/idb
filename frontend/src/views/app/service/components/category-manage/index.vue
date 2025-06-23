<template>
  <a-drawer
    :visible="visible"
    width="500px"
    :title="$t('app.service.category.manage.title')"
    placement="right"
    :footer="false"
    @cancel="handleClose"
  >
    <idb-table
      v-if="visible"
      ref="gridRef"
      :columns="columns"
      :params="{ type: props.type }"
      :fetch="getServiceCategoryListApi"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleCreate">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.service.category.action.create') }}
        </a-button>
      </template>
      <template #operation="{ record }">
        <div class="operation">
          <a-button type="text" size="small" @click="handleEdit(record)">
            {{ $t('common.edit') }}
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
  </a-drawer>
  <category-form-modal ref="formRef" :type="type" @ok="reload" />
</template>

<script setup lang="ts">
  import { GlobalComponents, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    deleteServiceCategoryApi,
    getServiceCategoryListApi,
  } from '@/api/service';
  import { SERVICE_TYPE } from '@/config/enum';
  import { useConfirm } from '@/composables/confirm';
  import { Message } from '@arco-design/web-vue';
  import CategoryFormModal from './form-modal.vue';

  const props = defineProps<{
    type: SERVICE_TYPE;
  }>();

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.service.category.column.name'),
      width: 200,
      ellipsis: true,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 140,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const { confirm } = useConfirm();
  const formRef = ref();

  const reload = () => {
    gridRef.value?.reload();
  };

  const handleCreate = () => {
    formRef.value?.show();
  };

  const handleEdit = (category: { name: string }) => {
    formRef.value?.setData({ name: category.name });
    formRef.value?.show();
  };

  const handleDelete = async (category: { name: string }) => {
    try {
      await confirm(
        t('app.service.category.delete.confirm', { name: category.name })
      );
      await deleteServiceCategoryApi({
        type: props.type,
        category: category.name,
      });
      Message.success(t('app.service.category.message.delete_success'));
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
    gap: 8px;
    justify-content: center;
  }
</style>
