<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="
      isEdit
        ? $t('app.script.form.title.edit')
        : $t('app.script.form.title.add')
    "
    unmountOnClose
    :ok-loading="submitLoading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-form :model="formState" :rules="rules">
        <a-form-item field="name" :label="$t('app.script.form.name.label')">
          <a-input
            v-model="formState.name"
            class="w-[368px]"
            :placeholder="$t('app.script.form.name.placeholder')"
          />
        </a-form-item>
        <a-form-item field="type" :label="$t('app.script.form.type.label')">
          <a-radio-group v-model="formState.type" :options="typeOptions" />
        </a-form-item>
        <a-form-item
          field="category"
          :label="$t('app.script.form.category.label')"
        >
          <a-select
            v-model="formState.category"
            class="w-[368px]"
            :placeholder="$t('app.script.form.category.placeholder')"
            :loading="categoryLoading"
            :options="categoryOptions"
            allow-clear
            allow-create
          />
        </a-form-item>
        <a-form-item
          field="content"
          :label="$t('app.script.form.content.label')"
        >
          <a-textarea
            v-model="formState.content"
            :placeholder="$t('app.script.form.content.placeholder')"
            :auto-size="{ minRows: 10 }"
            width="100%"
          />
        </a-form-item>
        <a-form-item field="mark" :label="$t('app.script.form.mark.label')">
          <a-textarea
            v-model="formState.mark"
            :placeholder="$t('app.script.form.mark.placeholder')"
            :auto-size="{ minRows: 5 }"
          />
        </a-form-item>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, toRaw, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import {
    createScriptApi,
    getScriptCategoryListApi,
    getScriptDetailApi,
  } from '@/api/script';
  import { ScriptType } from '@/config/enum';
  import { RadioOption } from '@arco-design/web-vue/es/radio/interface';

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const formState = reactive({
    name: '',
    type: ScriptType.Local,
    category: undefined as string | undefined,
    content: '',
    mark: '',
  });

  const rules = {};

  const typeOptions = ref<RadioOption[]>([
    {
      label: t('app.script.enum.type.local'),
      value: ScriptType.Local,
    },
    {
      label: t('app.script.enum.type.global'),
      value: ScriptType.Global,
    },
  ]);
  const categoryLoading = ref(false);
  const categoryOptions = ref<SelectOption[]>([]);
  const loadGroupOptions = async () => {
    categoryLoading.value = true;
    try {
      const ret = await getScriptCategoryListApi({
        page: 1,
        page_size: 1000,
        type: formState.type,
      });
      categoryOptions.value = [
        ...ret.items.map((cat) => ({
          label: cat,
          value: cat,
        })),
      ];
    } catch (err: any) {
      Message.error(err);
    } finally {
      categoryLoading.value = false;
    }
  };
  watch(
    () => formState.type,
    () => {
      loadGroupOptions();
    }
  );

  const paramsRef = ref<{ id: number }>();
  const isEdit = computed(() => !!paramsRef.value?.id);
  const setParams = (params: { id: number }) => {
    paramsRef.value = params;
  };
  const loading = ref(false);
  async function load() {
    loading.value = true;
    try {
      const data = await getScriptDetailApi(toRaw(paramsRef.value!));
      Object.assign(formState, data);
    } finally {
      loading.value = false;
    }
  }

  const visible = ref(false);
  const submitLoading = ref(false);
  const handleOk = async () => {
    submitLoading.value = true;
    try {
      await createScriptApi({
        name: formState.name,
        type: formState.type,
        category: formState.category,
        content: formState.content,
        mark: formState.mark,
      });
      visible.value = false;
      Message.success(t('app.script.form.success'));
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
    loadGroupOptions();
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
