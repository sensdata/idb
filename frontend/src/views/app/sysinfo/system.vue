<template>
  <a-spin :loading="loading">
    <div class="box">
      <!-- 系统信息 -->
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.host_name') }}</div>
        <div class="col2">{{ data.host_name }}</div>
        <div class="col3 col-actions">
          <a-button type="primary" size="mini" @click="handleModifyHostName">
            {{ $t('common.modify') }}
          </a-button>
          <a-button size="mini" @click="handleRefresh">
            {{ $t('common.refresh') }}
          </a-button>
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.fqdn') }}</div>
        <div class="col2">{{ data.fqdn || '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.distribution') }}</div>
        <div class="col2">{{ distributionText }}</div>
      </div>

      <div class="line">
        <div class="col1">{{
          $t('app.sysinfo.system.distribution_version')
        }}</div>
        <div class="col2">{{ distributionVersionText }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.kernel') }}</div>
        <div class="col2">{{ data.kernel || '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.arch') }}</div>
        <div class="col2">{{ data.arch || '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.os') }}</div>
        <div class="col2">{{ data.os || '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.uptime') }}</div>
        <div class="col2">{{ formatSeconds(data.uptime) }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.machine_id') }}</div>
        <div class="col2">{{ data.machine_id || '-' }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.virtual') }}</div>
        <div class="col2">
          {{ virtualText || $t('app.sysinfo.system.not_virtual') }}
        </div>
      </div>
    </div>
  </a-spin>
  <host-name-modify ref="hostNameModifyRef" @ok="handleHostNameUpdateSuccess" />
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoSystemApi, SysInfoSystemRes } from '@/api/sysinfo';
  import useLoading from '@/composables/loading';
  import { formatSeconds } from '@/utils/format';
  import HostNameModify from '@/views/app/sysinfo/components/host-name-modify/index.vue';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const hostNameModifyRef = ref<InstanceType<typeof HostNameModify>>();
  const data = ref<SysInfoSystemRes>({
    arch: '',
    distribution: '',
    distribution_version: '',
    fqdn: '',
    host_name: '',
    kernel: '',
    machine_id: '',
    os: '',
    platform: '',
    uptime: 0,
    virtual: '',
    version: '',
    vertual: '',
  });
  const distributionText = computed(
    () => data.value.distribution || data.value.platform || '-'
  );
  const distributionVersionText = computed(
    () => data.value.distribution_version || data.value.version || '-'
  );
  const virtualText = computed(
    () => data.value.virtual || data.value.vertual || ''
  );

  // 获取系统信息数据
  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoSystemApi();
      data.value = res;
    } catch (err: any) {
      Message.error(err.message || t('app.sysinfo.system.fetch_failed'));
    } finally {
      setLoading(false);
    }
  };

  // 处理主机名修改
  const handleModifyHostName = () => {
    if (hostNameModifyRef.value && data.value.host_name) {
      hostNameModifyRef.value.setHostName(data.value.host_name);
      hostNameModifyRef.value.show();
    }
  };

  // 主机名更新成功后的处理
  const handleHostNameUpdateSuccess = async () => {
    await fetchData();
    Message.success(t('app.sysinfo.system.refresh_after_host_update'));
  };

  const handleRefresh = async () => {
    await fetchData();
    Message.success(t('app.sysinfo.system.refresh_success'));
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
    width: 50px;
    margin-left: 30px;
  }

  .col-actions {
    display: flex;
    gap: 8px;
    width: auto;
  }
</style>
