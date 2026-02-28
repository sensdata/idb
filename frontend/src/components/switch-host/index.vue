<template>
  <div class="box">
    <div class="host-switch">
      <a-tooltip position="bottom" :content="$t('components.switchHost.tips')">
        <button class="btn color-primary" @click="handleClick">
          <IconHome />
          <span class="host-name truncate">{{ currentHost?.name }}</span>
        </button>
      </a-tooltip>
    </div>
    <div v-if="currentModuleName" class="current-module truncate">
      {{ currentModuleName }}
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
  import { useRoute } from 'vue-router';
  import { useI18n } from 'vue-i18n';
  import { HostEntity } from '@/entity/Host';
  import { getHostListApi } from '@/api/host';
  import { useHostStore } from '@/store';
  import usetCurrentHost from '@/composables/current-host';

  const { t } = useI18n();

  const gridRef = ref();
  const route = useRoute();
  const hostStore = useHostStore();
  const { switchHost } = usetCurrentHost();
  const currentHost = computed(() => hostStore.current);
  const currentModuleName = computed(() => {
    const localeKey = route.meta?.locale;
    if (typeof localeKey === 'string' && localeKey.length > 0) {
      return t(localeKey);
    }
    return '';
  });

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
    gap: 12px;
    align-items: center;
    justify-content: flex-start;
    padding: 10px 16px;
    background-color: var(--color-bg-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  .host-switch {
    display: flex;
    align-items: center;
    max-width: 60%;
  }

  .current-module {
    display: inline-flex;
    align-items: center;
    max-width: 36%;
    height: 36px;
    padding: 0 12px;
    overflow: hidden;
    font-size: 14px;
    font-weight: 500;
    line-height: 36px;
    color: var(--color-text-2);
    text-align: left;
    background-color: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 999px;
    transition: all 0.2s ease;
  }

  .current-module::before {
    width: 6px;
    height: 6px;
    margin-right: 8px;
    content: '';
    background-color: var(--color-primary-light-4);
    border-radius: 999px;
  }

  .btn {
    display: inline-flex;
    gap: 8px;
    align-items: center;
    justify-content: center;
    max-width: 100%;
    height: 36px;
    padding: 0 14px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    background-color: var(--color-fill-2);
    border: 1px solid var(--color-border-2);
    border-radius: 18px;
    transition: all 0.2s ease;
  }

  .btn:hover {
    color: var(--color-primary-6);
    background-color: var(--color-fill-3);
    border-color: var(--color-border-3);
  }

  .btn:active {
    transform: translateY(1px);
  }

  .host-name {
    max-width: 220px;
  }

  @media (width <= 1280px) {
    .host-switch {
      max-width: 55%;
    }
    .current-module {
      max-width: 42%;
    }
  }
</style>
