<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.setting.log.title')"
    destroy-on-close
    @close="handleClose"
    @before-ok="handleBeforeOk"
  >
    <a-spin :loading="loading" class="w-full">
      <a-alert class="mb-4" type="info" :show-icon="true">
        <div class="guide-title">
          {{ t('app.docker.setting.log.guide.title') }}
        </div>
        <div class="guide-desc">
          {{ t('app.docker.setting.log.guide.desc') }}
        </div>
      </a-alert>
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="log_max_size"
              :label="t('app.docker.setting.log.max_size')"
            >
              <template #extra>
                <span class="field-help">
                  <icon-question-circle />
                  {{ t('app.docker.setting.log.max_size.help') }}
                </span>
              </template>
              <a-input
                v-model="form.log_max_size"
                :placeholder="t('app.docker.setting.log.max_size.placeholder')"
              />
            </a-form-item>
            <a-form-item
              field="log_max_file"
              :label="t('app.docker.setting.log.max_file')"
            >
              <template #extra>
                <span class="field-help">
                  <icon-question-circle />
                  {{ t('app.docker.setting.log.max_file.help') }}
                </span>
              </template>
              <a-input
                v-model="form.log_max_file"
                :placeholder="t('app.docker.setting.log.max_file.placeholder')"
              />
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, toRaw } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { updateLogOptionApi } from '@/api/docker';
  import { useConfirm } from '@/composables/confirm';

  const { t } = useI18n();
  const emit = defineEmits(['ok']);

  const loading = ref(false);
  const visible = ref(false);
  const formRef = ref();

  const form = reactive({
    log_max_size: '',
    log_max_file: '',
  });

  const rules = reactive({
    log_max_size: [
      {
        required: true,
        message: t('app.docker.setting.log.max_size.required'),
        trigger: 'blur',
      },
      {
        validator(value: string, callback: any) {
          // 支持如 10m/100m/1g/100k 等格式
          if (!/^\d+(k|m|g|B|KB|MB|GB)$/i.test(value)) {
            callback(t('app.docker.setting.log.max_size.format'));
            return;
          }
          callback();
        },
        trigger: 'blur',
      },
    ],
    log_max_file: [
      {
        required: true,
        message: t('app.docker.setting.log.max_file.required'),
        trigger: 'blur',
      },
      {
        validator(value: string, callback: any) {
          if (!/^\d+$/.test(value)) {
            callback(t('app.docker.setting.log.max_file.format'));
            return;
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
    form.log_max_size = params.log_max_size || '';
    form.log_max_file = params.log_max_file || '';
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
    const params = toRaw(form);
    if (
      await confirm({
        title: t('app.docker.setting.log.confirm.title'),
        content: t('app.docker.setting.log.confirm.content'),
      })
    ) {
      loading.value = true;
      try {
        await updateLogOptionApi(params);
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
