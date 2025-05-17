<template>
  <div class="category-tree">
    <div v-if="items.length === 0" class="empty-text">
      {{ $t('app.crontab.category.tree.empty') }}
      <span class="color-primary" @click="handleCreate">
        {{ $t('app.crontab.category.tree.create') }}
      </span>
    </div>
    <template v-else>
      <div
        v-for="cat of items"
        :key="cat || 'all'"
        class="item"
        :class="{ selected: modelValue === cat }"
        :data-category="cat"
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
  import { getCrontabCategoryListApi } from '@/api/crontab';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { CRONTAB_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';
  import CategoryFormModal from '../category-manage/form-modal.vue';

  const props = defineProps<{
    type: CRONTAB_TYPE;
  }>();

  // 注意：父组件使用的是v-model:selected，所以这里我们需要modelName为'selected'
  const modelValue = defineModel('selected', {
    type: String,
    required: false,
  });

  const items = ref<string[]>([]);

  // 加载分类列表
  const loadCategories = async () => {
    try {
      const ret = await getCrontabCategoryListApi({
        page: 1,
        page_size: 1000,
        type: props.type,
      });

      const newItems = [...ret.items.map((item) => item.name)];

      // 如果当前选中的分类不在列表中，添加到列表中
      if (modelValue.value && !newItems.includes(modelValue.value)) {
        newItems.push(modelValue.value);
      }

      items.value = newItems;

      // 如果没有选择任何分类且列表不为空，选择第一个分类
      if (!modelValue.value && newItems.length > 0) {
        modelValue.value = newItems[0];
      }

      return items.value;
    } catch (err: any) {
      Message.error(err?.message);
      return [];
    }
  };

  // 监听modelValue变化，确保新选择的分类在列表中存在
  watch(
    () => modelValue.value,
    (newCategory) => {
      if (newCategory && !items.value.includes(newCategory)) {
        // 如果选择了一个不在当前列表中的分类，添加到列表中
        items.value = [...items.value, newCategory];
      }
    }
  );

  onMounted(() => {
    loadCategories();
  });

  // 处理点击分类项
  function handleClick(cat: string) {
    if (cat === modelValue.value) return;
    modelValue.value = cat;
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
    selectCategory: (category: string) => {
      if (!category) return;

      // 如果分类不在列表中，添加它
      if (!items.value.includes(category)) {
        items.value = [...items.value, category];
      }

      // 更新选中值
      modelValue.value = category;
    },
    // 暴露分类列表，方便外部检查
    items,
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
