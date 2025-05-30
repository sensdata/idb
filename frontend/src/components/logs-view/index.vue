<template>
  <codemirror
    ref="cmRef"
    v-model="value"
    :autofocus="true"
    :indent-with-tab="true"
    :tabSize="4"
    :lineWrapping="true"
    :matchBrackets="true"
    theme="cobalt"
    :styleActiveLine="true"
    :extensions="extensions"
    :disabled="true"
  />
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import { javascript } from '@codemirror/lang-javascript';
  import { oneDark } from '@codemirror/theme-one-dark';

  const props = defineProps({
    content: {
      type: String,
      default: '',
    },
  });

  const cmRef = ref<InstanceType<typeof Codemirror>>();
  const scrollBottom = () => {
    const sc = cmRef.value?.$el.querySelector('.cm-scroller');
    if (sc) {
      sc.scrollTo(0, sc.scrollHeight);
    }
  };

  const value = ref(props.content);
  watch(
    () => props.content,
    (newVal) => {
      value.value = newVal;
    },
    {
      immediate: true,
    }
  );
  const extensions = [javascript(), oneDark];

  defineExpose({
    scrollBottom,
  });
</script>
