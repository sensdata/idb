<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('components.file.propertyDrawer.title')"
    unmountOnClose
    :footer="false"
    @cancel="handleCancel"
  >
    <a-descriptions :data="data" :column="1" size="large" bordered />
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { formatFileSize, formatTime } from '@/utils/format';
  import useLoading from '@/composables/loading';
  import { getFileSizeApi } from '@/api/file';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const fileInfoRef = ref<FileInfoEntity>();
  const sizeLockRef = ref(false);

  const loadSize = async () => {
    if (!fileInfoRef.value) {
      return;
    }
    if (fileInfoRef.value.size || sizeLockRef.value) {
      return;
    }
    if (fileInfoRef.value.is_dir) {
      setLoading(true);
      try {
        const res = await getFileSizeApi({
          source: fileInfoRef.value.path,
        });
        fileInfoRef.value.size = res.size;
        sizeLockRef.value = true;
      } finally {
        setLoading(false);
      }
    }
  };

  const data = computed(() => {
    const fileInfo = fileInfoRef.value;
    if (!fileInfo) {
      return [];
    }
    return [
      {
        label: t('components.file.propertyDrawer.name'),
        value: fileInfo.name,
      },
      {
        label: t('components.file.propertyDrawer.path'),
        value: fileInfo.path,
      },
      {
        label: t('components.file.propertyDrawer.size'),
        value: () => {
          if (loading.value) {
            return h(resolveComponent('a-spin'));
          }
          if (
            fileInfo.is_dir &&
            !fileInfoRef.value?.size &&
            !sizeLockRef.value
          ) {
            return h(
              'span',
              {
                class: 'color-primary cursor-pointer',
                onClick: loadSize,
              },
              t('components.file.propertyDrawer.calculate')
            );
          }
          return formatFileSize(fileInfo.size);
        },
      },
      {
        label: t('components.file.propertyDrawer.mode'),
        value: fileInfo.mode,
      },
      {
        label: t('components.file.propertyDrawer.user'),
        value: fileInfo.user,
      },
      {
        label: t('components.file.propertyDrawer.group'),
        value: fileInfo.group,
      },
      {
        label: t('components.file.propertyDrawer.mod_time'),
        value: formatTime(fileInfo.mod_time),
      },
    ];
  });

  const setData = (f: FileInfoEntity) => {
    fileInfoRef.value = f;
    sizeLockRef.value = false;
  };

  const visible = ref(false);
  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  const handleCancel = () => {
    hide();
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
