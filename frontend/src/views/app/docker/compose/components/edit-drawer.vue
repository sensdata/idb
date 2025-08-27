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
    <code-editor
      v-if="formState.active === 'compose_content'"
      v-model="formState.compose_content"
      :tab-size="4"
      :extensions="yamlExtensions"
      :autofocus="true"
      :indent-with-tab="true"
      :file="composeFile"
    />
    <code-editor
      v-else
      v-model="formState.env_content"
      :tab-size="4"
      :extensions="propExtension"
      :autofocus="true"
      :indent-with-tab="true"
      :file="envFile"
    />
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, computed } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { Message } from '@arco-design/web-vue';
  import { StreamLanguage } from '@codemirror/language';
  import { yaml } from '@codemirror/legacy-modes/mode/yaml';
  import { properties } from '@codemirror/legacy-modes/mode/properties';
  import CodeEditor from '@/components/code-editor/index.vue';
  import { getComposeDetailApi, updateComposeApi } from '@/api/docker';

  const { t } = useI18n();

  const emit = defineEmits(['ok']);

  // 创建文件对象
  const composeFile = computed(() => ({
    name: 'docker-compose.yml',
    path: '/tmp/docker-compose.yml',
  }));

  const envFile = computed(() => ({
    name: '.env',
    path: '/tmp/.env',
  }));

  const yamlExtensions = [StreamLanguage.define(yaml)];
  const propExtension = [StreamLanguage.define(properties)];

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
