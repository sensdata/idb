<template>
  <div class="category-tree">
    <div v-if="items.length === 0" class="empty-text">
      {{ $t('app.logrotate.category.tree.empty') }}
      <span class="color-primary" @click="handleCreate">
        {{ $t('app.logrotate.category.tree.create') }}
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
  /**
   * Logrotate 分类树组件
   *
   * 功能：
   * - 显示日志轮转配置的分类列表
   * - 支持选择分类
   * - 支持创建新分类
   * - 自动同步分类列表
   */

  import { onMounted, ref, watch } from 'vue';
  import { getLogrotateCategoriesApi } from '@/api/logrotate';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { LOGROTATE_TYPE } from '@/config/enum';
  import { Message } from '@arco-design/web-vue';
  import useCurrentHost from '@/hooks/current-host';
  import { useLogger } from '@/hooks/use-logger';
  import CategoryFormModal from '../category-manage/form-modal.vue';

  // 日志记录
  const { logInfo, logError } = useLogger('CategoryTree');

  // Props 定义
  const props = defineProps<{
    type: LOGROTATE_TYPE;
  }>();

  // 双向绑定：父组件使用 v-model:selected
  const modelValue = defineModel('selected', {
    type: String,
    required: false,
  });

  // Composables
  const { currentHostId } = useCurrentHost();

  // 响应式数据
  const items = ref<string[]>([]);
  const formRef = ref<InstanceType<typeof CategoryFormModal>>();
  const isLoading = ref(false);

  /**
   * 加载分类列表
   */
  const loadCategories = async () => {
    const hostId = currentHostId.value;
    if (!hostId) {
      Message.error('Host ID is required');
      return;
    }

    // 防止重复加载
    if (isLoading.value) {
      logInfo('分类正在加载中，跳过重复请求');
      return;
    }

    logInfo('开始加载分类列表');
    isLoading.value = true;
    try {
      const ret = await getLogrotateCategoriesApi(props.type, 1, 1000, hostId);
      logInfo(`分类 API 返回数据:`, ret);

      const newItems = [...ret.items.map((item) => item.name)];
      logInfo(`处理后的分类列表:`, newItems);

      // 如果当前选中的分类不在列表中，添加到列表中
      if (modelValue.value && !newItems.includes(modelValue.value)) {
        newItems.push(modelValue.value);
        logInfo(`添加当前选中分类到列表: ${modelValue.value}`);
      }

      items.value = newItems;
      logInfo(`分类列表已更新，当前选中: ${modelValue.value}`);

      // 如果没有选择任何分类且列表不为空，选择第一个分类
      if (!modelValue.value && newItems.length > 0) {
        logInfo(`自动选择第一个分类: ${newItems[0]}`);
        modelValue.value = newItems[0];
      }
    } catch (err: any) {
      logError('加载分类失败', err);
      Message.error(err?.message || 'Failed to load categories');
    } finally {
      isLoading.value = false;
      logInfo('分类加载完成');
    }
  };

  /**
   * 监听modelValue变化，确保新选择的分类在列表中存在
   */
  watch(
    () => modelValue.value,
    (newCategory, oldCategory) => {
      logInfo(`分类选择变化: ${oldCategory} -> ${newCategory}`);
      if (newCategory && !items.value.includes(newCategory)) {
        // 如果选择了一个不在当前列表中的分类，添加到列表中
        items.value = [...items.value, newCategory];
        logInfo(`添加新分类到列表: ${newCategory}`);
      }
    }
  );

  // 生命周期
  onMounted(() => {
    loadCategories();
  });

  /**
   * 处理点击分类项
   */
  function handleClick(cat: string) {
    if (cat === modelValue.value) return;
    modelValue.value = cat;
  }

  /**
   * 处理创建分类
   */
  function handleCreate() {
    formRef.value?.show();
  }

  /**
   * 处理创建分类成功回调
   */
  function handleCreateOk() {
    loadCategories();
  }

  /**
   * 选择指定分类
   */
  const selectCategory = (category: string) => {
    if (!category) {
      return;
    }

    // 如果分类不在列表中，添加它
    if (!items.value.includes(category)) {
      items.value = [...items.value, category];
    }

    // 更新选中值
    modelValue.value = category;
  };

  // 暴露给父组件的方法和数据
  defineExpose({
    reload: loadCategories,
    refresh: loadCategories,
    selectCategory,
    items,
  });
</script>

<style scoped>
  .category-tree {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 16px;
    background: var(--color-bg-2);
    border-radius: 4px;
  }

  .empty-text {
    padding: 20px;
    font-size: 14px;
    color: var(--color-text-3);
    text-align: center;
  }

  .color-primary {
    margin-left: 4px;
    color: var(--color-primary-6);
    cursor: pointer;
    transition: color 0.2s ease;
  }

  .color-primary:hover {
    color: var(--color-primary-5);
  }

  .item {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    margin-bottom: 4px;
    cursor: pointer;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .item:hover {
    background: var(--color-fill-2);
  }

  .item.selected {
    color: var(--color-primary-6);
    background: var(--color-primary-light-1);
  }

  .item-icon {
    flex-shrink: 0;
    width: 16px;
    height: 16px;
    margin-right: 8px;
  }

  .item-text {
    flex: 1;
    font-size: 14px;
  }

  .truncate {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
