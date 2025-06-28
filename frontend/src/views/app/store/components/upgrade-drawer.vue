<template>
  <a-drawer
    :width="720"
    :visible="visible"
    :title="$t('app.store.app.upgrade.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        :label="$t('app.store.app.install.version')"
        field="version_id"
        :rules="[
          {
            required: true,
            message: $t('app.store.app.install.version.required'),
          },
        ]"
      >
        <a-select
          v-model="formState.version_id"
          :options="versionOptions"
          :placeholder="$t('app.store.app.install.version.placeholder')"
        />
      </a-form-item>
      <a-form-item field="mode">
        <a-radio-group v-model="formState.mode" type="button">
          <a-radio value="form">
            {{ $t('app.store.app.install.mode.form') }}
          </a-radio>
          <a-radio value="yaml">{{
            $t('app.store.app.install.mode.yaml')
          }}</a-radio>
        </a-radio-group>
      </a-form-item>
      <template v-if="formState.mode === 'form'">
        <template v-for="field in dynamicFields" :key="field.Name">
          <a-form-item
            :label="field.Label"
            :field="field.Name"
            :required="field.Required"
            :tooltip="field.Hint"
            :rules="getFieldRules(field)"
          >
            <component
              :is="getFieldComponent(field)"
              v-model="formState[field.Name]"
              v-bind="getFieldProps(field)"
            />
          </a-form-item>
        </template>
      </template>
      <template v-if="formState.mode === 'yaml'">
        <codemirror
          v-model="composeContent"
          theme="cobalt"
          :style="{ width: '100%', height: '400px' }"
          :tabSize="4"
          :extensions="extensions"
          autofocus
          indent-with-tab
          line-wrapping
          match-brackets
          style-active-line
        />
      </template>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { Message } from '@arco-design/web-vue';
  import { getAppDetailApi, upgradeAppApi } from '@/api/store';
  import type { AppEntity, AppFormField } from '@/entity/App';
  import { Codemirror } from 'vue-codemirror';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { useConfirm } from '@/composables/confirm';

  const { t } = useI18n();

  const emit = defineEmits(['ok']);

  const extensions = [StreamLanguage.define(yaml), oneDark];

  const formRef = ref();
  const formState = reactive<Record<string, any>>({
    mode: 'form',
    version_id: undefined,
  });
  const rules = reactive<Record<string, any>>({});

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);

  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };
  const handleCancel = () => {
    visible.value = false;
  };

  const versionOptions = ref<{ label: string; value: number }[]>([]);
  const version = computed(() => {
    return versionOptions.value.find((v) => v.value === formState.version_id)
      ?.label;
  });
  const appDetail = ref<AppEntity | null>(null);
  const dynamicFields = ref<AppFormField[]>([]);
  const composeContent = ref<string>('');
  watch(
    () => formState.version_id,
    () => {
      if (!appDetail.value) {
        return;
      }
      const v = appDetail.value.versions.find(
        (ver) => ver.id === formState.version_id
      );
      composeContent.value = v?.compose_content || '';
    }
  );

  const params = reactive<{
    id?: number;
  }>({});
  const setParams = (newParams: any) => {
    Object.assign(params, newParams);
  };

  const getFieldComponent = (field: AppFormField) => {
    switch (field.Type) {
      case 'select':
        return 'a-select';
      case 'textarea':
        return 'a-textarea';
      default:
        return 'a-input';
    }
  };

  const getFieldProps = (field: AppFormField) => {
    if (field.Type === 'select') {
      return {
        options:
          field.Options?.map((opt) => ({ label: opt, value: opt })) || [],
        placeholder: field.Hint || '',
      };
    }
    return {
      placeholder: field.Hint || '',
      type: field.Type === 'password' ? 'password' : 'text',
    };
  };

  const getFieldRules = (field: AppFormField) => {
    const fieldRules: any[] = [];
    if (field.Required) {
      fieldRules.push({
        required: true,
        message: t('app.store.app.install.required', { label: field.Label }),
      });
    }
    if (field.Validation) {
      if (field.Validation.MinLength) {
        fieldRules.push({
          min: field.Validation.MinLength,
          message: t('app.store.app.install.minLength', {
            min: field.Validation.MinLength,
          }),
        });
      }
      if (field.Validation.MaxLength) {
        fieldRules.push({
          max: field.Validation.MaxLength,
          message: t('app.store.app.install.maxLength', {
            max: field.Validation.MaxLength,
          }),
        });
      }
      if (field.Validation.Pattern) {
        fieldRules.push({
          pattern: new RegExp(field.Validation.Pattern),
          message: t('app.store.app.install.pattern'),
        });
      }
    }
    return fieldRules;
  };

  const load = async () => {
    setLoading(true);
    try {
      const data = await getAppDetailApi({ id: params.id! });
      appDetail.value = data;
      formState.display_name = data.display_name;
      formState.version_id = data.versions[0]?.id;
      versionOptions.value = data.versions.map((v) => ({
        label: v.version,
        value: v.id,
      }));
      dynamicFields.value = data.form?.Fields || [];
      dynamicFields.value.forEach((field) => {
        formState[field.Name] = field.Default || '';
        rules[field.Name] = getFieldRules(field);
      });
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      setLoading(false);
    }
  };

  const getData = () => {
    return {
      id: appDetail.value!.id,
      version_id: formState.version_id,
      extra_params: [],
      compose_content: formState.mode === 'yaml' ? composeContent.value : '',
      form_params:
        formState.mode === 'yaml'
          ? []
          : dynamicFields.value.map((field) => ({
              key: field.Name,
              value: formState[field.Name],
            })),
    };
  };

  const validate = async () => {
    return formRef.value?.validate().then((errors: any) => {
      return !errors;
    });
  };

  const { confirm } = useConfirm();
  const handleOk = async () => {
    if (!(await validate())) {
      return;
    }

    if (
      await confirm(
        t('app.store.app.upgrade.confirm', { version: version.value })
      )
    ) {
      try {
        const data = getData();
        setLoading(true);
        await upgradeAppApi(data);
        Message.success(t('app.store.app.upgrade.success'));
        emit('ok');
        hide();
      } catch (err: any) {
        Message.error(err?.message);
      } finally {
        setLoading(false);
      }
    }
  };

  defineExpose({
    show,
    hide,
    setParams,
    load,
  });
</script>
