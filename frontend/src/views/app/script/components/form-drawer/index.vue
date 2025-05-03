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
      <a-form
        ref="formRef"
        :model="formState"
        :rules="rules"
        class="script-form"
      >
        <a-form-item
          field="name"
          :label="$t('app.script.form.name.label')"
          required
        >
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
          required
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
          :label-col-props="{ span: 24 }"
          :wrapper-col-props="{ span: 24 }"
        >
          <div class="editor-container">
            <codemirror
              v-model="formState.content"
              :autofocus="true"
              :indent-with-tab="true"
              :tab-size="2"
              :style="{ height: '100%', width: '100%' }"
              :placeholder="$t('app.script.form.content.placeholder')"
              :extensions="extensions"
              @ready="handleEditorReady"
            />
          </div>
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
  import { computed, reactive, ref, watch, shallowRef, nextTick } from 'vue';
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
  import { Codemirror } from 'vue-codemirror';
  import { EditorView, lineNumbers } from '@codemirror/view';
  import { EditorState } from '@codemirror/state';
  import { oneDark } from '@codemirror/theme-one-dark';
  import { StreamLanguage } from '@codemirror/language';
  import { shell } from '@codemirror/legacy-modes/mode/shell';
  import { autocompletion } from '@codemirror/autocomplete';
  import useUserStore from '@/store/modules/user';

  const props = defineProps<{
    type: SCRIPT_TYPE;
  }>();

  const emit = defineEmits(['ok', 'categoryChange']);

  const { t } = useI18n();

  const formRef = ref();
  const formState = reactive({
    name: '',
    type: props.type,
    category: undefined as string | undefined,
    content: '',
    // mark: '',
  });

  const rules = {
    name: [{ required: true, message: t('app.script.form.name.required') }],
    category: [
      { required: true, message: t('app.script.form.category.required') },
    ],
  };

  // 添加重置表单的方法
  const resetForm = () => {
    const userStore = useUserStore();
    const username = userStore.name || 'unknown';
    const currentDate = new Date().toISOString().split('T')[0];

    formState.name = '';
    // 添加默认的shell脚本头部，使用英文注释
    formState.content = `#!/bin/bash

# Description: 
# Author: ${username}
# Created: ${currentDate}
# Last modified: ${currentDate}

`;
    formState.category = undefined;
    // 保留type，因为这是根据props传递的，不需要重置
    formState.type = props.type;
  };

  // CodeMirror setup
  const editorView = shallowRef();
  const extensions = [
    StreamLanguage.define(shell),
    EditorView.lineWrapping,
    oneDark,
    lineNumbers(),
    // 禁用自动完成功能
    autocompletion({ override: [] }),
    // 使用自定义主题，确保编辑器适应容器
    EditorView.theme({
      '&': {
        height: '100%',
        width: '100%',
      },
      '.cm-content': {
        width: '100%',
      },
      '.cm-scroller': {
        width: '100%',
      },
      '.cm-line': {
        width: '100%',
        minWidth: '100%',
      },
    }),
    EditorState.lineSeparator.of('\n'),
  ];

  const handleEditorReady = (payload: { view: EditorView }) => {
    editorView.value = payload.view;
    nextTick(() => {
      if (editorView.value) {
        editorView.value.requestMeasure();
      }
    });
  };

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

  const paramsRef = ref<{ name: string; category: string }>();
  const isEdit = computed(() => !!paramsRef.value?.name);
  const setParams = (params: { name: string; category: string }) => {
    paramsRef.value = params;
  };
  const clearParams = () => {
    paramsRef.value = undefined;
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
        // 如果是新建脚本并指定了新分类，通知父组件刷新分类树
        if (formState.category) {
          emit('categoryChange', formState.category);
        }
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
    // 取消时也清除参数
    clearParams();
  };

  const show = () => {
    // 如果不是编辑模式，重置表单
    if (!paramsRef.value?.name) {
      resetForm();
    }
    visible.value = true;
    loadGroupOptions();
  };

  const hide = () => {
    visible.value = false;
    // 关闭抽屉时清除参数，确保下次打开时不会保留编辑状态
    clearParams();
  };

  // 添加监听visible变化
  watch(
    () => visible.value,
    (newVisible) => {
      // 当抽屉打开且不是编辑模式时，重置表单
      if (newVisible && !paramsRef.value?.name) {
        resetForm();
      }
    }
  );

  defineExpose({
    setParams,
    clearParams,
    load,
    show,
    hide,
  });
</script>

<style scoped>
  .script-form {
    width: 100%;
  }

  .script-form :deep(.arco-form-item-label-required-symbol) {
    margin-right: 4px;
    color: var(--color-danger);
  }

  .editor-container {
    position: relative;
    display: flex;
    width: 100%;
    height: 500px;
    overflow: hidden;
    background-color: #282c34;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  :deep(.cm-editor) {
    flex: 1;
    width: 100%;
    height: 100%;
  }

  :deep(.cm-scroller) {
    width: 100% !important;
    height: 100%;
    overflow: auto;
  }

  :deep(.cm-content) {
    width: 100%;
    min-height: 100%;
    padding: 4px 8px;
    font-size: 14px;
    font-family: monospace;
    line-height: 1.5;
  }

  :deep(.cm-line) {
    width: 100%;
    padding: 0 4px;
    white-space: pre;
  }

  :deep(.cm-gutters) {
    min-width: 40px;
    background-color: #282c34;
    border-right: 1px solid #3e4451;
  }
</style>
