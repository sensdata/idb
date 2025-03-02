<template>
  <a-spin :loading="loading">
    <div class="box">
      <!-- 系统信息 -->
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.host_name') }}</div>
        <div class="col2">{{ data.host_name }}</div>
        <div class="col3">
          <a-button type="primary" size="mini" @click="handleModifyHostName">
            {{ $t('common.modify') }}
          </a-button>
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.platform') }}</div>
        <div class="col2">{{ data.platform }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.version') }}</div>
        <div class="col2">{{ data.version }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.kernel') }}</div>
        <div class="col2">{{ data.kernel }}</div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.system.virtual') }}</div>
        <div class="col2">
          {{ data.vertual || $t('app.sysinfo.system.not_virtual') }}
        </div>
      </div>
    </div>
  </a-spin>
  <host-name-modify ref="hostNameModifyRef" @ok="handleHostNameUpdateSuccess" />
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoSystemApi, SysInfoSystemRes } from '@/api/sysinfo';
  import useLoading from '@/hooks/loading';
  import HostNameModify from '@/views/app/sysinfo/components/host-name-modify/index.vue';

  const { loading, setLoading } = useLoading(true);
  const hostNameModifyRef = ref<InstanceType<typeof HostNameModify>>();
  const data = ref<SysInfoSystemRes>({
    host_name: '',
    kernel: '',
    platform: '',
    version: '',
    vertual: '',
  });

  // 获取系统信息数据
  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoSystemApi();
      data.value = res;
    } catch (err: any) {
      Message.error(err.message || 'Failed to fetch system information');
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
  const handleHostNameUpdateSuccess = () => {
    // 立即刷新数据
    fetchData();
  };

  onMounted(() => {
    fetchData();
  });
</script>

<style scoped lang="less">
  .box {
    width: 940px;
    margin: 0 auto;
    padding: 0 16px;
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
    color: var(--color-text-2);
    font-size: 14px;
    text-align: right;
  }

  .col2 {
    flex: 1;
    min-width: 200px;
    color: var(--color-text-1);
    font-size: 14px;
  }

  .col3 {
    width: 50px;
    margin-left: 30px;
  }
</style>
