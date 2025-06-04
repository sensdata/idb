<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="queryImagesApi">
    <template #leftActions>
      <div class="flex gap-2">
        <a-button type="primary" @click="onPullImageClick">{{
          $t('app.docker.image.list.action.pull')
        }}</a-button>
        <a-button @click="onImportImageClick">{{
          $t('app.docker.image.list.action.import')
        }}</a-button>
        <a-button @click="onBuildImageClick">{{
          $t('app.docker.image.list.action.build')
        }}</a-button>
      </div>
    </template>
    <template #state="{ record }">
      <a-tag v-if="record.is_used" color="green">
        {{ $t('app.docker.image.list.state.used') }}
      </a-tag>
      <a-tag v-else color="gray">
        {{ $t('app.docker.image.list.state.unused') }}
      </a-tag>
    </template>
    <template #tags="{ record }">
      <div
        v-for="(tag, index) in record.tags"
        :key="tag"
        :class="{
          'mt-2': index > 0,
        }"
      >
        <a-tag bordered>{{ tag }}</a-tag>
      </div>
    </template>
    <template #operation="{ record }">
      <idb-table-operation
        type="button"
        :options="getOperationOptions(record)"
      />
    </template>
  </idb-table>
  <pull-image-drawer ref="pullImageRef" @success="reload" />
  <import-image-drawer ref="importImageRef" @success="reload" />
  <build-image-drawer ref="buildImageRef" @success="reload" />
  <tag-drawer ref="tagRef" @success="reload" />
  <push-drawer ref="pushRef" @success="reload" />
  <export-drawer ref="exportRef" @success="reload" />
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { formatTime } from '@/utils/format';
  import { queryImagesApi, batchDeleteImagesApi } from '@/api/docker';
  import PullImageDrawer from './components/pull-image-drawer.vue';
  import ImportImageDrawer from './components/import-image-drawer.vue';
  import BuildImageDrawer from './components/build-image-drawer.vue';
  import TagDrawer from './components/tag-drawer.vue';
  import PushDrawer from './components/push-drawer.vue';
  import ExportDrawer from './components/export-drawer.vue';

  const { t } = useI18n();
  const gridRef = ref();
  const reload = () => gridRef.value?.reload();

  const pullImageRef = ref<InstanceType<typeof PullImageDrawer>>();
  const importImageRef = ref();
  const buildImageRef = ref();
  const tagRef = ref();
  const pushRef = ref();
  const exportRef = ref();

  const columns = [
    {
      dataIndex: 'id',
      title: t('app.docker.image.list.column.id'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return record.id.startsWith('sha256:')
          ? record.id.substring(7, 19)
          : record.id;
      },
    },
    {
      dataIndex: 'state',
      title: t('app.docker.image.list.column.state'),
      width: 110,
      slotName: 'state',
    },
    {
      dataIndex: 'tags',
      title: t('app.docker.image.list.column.tags'),
      width: 200,
      slotName: 'tags',
    },
    {
      dataIndex: 'size',
      title: t('app.docker.image.list.column.size'),
      width: 120,
    },
    {
      dataIndex: 'created_at',
      title: t('app.docker.image.list.column.created'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return formatTime(record.created_at);
      },
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      align: 'center' as const,
      width: 180,
      slotName: 'operation',
    },
  ];

  const onPullImageClick = () => pullImageRef.value?.show();
  const onImportImageClick = () => importImageRef.value?.show();
  const onBuildImageClick = () => buildImageRef.value?.show();

  const getOperationOptions = (record: any) => [
    {
      text: t('app.docker.image.list.operation.tag'),
      click: () => tagRef.value?.show(record),
    },
    {
      text: t('app.docker.image.list.operation.push'),
      click: () => pushRef.value?.show(record),
    },
    {
      text: t('app.docker.image.list.operation.export'),
      click: () => exportRef.value?.show(record),
    },
    {
      text: t('app.docker.image.list.operation.delete'),
      confirm: t('app.docker.image.list.operation.delete.confirm'),
      click: async () => {
        try {
          await batchDeleteImagesApi({ sources: record.id, force: false });
          Message.success(t('app.docker.image.list.operation.delete.success'));
          reload();
        } catch (e: any) {
          Message.error(
            e.message || t('app.docker.image.list.operation.delete.failed')
          );
        }
      },
    },
  ];
</script>

<style scoped></style>
