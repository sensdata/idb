<template>
  <span
    class="arco-tag arco-tag-size-medium arco-tag-checked"
    :class="{ 'link-activated': itemData.fullPath === $route.fullPath }"
    @click="goto(itemData)"
  >
    <span class="tag-link">
      {{ $t(itemData.title) }}
    </span>
    <span
      class="arco-icon-hover arco-tag-icon-hover arco-icon-hover-size-medium arco-tag-close-btn"
      @click.stop="tagClose(itemData, index)"
    >
      <icon-close />
    </span>
  </span>
</template>

<script lang="ts" setup>
  import { PropType, computed } from 'vue';
  import { useRouter, useRoute } from 'vue-router';
  import { useTabBarStore } from '@/store';
  import type { TagProps } from '@/store/modules/tab-bar/types';

  const props = defineProps({
    itemData: {
      type: Object as PropType<TagProps>,
      default() {
        return [];
      },
    },
    index: {
      type: Number,
      default: 0,
    },
  });

  const router = useRouter();
  const route = useRoute();
  const tabBarStore = useTabBarStore();

  const goto = (tag: TagProps) => {
    router.push({ ...tag });
  };
  const tagList = computed(() => {
    return tabBarStore.getTabList;
  });

  const tagClose = (tag: TagProps, idx: number) => {
    tabBarStore.deleteTag(idx, tag);
    if (props.itemData.fullPath === route.fullPath) {
      const latest = tagList.value[idx - 1]; // 获取队列的前一个tab
      router.push({ name: latest.name });
    }
  };
</script>

<style scoped lang="less">
  .tag-link {
    color: var(--color-text-2);
    text-decoration: none;
  }
  .link-activated {
    color: rgb(var(--link-6));
    .tag-link {
      color: rgb(var(--link-6));
    }
    & + .arco-tag-close-btn {
      color: rgb(var(--link-6));
    }
  }
</style>
