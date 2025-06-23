<template>
  <div class="category-tree">
    <div v-if="items.length === 0" class="empty-text">
      {{ $t('app.service.category.tree.empty') }}
      <span class="color-primary" @click="handleCreate">
        {{ $t('app.service.category.tree.create') }}
      </span>
    </div>
    <template v-else>
      <div
        v-for="cat of items"
        :key="cat"
        class="item"
        :class="{ selected: selected === cat }"
        @click="handleClick(cat)"
      >
        <div class="item-icon">
          <folder-icon />
        </div>
        <div class="item-text truncate">{{ cat }}</div>
      </div>
    </template>
  </div>
  <category-form-modal ref="formRef" :type="props.type" @ok="handleCreateOk" />
</template>

<script lang="ts" setup>
  import { onMounted, ref, watch } from 'vue';
  import { getServiceCategoryListApi } from '@/api/service';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { SERVICE_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';
  import useCurrentHost from '@/composables/current-host';
  import CategoryFormModal from '../category-manage/form-modal.vue';

  const props = defineProps<{
    type: SERVICE_TYPE;
  }>();

  const selected = defineModel('selected', {
    type: String,
    required: false,
  });

  const { currentHostId } = useCurrentHost();
  const items = ref<string[]>([]);

  const loadCategories = async () => {
    try {
      // 检查必要参数
      if (!currentHostId.value) {
        items.value = [];
        selected.value = '';
        return;
      }

      const ret = await getServiceCategoryListApi({
        page: 1,
        page_size: 1000,
        type: props.type,
        host: currentHostId.value,
      });
      items.value = [...ret.items.map((item) => item.name)];
      if (items.value.length > 0) {
        // 如果当前选择的分类仍然存在，保持选择；否则选择第一个
        if (!selected.value || !items.value.includes(selected.value)) {
          selected.value = items.value[0];
        }
      } else {
        selected.value = '';
      }
    } catch (err: any) {
      Message.error(err?.message);
    }
  };

  // 监听type和主机ID变化
  watch(
    [() => props.type, () => currentHostId.value],
    () => {
      loadCategories();
    },
    { immediate: true }
  );

  onMounted(() => {
    loadCategories();
  });

  function handleClick(cat: string) {
    selected.value = cat;
  }

  const formRef = ref<InstanceType<typeof CategoryFormModal>>();
  function handleCreate() {
    formRef.value?.show();
  }

  function handleCreateOk() {
    loadCategories();
  }

  const refresh = () => {
    loadCategories();
  };

  defineExpose({
    refresh,
    reload: loadCategories,
  });
</script>

<style scoped>
  .category-tree {
    padding-left: 8px;
  }

  .empty-text {
    padding: 10px 0;
  }

  .color-primary {
    color: var(--color-primary-6);
    cursor: pointer;
  }

  .color-primary:hover {
    color: var(--color-primary-5);
  }

  .item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    padding-left: 10px;
    margin-bottom: 8px;
    line-height: 32px;
    cursor: pointer;
    border-radius: 4px;
  }

  .item:hover {
    background-color: var(--color-fill-1);
  }

  .item.selected {
    background-color: var(--color-fill-2);
  }

  .item.selected::before {
    position: absolute;
    top: 12.5%;
    left: -8px;
    width: 4px;
    height: 75%;
    content: '';
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
  }

  .item-icon {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 5px 0;
  }

  .item-icon svg {
    width: 14px;
    height: 14px;
  }

  .item-text {
    flex: 1;
    min-width: 0;
    margin-left: 8px;
    font-size: 14px;
    line-height: 22px;
  }

  .truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
