<template>
  <div class="service-raw">
    <div class="raw-content">
      <codemirror
        :model-value="content"
        :placeholder="$t('app.service.form.field.content')"
        :autofocus="true"
        :indent-with-tab="true"
        :tab-size="2"
        :extensions="serviceExtensions"
        class="service-codemirror"
        @update:model-value="handleContentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { Codemirror } from 'vue-codemirror';
  import useEditorConfig from '@/components/code-editor/composables/use-editor-config';
  import { useLogger } from '@/composables/use-logger';
  import { SERVICE_TYPE } from '@/config/enum';
  import { ServiceEntity } from '@/entity/Service';

  const props = defineProps<{
    type: SERVICE_TYPE;
    category: string;
    isEdit: boolean;
    record?: ServiceEntity;
  }>();

  const emit = defineEmits<{
    'categoryChange': [category: string];
    'update:content': [content: string];
    'change': [];
  }>();

  const content = defineModel<string>('content', { default: '' });

  // 日志记录
  const { logDebug } = useLogger('ServiceRaw');

  // 使用编辑器配置获取service扩展
  const { getServiceExtensions } = useEditorConfig(ref(null));
  const serviceExtensions = getServiceExtensions();

  // 处理内容变化
  const handleContentChange = (value: string) => {
    logDebug('内容变化，长度:', value.length);

    content.value = value;
    // 每次内容变化时通知父组件
    emit('change');
  };

  // 验证表单
  const validate = async () => {
    try {
      if (!content.value.trim()) {
        throw new Error('Content is required');
      }
      return true;
    } catch {
      return false;
    }
  };

  // 获取提交数据
  const getSubmitData = () => {
    return {
      type: props.type,
      category: props.category, // 使用从props传入的category
      name: props.record?.name || '', // 使用从props传入的记录名称
      content: content.value,
    };
  };

  defineExpose({
    validate,
    getSubmitData,
  });
</script>

<style scoped>
  .service-raw {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 400px;
  }

  .raw-content {
    flex: 1;
    min-height: 400px;
    overflow: auto;
  }

  .service-codemirror {
    height: 100%;
    min-height: 400px;
  }

  /* Make the editor take full available height */
  .service-codemirror :deep(.cm-editor) {
    height: 100%;
    min-height: 400px;
    overflow: auto;
  }

  /* Ensure the content area is scrollable and has proper height */
  .service-codemirror :deep(.cm-scroller) {
    min-height: 400px;
    overflow: auto;
  }

  .service-codemirror :deep(.cm-content) {
    min-height: 400px;
    padding: 12px;
  }

  .service-codemirror :deep(.cm-focused) {
    outline: none;
  }
</style>
