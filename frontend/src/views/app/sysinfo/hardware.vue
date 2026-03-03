<template>
  <a-spin :loading="loading">
    <div class="box">
      <div class="line line-head">
        <div class="col1">{{ $t('app.sysinfo.hardware.updated_at') }}</div>
        <div class="col2">{{ formatTime(lastUpdatedAt || undefined) }}</div>
        <div class="col3">
          <a-button size="mini" @click="handleRefresh">
            {{ $t('common.refresh') }}
          </a-button>
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.cpu_count') }}</div>
        <div class="col2">{{ data.cpu_count }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.cpu_cores') }}</div>
        <div class="col2">{{ data.cpu_cores }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.processor') }}</div>
        <div class="col2">{{ data.processor }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.module_names') }}</div>
        <div class="col2">
          <a-table
            :columns="cpuModelColumns"
            :data="cpuModelRows"
            :pagination="false"
            size="small"
            :scroll="{ x: true }"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.memory') }}</div>
        <div class="col2">{{ data.memory }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.memory_slots') }}</div>
        <div class="col2">{{ data.memory_slots ?? '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.memory_modules') }}</div>
        <div class="col2">
          <a-table
            :columns="memoryModuleColumns"
            :data="data.memory_modules || []"
            :pagination="false"
            size="small"
            :scroll="{ x: true }"
          />
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.disk_count') }}</div>
        <div class="col2">{{ data.disk_count ?? '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.disks') }}</div>
        <div class="col2">
          <a-table
            :columns="diskColumns"
            :data="data.disks || []"
            :pagination="false"
            size="small"
            :scroll="{ x: true }"
          />
        </div>
      </div>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoHardwareApi, SysInfoHardwareRes } from '@/api/sysinfo';
  import useLoading from '@/composables/loading';
  import { formatTime } from '@/utils/format';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const lastUpdatedAt = ref<string | null>(null);

  const emptyData = (): SysInfoHardwareRes => ({
    cpu_count: 0,
    cpu_cores: 0,
    cpu_models: [],
    disk_count: 0,
    disks: [],
    processor: 0,
    memory_modules: [],
    memory_slots: 0,
    memory: '-',
    module_names: [],
  });

  const data = ref<SysInfoHardwareRes>(emptyData());

  const normalizeRows = <T extends object>(rows?: T[]) =>
    (rows || []).map(
      (row) =>
        Object.fromEntries(
          Object.entries(row).map(([k, v]) => [
            k,
            v === undefined || v === null || v === '' ? '-' : v,
          ])
        ) as T
    );

  const applyData = (res: SysInfoHardwareRes) => {
    data.value = {
      ...res,
      cpu_models: normalizeRows(res.cpu_models),
      disks: normalizeRows(res.disks),
      memory_modules: normalizeRows(res.memory_modules),
    };
    lastUpdatedAt.value = res.updated_at || null;
  };

  const cpuModelRows = computed(() => {
    const rows = data.value.cpu_models || [];
    if (rows.length) return rows;
    return (data.value.module_names || []).map((name) => ({
      model: name,
      count: 1,
    }));
  });

  const cpuModelColumns = computed(() => [
    {
      title: t('app.sysinfo.hardware.cpu_model'),
      dataIndex: 'model',
    },
    {
      title: t('app.sysinfo.hardware.cpu_model_count'),
      dataIndex: 'count',
      width: 140,
    },
  ]);

  const memoryModuleColumns = computed(() => [
    {
      title: t('app.sysinfo.hardware.memory_locator'),
      dataIndex: 'locator',
      width: 160,
    },
    {
      title: t('app.sysinfo.hardware.memory_module_size'),
      dataIndex: 'size',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.memory_module_type'),
      dataIndex: 'type',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.memory_module_speed'),
      dataIndex: 'speed',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.memory_module_manufacturer'),
      dataIndex: 'manufacturer',
      width: 140,
    },
    {
      title: t('app.sysinfo.hardware.memory_module_part_number'),
      dataIndex: 'part_number',
      width: 180,
    },
  ]);

  const diskColumns = computed(() => [
    {
      title: t('app.sysinfo.hardware.disk_name'),
      dataIndex: 'name',
      width: 180,
    },
    {
      title: t('app.sysinfo.hardware.disk_model'),
      dataIndex: 'model',
      width: 220,
    },
    {
      title: t('app.sysinfo.hardware.disk_type'),
      dataIndex: 'type',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.disk_health'),
      dataIndex: 'health',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.disk_life'),
      dataIndex: 'life_used',
      width: 120,
    },
    {
      title: t('app.sysinfo.hardware.disk_size'),
      dataIndex: 'size',
      width: 120,
    },
  ]);

  const fetchData = async (): Promise<boolean> => {
    try {
      setLoading(true);
      const res = await getSysInfoHardwareApi();
      applyData(res);
      return true;
    } catch (err: any) {
      if (!lastUpdatedAt.value) {
        data.value = emptyData();
        lastUpdatedAt.value = null;
      }
      Message.error(err.message || t('app.sysinfo.hardware.fetch_failed'));
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    const ok = await fetchData();
    if (ok) {
      Message.success(t('app.sysinfo.hardware.refresh_success'));
    }
  };

  onMounted(() => {
    fetchData();
  });
</script>

<style scoped lang="less">
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
    width: 100%;
    padding: 16px 0;
    border-bottom: 1px solid var(--color-border);
    &:last-child {
      border-bottom: none;
    }
  }

  .col1 {
    width: 120px;
    margin-right: 40px;
    font-size: 14px;
    color: var(--color-text-2);
    text-align: right;
  }

  .col2 {
    flex: 1;
    min-width: 200px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .col3 {
    width: 120px;
    margin-left: 30px;
  }

  .line-head {
    align-items: center;
  }
</style>
