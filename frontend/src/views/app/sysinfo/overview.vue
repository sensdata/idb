<template>
  <a-spin :loading="loading">
    <div class="box">
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.server_time') }}</div>
        <div class="col2">{{ formatTime(data.server_time) }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-space>
            <a-button type="primary" size="mini">{{
              $t('app.sysinfo.button.modify')
            }}</a-button>
            <a-button type="primary" size="mini">{{
              $t('app.sysinfo.button.sync_time')
            }}</a-button>
          </a-space>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.server_time_zone') }}</div>
        <div class="col2">{{ data.server_time_zone }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-button type="primary" size="mini">{{
            $t('app.sysinfo.button.modify')
          }}</a-button>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.boot_time') }}</div>
        <div class="col2"> {{ formatTime(data.boot_time) }} </div>
        <div class="col3"></div>
        <div class="col4">
          <a-tag color="blue">{{ $t('app.sysinfo.tag.busy') }}</a-tag>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.run_time') }}</div>
        <div class="col2"> {{ formatSeconds(data.run_time) }} </div>
        <div class="col3"></div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.idle_time') }}</div>
        <div class="col2">{{ formatSeconds(data.idle_time) }}</div>
        <div class="col3"></div>
        <div class="col4">
          <a-tag color="green">{{
            // todo
            $t('app.sysinfo.tag.label_free', { free: '92.41' })
          }}</a-tag>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.cpu_usage') }}</div>
        <div class="col2">{{ data.cpu_usage }}</div>
        <div class="col3"></div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.current_load') }}</div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.tag.count1') }}</div>
            <div class="col3">{{ data.current_load?.process_count1 }}</div>
            <div class="col4">
              <template v-if="data.current_load?.process_count1">
                <a-tag
                  v-if="data.current_load.process_count1 > 50"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.busy') }}</a-tag
                >
                <a-tag
                  v-else-if="data.current_load.process_count1 > 30"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.normal') }}</a-tag
                >
                <a-tag v-else color="green">{{
                  $t('app.sysinfo.tag.free')
                }}</a-tag>
              </template>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.process_count5') }}</div>
            <div class="col3">{{ data.current_load?.process_count5 }}</div>
            <div class="col4">
              <template v-if="data.current_load?.process_count5">
                <a-tag
                  v-if="data.current_load.process_count5 > 50"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.busy') }}</a-tag
                >
                <a-tag
                  v-else-if="data.current_load.process_count5 > 30"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.normal') }}</a-tag
                >
                <a-tag v-else color="green">{{
                  $t('app.sysinfo.tag.free')
                }}</a-tag>
              </template>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.process_count15') }}</div>
            <div class="col3">{{ data.current_load?.process_count15 }}</div>
            <div class="col4">
              <template v-if="data.current_load?.process_count15">
                <a-tag
                  v-if="data.current_load.process_count15 > 50"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.busy') }}</a-tag
                >
                <a-tag
                  v-else-if="data.current_load.process_count15 > 30"
                  color="blue"
                  >{{ $t('app.sysinfo.tag.normal') }}</a-tag
                >
                <a-tag v-else color="green">{{
                  $t('app.sysinfo.tag.free')
                }}</a-tag>
              </template>
            </div>
          </div>
        </div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.memory_usage') }}</div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.memory_total') }}
              <a-tooltip :content="$t('app.sysinfo.memory_total_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip>
            </div>
            <div class="col3">{{ data.memory_usage?.total }}</div>
            <div class="col4"></div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.memory_free') }}</div>
            <div class="col3">{{ data.memory_usage?.free }}</div>
            <div class="col4">
              <a-tag v-if="data.memory_usage?.free_rate" color="gold">{{
                $t('app.sysinfo.tag.leave_unused', {
                  used: data.memory_usage.free_rate + '%',
                })
              }}</a-tag>
            </div>
          </div>
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.memory_used') }}
              <a-tooltip :content="$t('app.sysinfo.memory_used_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip>
            </div>
            <div class="col3">{{ data.memory_usage?.used }}</div>
            <div class="col4"></div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.memory_real_used') }}</div>
            <div class="col3"> {{ data.memory_usage?.real_used }}</div>
            <div class="col4">
              <a-link class="text-sm">{{
                $t('app.sysinfo.button.view_memory')
              }}</a-link>
            </div>
          </div>
          <div class="subline">
            <div class="col2">
              {{ $t('app.sysinfo.memory_buffered') }}
              <!-- <a-tooltip :content="$t('app.sysinfo.memory_buffered_tips')">
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
              {{ $t('app.sysinfo.memory_cached') }}
              <!-- <a-tooltip :content="$t('app.sysinfo.memory_cached_tips')">
                <icon-question-circle-fill
                  class="color-primary cursor-pointer"
                />
              </a-tooltip> -->
            </div>
            <div class="col3">{{ data.memory_usage?.cached }}</div>
            <div class="col4">
              <a-space>
                <a-button type="primary" size="mini">{{
                  $t('app.sysinfo.button.clear_cache')
                }}</a-button>
                <a-button type="primary" size="mini">{{
                  $t('app.sysinfo.button.auto_clear_setting')
                }}</a-button>
              </a-space>
            </div>
          </div>
        </div>
      </div>
      <div class="line no-border">
        <div class="col1">{{ $t('app.sysinfo.virtual_memory') }}</div>
        <div class="col2">
          <!-- todo -->
          {{ $t('app.sysinfo.no_virtual_memory') }}
        </div>
        <div class="col3"></div>
        <div class="col4">
          <a-button type="primary" size="mini">{{
            $t('app.sysinfo.button.create_virtual_memory')
          }}</a-button>
        </div>
      </div>
      <div class="line">
        <div class="col1">({{ $t('app.sysinfo.swap_usage') }})</div>
        <div class="col2">/</div>
        <div class="col3">{{ data.swap_usage?.total }}</div>
        <div class="col4"></div>
      </div>
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.storage') }}</div>
        <div class="colspan mb-6">
          <a-table
            :columns="storageColumns"
            :data="data.storage || []"
            :pagination="false"
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
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import { onMounted, reactive } from 'vue';
  import { formatSeconds, formatTime } from '@/utils/format';
  import useLoading from '@/hooks/loading';
  import { getSysInfoOverviewtApi, SysInfoOverviewRes } from '@/api/sysinfo';

  const { t } = useI18n();

  const storageColumns = [
    {
      title: t('app.sysinfo.storage_mount_point'),
      dataIndex: 'name',
      width: 200,
    },
    {
      title: t('app.sysinfo.storage_total'),
      dataIndex: 'total',
      width: 120,
    },
    {
      title: t('app.sysinfo.storage_used'),
      dataIndex: 'used',
      width: 120,
    },
    {
      title: t('app.sysinfo.storage_free'),
      dataIndex: 'free',
      width: 120,
    },
    {
      title: t('app.sysinfo.storage_used_rate'),
      dataIndex: 'used_rate',
      slotName: 'rate',
      width: 120,
    },
  ];
  const data = reactive<Partial<SysInfoOverviewRes>>({
    server_time: '2024-12-12 23:02:03',
    server_time_zone: 'Asia/Shanghai',
    boot_time: '2024-12-12 13:49:36',
    run_time: 83211,
    idle_time: 86548,
    cpu_usage: '0.75%',
    current_load: {
      process_count1: 2,
      process_count5: 5,
      process_count15: 6,
    },
    memory_usage: {
      physical: '1.63G',
      kernel: '191.39M',
      total: '1.46G',
      free: '759.88M',
      free_rate: 56.07,
      used: '734.97M',
      used_rate: 43.93,
      buffered: '64.00M',
      cached: '778.39M',
      real_used: '734.97M',
    },
    swap_usage: {
      total: '0B',
      free: '0B',
      free_rate: 100,
      used: '0B',
      used_rate: 0,
    },
    storage: [
      {
        name: '/',
        total: '39.01G',
        free: '30.38G',
        used: '6.82G',
        used_rate: 18.34,
      },
      {
        name: '/snap/lxd/22923',
        total: '80.00M',
        free: '0B',
        used: '80.00M',
        used_rate: 100,
      },
      {
        name: '/snap/core20/1405',
        total: '62.00M',
        free: '0B',
        used: '62.00M',
        used_rate: 100,
      },
      {
        name: '/boot/efi',
        total: '196.89M',
        free: '190.84M',
        used: '6.05M',
        used_rate: 3.07,
      },
      {
        name: '/snap/snapd/23258',
        total: '44.38M',
        free: '0B',
        used: '44.38M',
        used_rate: 100,
      },
      {
        name: '/snap/core20/2434',
        total: '63.75M',
        free: '0B',
        used: '63.75M',
        used_rate: 100,
      },
      {
        name: '/snap/lxd/31333',
        total: '89.50M',
        free: '0B',
        used: '89.50M',
        used_rate: 100,
      },
    ],
  });

  const getStorageUsedColor = (rate: number) => {
    if (rate >= 80) {
      return 'blue';
    }
    if (rate >= 60) {
      return 'cyan';
    }
    return 'green';
  };

  const { loading, setLoading } = useLoading(false);

  const load = async () => {
    setLoading(true);
    try {
      const res = await getSysInfoOverviewtApi();
      Object.assign(data, res);
    } finally {
      setLoading(false);
    }
  };

  onMounted(() => {
    load();
  });
</script>

<style lang="less" scoped>
  .box {
    border: 1px solid var(--color-border-2);
    padding: 0 16px;
    width: 940px;
    margin: 0 auto;
  }
  .line {
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
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
    width: 100%;
    display: flex;
    justify-content: flex-start;
    align-items: top;
    margin-bottom: 14px;
    &:last-child {
      margin-bottom: 0;
    }
  }
  .col1 {
    width: 160px;
    text-align: right;
    color: var(--color-text-2);
    font-size: 14px;
    margin-right: 40px;
  }
  .col2 {
    width: 160px;
    margin-right: 40px;
    color: var(--color-text-1);
    font-size: 14px;
  }
  .col3 {
    width: 50px;
    margin-right: 30px;
    color: var(--color-text-1);
    font-size: 14px;
  }
  .col4 {
    min-width: 160px;
  }
</style>
