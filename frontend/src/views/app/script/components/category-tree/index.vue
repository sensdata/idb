<template>
  <div class="category-tree">
    <div v-if="items.length === 0" class="empty-text">
      {{ $t('app.script.category.tree.empty') }}
      <span class="color-primary" @click="handleCreate">
        {{ $t('app.script.category.tree.create') }}
      </span>
    </div>
    <template v-else>
      <div
        v-for="cat of items"
        :key="cat || 'all'"
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
  import { onMounted, ref } from 'vue';
  import { getScriptCategoryListApi } from '@/api/script';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';
  import CategoryFormModal from '../category-manage/form-modal.vue';

  const props = defineProps<{
    type: SCRIPT_TYPE;
  }>();

  const selected = defineModel('selected', {
    type: String,
    required: false,
  });

  const items = ref<string[]>([]);
  const loadCategories = async () => {
    try {
      const ret = await getScriptCategoryListApi({
        page: 1,
        page_size: 1000,
        type: props.type,
      });
      items.value = [...ret.items.map((item) => item.name)];
      if (items.value.length > 0) {
        selected.value = items.value[0];
      } else {
        selected.value = '';
      }
    } catch (err: any) {
      Message.error(err?.message);
    }
  };
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

  defineExpose({
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

  .item {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 32px;
    margin-bottom: 8px;
    padding-left: 10px;
    line-height: 32px;
    border-radius: 4px;
    cursor: pointer;
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
    background-color: rgb(var(--primary-6));
    border-radius: 11px;
    content: '';
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
</style>
