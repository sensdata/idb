<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.setting.mirror.title')"
    destroy-on-close
    @close="handleClose"
    @before-ok="handleBeforeOk"
  >
    <a-spin :loading="loading" class="w-full">
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="registry_mirrors"
              :label="t('app.docker.setting.mirror.mirrors')"
            >
              <a-textarea
                v-model="form.registry_mirrors"
                :placeholder="
                  t('app.docker.setting.mirror.mirrors.placeholder')
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
    registry_mirrors: '',
  });

  const rules = reactive({
    registry_mirrors: [
      {
        required: true,
        message: t('app.docker.setting.mirror.mirrors.required'),
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
          const urlReg = /^(https?:\/\/|docker:\/\/)[\w\-.:/]+$/i;
          for (const url of urls) {
            if (!urlReg.test(url)) {
              callback(t('app.docker.setting.mirror.mirrors.format'));
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
    if (Array.isArray(params.mirrors)) {
      form.registry_mirrors = params.registry_mirrors.join('\n');
    } else {
      form.registry_mirrors = params.registry_mirrors || '';
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
    const mirrors = form.registry_mirrors
      .split('\n')
      .map((v) => v.trim())
      .filter(Boolean);
    if (
      await confirm({
        title: t('app.docker.setting.mirror.confirm.title'),
        content: t('app.docker.setting.mirror.confirm.content'),
      })
    ) {
      loading.value = true;
      try {
        await updateDockerConfApi({
          key: 'registry_mirrors',
          value: mirrors,
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
