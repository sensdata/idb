# IDB-Tree 通用树形组件

一个支持分类管理的通用树形组件，用于展示和管理分层数据结构。

## 快速开始

### 基础树形展示

```vue
<template>
  <idb-tree
    :items="treeItems"
    v-model:selected="selected"
    @select="handleSelect"
  />
</template>

<script setup>
import { ref } from 'vue';
import IdbTree from '@/components/idb-tree/index.vue';

const treeItems = ref([
  { id: 1, label: '文件夹1', icon: 'folder' },
  { id: 2, label: '文件夹2', icon: 'folder' },
]);

const selected = ref(null);

const handleSelect = (item) => {
  console.log('选中:', item);
};
</script>
```

### 分类管理模式

```vue
<template>
  <category-tree
    v-model:selected-category="selectedCategory"
    :category-config="categoryConfig"
    :enable-category-management="true"
    :host-id="currentHostId"
    :categories="categories"
    @category-created="handleCategoryCreated"
    @category-updated="handleCategoryUpdated"
    @category-deleted="handleCategoryDeleted"
  />
</template>

<script setup>
import { computed } from 'vue';
import CategoryTree from '@/components/idb-tree/category-tree.vue';
import { createLogrotateCategoryConfig } from './adapters/category-adapter';

const categoryConfig = computed(() => 
  createLogrotateCategoryConfig('local')
);

const selectedCategory = ref('');
const categories = ref(['默认分类', '系统分类']);

const handleCategoryCreated = (name) => {
  console.log('创建分类:', name);
};

const handleCategoryUpdated = (oldName, newName) => {
  console.log('更新分类:', oldName, '->', newName);
};

const handleCategoryDeleted = (name) => {
  console.log('删除分类:', name);
};
</script>
```

## 业务适配器

创建业务适配器来处理具体的 API 调用：

```typescript
// adapters/category-adapter.ts
import { CategoryApiAdapter, CategoryManagerConfig } from '@/components/idb-tree/types/category';
import { getLogrotateCategoriesApi, createLogrotateCategoryApi } from '@/api/logrotate';

export class LogrotateCategoryApiAdapter implements CategoryApiAdapter {
  constructor(private type: string) {}

  async getCategories(params) {
    const response = await getLogrotateCategoriesApi(
      this.type,
      params.page || 1,
      params.pageSize || 100,
      params.host
    );
    
    return {
      items: response.items.map(item => ({
        name: item.name,
        type: 'logrotate',
        count: item.count || 0,
      })),
      total: response.total || 0,
    };
  }

  async createCategory(params) {
    await createLogrotateCategoryApi(this.type, params.name, params.host);
  }

  async updateCategory(params) {
    await updateLogrotateCategoryApi(this.type, params.oldName, params.newName, params.host);
  }

  async deleteCategory(params) {
    await deleteLogrotateCategoryApi(this.type, params.name, params.host);
  }
}

export function createLogrotateCategoryConfig(type) {
  return {
    type: 'logrotate',
    apiAdapter: new LogrotateCategoryApiAdapter(type),
    allowCreate: true,
    allowEdit: true,
    allowDelete: true,
  };
}
```

## 完整示例

在 Logrotate 模块中的完整使用：

```vue
<template>
  <app-sidebar-layout>
    <template #sidebar>
      <category-tree
        ref="categoryTreeRef"
        v-model:selected-category="params.category"
        :category-config="categoryConfig"
        :enable-category-management="true"
        :host-id="currentHostId"
        :categories="categoryItems"
        @category-created="handleCategoryCreated"
        @category-updated="handleCategoryUpdated"
        @category-deleted="handleCategoryDeleted"
      />
    </template>
    <template #main>
      <idb-table
        :params="params"
        :columns="columns"
        :fetch="fetchData"
      />
    </template>
  </app-sidebar-layout>
</template>

<script setup>
import { computed, ref } from 'vue';
import CategoryTree from '@/components/idb-tree/category-tree.vue';
import { createLogrotateCategoryConfig } from './adapters/category-adapter';

const props = defineProps({
  type: { type: String, required: true }
});

const categoryConfig = computed(() => 
  createLogrotateCategoryConfig(props.type)
);

const params = ref({ category: '' });
const categoryItems = ref([]);

const handleCategoryCreated = async (name) => {
  await loadCategories();
  gridRef.value?.reload();
};

const handleCategoryUpdated = async (oldName, newName) => {
  await loadCategories();
  gridRef.value?.reload();
};

const handleCategoryDeleted = async (name) => {
  await loadCategories();
  gridRef.value?.reload();
};
</script>
```

## API 参考

### CategoryTree Props

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `selectedCategory` | `string` | `''` | 选中的分类 |
| `categoryConfig` | `CategoryManagerConfig` | - | 分类管理配置 |
| `enableCategoryManagement` | `boolean` | `false` | 是否启用分类管理 |
| `hostId` | `number` | - | 主机ID |
| `categories` | `string[]` | `[]` | 分类列表 |

### CategoryTree Events

| 事件 | 参数 | 说明 |
|------|------|------|
| `category-created` | `(name: string)` | 分类创建成功 |
| `category-updated` | `(oldName: string, newName: string)` | 分类更新成功 |
| `category-deleted` | `(name: string)` | 分类删除成功 |

### CategoryApiAdapter 接口

```typescript
interface CategoryApiAdapter {
  getCategories(params: CategoryListParams): Promise<CategoryListResult>;
  createCategory(params: CategoryCreateParams): Promise<void>;
  updateCategory(params: CategoryUpdateParams): Promise<void>;
  deleteCategory(params: CategoryDeleteParams): Promise<void>;
}
```
