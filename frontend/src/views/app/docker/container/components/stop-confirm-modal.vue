<template>
  <a-modal
    v-model:visible="visible"
    :width="400"
    :title="t('common.confirm.title')"
    unmount-on-close
    @ok="onOk"
    @cancel="onCancel"
  >
    <p>
      {{ t('app.docker.container.stop.confirm.message') }}
    </p>
    <a-checkbox v-model="force">{{
      t('app.docker.container.stop.confirm.force')
    }}</a-checkbox>
  </a-modal>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';

  const emit = defineEmits(['confirm']);
  const { t } = useI18n();
  const visible = ref(false);

  const force = ref(false);
  const nameRef = ref('');
  const show = (name: string) => {
    visible.value = true;
    nameRef.value = name;
    force.value = false;
  };
  const hide = () => {
    visible.value = false;
  };
  const onOk = async () => {
    emit('confirm', {
      name: nameRef.value,
      force: force.value,
    });
  };
  const onCancel = hide;

  defineExpose({ show });
</script>
