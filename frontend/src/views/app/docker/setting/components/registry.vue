<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.setting.registry.title')"
    destroy-on-close
    @close="handleClose"
    @before-ok="handleBeforeOk"
  >
    <a-spin :loading="loading" class="w-full">
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="insecure_registries"
              :label="t('app.docker.setting.registry.registries')"
            >
              <a-textarea
                v-model="form.insecure_registries"
                :placeholder="
                  t('app.docker.setting.registry.registries.placeholder')
                "
                :auto-size="{ minRows: 3, maxRows: 8 }"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { updateDockerConfApi } from '@/api/docker';
  import { useConfirm } from '@/composables/confirm';

  const { t } = useI18n();
  const emit = defineEmits(['ok']);

  const loading = ref(false);
  const visible = ref(false);
  const formRef = ref();

  const form = reactive({
    insecure_registries: '',
  });

  const rules = reactive({
    insecure_registries: [
      {
        required: true,
        message: t('app.docker.setting.registry.registries.required'),
        trigger: 'blur',
      },
      {
        validator(value: string, callback: any) {
          if (!value) {
            callback();
            return;
          }
          const urls = value
            .split('\n')
            .map((v) => v.trim())
            .filter(Boolean);
          // 允许 http(s)://、IP:PORT、域名:PORT
          const urlReg = /^(https?:\/\/)?([\w\-.]+)(:\d+)?$/i;
          for (const url of urls) {
            if (!urlReg.test(url)) {
              callback(t('app.docker.setting.registry.registries.format'));
              return;
            }
          }
          callback();
        },
        trigger: 'blur',
      },
    ],
  });

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  const setData = (params: any) => {
    if (Array.isArray(params.insecure_registries)) {
      form.insecure_registries = params.insecure_registries.join('\r\n');
    } else {
      form.insecure_registries = params.insecure_registries || '';
    }
  };

  const handleClose = () => {
    visible.value = false;
  };

  const { confirm } = useConfirm();

  const handleBeforeOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) {
      return false;
    }
    const registries = form.insecure_registries
      .split('\r\n')
      .map((v) => v.trim())
      .filter(Boolean);
    if (
      await confirm({
        title: t('app.docker.setting.registry.confirm.title'),
        content: t('app.docker.setting.registry.confirm.content'),
      })
    ) {
      loading.value = true;
      try {
        await updateDockerConfApi({
          key: 'insecure_registries',
          value: registries,
        });
        emit('ok');
      } catch (err: any) {
        loading.value = false;
        Message.error(err.message);
        return false;
      }
      return true;
    }
    return false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
