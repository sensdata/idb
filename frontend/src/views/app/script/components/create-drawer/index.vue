<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="$t('app.script.createDrawer.title')"
    unmountOnClose
    :ok-loading="loading"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-form :model="formState" :rules="rules">
      <a-form-item
        field="name"
        :label="$t('app.script.createDrawer.name.label')"
      >
        <a-input
          v-model="formState.name"
          class="w-[368px]"
          :placeholder="$t('app.script.createDrawer.name.placeholder')"
        />
      </a-form-item>
      <a-form-item
        field="type"
        :label="$t('app.script.createDrawer.type.label')"
      >
        <a-radio-group v-model="formState.type" :options="typeOptions" />
      </a-form-item>
      <a-form-item
        field="category"
        :label="$t('app.script.createDrawer.category.label')"
      >
        <a-select
          v-model="formState.category"
          class="w-[368px]"
          :placeholder="$t('app.script.createDrawer.category.placeholder')"
          :loading="categoryLoading"
          :options="categoryOptions"
          allow-clear
          allow-create
        />
      </a-form-item>
      <a-form-item
        field="content"
        :label="$t('app.script.createDrawer.content.label')"
      >
        <a-textarea
          v-model="formState.content"
          :placeholder="$t('app.script.createDrawer.content.placeholder')"
          :auto-size="{ minRows: 10 }"
        />
      </a-form-item>
      <a-form-item
        field="mark"
        :label="$t('app.script.createDrawer.mark.label')"
      >
        <a-textarea
          v-model="formState.mark"
          :placeholder="$t('app.script.createDrawer.mark.placeholder')"
          :auto-size="{ minRows: 5 }"
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message, SelectOption } from '@arco-design/web-vue';
  import useLoading from '@/hooks/loading';
  import { createScriptApi, getScriptCategoryListApi } from '@/api/script';
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

  const visible = ref(false);
  const { loading, setLoading } = useLoading(false);
  const handleOk = async () => {
    setLoading(true);
    try {
      await createScriptApi({
        name: formState.name,
        type: formState.type,
        category: formState.category,
        content: formState.content,
        mark: formState.mark,
      });
      visible.value = false;
      Message.success(t('app.script.createDrawer.success'));
      emit('ok');
    } finally {
      setLoading(false);
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
    show,
    hide,
  });
</script>
