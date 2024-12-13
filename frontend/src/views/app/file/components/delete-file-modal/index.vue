<template>
  <a-modal
    v-model:visible="visible"
    :ok-loading="loading"
    :title="$t('app.file.deleteFileModal.title')"
    @ok="handleOk"
    @cancel="handleCancel"
  >
    <a-alert type="warning">{{ $t('app.file.deleteFileModal.alert') }}</a-alert>
    <div class="mt-4">
      <a-checkbox v-model="formState.force_delete">
        {{ $t('app.file.deleteFileModal.force_delete') }}
      </a-checkbox>
    </div>
    <div class="mt-4 mb-2">
      <a-checkbox v-model="formState.permanently_delete">
        {{ $t('app.file.deleteFileModal.permanently_delete') }}
      </a-checkbox>
    </div>
    <div class="list">
      <div
        v-for="source in formState.sources"
        :key="source.path"
        class="flex items-center"
      >
        <folder-icon v-if="source.is_dir" />
        <file-icon v-else />
        <span>{{ source.name }}</span>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import useLoading from '@/hooks/loading';
  import { batchDeleteFileApi } from '@/api/file';
  import { pick } from 'lodash';

  const emit = defineEmits(['ok']);

  const visible = ref(false);
  const formState = reactive({
    sources: [] as Array<{
      name: string;
      path: string;
      is_dir: boolean;
    }>,
    force_delete: false,
    permanently_delete: false,
  });

  const { loading, setLoading } = useLoading(false);

  const setData = (files: FileInfoEntity[]) => {
    formState.sources = files.slice(0);
    formState.permanently_delete = false;
    formState.force_delete = false;
  };

  const handleOk = async () => {
    setLoading(true);
    try {
      await batchDeleteFileApi({
        permanently_delete: formState.permanently_delete,
        force_delete: formState.force_delete,
        sources: formState.sources.map((source) =>
          pick(source, ['path', 'is_dir'])
        ),
      });
      visible.value = false;
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

<style scoped>
  .list {
    padding-left: 4px;
    line-height: 24px;
  }

  .list svg {
    width: 14px;
    height: 14px;
    margin-right: 8px;
  }
</style>
