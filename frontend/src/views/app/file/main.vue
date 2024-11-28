<template>
  <div class="file-page">
    <address-bar :path="pwd" :items="data.items" @goto="handleGoto" />
    <div class="file-layout">
      <div class="file-sidebar">
        <file-tree :items="tree" />
      </div>
      <div class="file-main">
        <div class="action-wrap mb-4">
          <a-button-group class="idb-button-group">
            <a-button>
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
            <a-button>
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
        </idb-table>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { storeToRefs } from 'pinia';
  import { computed, GlobalComponents, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { FileInfoEntity } from '@/entity/FileInfo';
  import AddressBar from '@/components/address-bar/index.vue';
  import FileTree from '@/components/file-tree/index.vue';
  import useFileStore from './store/file-store';

  const { t } = useI18n();
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();
  const store = useFileStore();
  const { pwd, tree, data, pasteVisible, selected } = storeToRefs(store);

  const columns = [
    {
      dataIndex: 'name',
      title: t('app.file.list.column.name'),
      width: 150,
      ellipsis: true,
    },
    {
      dataIndex: 'size',
      title: t('app.file.list.column.size'),
      width: 100,
      render: ({ record }: { record: FileInfoEntity }) => {
        return record.size;
      },
    },
    {
      dataIndex: 'mod_time',
      title: t('app.file.list.column.mod_time'),
      width: 120,
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
      slotName: 'operation',
    },
  ];

  const showHidden = computed({
    get: () => store.showHidden,
    set: (val) =>
      store.$patch({
        showHidden: val,
      }),
  });

  const handleClearSelected = () => {
    store.clearSelected();
    gridRef.value?.clearSelected();
  };

  const handleCreate = (key: any) => {
    console.log(key);
  };

  const handleGoto = (path: string) => {
    console.log(path);
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
    width: 300px;
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
</style>
