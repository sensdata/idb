<template>
  <a-drawer
    :width="600"
    :visible="visible"
    :title="$t('app.file.propertyDrawer.title')"
    unmountOnClose
    :footer="false"
  >
    <a-descriptions :data="data" bordered />
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, h, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { formatFileSize } from '@/utils/format';
  import useLoading from '@/hooks/loading';
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
        const size = await getFileSizeApi({
          source: fileInfoRef.value.path,
        });
        fileInfoRef.value.size = size;
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
        label: t('app.file.propertyDrawer.name'),
        value: fileInfo.name,
      },
      {
        label: t('app.file.propertyDrawer.path'),
        value: fileInfo.path,
      },
      {
        label: t('app.file.propertyDrawer.size'),
        value: () => {
          if (loading) {
            return h('a-spin');
          }
          if (fileInfo.is_dir && !sizeLockRef.value) {
            return h(
              'span',
              {
                class: 'color-primary cursor-pointer',
                onClick: loadSize,
              },
              t('app.file.propertyDrawer.calculate')
            );
          }
          return formatFileSize(fileInfo.size);
        },
      },
      {
        label: t('app.file.propertyDrawer.mode'),
        value: fileInfo.mode,
      },
      {
        label: t('app.file.propertyDrawer.user'),
        value: fileInfo.user,
      },
      {
        label: t('app.file.propertyDrawer.group'),
        value: fileInfo.group,
      },
      {
        label: t('app.file.propertyDrawer.mtime'),
        value: fileInfo.mod_time,
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

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
