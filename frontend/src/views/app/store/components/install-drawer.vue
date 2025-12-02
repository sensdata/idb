<template>
  <a-drawer
    :width="720"
    :visible="visible"
    :title="$t('app.store.app.install.title')"
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
        <a-tabs default-active-key="compose">
          <a-tab-pane key="compose" title="docker-compose.yml">
            <div style="width: 100%; height: 400px">
              <code-editor
                v-model="composeContent"
                :tab-size="4"
                :extensions="extensions"
                :autofocus="true"
                :indent-with-tab="true"
                :file="yamlFile"
              />
            </div>
          </a-tab-pane>
          <a-tab-pane key="env" title=".env">
            <div style="width: 100%; height: 400px">
              <code-editor
                v-model="envContent"
                :tab-size="4"
                :extensions="extensions"
                :indent-with-tab="true"
                :file="envFile"
              />
            </div>
          </a-tab-pane>
        </a-tabs>
      </template>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import useLoading from '@/composables/loading';
  import { Message, Modal } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { getAppDetailApi, installAppApi } from '@/api/store';
  import type { AppEntity, AppFormField } from '@/entity/App';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import CodeEditor from '@/components/code-editor/index.vue';

  const { t } = useI18n();
  const router = useRouter();

  const emit = defineEmits(['ok']);

  // 创建 YAML 文件对象
  const yamlFile = computed(() => ({
    name: 'docker-compose.yml',
    path: '/tmp/docker-compose.yml',
  }));

  // 创建 ENV 文件对象
  const envFile = computed(() => ({
    name: '.env',
    path: '/tmp/.env',
  }));

  const extensions = [StreamLanguage.define(yaml)];

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
  const appDetail = ref<AppEntity | null>(null);
  const dynamicFields = ref<AppFormField[]>([]);
  const composeContent = ref<string>('');
  const envContent = ref<string>('');
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
      envContent.value = v?.env_content || '';
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
        label: v.version + '.' + v.update_version,
        value: v.id,
      }));
      dynamicFields.value = data.form?.Fields || [];
      dynamicFields.value.forEach((field) => {
        formState[field.Name] = field.Default || '';
        rules[field.Name] = getFieldRules(field);
      });
    } catch (err: any) {
      await showErrorWithDockerCheck(err?.message, err);
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
      env_content: formState.mode === 'yaml' ? envContent.value : '',
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

  const handleOk = async () => {
    try {
      if (!(await validate())) {
        return;
      }
      setLoading(true);
      const data = getData();
      await installAppApi(data);

      // 显示安装成功的确认对话框
      Modal.confirm({
        title: t('app.store.app.install.success.confirm.title'),
        content: t('app.store.app.install.success.confirm.content'),
        okText: t('app.store.app.install.success.confirm.ok'),
        cancelText: t('app.store.app.install.success.confirm.cancel'),
        onOk: () => {
          // 跳转到 Compose 管理页面，方便后续编辑 compose 和 .env
          router.push('/app/docker/compose');
        },
        onCancel: () => {
          // 用户选择留在当前页面，显示成功消息
          Message.success(t('app.store.app.install.success'));
        },
      });

      emit('ok');
      hide();
    } catch (err: any) {
      await showErrorWithDockerCheck(err?.message, err);
    } finally {
      setLoading(false);
    }
  };

  defineExpose({
    show,
    hide,
    setParams,
    load,
  });
</script>
