<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="$t('components.file.uploadFilesDrawer.title')"
    :footer="false"
    unmountOnClose
    @cancel="handleCancel"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :label-col-props="{ span: 6 }"
      :wrapper-col-props="{ span: 18 }"
      label-align="left"
    >
      <a-form-item
        field="directory"
        :label="$t('components.file.uploadFilesDrawer.directory')"
      >
        <file-selector
          v-model="formState.directory"
          type="directory"
          :host="props.host"
          :placeholder="
            $t('components.file.uploadFilesDrawer.directory.placeholder')
          "
        />
      </a-form-item>
      <!-- <a-form-item field="overwrite" label=" ">
        <a-checkbox v-model="formState.overwrite">
          {{ $t('components.file.uploadFilesDrawer.overwrite') }}
        </a-checkbox>
      </a-form-item> -->
      <a-form-item
        field="upload"
        :label="$t('components.file.uploadFilesDrawer.dragTip')"
      >
        <a-upload
          v-model:file-list="fileList"
          :headers="{ Authorization: getToken()! }"
          :action="action"
          :data="{ dest: formState.directory }"
          :multiple="true"
          draggable
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, reactive, ref } from 'vue';
  import type { FileItem } from '@arco-design/web-vue/es/upload/interfaces';
  import { useHostStore } from '@/store';
  import { resolveApiUrl } from '@/helper/api-helper';
  import { getToken } from '@/helper/auth';
  import FileSelector from '@/components/file/file-selector/index.vue';

  const emit = defineEmits(['ok']);

  const hostStore = useHostStore();
  const props = defineProps<{ host?: number }>();

  const formRef = ref();
  const formState = reactive({
    directory: '',
    // todo: api
    // overwrite: false,
  });

  const visible = ref(false);
  const fileList = ref<FileItem[]>([]);
  const action = computed(() => {
    const host = props.host ?? hostStore.currentId ?? hostStore.defaultId;
    return resolveApiUrl(`/files/${host}/upload`);
  });

  const setData = (data: { directory: string }) => {
    formState.directory = data.directory;
  };

  const handleCancel = () => {
    visible.value = false;
    emit('ok');
  };

  const show = () => {
    fileList.value = [];
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>

<style lang="less" scoped>
  :deep(.arco-form-item-content) {
    .file-selector {
      display: block;
      width: 100%;
    }
  }
</style>
