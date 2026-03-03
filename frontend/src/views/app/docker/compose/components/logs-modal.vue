<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.docker.compose.logsModal.title')"
    :footer="false"
    :width="1200"
    destroy-on-close
    @cancel="handleCancel"
  >
    <logs-view ref="logsRef" style="height: 85vh" :content="contentRef" />
  </a-modal>
</template>

<script setup lang="ts">
  import { computed, nextTick, onBeforeUnmount, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import LogsView from '@/components/logs-view/index.vue';
  import { resolveApiUrl } from '@/helper/api-helper';
  import { useHostStore } from '@/store';
  import { Message } from '@arco-design/web-vue';
  import { debounce } from 'lodash';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const visible = ref(false);
  const contentArrayRef = ref<string[]>([]);
  const hostStore = useHostStore();
  const hostId = computed(() => hostStore.currentId);
  const logsRef = ref<InstanceType<typeof LogsView>>();
  const eventSourceRef = ref<EventSource | null>(null);

  const contentRef = ref('');
  const updateContent = () => {
    contentRef.value = contentArrayRef.value.join('');
    nextTick(() => {
      if (logsRef.value) {
        logsRef.value.scrollBottom();
      }
    });
  };

  const updateContentWithDebounce = debounce(updateContent, 300, {
    maxWait: 1000,
  });

  const closeEventSource = () => {
    eventSourceRef.value?.close();
    eventSourceRef.value = null;
    updateContentWithDebounce.cancel();
  };

  function connect(configFiles: string) {
    closeEventSource();
    contentArrayRef.value = [];
    contentRef.value = '';
    const url = resolveApiUrl('/docker/{host}/compose/logs/tail', {
      host: hostId.value,
      config_files: configFiles,
    });
    const eventSource = new EventSource(url);
    eventSourceRef.value = eventSource;

    const handleClose = () => {
      closeEventSource();
      emit('ok');
    };

    eventSource.addEventListener('close', handleClose);
    eventSource.addEventListener('end', handleClose);

    eventSource.addEventListener('log', (event: Event) => {
      if (event instanceof MessageEvent) {
        if (event.data) {
          const rawData = event.data;
          contentArrayRef.value.push(rawData);
          if (contentArrayRef.value.length > 500) {
            contentArrayRef.value = contentArrayRef.value.slice(-500);
          }
          if (contentArrayRef.value.length === 1) {
            updateContent();
          } else {
            updateContentWithDebounce();
          }
        }
      }
    });
    eventSource.addEventListener('error', (event: Event) => {
      if (event.type === 'error') {
        Message.error(t('app.docker.compose.logsModal.error'));
      }
    });
  }

  function show() {
    visible.value = true;
  }

  function handleCancel() {
    closeEventSource();
    visible.value = false;
  }

  onBeforeUnmount(() => {
    closeEventSource();
  });

  defineExpose({
    show,
    connect,
  });
</script>
