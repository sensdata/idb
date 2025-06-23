<template>
  <a-drawer
    :width="720"
    :visible="visible"
    :title="$t('app.docker.compose.edit.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-radio-group v-model="formState.active" type="button">
      <a-radio value="compose_content">
        {{ $t('app.docker.compose.edit.compose_content') }}
      </a-radio>
      <a-radio value="env_content">{{
        $t('app.docker.compose.edit.env_content')
      }}</a-radio>
    </a-radio-group>
    <codemirror
      v-if="formState.active === 'compose_content'"
      v-model="formState.compose_content"
      theme="cobalt"
      :tabSize="4"
      :extensions="yamlExtensions"
      autofocus
      indent-with-tab
      line-wrapping
      match-brackets
      style-active-line
    />
    <codemirror
      v-else
      v-model="formState.env_content"
      theme="cobalt"
      :tabSize="4"
      :extensions="propExtension"
      autofocus
      indent-with-tab
      line-wrapping
      match-brackets
      style-active-line
    />
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { Message } from '@arco-design/web-vue';
  import { Codemirror } from 'vue-codemirror';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import { properties } from '@codemirror/legacy-modes/mode/properties';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { getComposeDetailApi, updateComposeApi } from '@/api/docker';

  const { t } = useI18n();

  const emit = defineEmits(['ok']);

  const yamlExtensions = [StreamLanguage.define(yaml), oneDark];
  const propExtension = [StreamLanguage.define(properties), oneDark];

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

  const params = reactive<{
    name?: string;
  }>({});
  const setParams = (newParams: any) => {
    Object.assign(params, newParams);
  };

  const formState = reactive({
    active: 'compose_content',
    env_content: '',
    compose_content: '',
  });
  const load = async () => {
    setLoading(true);
    try {
      const data = await getComposeDetailApi({
        name: params.name!,
      });
      formState.env_content = data.env_content;
      formState.compose_content = data.compose_content;
    } catch (err: any) {
      Message.error(err?.message);
    } finally {
      setLoading(false);
    }
  };

  const getData = () => {
    return {
      name: params.name,
      compose_content: formState.compose_content,
      env_content: formState.env_content,
    };
  };

  const validate = () => {
    const data = getData();
    if (!data.compose_content) {
      Message.error(t('app.docker.compose.edit.compose_content_required'));
      return false;
    }

    return true;
  };

  const handleOk = async () => {
    try {
      if (!validate()) {
        return;
      }
      setLoading(true);
      const data = getData();
      await updateComposeApi(data);
      Message.success(t('app.docker.compose.edit.success'));
      emit('ok');
      hide();
    } catch (err: any) {
      Message.error(err?.message);
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
