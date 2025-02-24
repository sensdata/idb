<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.form.title.create')"
    width="540px"
    :ok-loading="loading"
    @cancel="handleCancel"
    @before-ok="handleBeforeOk"
  >
    <a-form ref="formRef" :model="model" :rules="rules">
      <a-form-item field="name" :label="$t('manage.host.form.name.label')">
        <a-input
          v-model="model.name"
          :placeholder="$t('manage.host.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item field="group" :label="$t('manage.host.form.group.label')">
        <a-select
          v-model="model.group_id"
          :placeholder="$t('manage.host.form.group.placeholder')"
          :loading="groupLoading"
          :options="groupOptions"
          allow-clear
        >
          <template #footer>
            <a-link class="group-add" @click="handleAddGroup">
              <icon-plus class="icon-plus" />
              {{ $t('manage.host.form.group.add') }}
            </a-link>
          </template>
        </a-select>
      </a-form-item>
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
  <group-form ref="groupFormRef" @ok="handleGroupFormOk" />
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import { AUTH_MODE } from '@/config/enum';
  import {
    CreateHostParams,
    createHostApi,
    getHostGroupListApi,
    testHostSSHApi,
    testHostAgentApi,
    installHostAgentApi,
  } from '@/api/host';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { useConfirm } from '@/hooks/confirm';
  import FileSelector from '@/components/file/file-selector/index.vue';
  import GroupForm from './group-form.vue';

  interface TestResult {
    success: boolean;
    message: string;
  }

  const emit = defineEmits(['ok', 'success']);

  const { t } = useI18n();

  const { confirm } = useConfirm();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();
  const { loading: testLoading, setLoading: setTestLoading } = useLoading();

  const formRef = ref();
  const fileSelectorRef = ref<InstanceType<typeof FileSelector>>();
  const testResult = ref<TestResult | null>(null);
  const model = reactive({
    name: '',
    addr: '',
    port: 22,
    user: '',
    group_id: 0,
    auth_mode: AUTH_MODE.Password,
    password: '',
    private_key: '',
    pass_phrase: '',
  });

  watch(visible, (newVal) => {
    if (!newVal) {
      fileSelectorRef.value?.closePopover();
    }
  });

  const rules = computed(() => ({
    name: [{ required: true, message: t('manage.host.form.name.required') }],
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

  const groupOptions = ref<SelectOption[]>([]);
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
    formRef.value.resetFields();
    formRef.value.clearValidate();
    testResult.value = null;
  };

  const { loading: groupLoading, setLoading: setGroupLoading } = useLoading();
  const loadGroupOptions = async () => {
    setGroupLoading(true);
    try {
      const ret = await getHostGroupListApi({
        page: 1,
        page_size: 1000,
      });
      groupOptions.value = [
        {
          label: t('manage.host.form.group.default'),
          value: 0,
        },
        ...ret.items.map((item: any) => ({
          label: item.group_name,
          value: item.id,
        })),
      ];
    } catch (err: any) {
      Message.error(err);
    } finally {
      setGroupLoading(false);
    }
  };

  const loadOptions = async () => {
    await loadGroupOptions();
  };

  const groupFormRef = ref<InstanceType<typeof GroupForm>>();
  const handleAddGroup = () => {
    const form = groupFormRef.value;
    form?.reset();
    form?.show();
  };

  const handleGroupFormOk = () => {
    loadGroupOptions();
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
      const result = await testHostSSHApi(data as CreateHostParams);

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

  const installAgent = async (hostId: number) => {
    let loadingInstance: any = null;
    try {
      loadingInstance = Message.loading({
        content: t('manage.host.agent.installing'),
        duration: 0,
      });
      const result = await installHostAgentApi(hostId);
      loadingInstance.close();
      if (result.success) {
        Message.success(t('manage.host.agent.installSuccess'));
      } else {
        Message.error(t('manage.host.agent.installFailed'));
      }
    } catch (error) {
      if (loadingInstance) {
        loadingInstance.close();
      }
      Message.error(t('manage.host.agent.installFailed'));
      console.error('Failed to install agent:', error);
    }
  };

  const checkAgentInstall = async (hostId: number) => {
    try {
      const result = await testHostAgentApi(hostId);
      if (!result.installed) {
        const confirmResult = await confirm(
          t('manage.host.agent.notInstalled')
        );
        if (confirmResult) {
          await installAgent(hostId);
        }
      }
    } catch (error) {
      console.error('Failed to check agent:', error);
    }
    emit('success');
  };

  const handleBeforeOk = async (done: any) => {
    if (await validate()) {
      try {
        showLoading();
        const data = getData();
        const res = await createHostApi(data as CreateHostParams);
        await checkAgentInstall(res.id);
        done();
        Message.success(t('manage.host.form.save.success'));
        emit('ok');
        return true;
      } finally {
        hideLoading();
      }
    }
    return false;
  };

  const handleCancel = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    reset,
    loadOptions,
  });
</script>

<style scoped lang="less">
  .group-add {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    padding: 8px 12px;
    .icon-plus {
      margin-right: 4px;
    }
  }
</style>
