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
      v-model="value"
      :placeholder="$t('components.addressBar.input.placeholder')"
      class="address-bar"
      allow-clear
      ref="inputRef"
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
  <!-- <div class="address-bar">
    <a-auto-complete
      v-model="value"
      :data="formatedOptions"
      :style="{ width: '720px' }"
    />
  </div> -->
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';

  const props = defineProps<{
    path: string;
    items: Array<{
      name: string;
      path: string;
    }>;
  }>();

  const inputRef = ref();
  const breadcrumbItems = computed(() => {
    const arr = props.path.split('/');
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

  const formatedOptions = computed(() => {
    return props.items.map((item) => ({
      value: item.name,
      label: item.name,
    }));
  });

  const validOptions = computed(() => {
    return formatedOptions.value.filter((item) =>
      item.label.includes(value.value)
    );
  });

  const popupVisible = ref(false);
  const computedPopupVisible = computed(
    () => popupVisible.value && validOptions.value.length > 0
  );

  function handlePopupVisibleChange(visible: boolean) {
    popupVisible.value = visible;
  }

  function handleSelect(item: any) {
    inputRef.value?.blur();
    value.value = item.value;
  }

  function handleHome() {
    console.log('home');
  }

  function handleGo() {
    console.log('go', value.value);
  }
</script>

<style scoped>
  .address-bar {
    padding-left: 0;
    padding-right: 0;
  }
  .address-bar :deep(.arco-input-prefix) {
    padding-right: 4px;
  }
  .before {
    display: flex;
    width: 36px;
    height: 32px;
    cursor: pointer;
    align-items: center;
    justify-content: center;
    background-color: var(--color-fill-2);
    border-right: 1px solid var(--color-border-2);
  }
  .address-bar :deep(.arco-input-suffix) {
    padding: 0;
  }
  .after {
    display: flex;
    width: 36px;
    height: 32px;
    cursor: pointer;
    align-items: center;
    justify-content: center;
    border-left: 1px solid var(--color-border-2);
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
    cursor: pointer;
    color: rgb(var(--link-6));
  }
  .breadcrumb :deep(.arco-breadcrumb-item-separator) {
    margin: 0;
  }
</style>
