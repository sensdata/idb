<template>
  <a-spin :loading="loading">
    <div class="box">
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.server_time') }}</div>
        <div class="col2">{{ formatTime(data.server_time) }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-space>
            <a-button type="primary" size="mini" @click="handleModifyTime">{{
              $t('common.modify')
            }}</a-button>
            <a-button
              type="primary"
              size="mini"
              :loading="isSyncingTime"
              @click="handleSyncTime"
              >{{ $t('app.sysinfo.overview.button.sync_time') }}</a-button
            >
            <span
              v-if="syncTimeStatus"
              :class="{
                'sync-success': syncTimeStatus === 'success',
                'sync-syncing': syncTimeStatus === 'syncing',
              }"
            >
              {{
                syncTimeStatus === 'syncing'
                  ? $t('app.sysinfo.overview.sync.syncing')
                  : $t('app.sysinfo.overview.sync.success')
              }}
            </span>
          </a-space>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.data_status') }}</div>
        <div class="col2">
          {{ formatTime(lastUpdatedAt || undefined) }}
        </div>
        <div class="col3"></div>
        <div class="col4">
          <a-tag :color="dataFresh ? 'green' : 'red'">
            {{
              dataFresh
                ? $t('app.sysinfo.overview.data_fresh')
                : $t('app.sysinfo.overview.data_stale')
            }}
          </a-tag>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{
          $t('app.sysinfo.overview.server_time_zone')
        }}</div>
        <div class="col2">{{ data.server_time_zone }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-button type="primary" size="mini" @click="handleModifyTimeZone">{{
            $t('common.modify')
          }}</a-button>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.boot_time') }}</div>
        <div class="col2"> {{ formatTime(data.boot_time) }} </div>
        <div class="col3"></div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.run_time') }}</div>
        <div class="col2"> {{ formatSeconds(data.run_time) }} </div>
        <div class="col3"></div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.idle_time') }}</div>
        <div class="col2">{{ formatSeconds(data.idle_time) }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-tag :color="'rgb(var(--success-6))'">{{
            $t('app.sysinfo.overview.tag.label_free', {
              free: (data.idle_rate || 0) + '%',
            })
          }}</a-tag>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.cpu_usage') }}</div>
        <div class="col2">{{ data.cpu_usage }}</div>
        <div class="col3"></div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.current_load') }}</div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.overview.tag.count1') }}</div>
            <div class="col3">{{ data.current_load?.process_count1 }}</div>
            <div class="col4">
              <ProcessCountTag :count="data.current_load?.process_count1" />
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{
              $t('app.sysinfo.overview.process_count5')
            }}</div>
            <div class="col3">{{ data.current_load?.process_count5 }}</div>
            <div class="col4">
              <ProcessCountTag :count="data.current_load?.process_count5" />
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{
              $t('app.sysinfo.overview.process_count15')
            }}</div>
            <div class="col3">{{ data.current_load?.process_count15 }}</div>
            <div class="col4">
              <ProcessCountTag :count="data.current_load?.process_count15" />
            </div>
          </div>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.memory_usage') }}</div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.overview.memory_total') }}
              <a-tooltip
                :content="$t('app.sysinfo.overview.memory_total_tips')"
              >
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip>
            </div>
            <div class="col3">{{ data.memory_usage?.total }}</div>
            <div class="col4"></div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.overview.memory_free') }}</div>
            <div class="col3">{{ data.memory_usage?.free }}</div>
            <div class="col4">
              <a-tag
                v-if="data.memory_usage?.free_rate !== undefined"
                :color="'rgb(var(--warning-6))'"
                >{{
                  $t('app.sysinfo.overview.tag.leave_unused', {
                    used: data.memory_usage.free_rate.toFixed(2) + '%',
                  })
                }}</a-tag
              >
            </div>
          </div>
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.overview.memory_used') }}
              <a-tooltip :content="$t('app.sysinfo.overview.memory_used_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip>
            </div>
            <div class="col3">{{ data.memory_usage?.used }}</div>
            <div class="col4"></div>
          </div>
          <div class="subline">
            <div class="col2">{{
              $t('app.sysinfo.overview.memory_real_used')
            }}</div>
            <div class="col3"> {{ data.memory_usage?.real_used }}</div>
            <div class="col4">
              <a-link class="text-sm" @click="handleViewMemoryDetail">{{
                $t('app.sysinfo.overview.button.view_memory')
              }}</a-link>
            </div>
          </div>
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.overview.memory_buffered') }}
              <!-- <a-tooltip :content="$t('app.sysinfo.overview.memory_buffered_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip> -->
            </div>
            <div class="col3"> {{ data.memory_usage?.buffered }}</div>
            <div class="col4"></div>
          </div>
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.overview.memory_cached') }}
              <!-- <a-tooltip :content="$t('app.sysinfo.overview.memory_cached_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip> -->
            </div>
            <div class="col3">{{ data.memory_usage?.cached }}</div>
            <div class="col4">
              <a-space>
                <a-button
                  type="primary"
                  size="mini"
                  @click="handleClearCache"
                  >{{ $t('app.sysinfo.overview.button.clear_cache') }}</a-button
                >
                <a-button
                  type="primary"
                  size="mini"
                  @click="handleAutoClearCache"
                  >{{
                    $t('app.sysinfo.overview.button.auto_clear_setting')
                  }}</a-button
                >
              </a-space>
            </div>
          </div>
        </div>
      </div>
      <div class="line no-border">
        <div class="col1">
          {{ $t('app.sysinfo.overview.virtual_memory') }}
          <br />
          ({{ $t('app.sysinfo.overview.swap_usage') }})
        </div>
        <div class="colspan">
          <div
            v-if="!data.swap_usage || data.swap_usage.total === '0B'"
            class="subline"
          >
            <div class="col2">{{
              $t('app.sysinfo.overview.no_virtual_memory')
            }}</div>
            <div class="col3"></div>
            <div class="col4">
              <a-button type="primary" size="mini" @click="handleCreateSwap">{{
                $t('app.sysinfo.overview.button.create_virtual_memory')
              }}</a-button>
            </div>
          </div>
          <template v-else>
            <div class="subline">
              <div class="col2">{{
                $t('app.sysinfo.overview.swap_total')
              }}</div>
              <div class="col3">{{ data.swap_usage?.total }}</div>
              <div class="col4"></div>
            </div>
            <div class="subline">
              <div class="col2">{{ $t('app.sysinfo.overview.swap_used') }}</div>
              <div class="col3">{{ data.swap_usage?.used }}</div>
              <div class="col4"></div>
            </div>
            <div class="subline">
              <div class="col2">{{ $t('app.sysinfo.overview.swap_free') }}</div>
              <div class="col3">{{ data.swap_usage?.free }}</div>
              <div class="col4">
                <a-tag
                  :color="getStorageUsedColor(data.swap_usage!.used_rate)"
                  class="italic"
                >
                  {{ data.swap_usage!.used_rate }}%
                </a-tag>
              </div>
            </div>
            <div class="subline">
              <div class="col2"></div>
              <div class="col3">
                <a-button
                  type="primary"
                  size="mini"
                  @click="handleDeleteSwap"
                  >{{ $t('app.sysinfo.overview.button.delete_swap') }}</a-button
                >
              </div>
            </div>
          </template>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.overview.storage') }}</div>
        <div class="colspan mb-6">
          <a-table
            :columns="storageColumns"
            :data="data.storage || []"
            :pagination="false"
            size="small"
          >
            <template #rate="{ record }">
              <a-tag
                :color="getStorageUsedColor(record.used_rate)"
                class="italic"
              >
                {{ record.used_rate }}%
              </a-tag>
            </template>
          </a-table>
        </div>
      </div>
    </div>
  </a-spin>
  <time-modify ref="timeModifyRef" @ok="load" />
  <create-swap-modal ref="createSwapModalRef" @ok="load" />
  <timezone-modify ref="timezoneModifyRef" @ok="load" />
  <auto-clear-cache ref="autoClearCacheRef" @ok="load" />
  <a-modal
    v-model:visible="memoryDetailVisible"
    :title="$t('app.sysinfo.overview.memory_detail.title')"
    :footer="false"
    width="520px"
  >
    <a-descriptions
      :column="1"
      size="medium"
      layout="horizontal"
      :label-style="{ width: '220px' }"
    >
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.total')"
      >
        {{ data.memory_usage?.total || '-' }}
      </a-descriptions-item>
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.used')"
      >
        {{ data.memory_usage?.used || '-' }}
      </a-descriptions-item>
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.free')"
      >
        {{ data.memory_usage?.free || '-' }}
      </a-descriptions-item>
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.process')"
      >
        {{ data.memory_usage?.real_used || '-' }}
      </a-descriptions-item>
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.buffered')"
      >
        {{ data.memory_usage?.buffered || '-' }}
      </a-descriptions-item>
      <a-descriptions-item
        :label="$t('app.sysinfo.overview.memory_detail.cached')"
      >
        {{ data.memory_usage?.cached || '-' }}
      </a-descriptions-item>
    </a-descriptions>
  </a-modal>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import {
    computed,
    h,
    onBeforeUnmount,
    onMounted,
    reactive,
    ref,
    resolveComponent,
  } from 'vue';
  import { formatSeconds, formatTime } from '@/utils/format';
  import useLoading from '@/composables/loading';
  import {
    getSysInfoOverviewApi,
    SysInfoOverviewRes,
    syncTimeApi,
    deleteSwapApi,
    clearMemoryCacheApi,
  } from '@/api/sysinfo';
  import { useConfirm } from '@/composables/confirm';
  import { Message } from '@arco-design/web-vue';
  import TimeModify from '@/components/time-modify/index.vue';
  import CreateSwapModal from './components/create-swap-modal/index.vue';
  import TimezoneModify from './components/timezone-modify/index.vue';
  import AutoClearCache from './components/auto-clear-cache/index.vue';

  const { t } = useI18n();
  const { confirm } = useConfirm();
  const { loading, setLoading } = useLoading(false);

  const data = reactive<Partial<SysInfoOverviewRes>>({});

  const storageColumns = [
    {
      title: t('app.sysinfo.overview.storage_mount_point'),
      dataIndex: 'name',
      width: 200,
    },
    {
      title: t('app.sysinfo.overview.storage_total'),
      dataIndex: 'total',
      width: 120,
    },
    {
      title: t('app.sysinfo.overview.storage_used'),
      dataIndex: 'used',
      width: 120,
    },
    {
      title: t('app.sysinfo.overview.storage_free'),
      dataIndex: 'free',
      width: 120,
    },
    {
      title: t('app.sysinfo.overview.storage_used_rate'),
      dataIndex: 'used_rate',
      slotName: 'rate',
      width: 120,
    },
  ];

  const isSyncingTime = ref(false);
  const syncTimeStatus = ref<'syncing' | 'success' | null>(null);
  const memoryDetailVisible = ref(false);
  const lastUpdatedAt = ref<number | null>(null);
  const nowTick = ref(Date.now());
  const polling = ref(false);
  let syncStatusTimer: ReturnType<typeof setTimeout> | null = null;

  const timeModifyRef = ref<InstanceType<typeof TimeModify>>();
  const createSwapModalRef = ref<InstanceType<typeof CreateSwapModal>>();
  const timezoneModifyRef = ref<InstanceType<typeof TimezoneModify>>();
  const autoClearCacheRef = ref<InstanceType<typeof AutoClearCache>>();

  const dataFresh = computed(() => {
    if (!lastUpdatedAt.value) return false;
    return nowTick.value - lastUpdatedAt.value <= 12000;
  });

  const ProcessCountTag = ({ count }: { count?: string }) => {
    if (!count) {
      return null;
    }
    const rate = parseFloat(count);
    if (Number.isNaN(rate)) {
      return null;
    }
    if (rate >= 85) {
      return h(
        resolveComponent('a-tag'),
        {
          color: 'red',
        },
        () => t('app.sysinfo.overview.tag.busy')
      );
    }
    if (rate >= 70) {
      return h(
        resolveComponent('a-tag'),
        {
          color: 'orange',
        },
        () => t('app.sysinfo.overview.tag.normal')
      );
    }
    return h(
      resolveComponent('a-tag'),
      {
        color: 'green',
      },
      () => t('app.sysinfo.overview.tag.free')
    );
  };

  const load = async (silent = false) => {
    if (silent && polling.value) return;

    if (!silent) {
      setLoading(true);
    }
    try {
      polling.value = true;
      const res = await getSysInfoOverviewApi();
      Object.assign(data, res);
      lastUpdatedAt.value = Date.now();
      nowTick.value = lastUpdatedAt.value;
    } catch (err: any) {
      if (!silent) {
        Message.error(err?.message);
      }
    } finally {
      polling.value = false;
      if (!silent) {
        setLoading(false);
      }
    }
  };

  // 定时刷新数据
  let timer: number | null = null;
  const startTimer = () => {
    if (timer) return;
    timer = window.setInterval(() => {
      nowTick.value = Date.now();
      if (document.visibilityState !== 'visible') return;
      load(true);
    }, 5000);
  };

  const stopTimer = () => {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  };

  const handleVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      load(true);
      startTimer();
      return;
    }
    stopTimer();
  };

  const getStorageUsedColor = (rate: number) => {
    if (rate >= 85) {
      return 'red';
    }
    if (rate >= 70) {
      return 'orange';
    }
    return 'green';
  };

  const handleModifyTime = () => {
    if (timeModifyRef.value) {
      timeModifyRef.value.setCurrentTime(data.server_time || '');
      timeModifyRef.value.show();
    }
  };

  const handleSyncTime = async () => {
    if (isSyncingTime.value) return;

    isSyncingTime.value = true;
    syncTimeStatus.value = 'syncing';

    try {
      await syncTimeApi();
      syncTimeStatus.value = 'success';
      await load();

      if (syncStatusTimer) {
        clearTimeout(syncStatusTimer);
      }
      syncStatusTimer = setTimeout(() => {
        syncTimeStatus.value = null;
      }, 3000);
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.overview.sync.failed'));
      syncTimeStatus.value = null;
    } finally {
      isSyncingTime.value = false;
    }
  };

  const handleModifyTimeZone = () => {
    if (timezoneModifyRef.value) {
      timezoneModifyRef.value.setTimeZone(data.server_time_zone || '');
      timezoneModifyRef.value.show();
    }
  };

  const handleClearCache = async () => {
    try {
      await clearMemoryCacheApi();
      Message.success(t('app.sysinfo.overview.clear_cache_success'));
      await load();
    } catch (err: any) {
      Message.error(
        err.message || t('app.sysinfo.overview.clear_cache_failed')
      );
    }
  };

  const handleAutoClearCache = () => {
    if (autoClearCacheRef.value) {
      autoClearCacheRef.value.reset();
      autoClearCacheRef.value.load();
      autoClearCacheRef.value.show();
    }
  };

  const handleCreateSwap = () => {
    if (createSwapModalRef.value) {
      createSwapModalRef.value.reset();
      createSwapModalRef.value.show();
    }
  };

  const handleViewMemoryDetail = () => {
    memoryDetailVisible.value = true;
  };

  const handleDeleteSwap = async () => {
    try {
      if (
        await confirm({
          title: t('app.sysinfo.overview.delete_swap_confirm_title'),
          content: t('app.sysinfo.overview.delete_swap_confirm_content'),
        })
      ) {
        setLoading(true);
        await deleteSwapApi();
        Message.success(t('app.sysinfo.overview.delete_swap_success'));
        await load();
      }
    } catch (err: any) {
      Message.error(
        err.message || t('app.sysinfo.overview.delete_swap_failed')
      );
    } finally {
      setLoading(false);
    }
  };

  onMounted(() => {
    load();
    startTimer();
    document.addEventListener('visibilitychange', handleVisibilityChange);
  });

  onBeforeUnmount(() => {
    stopTimer();
    document.removeEventListener('visibilitychange', handleVisibilityChange);
    if (syncStatusTimer) {
      clearTimeout(syncStatusTimer);
      syncStatusTimer = null;
    }
  });
</script>

<style lang="less" scoped>
  .box {
    width: 940px;
    padding: 0 16px;
    margin: 0 auto;
    border: 1px solid var(--color-border-2);
  }

  .line {
    display: flex;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 12px 40px;
    line-height: 24px;
    border-bottom: 1px solid var(--color-border-2);
    &:last-child {
      border-bottom: none;
    }
  }

  .no-border {
    border-bottom: none;
  }

  .colspan {
    flex: 1;
  }

  .subline {
    display: flex;
    align-items: top;
    justify-content: flex-start;
    width: 100%;
    margin-bottom: 14px;
    &:last-child {
      margin-bottom: 0;
    }
  }

  .col1 {
    width: 160px;
    margin-right: 40px;
    font-size: 14px;
    color: var(--color-text-2);
    text-align: right;
  }

  .col2 {
    width: 160px;
    margin-right: 40px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .col3 {
    width: 50px;
    margin-right: 30px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .col4 {
    min-width: 160px;
  }

  .sync-syncing {
    color: var(--color-text-2);
  }

  .sync-success {
    color: rgb(var(--red-6));
  }
</style>
