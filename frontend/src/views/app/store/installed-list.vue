<template>
  <a-spin :loading="loading" class="w-full">
    <div class="list">
      <div class="list-header">
        <a-input-search
          v-model="searchValue"
          class="w-[240px]"
          :placeholder="$t('common.search')"
          :loading="loading"
          allow-clear
          @clear="() => onSearch('')"
          @search="onSearch"
          @press-enter="onSearchEnter"
        />
      </div>
    </div>
    <div class="list-body">
      <a-grid :cols="3" :col-gap="24" :row-gap="24">
        <a-grid-item v-for="item of items" :key="item.id">
          <app-card
            :app="item"
            :manage-loading="managingAppId === item.id"
            @upgrade="handleUpgrade"
            @uninstall="handleUninstall"
            @manage="handleManageDatabase"
          />
        </a-grid-item>
      </a-grid>
    </div>
    <div class="list-footer mt-3">
      <a-pagination
        :default-current="pagination.page"
        :default-page-size="pagination.page_size"
        :total="pagination.total"
        show-total
      />
    </div>
  </a-spin>
  <!-- <upgrade-drawer ref="upgradeRef" /> -->
  <upgrade-log ref="upgradeLogRef" @ok="load" />
  <uninstall-log ref="uninstallLogRef" @ok="load" />
  <database-manager-drawer ref="databaseManagerRef" />
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref, toRaw } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { AppSimpleEntity } from '@/entity/App';
  import {
    getInstalledAppListApi,
    uninstallAppApi,
    upgradeAppApi,
  } from '@/api/store';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { useConfirm } from '@/composables/confirm';
  import AppCard from './components/app-card.vue';
  import UpgradeLog from './components/upgrade-log.vue';
  import UninstallLog from './components/uninstall-log.vue';
  import DatabaseManagerDrawer from './components/database-manager-drawer.vue';

  const { t } = useI18n();
  const pagination = reactive({
    page: 1,
    page_size: 10,
    total: 0,
  });

  const items = ref<AppSimpleEntity[]>([]);
  const managingAppId = ref<number | null>(null);

  // const upgradeRef = ref<InstanceType<typeof UpgradeDrawer>>();

  const { loading, showLoading, hideLoading } = useLoading();

  const params = reactive({
    name: '',
    page: pagination.page,
    page_size: pagination.page_size,
  });
  const load = async () => {
    if (loading.value) {
      return;
    }
    showLoading();
    try {
      const data = await getInstalledAppListApi(toRaw(params));
      items.value = data.items;
      pagination.page = data.page;
      pagination.page_size = data.page_size;
      pagination.total = data.total!;
    } catch (err: any) {
      await showErrorWithDockerCheck(err.message, err);
    } finally {
      hideLoading();
    }
  };
  onMounted(() => {
    load();
  });

  const searchValue = ref('');
  const onSearch = (value: string) => {
    searchValue.value = value;
    params.name = value;
    load();
  };
  const onSearchEnter = () => {
    onSearch(searchValue.value);
  };

  const { confirm } = useConfirm();
  const upgradeLogRef = ref<InstanceType<typeof UpgradeLog>>();
  const handleUpgrade = async (item: AppSimpleEntity) => {
    // upgradeRef.value?.setParams({ id: item.id });
    // upgradeRef.value?.load();
    // upgradeRef.value?.show();
    // 暂时没有输入，先直接一键升级
    const upgradeVersion = item.versions.find((v) => v.can_upgrade)!;
    if (!upgradeVersion) {
      console.error('no upgrade version');
      return;
    }

    if (
      await confirm(
        t('app.store.app.upgrade.confirm', {
          version: upgradeVersion.version + '.' + upgradeVersion.update_version,
        })
      )
    ) {
      try {
        loading.value = true;
        const res = await upgradeAppApi({
          id: item.id,
          upgrade_version_id: upgradeVersion.id,
          compose_name: item.name,
        });
        upgradeLogRef.value?.logFileLogs(res.log_host, res.log_path);
      } catch (err: any) {
        await showErrorWithDockerCheck(err?.message, err);
      } finally {
        loading.value = false;
      }
    }
  };

  const uninstallLogRef = ref<InstanceType<typeof UninstallLog>>();
  const handleUninstall = async (item: AppSimpleEntity) => {
    if (await confirm(t('app.store.app.uninstall.confirm'))) {
      try {
        loading.value = true;
        const res = await uninstallAppApi({
          id: item.id,
          compose_name: item.name,
        });
        uninstallLogRef.value?.logFileLogs(res.log_host, res.log_path);
        uninstallLogRef.value?.show();
      } catch (err: any) {
        await showErrorWithDockerCheck(err?.message, err);
      } finally {
        loading.value = false;
      }
    }
  };

  // 获取数据库类型
  const getDatabaseType = (
    item: AppSimpleEntity
  ): 'mysql' | 'postgresql' | 'redis' | null => {
    const name = item.name.toLowerCase();
    if (name.includes('mysql')) return 'mysql';
    if (name.includes('postgresql') || name.includes('postgres'))
      return 'postgresql';
    if (name.includes('redis')) return 'redis';
    return null;
  };

  // 管理数据库
  const databaseManagerRef = ref<InstanceType<typeof DatabaseManagerDrawer>>();
  const handleManageDatabase = async (item: AppSimpleEntity) => {
    const dbType = getDatabaseType(item);
    if (!dbType) return;

    // 设置 loading 状态
    managingAppId.value = item.id;

    try {
      await databaseManagerRef.value?.show(dbType, item.name);
    } finally {
      // 清除 loading 状态
      managingAppId.value = null;
    }
  };

  defineExpose({
    load,
  });
</script>

<style scoped>
  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .list-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-bottom: 16px;
  }
</style>
