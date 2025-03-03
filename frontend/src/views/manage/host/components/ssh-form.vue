<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.ssh.form.title')"
    width="540px"
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="model" :rules="rules">
      <a-form-item field="addr" :label="$t('manage.host.form.addr.label')">
        <a-input
          v-model="model.addr"
          :placeholder="$t('manage.host.form.addr.placeholder')"
        />
      </a-form-item>
      <a-form-item field="port" :label="$t('manage.host.form.port.label')">
        <a-input-number
          v-model="model.port"
          :placeholder="$t('manage.host.form.port.placeholder')"
        />
      </a-form-item>
      <a-form-item field="user" :label="$t('manage.host.form.user.label')">
        <a-input
          v-model="model.user"
          :placeholder="$t('manage.host.form.user.placeholder')"
          autocomplete="off"
        />
      </a-form-item>
      <a-form-item
        field="auth_mode"
        :label="$t('manage.host.form.auth_mode.label')"
      >
        <a-radio-group v-model="model.auth_mode" :options="authModeOptions" />
      </a-form-item>
      <a-form-item
        v-if="model.auth_mode === AUTH_MODE.Password"
        field="password"
        :label="$t('manage.host.form.password.label')"
      >
        <a-input
          v-model="model.password"
          type="password"
          :placeholder="$t('manage.host.form.password.placeholder')"
          autocomplete="off"
        />
      </a-form-item>
      <template v-else>
        <a-form-item
          field="private_key"
          :label="$t('manage.host.form.private_key.label')"
        >
          <file-selector
            ref="fileSelectorRef"
            v-model="model.private_key"
            type="file"
            :placeholder="$t('manage.host.form.private_key.placeholder')"
          />
        </a-form-item>
        <a-form-item
          field="pass_phrase"
          :label="$t('manage.host.form.pass_phrase.label')"
        >
          <a-input
            v-model="model.pass_phrase"
            type="password"
            :placeholder="$t('manage.host.form.pass_phrase.placeholder')"
            autocomplete="off"
          />
        </a-form-item>
      </template>
      <a-form-item :label="$t('manage.host.form.test.label')">
        <div class="flex items-center">
          <a-button
            type="text"
            size="mini"
            :loading="testLoading"
            @click="handleTestConnection"
          >
            {{ $t('manage.host.form.test.button') }}
          </a-button>
          <span
            v-if="testResult"
            :class="[
              'ml-2',
              testResult.success ? 'text-green-600' : 'text-red-600',
            ]"
          >
            {{ testResult.message }}
          </span>
        </div>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { AUTH_MODE } from '@/config/enum';
  import { updateHostSSHApi, testHostSSHApi, getHostInfoApi } from '@/api/host';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import FileSelector from '@/components/file/file-selector/index.vue';

  interface TestResult {
    success: boolean;
    message: string;
  }

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();
  const { loading: testLoading, setLoading: setTestLoading } = useLoading();

  const formRef = ref();
  const fileSelectorRef = ref<InstanceType<typeof FileSelector>>();
  const testResult = ref<TestResult | null>(null);

  const model = reactive({
    id: 0,
    addr: '',
    port: 22,
    user: '',
    auth_mode: AUTH_MODE.Password,
    password: '',
    private_key: '',
    pass_phrase: '',
  });

  const rules = computed(() => ({
    addr: [{ required: true, message: t('manage.host.form.addr.required') }],
    port: [{ required: true, message: t('manage.host.form.port.required') }],
    user: [{ required: true, message: t('manage.host.form.user.required') }],
    auth_mode: [
      { required: true, message: t('manage.host.form.auth_mode.required') },
    ],
    password: [
      {
        required: model.auth_mode === AUTH_MODE.Password,
        message: t('manage.host.form.password.required'),
      },
    ],
    private_key: [
      {
        required: model.auth_mode !== AUTH_MODE.Password,
        message: t('manage.host.form.private_key.required'),
      },
    ],
  }));

  const authModeOptions = ref([
    {
      value: AUTH_MODE.Password,
      label: t('manage.host.enum.auth_mode.password') as string,
    },
    {
      value: AUTH_MODE.PrivateKey,
      label: t('manage.host.enum.auth_mode.private_key') as string,
    },
  ]);

  const getData = () => {
    const data = toRaw(model);
    return data;
  };

  const reset = () => {
    formRef.value?.resetFields();
    formRef.value?.clearValidate();
    testResult.value = null;
  };

  const validate = () => {
    return formRef.value.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleTestConnection = async () => {
    try {
      if (!(await validate())) {
        return;
      }

      setTestLoading(true);
      testResult.value = null;

      const data = getData();
      const result = await testHostSSHApi(data);

      testResult.value = {
        success: result.success,
        message: result.success
          ? t('manage.host.form.test.success')
          : t('manage.host.form.test.failed', { message: result.message }),
      };
    } catch (err: any) {
      testResult.value = {
        success: false,
        message: t('manage.host.form.test.error', { message: err.message }),
      };
    } finally {
      setTestLoading(false);
    }
  };

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      showLoading();
      const data = getData();
      await updateHostSSHApi(model.id, data);
      Message.success(t('manage.host.form.save.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      hideLoading();
    }
  };

  const handleCancel = () => {
    visible.value = false;
  };

  const load = async (hostId: number) => {
    try {
      showLoading();
      const data = await getHostInfoApi(hostId);
      model.id = data.id;
      model.addr = data.addr || '';
      model.port = data.port || 22;
      model.user = data.user || '';
      model.auth_mode = (data.auth_mode as AUTH_MODE) || AUTH_MODE.Password;
      model.password = data.password || '';
      model.private_key = data.private_key || '';
      model.pass_phrase = data.pass_phrase || '';
    } catch (err: any) {
      Message.error(t('manage.host.form.load.failed'));
    } finally {
      hideLoading();
    }
  };

  defineExpose({
    show,
    hide,
    reset,
    load,
  });
</script>
