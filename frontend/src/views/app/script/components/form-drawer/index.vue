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
        layout="vertical"
      >
        <div class="form-section">
          <a-form-item field="name" :label="$t('app.script.form.name.label')">
            <a-input
              v-model="formState.name"
              class="form-input"
              :placeholder="$t('app.script.form.name.placeholder')"
            />
          </a-form-item>

          <a-form-item
            v-if="!isEdit"
            field="type"
            :label="$t('app.script.form.type.label')"
            class="form-item-type"
          >
            <a-radio-group
              v-model="formState.type"
              :options="typeOptions"
              class="type-radio-group"
            />
          </a-form-item>

          <a-form-item
            field="category"
            :label="$t('app.script.form.category.label')"
          >
            <a-select
              v-model="formState.category"
              class="form-input"
              :placeholder="$t('app.script.form.category.placeholder')"
              :loading="categoryLoading"
              :options="categoryOptions"
              allow-clear
              allow-create
            />
          </a-form-item>
        </div>

        <div class="content-section">
          <a-form-item
            field="content"
            :label="$t('app.script.form.content.label')"
            class="content-form-item"
          >
            <div class="editor-container">
              <CodeEditor
                v-model="formState.content"
                :file="editorFile"
                :extensions="lightThemeExtensions"
                :autofocus="true"
                :indent-with-tab="true"
                :tab-size="2"
                @editor-ready="handleEditorReady"
              />
            </div>
          </a-form-item>
        </div>
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
  import CodeEditor from '@/components/code-editor/index.vue';
  import { githubLight } from '@fsegurai/codemirror-theme-github-light';
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
    const username = userStore.name || 'admin';

    formState.name = '';
    // 添加默认的shell脚本头部，格式与图片中保持一致
    formState.content = `#!/bin/bash
#Description:
#Author:${username}
`;
    formState.category = undefined;
    // 保留type，因为这是根据props传递的，不需要重置
    formState.type = props.type;
  };

  // 为 CodeEditor 组件创建虚拟文件对象，用于语法高亮
  const editorFile = computed(() => ({
    name: 'script.sh',
    path: '/tmp/script.sh',
  }));

  // 添加浅色主题扩展，提供语法高亮颜色
  const lightThemeExtensions = computed(() => [githubLight]);

  const handleEditorReady = (payload: { view: any }) => {
    // 编辑器准备完成的回调，可以在这里做一些初始化操作
    console.log('Editor ready:', payload);
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
    padding: 0 4px;
  }

  /* 表单分区 */
  .form-section {
    margin-bottom: 0;
  }

  .content-section {
    margin-top: 0;
  }

  /* 表单项样式 */
  .script-form :deep(.arco-form-item) {
    margin-bottom: 20px;
  }

  .script-form :deep(.arco-form-item-label) {
    padding: 0;
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text-1);
  }

  /* 必填标记样式 - 使用ArcoDesign内置样式 */
  .script-form :deep(.arco-form-item-label-required-symbol) {
    margin-right: 4px;
    font-weight: bold;
    color: #f53f3f;
  }

  /* 统一输入框样式 */
  .form-input {
    width: 100%;
  }

  .form-input :deep(.arco-input),
  .form-input :deep(.arco-select-view) {
    font-size: 14px;
    border-radius: 6px;
    transition: all 0.2s;
  }

  .form-input :deep(.arco-input:focus),
  .form-input :deep(.arco-select-view-focus) {
    border-color: var(--color-primary);
  }

  .form-input :deep(.arco-input:hover),
  .form-input :deep(.arco-select-view:hover) {
    border-color: var(--color-primary);
  }

  /* Type单选按钮组样式 */
  .form-item-type {
    margin-bottom: 20px;
  }

  .type-radio-group :deep(.arco-radio) {
    margin-right: 24px;
  }

  .type-radio-group :deep(.arco-radio-label) {
    padding-left: 8px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .type-radio-group :deep(.arco-radio-button) {
    border-radius: 6px;
  }

  /* 内容编辑器样式 */
  .content-form-item {
    margin-bottom: 0;
  }

  .content-form-item :deep(.arco-form-item-label) {
    margin-bottom: 12px;
  }

  .editor-container {
    position: relative;
    display: flex;
    width: 100%;
    height: 450px;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    transition: border-color 0.2s ease;
  }

  /* 编辑器聚焦时的紫色边框 */
  .editor-container :deep(.cm-focused) {
    outline: none;
  }

  .editor-container:has(.cm-focused) {
    border-color: #722ed1;
  }

  /* 兼容性回退 - 如果浏览器不支持:has()选择器 */
  .editor-container :deep(.cm-editor.cm-focused) {
    border-color: #722ed1;
  }

  .editor-container :deep(.cm-editor) {
    border-radius: 6px;
  }

  /* 选择器下拉框样式优化 */
  .form-input :deep(.arco-select-option) {
    padding: 8px 12px;
    font-size: 14px;
  }

  .form-input :deep(.arco-select-option:hover) {
    background-color: var(--color-fill-2);
  }

  .form-input :deep(.arco-select-option-selected) {
    color: var(--color-primary);
    background-color: var(--color-primary-light-1);
  }

  /* 加载状态样式 */
  .form-input :deep(.arco-select-loading) {
    color: var(--color-text-3);
  }

  /* 响应式布局 */
  @media (width <= 768px) {
    .editor-container {
      height: 350px;
    }
  }
</style>
