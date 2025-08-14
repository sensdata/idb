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
          <a-card hoverable>
            <div class="item-box flex gap-3 h-26">
              <a-avatar
                shape="square"
                class="item-avatar"
                :size="72"
                :style="{
                  backgroundColor: getHexColorByChar(item.display_name),
                }"
              >
                {{ item.display_name.charAt(0) }}
              </a-avatar>
              <div class="item-main flex-1">
                <h3 class="mt-0 mb-3">
                  {{ item.display_name }}
                </h3>
                <!-- 未安装应用显示描述和分类 -->
                <template v-if="item.status === 'uninstalled'">
                  <div class="mb-4 text-sm text-gray-500 line-clamp-2">
                    {{ item.description }}
                  </div>
                  <a-tag color="arcoblue">{{ item.category }}</a-tag>
                </template>
                <!-- 已安装应用显示版本和安装时间 -->
                <template v-else>
                  <a-tag bordered class="text-gray-600 mb-2">
                    {{ $t('app.store.app.list.version') }}:
                    {{ item.current_version }}
                  </a-tag>
                  <div class="text-gray-500 text-sm mb-2">
                    {{ $t('app.store.app.list.install_at') }}:
                    {{
                      item.versions && item.versions[0]
                        ? item.versions[0].created_at
                        : ''
                    }}
                  </div>
                </template>
              </div>
              <div class="item-actions flex flex-col gap-3">
                <!-- 未安装应用显示安装按钮 -->
                <a-button
                  v-if="item.status === 'uninstalled'"
                  type="primary"
                  shape="round"
                  size="small"
                  @click="handleInstall(item)"
                >
                  {{ $t('app.store.app.list.install') }}
                </a-button>
                <!-- 已安装应用显示升级和卸载按钮 -->
                <template v-else>
                  <a-button
                    type="primary"
                    shape="round"
                    size="small"
                    :disabled="!item.has_upgrade"
                    @click="handleUpgrade(item)"
                  >
                    {{ $t('app.store.app.list.upgrade') }}
                  </a-button>
                  <a-button
                    type="primary"
                    shape="round"
                    status="danger"
                    size="small"
                    @click="handleUninstall(item)"
                  >
                    {{ $t('app.store.app.list.uninstall') }}
                  </a-button>
                </template>
              </div>
            </div>
          </a-card>
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
  import { getAppListApi, upgradeAppApi, uninstallAppApi } from '@/api/store';
  import { Message } from '@arco-design/web-vue';
  import { getHexColorByChar } from '@/helper/utils';
  import { useConfirm } from '@/composables/confirm';
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
      Message.error(err.message);
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
        const res = await upgradeAppApi({
          id: item.id,
          upgrade_version_id: upgradeVersion.id,
          compose_name: item.name,
        });
        upgradeLogRef.value?.logFileLogs(res.log_host, res.log_path);
        upgradeLogRef.value?.show();
      } catch (err: any) {
        Message.error(err?.message);
      } finally {
        hideLoading();
      }
    }
  };

  const handleUninstall = async (item: AppSimpleEntity) => {
    if (await confirm(t('app.store.app.uninstall.confirm'))) {
      try {
        showLoading();
        const res = await uninstallAppApi({
          id: item.id,
          compose_name: item.name,
        });
        uninstallLogRef.value?.logFileLogs(res.log_host, res.log_path);
        uninstallLogRef.value?.show();
      } catch (err: any) {
        Message.error(err?.message);
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
    margin-bottom: 16px;
  }

  .list-footer {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    margin-bottom: 16px;
  }
</style>
