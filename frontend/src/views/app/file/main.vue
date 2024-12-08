<template>
  <div class="file-page">
    <address-bar :path="pwd" :items="data.items" @goto="store.handleGoto" />
    <div class="file-layout">
      <div class="file-sidebar">
        <file-tree
          :items="tree"
          :selected="current"
          :selected-change="store.handleTreeItemSelect"
          :open-change="store.handleTreeItemOpenChange"
        />
      </div>
      <div class="file-main">
        <div class="action-wrap mb-4">
          <a-button-group class="idb-button-group">
            <a-button @click="store.handleBack">
              <icon-left />
              <span class="ml-2">{{ t('app.file.list.action.back') }}</span>
            </a-button>
            <template v-if="!selected.length">
              <a-dropdown position="bl" @select="handleCreate">
                <a-button>
                  <icon-plus />
                  <span class="mx-2">{{
                    $t('app.file.list.action.create')
                  }}</span>
                  <icon-caret-down />
                </a-button>

                <template #content>
                  <a-doption value="createFolder">
                    <template #icon>
                      <icon-plus />
                    </template>
                    <template #default>
                      <span class="ml-2">{{
                        $t('app.file.list.action.createFolder')
                      }}</span>
                    </template>
                  </a-doption>
                  <a-doption value="createFile">
                    <template #icon>
                      <icon-plus />
                    </template>
                    <template #default>
                      <span class="ml-2">{{
                        $t('app.file.list.action.createFile')
                      }}</span>
                    </template>
                  </a-doption>
                </template>
              </a-dropdown>
              <a-button>
                <icon-upload />
                <span class="ml-2">{{
                  $t('app.file.list.action.upload')
                }}</span>
              </a-button>
            </template>
            <template v-else>
              <a-button>
                <icon-download />
                <span class="ml-2">{{
                  $t('app.file.list.action.download')
                }}</span>
              </a-button>
              <a-button @click="store.handleCopy">
                <icon-copy />
                <span class="ml-2">{{ $t('app.file.list.action.copy') }}</span>
              </a-button>
              <a-button @click="store.handleCut">
                <icon-scissor />
                <span class="ml-2">{{ $t('app.file.list.action.cut') }}</span>
              </a-button>
            </template>
            <a-button>
              <icon-code-square />
              <span class="ml-2">{{
                $t('app.file.list.action.terminal')
              }}</span>
            </a-button>
          </a-button-group>
          <a-button-group v-if="pasteVisible" class="idb-button-group ml-4">
            <a-button @click="store.handlePaste">
              <icon-paste />
              <span class="ml-2">
                {{
                  $t('app.file.list.action.paste', { count: selected.length })
                }}
              </span>
            </a-button>
            <a-button @click="handleClearSelected">
              <template #icon>
                <icon-close />
              </template>
            </a-button>
          </a-button-group>
        </div>
        <idb-table
          ref="gridRef"
          :loading="loading"
          :columns="columns"
          :data-source="data"
          has-batch
          row-key="path"
          @selected-change="store.handleSelected"
        >
          <template #leftActions>
            <a-checkbox v-model="showHidden">{{
              $t('app.file.list.filter.showHidden')
            }}</a-checkbox>
          </template>
          <template #name="{ record }: { record: FileItem }">
            <div
              class="name-cell flex items-center"
              @click="store.handleOpen(record)"
            >
              <folder-icon v-if="record.is_dir" />
              <file-icon v-else />
              <span class="color-primary cursor-pointer">{{
                record.name
              }}</span>
            </div>
          </template>
          <template #operation="{ record }: { record: FileItem }">
            <a-dropdown
              :popup-max-height="false"
              @select="handleOperation($event, record)"
            >
              <a-button type="text">
                <icon-settings />
                <icon-caret-down class="ml-4" />
              </a-button>
              <template #content>
                <a-doption value="open">
                  {{ $t('app.file.list.operation.open') }}
                </a-doption>
                <a-doption value="mode">
                  {{ $t('app.file.list.operation.mode') }}
                </a-doption>
                <a-doption value="rename">
                  {{ $t('app.file.list.operation.rename') }}
                </a-doption>
                <a-doption value="copyPath">
                  {{ $t('app.file.list.operation.copyPath') }}
                </a-doption>
                <a-doption value="property">
                  {{ $t('app.file.list.operation.property') }}
                </a-doption>
                <a-doption value="delete">
                  {{ $t('app.file.list.operation.delete') }}
                </a-doption>
              </template>
            </a-dropdown>
          </template>
        </idb-table>
      </div>
    </div>
    <mode-drawer ref="modeDrawerRef" />
  </div>
</template>

<script lang="ts" setup>
  import { storeToRefs } from 'pinia';
  import { computed, GlobalComponents, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import { formatFileSize } from '@/utils/format';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import FileIcon from '@/assets/icons/drive-file.svg';
  import AddressBar from '@/components/address-bar/index.vue';
  import FileTree from './components/file-tree/index.vue';
  import ModeDrawer from './components/mode-drawer/index.vue';
  import useFileStore from './store/file-store';
  import { FileItem } from './types/file-item';

  const { t } = useI18n();
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const modeDrawerRef = ref<InstanceType<typeof ModeDrawer>>();
  const store = useFileStore();
  const { current, pwd, tree, data, loading, pasteVisible, selected } =
    storeToRefs(store);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.file.list.column.name'),
      width: 150,
      ellipsis: true,
      slotName: 'name',
    },
    {
      dataIndex: 'size',
      title: t('app.file.list.column.size'),
      width: 100,
      render: ({ record }: { record: FileInfoEntity }) => {
        if (record.is_dir) {
          return '-';
        }
        return formatFileSize(record.size);
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.file.list.column.mod_time'),
      width: 180,
    },
    {
      dataIndex: 'mode',
      title: t('app.file.list.column.mode'),
      width: 100,
    },
    {
      dataIndex: 'user',
      title: t('app.file.list.column.user'),
      width: 150,
    },
    {
      dataIndex: 'group',
      title: t('app.file.list.column.group'),
      width: 110,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const showHidden = computed({
    get: () => store.showHidden,
    set: (val) => store.handleShowHiddenChange(val),
  });

  const handleClearSelected = () => {
    store.clearSelected();
    gridRef.value?.clearSelected();
  };

  const handleCreate = (key: any) => {
    switch (key) {
      case 'createFolder':
        store.handleCreateFolder();
        break;
      case 'createFile':
        store.handleCreateFile();
        break;
      default:
        break;
    }
  };

  const handleMode = (record: FileItem) => {
    modeDrawerRef.value?.setData(record);
    modeDrawerRef.value?.show();
  };

  const handleRename = (record: FileItem) => {
    console.log('rename', record);
  };

  const handleCopyPath = (record: FileItem) => {
    console.log('copyPath', record);
  };

  const handleProperty = (record: FileItem) => {
    console.log('property', record);
  };

  const handleDelete = (record: FileItem) => {
    console.log('delete', record);
  };

  const handleOperation = (key: any, record: FileItem) => {
    switch (key) {
      case 'open':
        store.handleOpen(record);
        break;
      case 'mode':
        handleMode(record);
        break;
      case 'rename':
        handleRename(record);
        break;
      case 'copyPath':
        handleCopyPath(record);
        break;
      case 'property':
        handleProperty(record);
        break;
      case 'delete':
        handleDelete(record);
        break;
      default:
        break;
    }
  };
</script>

<style scoped>
  .file-layout {
    display: flex;
    align-items: stretch;
    height: calc(100vh - 240px);
    margin-top: 20px;
    border: 1px solid var(--color-border-2);
    border-radius: 4px;
  }

  .file-sidebar {
    width: 240px;
    height: 100%;
    padding: 4px 8px;
    overflow: auto;
    border-right: 1px solid var(--color-border-2);
  }

  .file-main {
    flex: 1;
    min-width: 0;
    height: 100%;
    padding: 20px;
  }

  .name-cell svg {
    width: 14px;
    height: 14px;
    margin-right: 8px;
    vertical-align: top;
  }
</style>
