<template>
  <a-drawer
    v-model:visible="visible"
    :footer="false"
    height="95vh"
    class="terminal-drawer"
    placement="bottom"
  >
    <template #title>
      <span>{{ $t('components.terminal.title') }}</span>
      <a-button type="primary" status="danger" size="mini" @click="handlePrune">
        {{ $t('components.terminal.session.prune') }}
      </a-button>
    </template>
    <terminal-tabs ref="tabsRef" />
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, watch, nextTick } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useHostStore } from '@/store';
  import { pruneTerminalSessionApi } from '@/api/terminal';
  import { Message } from '@arco-design/web-vue';
  import TerminalTabs from './terminal-tabs.vue';

  const { t } = useI18n();

  const visible = defineModel('visible', {
    type: Boolean,
    required: true,
  });
  const hostStore = useHostStore();
  const tabsRef = ref<InstanceType<typeof TerminalTabs> | null>(null);

  const firstShow = ref(true);
  watch(visible, (val) => {
    if (val && firstShow.value) {
      firstShow.value = false;
      tabsRef?.value?.addItem({
        type: 'attach',
        host: hostStore.current!,
      });
    } else if (val) {
      nextTick(() => {
        tabsRef?.value?.focus();
      });
    }
  });

  async function handlePrune() {
    await pruneTerminalSessionApi(hostStore.currentId!);
    Message.success(t('components.terminal.session.pruneSuccess'));
  }
</script>

<style>
  .terminal-drawer .arco-drawer-title {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 98%;
  }
</style>
