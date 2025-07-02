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
  import { nextTick, ref } from 'vue';
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
  const hostId = hostStore.currentId;
  const logsRef = ref<InstanceType<typeof LogsView>>();
  const eventSourceRef = ref();

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

  function connect(configFiles: string) {
    contentArrayRef.value = [];
    const url = resolveApiUrl(`/docker/${hostId}/compose/logs/tail`, {
      config_files: configFiles,
    });
    eventSourceRef.value = new EventSource(url);

    const handleClose = () => {
      eventSourceRef.value.close();
      emit('ok');
    };

    eventSourceRef.value.addEventListener('close', handleClose);
    eventSourceRef.value.addEventListener('end', handleClose);

    eventSourceRef.value.addEventListener('log', (event: Event) => {
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
    eventSourceRef.value.addEventListener('error', (event: Event) => {
      if (event.type === 'error') {
        Message.error(t('app.docker.compose.logsModal.error'));
      }
    });
  }

  function show() {
    visible.value = true;
  }

  function handleCancel() {
    eventSourceRef.value?.close();
    visible.value = false;
  }

  defineExpose({
    show,
    connect,
  });
</script>
