<template>
  <a-row v-if="items.length > 0">
    <a-col :flex="1">
      <a-form
        :model="formModel"
        :label-col-props="{ span: 6 }"
        :wrapper-col-props="{ span: 18 }"
        :label-align="labelAlign"
        :auto-label-width="items.length <= 3"
      >
        <a-row :gutter="50">
          <a-col v-for="item of items" :key="item.field" :span="8">
            <a-form-item :field="item.field" :label="$t(item.label)">
              <a-input
                v-if="item.type === 'input'"
                v-model="formModel[item.field]"
                :placeholder="item.placeholder ? $t(item.placeholder!) : ''"
                v-bind="getOtherProps(item)"
                @press-enter="search"
              />
              <a-input-number
                v-else-if="item.type === 'input-number'"
                v-model="formModel[item.field]"
                :placeholder="item.placeholder ? $t(item.placeholder!) : ''"
                v-bind="getOtherProps(item)"
                @press-enter="search"
              />
              <a-select
                v-else-if="item.type === 'select'"
                v-model="formModel[item.field]"
                :options="item.options?.map((item) => ({ ...item, label: $t(item.label!) }))"
                :placeholder="
                  $t(
                    item.placeholder ||
                      'components.idbTable.filter.selectDefault'
                  )
                "
                v-bind="getOtherProps(item)"
                @clear="handleItemClear(item)"
              />
              <a-range-picker
                v-else-if="item.type === 'range-picker'"
                v-model="formModel[item.field]"
                v-bind="getOtherProps(item)"
                style="width: 100%"
              />
              <a-date-picker
                v-else-if="item.type === 'date-picker'"
                v-model="formModel[item.field]"
                v-bind="getOtherProps(item)"
                style="width: 100%"
              />
              <a-month-picker
                v-else-if="item.type === 'month-picker'"
                v-model="formModel[item.field]"
                v-bind="getOtherProps(item)"
                style="width: 100%"
              />
              <component
                :is="item.component"
                v-else-if="item.type === 'component'"
                v-model="formModel[item.field]"
                v-bind="getCustomComponentProps(item)"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-col>
    <a-divider :style="`height: ${dividerHeight}px`" direction="vertical" />
    <a-col :flex="dividerHeight + 2 + 'px'" style="text-align: right">
      <a-space
        :direction="items.length > 3 ? 'vertical' : 'horizontal'"
        :size="18"
      >
        <a-button type="primary" @click="search">
          <template #icon>
            <icon-search />
          </template>
          {{ $t('components.idbTable.filter.search') }}
        </a-button>
        <a-button @click="reset">
          <template #icon>
            <icon-refresh />
          </template>
          {{ $t('components.idbTable.filter.reset') }}
        </a-button>
      </a-space>
    </a-col>
  </a-row>
</template>

<script lang="ts" setup>
  import { ref, computed, PropType, onMounted } from 'vue';
  import { cloneDeep, omit } from 'lodash';
  import { FilterItem } from './types';

  const props = defineProps({
    items: {
      type: Array as PropType<FilterItem[]>,
      default: () => [],
    },
    labelAlign: {
      type: String as PropType<'left' | 'right'>,
      default: 'left',
    },
  });

  const emit = defineEmits(['filter', 'init']);

  const dividerHeight = computed(() => {
    return Math.ceil(props.items.length / 3) * 52 - 20;
  });

  const getItemDefaultValue = (item: FilterItem) => {
    if (item.defaultValue !== undefined) {
      return item.defaultValue;
    }
    switch (item.type) {
      case 'input':
      case 'input-number':
        return '';
      case 'select':
        return undefined;
      case 'range-picker':
        return [undefined, undefined];
      case 'date-picker':
        return undefined;
      default:
        return undefined;
    }
  };

  const generateDefaultModel = () => {
    return cloneDeep(
      props.items.reduce((acc, cur) => {
        acc[cur.field] = getItemDefaultValue(cur);
        return acc;
      }, {} as Record<string, any>)
    );
  };

  const formModel = ref(generateDefaultModel());

  const toParams = () => {
    return props.items.reduce((acc, cur) => {
      if (cur.toParams) {
        acc[cur.field] = cur.toParams(formModel.value[cur.field], cur);
      } else {
        acc[cur.field] = formModel.value[cur.field];
      }
      return acc;
    }, {} as Record<string, any>);
  };

  const search = () => {
    emit('filter', toParams());
  };

  const reset = () => {
    formModel.value = generateDefaultModel();
    emit('filter', toParams());
  };

  const handleItemClear = (item: FilterItem) => {
    formModel.value = Object.assign(formModel.value, {
      [item.field]: getItemDefaultValue(item),
    });
  };

  const getOtherProps = (item: FilterItem) => {
    return omit(item, [
      'key',
      'label',
      'type',
      'component',
      'placeholder',
      'defaultValue',
      'options',
      'toParams',
    ]);
  };

  const getCustomComponentProps = (item: FilterItem) => {
    return omit(item, ['key', 'label', 'type', 'component', 'toParams']);
  };

  onMounted(() => {
    emit('init', toParams());
  });

  defineExpose({
    model: formModel,
    toParams,
    search,
    reset,
  });
</script>
