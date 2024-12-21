<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="
      isEdit
        ? $t('app.crontab.form.title.edit')
        : $t('app.crontab.form.title.add')
    "
    unmountOnClose
    :ok-loading="submitLoading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form ref="formRef" :model="formState" :rules="rules">
        <a-form-item field="name" :label="$t('app.crontab.form.name.label')">
          <a-input
            v-model="formState.name"
            class="w-[368px]"
            :placeholder="$t('app.crontab.form.name.placeholder')"
          />
        </a-form-item>
        <a-form-item field="type" :label="$t('app.crontab.form.type.label')">
          <a-radio-group v-model="formState.type" :options="typeOptions" />
        </a-form-item>
        <a-form-item
          field="period"
          :label="$t('app.crontab.form.period.label')"
        >
          <period-input v-model="formState.period_details" />
        </a-form-item>
        <a-form-item
          field="content"
          :label="$t('app.crontab.form.content.label')"
        >
          <a-textarea
            v-model="formState.content"
            :placeholder="$t('app.crontab.form.content.placeholder')"
            :auto-size="{ minRows: 10 }"
            width="100%"
          />
        </a-form-item>
        <a-form-item field="mark" :label="$t('app.crontab.form.mark.label')">
          <a-textarea
            v-model="formState.mark"
            :placeholder="$t('app.crontab.form.mark.placeholder')"
            :auto-size="{ minRows: 5 }"
          />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, toRaw } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { createCrontabApi, getCrontabDetailApi } from '@/api/crontab';
  import { CRONTAB_KIND, CRONTAB_TYPE } from '@/config/enum';
  import { RadioOption } from '@arco-design/web-vue/es/radio/interface';
  import PeriodInput from '../period-input/index.vue';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const formRef = ref();
  const formState = reactive({
    name: '',
    type: CRONTAB_TYPE.Local,
    kind: CRONTAB_KIND.Shell,
    content: '',
    period_details: [],
    mark: '',
  });

  const rules = {
    name: [{ required: true, message: t('app.crontab.form.name.required') }],
    type: [{ required: true, message: t('app.crontab.form.type.required') }],
    period: [
      {
        required: true,
        message: t('app.crontab.form.period.required'),
        type: 'array' as const,
      },
    ],
    content: [
      { required: true, message: t('app.crontab.form.content.required') },
    ],
  };

  const typeOptions = ref<RadioOption[]>([
    {
      label: t('app.crontab.enum.type.local'),
      value: CRONTAB_TYPE.Local,
    },
    {
      label: t('app.crontab.enum.type.global'),
      value: CRONTAB_TYPE.Global,
    },
  ]);

  const paramsRef = ref<{ id: number }>();
  const isEdit = computed(() => !!paramsRef.value?.id);
  const setParams = (params: { id: number }) => {
    paramsRef.value = params;
  };
  const loading = ref(false);
  async function load() {
    loading.value = true;
    try {
      const data = await getCrontabDetailApi(toRaw(paramsRef.value!));
      Object.assign(formState, data);
    } finally {
      loading.value = false;
    }
  }

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const visible = ref(false);
  const submitLoading = ref(false);
  const handleOk = async () => {
    if (!(await validate())) {
      return;
    }

    if (submitLoading.value) {
      return;
    }

    submitLoading.value = true;
    try {
      await createCrontabApi({
        name: formState.name,
        type: formState.type,
        content: formState.content,
        mark: formState.mark,
      });
      visible.value = false;
      Message.success(t('app.crontab.form.success'));
      emit('ok');
    } finally {
      submitLoading.value = false;
    }
  };
  const handleCancel = () => {
    visible.value = false;
  };

  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  defineExpose({
    setParams,
    load,
    show,
    hide,
  });
</script>
