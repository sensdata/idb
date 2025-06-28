<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('common.confirm.title')"
    unmount-on-close
    @ok="onOk"
    @cancel="onCancel"
  >
    <p>
      {{ t('app.compose.down.confirm.message') }}
    </p>
    <a-checkbox v-model="deleteVolumes">{{
      t('app.compose.down.confirm.deleteVolumes')
    }}</a-checkbox>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  const emit = defineEmits(['confirm']);
  const { t } = useI18n();
  const visible = ref(false);

  const deleteVolumes = ref(false);
  const composeName = ref('');
  const show = (name: string) => {
    visible.value = true;
    composeName.value = name;
    deleteVolumes.value = false;
  };
  const hide = () => {
    visible.value = false;
  };
  const onOk = async () => {
    emit('confirm', {
      name: composeName.value,
      delete_volumes: deleteVolumes.value,
    });
  };
  const onCancel = hide;

  defineExpose({ show });
</script>
