<template>
  <a-dropdown
    :popup-visible="computedPopupVisible"
    trigger="focus"
    auto-fit-popup-width
    prevent-focus
    :click-to-close="false"
    :popup-offset="4"
    @popup-visible-change="handlePopupVisibleChange"
    @select="handleSelect"
  >
    <a-input
      ref="inputRef"
      v-model="value"
      :placeholder="$t('components.addressBar.input.placeholder')"
      class="address-bar"
      allow-clear
      @clear="handleClear"
      @input="handleInputValueChange"
      @press-enter="handleGo"
    >
      <template #prefix>
        <div class="before" @click="handleHome" @mousedown.stop>
          <icon-home />
        </div>
        <a-breadcrumb :max-count="4" class="breadcrumb" @mousedown.stop>
          <a-breadcrumb-item
            v-for="bc of breadcrumbItems"
            :key="bc.path"
            :class="bc.class"
            @click="emit('goto', bc.path)"
          >
            {{ bc.name }}
          </a-breadcrumb-item>
        </a-breadcrumb>
      </template>
      <template #suffix>
        <div class="after" @mousedown.stop @click="handleGo">
          <icon-arrow-right />
        </div>
      </template>
    </a-input>
    <template #content>
      <a-doption v-for="item of validOptions" :key="item.value" :value="item">
        {{ item.label }}
      </a-doption>
    </template>
  </a-dropdown>
</template>

<script lang="ts" setup>
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { debounce } from 'lodash';
  import { computed, ref } from 'vue';

  const props = defineProps<{
    path: string;
    items?: FileInfoEntity[];
  }>();

  const emit = defineEmits(['goto', 'search', 'clear']);

  const inputRef = ref();
  const breadcrumbItems = computed(() => {
    const arr = props.path.split('/').filter((item) => !!item);
    const bcItems = arr.map((item, index) => {
      return {
        name: item,
        path: arr.slice(0, index + 1).join('/'),
        class: index === arr.length - 1 ? '' : 'link',
      };
    });
    bcItems.push({
      name: '',
      path: '',
      class: 'hidden',
    });

    return bcItems;
  });

  const value = ref('');

  const handleInputValueChange = debounce(() => {
    emit('search', {
      path: props.path,
      word: value.value,
    });
  }, 300);

  const handleClear = () => {
    value.value = '';
    emit('clear');
    emit('search', {
      path: props.path,
      word: '',
    });
  };

  const validOptions = computed(() => {
    return (props.items || []).map((item) => ({
      value: item.name,
      label: item.name,
    }));
  });

  const popupVisible = ref(false);
  const computedPopupVisible = computed(
    () => popupVisible.value && validOptions.value.length > 0
  );

  function handlePopupVisibleChange(visible: boolean) {
    popupVisible.value = visible;
  }

  function handleHome() {
    emit('goto', '/');
  }

  function handleGo() {
    const v = value.value.trim();
    if (!v) {
      return;
    }

    // if (!(props.items || []).some((item) => item.name === v)) {
    //   return;
    // }

    value.value = '';
    emit('goto', [props.path, v].join('/'));
  }

  function handleSelect(item: any) {
    inputRef.value?.blur();
    value.value = item.value;
    handleGo();
  }
</script>

<style scoped>
  .address-bar {
    padding-right: 0;
    padding-left: 0;
  }

  .address-bar :deep(.arco-input-prefix) {
    padding-right: 4px;
    .address-bar :deep(.arco-input-suffix) {
      padding: 0;
    }
  }

  .before {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 32px;
    background-color: var(--color-fill-2);
    border-right: 1px solid var(--color-border-2);
    cursor: pointer;
  }

  .after {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 36px;
    height: 32px;
    border-left: 1px solid var(--color-border-2);
    cursor: pointer;
  }

  .address-bar :deep(.arco-input-clear-btn) {
    margin-right: 10px;
  }

  .address-bar :deep(.arco-input) {
    margin-right: 8px;
  }

  .breadcrumb {
    margin-left: 4px;
  }

  .breadcrumb :deep(.arco-breadcrumb-item.link) {
    color: rgb(var(--link-6));
    cursor: pointer;
  }

  .breadcrumb :deep(.arco-breadcrumb-item-separator) {
    margin: 0;
  }
</style>
