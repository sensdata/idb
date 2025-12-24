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
        <a-pagination
          :default-current="pagination.page"
          :default-page-size="pagination.page_size"
          :total="pagination.total"
        />
      </div>
    </div>
    <div class="list-body">
      <a-grid :cols="3" :col-gap="24" :row-gap="24">
        <a-grid-item v-for="item of items" :key="item.id">
          <app-card
            :app="item"
            @install="handleInstall"
            @upgrade="handleUpgrade"
            @uninstall="handleUninstall"
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
  <install-drawer ref="installRef" @ok="load" />
  <upgrade-log ref="upgradeLogRef" @ok="load" />
  <uninstall-log ref="uninstallLogRef" @ok="load" />
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref, toRaw } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { AppSimpleEntity } from '@/entity/App';
  import {
    getAppListApi,
    getInstalledAppListApi,
    upgradeAppApi,
    uninstallAppApi,
  } from '@/api/store';
  import { showErrorWithDockerCheck } from '@/helper/show-error';
  import { useConfirm } from '@/composables/confirm';
  import AppCard from './components/app-card.vue';
  import InstallDrawer from './components/install-drawer.vue';
  import UpgradeLog from './components/upgrade-log.vue';
  import UninstallLog from './components/uninstall-log.vue';

  const { t } = useI18n();

  const pagination = reactive({
    page: 1,
    page_size: 10,
    total: 0,
  });

  const items = ref<AppSimpleEntity[]>([]);

  const installRef = ref<InstanceType<typeof InstallDrawer>>();
  const upgradeLogRef = ref<InstanceType<typeof UpgradeLog>>();
  const uninstallLogRef = ref<InstanceType<typeof UninstallLog>>();

  const { loading, showLoading, hideLoading } = useLoading();
  const { confirm } = useConfirm();

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
      const data = await getAppListApi(toRaw(params));
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

  const handleInstall = (item: AppSimpleEntity) => {
    installRef.value?.setParams({ id: item.id });
    installRef.value?.load();
    installRef.value?.show();
  };

  const handleUpgrade = async (item: AppSimpleEntity) => {
    // 找到可升级的版本
    const upgradeVersion = item.versions?.find((v) => v.can_upgrade);
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
        showLoading();

        // 在"All"标签中,item.name 是通用名称(如 "mysql"),需要获取实际的 compose 名称
        // 通过查询已安装应用列表来获取实际的 compose 名称
        const response = await getInstalledAppListApi({
          page: 1,
          page_size: 100,
        });
        const installedApp = response.items.find(
          (app) => app.display_name === item.display_name
        );

        if (!installedApp) {
          throw new Error('未找到已安装的应用');
        }

        const res = await upgradeAppApi({
          id: item.id,
          upgrade_version_id: upgradeVersion.id,
          compose_name: installedApp.name,
        });
        upgradeLogRef.value?.logFileLogs(res.log_host, res.log_path);
        upgradeLogRef.value?.show();
      } catch (err: any) {
        await showErrorWithDockerCheck(err?.message, err);
      } finally {
        hideLoading();
      }
    }
  };

  const handleUninstall = async (item: AppSimpleEntity) => {
    if (await confirm(t('app.store.app.uninstall.confirm'))) {
      try {
        showLoading();

        // 在"All"标签中,item.name 是通用名称(如 "mysql"),需要获取实际的 compose 名称
        // 通过查询已安装应用列表来获取实际的 compose 名称
        const response = await getInstalledAppListApi({
          page: 1,
          page_size: 100,
        });
        const installedApp = response.items.find(
          (app) => app.display_name === item.display_name
        );

        if (!installedApp) {
          throw new Error('未找到已安装的应用');
        }

        const res = await uninstallAppApi({
          id: item.id,
          compose_name: installedApp.name,
        });
        uninstallLogRef.value?.logFileLogs(res.log_host, res.log_path);
        uninstallLogRef.value?.show();
      } catch (err: any) {
        await showErrorWithDockerCheck(err?.message, err);
      } finally {
        hideLoading();
      }
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
    margin-bottom: 1.143rem;
  }

  .list-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-bottom: 1.143rem;
  }
</style>
