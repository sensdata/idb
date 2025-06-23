<template>
  <div
    v-if="options.length > 0"
    class="idb-table-operation"
    :class="`type-${type}`"
  >
    <a-dropdown-button
      v-if="type === 'dropdown'"
      :disabled="options[0].disabled"
      @select="handleClick"
      @click="handleClick(options[0], $event)"
    >
      {{ options[0].text }}
      <template #icon>
        <icon-down />
      </template>
      <template #content>
        <a-doption
          v-for="option of options.slice(1)"
          :key="option.text"
          :value="option"
          :disabled="option.disabled"
        >
          <div class="min-w-32">{{ option.text }}</div>
        </a-doption>
      </template>
    </a-dropdown-button>
    <a-button-group v-else-if="type === 'button-group'">
      <a-button
        v-for="item of options"
        :key="item.text"
        :="item"
        :type="item.type || 'primary'"
        :size="item.size || 'small'"
        @click="handleClick(item, $event)"
      >
        {{ item.text }}
      </a-button>
    </a-button-group>
    <template v-else>
      <a-button
        v-for="item of options"
        :key="item.text"
        :="item"
        :size="item.size || 'small'"
        :type="item.type || 'text'"
        @click="handleClick(item, $event)"
      >
        {{ item.text }}
      </a-button>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { useConfirm } from '@/composables/confirm';
  import { computed } from 'vue';

  interface OperationOption {
    text: string;
    disabled?: boolean;
    visible?: boolean;
    confirm?: string;
    size?: 'mini' | 'medium' | 'large' | 'small';
    type?: 'dashed' | 'text' | 'outline' | 'primary' | 'secondary'; // type为button时默认为text，type为button-group时默认为primary
    status?: 'normal' | 'success' | 'warning' | 'danger'; // 仅按钮类型
    click: (event: Event) => void;
  }

  const props = withDefaults(
    defineProps<{
      type?: 'dropdown' | 'button' | 'button-group';
      options: OperationOption[];
    }>(),
    {
      type: 'dropdown',
    }
  );

  const options = computed(() => {
    return props.options.filter((option) => option.visible !== false);
  });

  const { confirm } = useConfirm();
  const handleClick = async (option: any, event: Event) => {
    if (option.confirm && !(await confirm(option.confirm))) {
      return;
    }
    option.click(event);
  };
</script>

<style scoped>
  .idb-table-operation.type-button :deep(.arco-btn) {
    padding-right: 4px;
    padding-left: 4px;
  }
</style>
