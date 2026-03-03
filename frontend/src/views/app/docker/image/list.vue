<template>
  <div>
    <docker-install-guide
      class="mb-4"
      @status-change="handleDockerStatusChange"
      @install-complete="handleDockerInstallComplete"
    />
    <idb-table
      ref="gridRef"
      :columns="columns"
      :has-search="true"
      :fetch="queryImagesApi"
      :beforeFetchHook="beforeFetchHook"
    >
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
          <a-button status="danger" @click="onPruneClick">
            {{ t('app.docker.image.list.action.prune') }}
          </a-button>
        </div>
      </template>
      <template #state="{ record }">
        <a-tag v-if="record.is_used" color="green">
          {{ $t('app.docker.image.list.state.used') }}
        </a-tag>
        <a-tag v-else color="orange">
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
        <div class="image-operation-bar">
          <a-button
            v-for="option in getPrimaryOptions(record)"
            :key="option.key"
            type="text"
            size="small"
            @click="handleOptionClick(option)"
          >
            {{ option.text }}
          </a-button>
          <a-dropdown v-if="getMoreOptions(record).length > 0" trigger="click">
            <a-button type="text" size="small">
              {{ $t('common.table.operation') }}
              <icon-down />
            </a-button>
            <template #content>
              <a-doption
                v-for="option in getMoreOptions(record)"
                :key="option.key"
                @click="handleOptionClick(option)"
              >
                <span
                  :class="{
                    'danger-option': option.status === 'danger',
                  }"
                >
                  {{ option.text }}
                </span>
              </a-doption>
            </template>
          </a-dropdown>
        </div>
      </template>
    </idb-table>
    <pull-image-drawer ref="pullImageRef" @success="reload" />
    <import-image-drawer ref="importImageRef" @success="reload" />
    <build-image-drawer ref="buildImageRef" @success="reload" />
    <tag-drawer ref="tagRef" @success="reload" />
    <push-drawer ref="pushRef" @success="reload" />
    <export-drawer ref="exportRef" @success="reload" />
    <inspect-drawer ref="inspectRef" />
  </div>
</template>

<script setup lang="ts">
  import { h, ref, resolveComponent } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { formatTime } from '@/utils/format';
  import { useConfirm } from '@/composables/confirm';
  import {
    queryImagesApi,
    batchDeleteImagesApi,
    pruneApi,
    inspectApi,
  } from '@/api/docker';
  import PullImageDrawer from './components/pull-image-drawer.vue';
  import ImportImageDrawer from './components/import-image-drawer.vue';
  import BuildImageDrawer from './components/build-image-drawer.vue';
  import TagDrawer from './components/tag-drawer.vue';
  import PushDrawer from './components/push-drawer.vue';
  import ExportDrawer from './components/export-drawer.vue';
  import InspectDrawer from './components/inspect-drawer.vue';

  const { t } = useI18n();
  const gridRef = ref();
  const reload = () => gridRef.value?.reload();

  const pullImageRef = ref<InstanceType<typeof PullImageDrawer>>();
  const importImageRef = ref<InstanceType<typeof ImportImageDrawer>>();
  const buildImageRef = ref<InstanceType<typeof BuildImageDrawer>>();
  const tagRef = ref<InstanceType<typeof TagDrawer>>();
  const pushRef = ref<InstanceType<typeof PushDrawer>>();
  const exportRef = ref<InstanceType<typeof ExportDrawer>>();
  const inspectRef = ref<InstanceType<typeof InspectDrawer>>();

  async function handleInspect(record: any) {
    try {
      const data = await inspectApi({
        type: 'image',
        id: record.id!,
      });
      inspectRef.value?.show(data.info);
    } catch (err: any) {
      await showErrorWithDockerCheck(err?.message, err);
    }
  }

  const beforeFetchHook = (fetchParams: any) => {
    const nextParams = { ...fetchParams };
    if (typeof nextParams.search === 'string') {
      const keyword = nextParams.search.trim();
      nextParams.info = keyword || undefined;
    }
    delete nextParams.search;
    return nextParams;
  };

  const columns = [
    {
      dataIndex: 'id',
      title: t('app.docker.image.list.column.id'),
      width: 180,
      render: ({ record }: { record: any }) => {
        return h(
          resolveComponent('a-link'),
          {
            onClick: () => {
              handleInspect(record);
            },
            hoverable: false,
          },
          {
            default: () => {
              return record.id.startsWith('sha256:')
                ? record.id.substring(7, 19)
                : record.id;
            },
          }
        );
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
      align: 'left' as const,
      width: 120,
      slotName: 'operation',
    },
  ];

  const { confirm } = useConfirm();
  const onPullImageClick = () => pullImageRef.value?.show();
  const onImportImageClick = () => importImageRef.value?.show();
  const onBuildImageClick = () => buildImageRef.value?.show();
  const onPruneClick = async () => {
    if (!(await confirm(t('app.docker.image.prune.confirm')))) {
      return;
    }
    try {
      await pruneApi({ type: 'image', with_tag_all: true });
      Message.success(t('app.docker.image.prune.success'));
      reload();
    } catch (e: any) {
      await showErrorWithDockerCheck(e.message, e);
    }
  };

  interface ImageOperationOption {
    key: string;
    text: string;
    visible?: boolean;
    confirm?: string | null;
    status?: 'normal' | 'success' | 'warning' | 'danger';
    click: () => void | Promise<void>;
  }

  const handleDelete = async (record: any) => {
    try {
      const result = await batchDeleteImagesApi({
        sources: record.id,
        force: false,
      });
      if (result.success) {
        Message.success(
          t('app.docker.image.list.operation.delete.success', {
            command: result.command,
          })
        );
      } else {
        Message.error(t('app.docker.image.list.operation.delete.failed'));
      }
      reload();
    } catch (e: any) {
      await showErrorWithDockerCheck(e.message, e);
    }
  };

  const getOperationOptions = (record: any): ImageOperationOption[] => [
    {
      key: 'inspect',
      text: t('app.docker.image.list.operation.inspect'),
      click: () => handleInspect(record),
    },
    {
      key: 'push',
      text: t('app.docker.image.list.operation.push'),
      click: () => pushRef.value?.show(record),
    },
    {
      key: 'tag',
      text: t('app.docker.image.list.operation.tag'),
      click: () => tagRef.value?.show(record),
    },
    {
      key: 'export',
      text: t('app.docker.image.list.operation.export'),
      click: () => exportRef.value?.show(record),
    },
    {
      key: 'delete',
      text: t('app.docker.image.list.operation.delete'),
      status: 'danger',
      confirm: t('app.docker.image.list.operation.delete.confirm'),
      click: () => handleDelete(record),
    },
  ];

  const handleOptionClick = async (option: ImageOperationOption) => {
    if (option.confirm && !(await confirm(option.confirm))) {
      return;
    }
    await option.click();
  };

  const getPrimaryOptions = (record: any) => {
    const options = getOperationOptions(record).filter(
      (item) => item.visible !== false
    );
    return options.filter((item) => ['inspect', 'push'].includes(item.key));
  };

  const getMoreOptions = (record: any) => {
    const primaryKeys = new Set(
      getPrimaryOptions(record).map((item) => item.key)
    );
    return getOperationOptions(record).filter(
      (item) => item.visible !== false && !primaryKeys.has(item.key)
    );
  };

  // Docker 状态变化处理
  const handleDockerStatusChange = (status: string) => {
    // 如果 Docker 状态变化，可以重新加载镜像列表
    if (status === 'installed') {
      reload();
    }
  };

  // Docker 安装完成处理
  const handleDockerInstallComplete = () => {
    // Docker 安装完成后重新加载镜像列表
    reload();
  };
</script>

<style scoped>
  .image-operation-bar {
    display: flex;
    flex-wrap: nowrap;
    gap: 0.25rem;
    align-items: center;
    white-space: nowrap;
  }

  .danger-option {
    color: rgb(var(--danger-6));
  }
</style>
