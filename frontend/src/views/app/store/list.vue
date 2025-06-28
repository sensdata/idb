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
                <h3 class="mt-0 mb-2">
                  {{ item.display_name }}
                </h3>
                <div class="mb-4 text-sm text-gray-500 line-clamp-2">
                  {{ item.description }}
                </div>
                <a-tag color="arcoblue">{{ item.category }}</a-tag>
              </div>
              <div class="item-actions">
                <!-- <a-tag v-if="item.status === 'installed'" color="green">
                  {{ $t('app.store.app.list.installed') }}
                </a-tag> -->
                <a-button
                  type="primary"
                  shape="round"
                  size="small"
                  @click="handleInstall(item)"
                >
                  {{ $t('app.store.app.list.install') }}
                </a-button>
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
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref, toRaw } from 'vue';
  import useLoading from '@/composables/loading';
  import { AppSimpleEntity } from '@/entity/App';
  import { getAppListApi } from '@/api/store';
  import { Message } from '@arco-design/web-vue';
  import { getHexColorByChar } from '@/helper/utils';
  import InstallDrawer from './components/install-drawer.vue';

  const pagination = reactive({
    page: 1,
    page_size: 10,
    total: 0,
  });

  const items = ref<AppSimpleEntity[]>([]);

  const installRef = ref<InstanceType<typeof InstallDrawer>>();

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
