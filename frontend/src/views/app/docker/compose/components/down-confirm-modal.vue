<template>
  <a-modal
    v-model:visible="visible"
    :width="600"
    :title="t('common.confirm.title')"
    unmount-on-close
    @ok="onOk"
    @cancel="onCancel"
  >
    <p>
      {{ t('app.docker.compose.down.confirm.message') }}
    </p>
    <a-checkbox v-model="removeVolumes">{{
      t('app.docker.compose.down.confirm.removeVolumes')
    }}</a-checkbox>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  const emit = defineEmits(['confirm']);
  const { t } = useI18n();
  const visible = ref(false);

  const removeVolumes = ref(false);
  const composeName = ref('');
  const show = (name: string) => {
    visible.value = true;
    composeName.value = name;
    removeVolumes.value = false;
  };
  const hide = () => {
    visible.value = false;
  };
  const onOk = async () => {
    emit('confirm', {
      name: composeName.value,
      remove_volumes: removeVolumes.value,
    });
  };
  const onCancel = hide;

  defineExpose({ show });
</script>
