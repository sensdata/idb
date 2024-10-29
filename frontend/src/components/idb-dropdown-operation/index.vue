<template>
  <a-dropdown-button
    v-if="options.length > 0"
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
</template>

<script setup lang="ts">
  import { computed } from 'vue';

  interface OperationOption {
    text: string;
    disabled?: boolean;
    visible?: boolean;
    click: (event: Event) => void;
  }

  const props = defineProps<{
    options: OperationOption[];
  }>();

  const options = computed(() => {
    return props.options.filter((option) => option.visible !== false);
  });

  const handleClick = (option: any, event: Event) => {
    option.click(event);
  };
</script>
