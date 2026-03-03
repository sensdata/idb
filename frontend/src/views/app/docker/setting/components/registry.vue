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
      <a-alert class="mb-4" type="warning" :show-icon="true">
        <div class="guide-title">
          {{ t('app.docker.setting.registry.guide.title') }}
        </div>
        <div class="guide-desc">
          {{ t('app.docker.setting.registry.guide.desc') }}
        </div>
      </a-alert>
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="insecure_registries"
              :label="t('app.docker.setting.registry.registries')"
            >
              <template #extra>
                <span class="field-help">
                  <icon-question-circle />
                  {{ t('app.docker.setting.registry.registries.help') }}
                </span>
              </template>
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
      .split('\n')
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

<style scoped>
  .guide-title {
    margin-bottom: 0.25rem;
    font-weight: 600;
  }

  .guide-desc {
    font-size: 14px;
    line-height: 1.5;
  }

  .field-help {
    display: inline-flex;
    gap: 0.25rem;
    align-items: center;
    font-size: 13px;
    color: var(--color-text-3);
  }
</style>
