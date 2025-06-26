<template>
  <fixed-footer-bar v-if="visible">
    <template #left>
      <span>
        {{ $t('components.idbTable.batch.selectedPrefix') }}
        <strong class="selected-count">{{ selectedCount }}</strong>
        {{ $t('components.idbTable.batch.selectedSuffix') }}
      </span>
      <a-button type="text" class="cancel-selected" @click="onCancelSelected">{{
        $t('components.idbTable.batch.cancelSelected')
      }}</a-button>
    </template>
    <template #right>
      <slot
        name="batch"
        :selected-rows="selectedRows"
        :selected-row-keys="selectedRowKeys"
      />
    </template>
  </fixed-footer-bar>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';
  import FixedFooterBar from '@/components/fixed-footer-bar/index.vue';
  import type { BaseEntity } from '@/types/global';

  defineOptions({
    name: 'BatchActionBar',
  });

  interface Props {
    /** 是否启用批量操作功能 */
    hasBatch?: boolean;
    /** 选中的行数据 */
    selectedRows: BaseEntity[];
    /** 选中的行键值 */
    selectedRowKeys: number[];
  }

  const props = withDefaults(defineProps<Props>(), {
    hasBatch: true,
  });

  const emit = defineEmits<{
    /** 取消选择事件 */
    cancelSelected: [];
  }>();

  /** 是否显示批量操作栏 */
  const visible = computed(
    () => props.hasBatch && props.selectedRows.length > 0
  );

  /** 选中数量 */
  const selectedCount = computed(() => props.selectedRows.length);

  /** 取消选择处理函数 */
  const onCancelSelected = (): void => {
    emit('cancelSelected');
  };
</script>

<style scoped lang="less">
  .selected-count {
    margin: 0 5px;
    color: rgb(var(--primary-6));
    font-weight: bold;
  }

  .cancel-selected {
    margin-left: 20px;
  }
</style>
