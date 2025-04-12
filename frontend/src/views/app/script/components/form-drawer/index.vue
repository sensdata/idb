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
      <a-form ref="formRef" :model="formState" :rules="rules">
        <a-form-item field="name" :label="$t('app.script.form.name.label')">
          <a-input
            v-model="formState.name"
            class="w-[368px]"
            :placeholder="$t('app.script.form.name.placeholder')"
          />
        </a-form-item>
        <a-form-item
          v-if="!isEdit"
          field="type"
          :label="$t('app.script.form.type.label')"
        >
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
        <!-- <a-form-item field="mark" :label="$t('app.script.form.mark.label')">
          <a-textarea
            v-model="formState.mark"
            :placeholder="$t('app.script.form.mark.placeholder')"
            :auto-size="{ minRows: 5 }"
          />
        </a-form-item> -->
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import {
    createScriptApi,
    getScriptCategoryListApi,
    getScriptDetailApi,
    updateScriptApi,
  } from '@/api/script';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { RadioOption } from '@arco-design/web-vue/es/radio/interface';

  const props = defineProps<{
    type: SCRIPT_TYPE;
  }>();

  const emit = defineEmits(['ok']);

  const { t } = useI18n();

  const formRef = ref();
  const formState = reactive({
    name: '',
    type: props.type,
    category: undefined as string | undefined,
    content: '',
    // mark: '',
  });

  const rules = {};

  const typeOptions = ref<RadioOption[]>([
    {
      label: t('app.script.enum.type.local'),
      value: SCRIPT_TYPE.Local,
    },
    {
      label: t('app.script.enum.type.global'),
      value: SCRIPT_TYPE.Global,
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
          label: cat.name,
          value: cat.name,
        })),
      ];
    } catch (err: any) {
      Message.error(err?.message);
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

  const paramsRef = ref<{ name: string; category: string }>();
  const isEdit = computed(() => !!paramsRef.value?.name);
  const setParams = (params: { name: string; category: string }) => {
    paramsRef.value = params;
  };
  const loading = ref(false);
  async function load() {
    loading.value = true;
    try {
      const data = await getScriptDetailApi({
        name: paramsRef.value!.name,
        category: paramsRef.value!.category,
        type: props.type,
      });
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
      if (isEdit.value) {
        await updateScriptApi({
          type: props.type,
          ...paramsRef.value!,
          new_name: formState.name,
          new_category: formState.category!,
          content: formState.content,
        });
      } else {
        await createScriptApi({
          name: formState.name,
          type: formState.type,
          category: formState.category,
          content: formState.content,
        });
      }
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
