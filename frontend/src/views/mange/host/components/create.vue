<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('manage.host.form.title.create')"
    width="600px"
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
            <a-link class="group-add">
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
        v-if="model.auth_mode === AuthModeEnum.Password"
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
          <a-textarea
            v-model="model.private_key"
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
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { toRaw, reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import { AuthModeEnum } from '@/config/enum';
  import { createHostApi, CreateHostParams } from '@/api/host';
  import { getGroupListApi } from '@/api/group';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref();
  const model = reactive({
    name: '',
    addr: '',
    port: 22,
    user: '',
    group_id: 0,
    auth_mode: AuthModeEnum.Password,
    password: '',
    private_key: '',
    pass_phrase: '',
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
        required: model.auth_mode === AuthModeEnum.Password,
        message: t('manage.host.form.password.required'),
      },
    ],
    private_key: [
      {
        required: model.auth_mode !== AuthModeEnum.Password,
        message: t('manage.host.form.private_key.required'),
      },
    ],
  }));

  const groupOptions = ref<SelectOption[]>([]);
  const authModeOptions = ref([
    {
      value: AuthModeEnum.Password,
      label: t('manage.host.enum.auth_mode.password') as string,
    },
    {
      value: AuthModeEnum.PrivateKey,
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
  };

  const { loading: groupLoading, setLoading: setGroupLoading } = useLoading();
  const loadGroupOptions = async () => {
    setGroupLoading(true);
    try {
      const ret = await getGroupListApi({
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
          value: item.group_name,
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

  const validate = () => {
    return formRef.value.validate().then((errors: any) => {
      return !errors;
    });
  };

  const handleBeforeOk = async (done: any) => {
    if (await validate()) {
      try {
        showLoading();
        const data = getData();
        await createHostApi(data as CreateHostParams);
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
