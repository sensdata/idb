<template>
  <div class="box">
    <div class="host-switch">
      <a-tooltip position="bottom" :content="$t('components.switchHost.tips')">
        <button class="btn color-primary" @click="handleClick">
          <IconHome />
          <span>{{ currentHost?.name }}</span>
        </button>
      </a-tooltip>
    </div>
  </div>
  <a-drawer
    :width="640"
    :visible="drawerVisible"
    :footer="false"
    unmountOnClose
    @cancel="handleCancel"
  >
    <template #title>
      {{ $t('components.switchHost.currentlabel') }}
      <strong> {{ currentHost?.name }}({{ currentHost?.addr }}) </strong>
    </template>
    <a-input-search
      v-model="searchValue"
      class="mt-2"
      :placeholder="$t('components.switchHost.searchPlaceholder')"
      search-button
      allow-clear
      @clear="() => handleSearch('')"
      @search="handleSearch"
      @press-enter="handleSearchEnter"
    />
    <div class="mt-5">
      <idb-table
        ref="gridRef"
        :columns="columns"
        :fetch="getHostListApi"
        :hasToolbar="false"
      >
        <template #operation="{ record }: { record: HostEntity }">
          <a-button
            v-if="record.id === currentHost?.id"
            type="primary"
            disabled
            size="mini"
          >
            {{ $t('components.switchHost.operation.selected') }}
          </a-button>
          <a-button
            v-else
            type="primary"
            size="mini"
            @click="handleSelect(record)"
          >
            {{ $t('components.switchHost.operation.select') }}
          </a-button>
        </template>
      </idb-table>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { HostEntity } from '@/entity/Host';
  import { getHostListApi } from '@/api/host';
  import { useHostStore } from '@/store';
  import usetCurrentHost from '@/composables/current-host';

  const { t } = useI18n();

  const gridRef = ref();
  const hostStore = useHostStore();
  const { switchHost } = usetCurrentHost();
  const currentHost = computed(() => hostStore.current);

  const drawerVisible = ref(false);
  const handleClick = () => {
    drawerVisible.value = true;
  };
  const handleCancel = () => {
    drawerVisible.value = false;
  };

  // 搜索
  const searchValue = ref('');
  const handleSearch = (value: string) => {
    searchValue.value = value;
    gridRef.value?.load({
      search: value,
      page: 1,
    });
  };
  const handleSearchEnter = () => {
    handleSearch(searchValue.value);
  };
  const columns = [
    {
      dataIndex: 'addr',
      title: t('components.switchHost.column.addr'),
      width: 150,
    },
    {
      dataIndex: 'name',
      title: t('components.switchHost.column.name'),
      width: 150,
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      slotName: 'operation',
    },
  ];

  const handleSelect = (record: HostEntity) => {
    switchHost(record.id, true);
    handleCancel();
  };
</script>

<style scoped>
  .box {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 20px;
    border-bottom: 1px solid var(--color-border-2);
  }

  .host-switch {
    display: flex;
    align-items: center;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    height: 36px;
    padding: 10px 20px;
    cursor: pointer;
    background-color: #fff;
    border: 1px solid var(--color-border-2);
    border-radius: 18px;
  }

  .btn:hover {
    border-color: var(--color-border-3);
  }

  .btn span {
    margin-left: 8px;
  }
</style>
