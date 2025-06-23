<template>
  <a-drawer
    :visible="visible"
    width="500px"
    :title="$t('app.logrotate.category.manage.title')"
    placement="right"
    :footer="false"
    @cancel="handleClose"
  >
    <idb-table
      v-if="visible"
      ref="gridRef"
      :columns="columns"
      :params="{ type: props.type }"
      :fetch="fetchCategories"
    >
      <template #leftActions>
        <a-button type="primary" @click="handleCreate">
          <template #icon>
            <icon-plus />
          </template>
          {{ $t('app.logrotate.category.manage.create') }}
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
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { GlobalComponents } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import {
    deleteLogrotateCategoryApi,
    getLogrotateCategoriesApi,
  } from '@/api/logrotate';
  import { LOGROTATE_TYPE } from '@/config/enum';
  import { useConfirm } from '@/composables/confirm';
  import useCurrentHost from '@/composables/current-host';
  import CategoryFormModal from './form-modal.vue';

  const props = defineProps<{
    type: LOGROTATE_TYPE;
  }>();

  const emit = defineEmits<{
    ok: [];
  }>();

  const { t } = useI18n();
  const { currentHostId } = useCurrentHost();
  const { confirm } = useConfirm();

  // 响应式数据
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const formRef = ref<InstanceType<typeof CategoryFormModal>>();

  const columns = computed(() => [
    {
      dataIndex: 'name',
      title: t('app.logrotate.category.manage.column.name'),
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
  ]);

  // 获取分类数据
  const fetchCategories = async (params: {
    type: LOGROTATE_TYPE;
    page?: number;
    page_size?: number;
  }) => {
    const hostId = currentHostId.value;
    if (!hostId) {
      throw new Error('Host ID is required');
    }

    return getLogrotateCategoriesApi(
      params.type,
      params.page || 1,
      params.page_size || 100,
      hostId
    );
  };

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

  // 删除分类
  const handleDelete = async (category: { name: string }) => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error('Host ID is required');
      return;
    }

    try {
      await confirm(
        t('app.logrotate.category.manage.delete.content', {
          name: category.name,
        })
      );

      await deleteLogrotateCategoryApi(props.type, category.name, hostId);

      Message.success(
        t('app.logrotate.category.manage.message.delete_success')
      );

      // 先通知父组件分类已删除，确保在加载数据前更新分类树和选择
      emit('ok');

      // 然后重新加载当前表格
      reload();
    } catch (err: any) {
      if (err?.message) {
        Message.error(err.message);
      }
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
