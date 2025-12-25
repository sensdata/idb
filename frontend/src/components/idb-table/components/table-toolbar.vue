<template>
  <a-row v-if="hasToolbar" align="center" style="margin-bottom: 16px">
    <a-col :span="12">
      <a-space>
        <slot name="leftActions" />
      </a-space>
    </a-col>
    <a-col
      :span="12"
      style="display: flex; align-items: center; justify-content: end"
    >
      <slot name="rightActions" />
      <a-input-search
        v-if="hasSearch"
        v-model="searchValue"
        class="w-[240px] mr-4"
        :placeholder="$t('components.idbTable.search.placeholder')"
        :loading="loading"
        allow-clear
        @clear="onClear"
        @search="onSearch"
        @press-enter="onSearchEnter"
      />
      <a-button v-if="download">
        <template #icon>
          <icon-download />
        </template>
        {{ $t('components.idbTable.actions.download') }}
      </a-button>
      <a-tooltip :content="$t('components.idbTable.actions.refresh')">
        <div class="action-icon" @click="onRefresh"
          ><icon-refresh size="18"
        /></div>
      </a-tooltip>
      <a-dropdown @select="onSelectDensity">
        <a-tooltip :content="$t('components.idbTable.actions.density')">
          <div class="action-icon"><icon-line-height size="18" /></div>
        </a-tooltip>
        <template #content>
          <a-doption
            v-for="item in densityList"
            :key="item.value"
            :value="item.value"
            :class="{ active: item.value === size }"
          >
            <span>{{ item.name }}</span>
          </a-doption>
        </template>
      </a-dropdown>
      <a-tooltip :content="$t('components.idbTable.actions.columnSetting')">
        <a-popover
          trigger="click"
          position="br"
          @popup-visible-change="onPopupVisibleChange"
        >
          <div class="action-icon"><icon-settings size="18" /></div>
          <template #content>
            <div id="tableSetting">
              <div
                v-for="item in showColumns"
                :key="item.dataIndex"
                class="setting"
              >
                <div style="margin-right: 4px; cursor: move">
                  <icon-drag-arrow />
                </div>
                <div>
                  <a-checkbox
                    v-model="item.checked"
                    @change="(checked: boolean) => onToggleColumn(checked, item)"
                  >
                    {{ item.title }}
                  </a-checkbox>
                </div>
              </div>
            </div>
          </template>
        </a-popover>
      </a-tooltip>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
  import { ref, computed, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import Sortable from 'sortablejs';
  import { Column, SizeProps, DownloadFn } from '../types';

  defineOptions({
    name: 'TableToolbar',
  });

  const props = withDefaults(
    defineProps<{
      hasToolbar?: boolean;
      hasSearch?: boolean;
      loading: boolean;
      download?: DownloadFn;
      size: SizeProps;
      showColumns: Column[];
    }>(),
    {
      hasToolbar: true,
      hasSearch: false,
    }
  );

  const emit = defineEmits<{
    search: [value: string];
    refresh: [];
    toggleColumn: [checked: boolean, column: Column];
    selectDensity: [size: SizeProps];
    columnsReordered: [columns: Column[]];
  }>();

  const { t } = useI18n();
  const searchValue = ref('');

  // 密度选项
  const densityList = computed(() => [
    { name: t('components.idbTable.size.mini'), value: 'mini' },
    { name: t('components.idbTable.size.small'), value: 'small' },
    { name: t('components.idbTable.size.medium'), value: 'medium' },
    { name: t('components.idbTable.size.large'), value: 'large' },
  ]);

  // 搜索相关
  const onSearch = (value: string) => {
    searchValue.value = value;
    emit('search', value);
  };

  const onClear = () => {
    onSearch('');
  };

  const onSearchEnter = () => {
    onSearch(searchValue.value);
  };

  // 刷新
  const onRefresh = () => {
    emit('refresh');
  };

  // 列设置
  const onToggleColumn = (
    checked: boolean | (string | boolean | number)[],
    column: Column
  ) => {
    emit('toggleColumn', checked as boolean, column);
  };

  // 密度设置
  const onSelectDensity = (
    val: string | number | Record<string, any> | undefined
  ) => {
    emit('selectDensity', val as SizeProps);
  };

  // 列排序
  const onPopupVisibleChange = (val: boolean) => {
    if (val) {
      nextTick(() => {
        const el = document.getElementById('tableSetting') as HTMLElement;
        // eslint-disable-next-line no-new
        new Sortable(el, {
          onEnd(e: any) {
            const { oldIndex, newIndex } = e;
            const newColumns = [...props.showColumns];
            if (oldIndex > -1 && newIndex > -1) {
              newColumns.splice(
                oldIndex,
                1,
                newColumns.splice(newIndex, 1, newColumns[oldIndex]).pop()!
              );
            }
            emit('columnsReordered', newColumns);
          },
        });
      });
    }
  };

  // 暴露方法
  defineExpose({
    clearSearch: () => {
      searchValue.value = '';
    },
  });
</script>

<style scoped lang="less">
  .action-icon {
    margin-left: 12px;
    color: var(--color-text-2);
    cursor: pointer;
    transition: color 0.2s ease;
  }

  .action-icon:hover {
    color: var(--color-text-1);
  }

  .active {
    color: rgb(var(--primary-6));
    background-color: var(--color-primary-light-1);
  }

  .setting {
    display: flex;
    align-items: center;
    min-width: 160px;
  }
</style>
