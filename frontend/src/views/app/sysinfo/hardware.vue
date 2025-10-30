<template>
  <a-spin :loading="loading">
    <div class="box">
      <!-- CPU信息 -->
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
          <div v-for="(name, index) in data.module_names" :key="index">
            {{ name }}
          </div>
        </div>
      </div>

      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.hardware.memory') }}</div>
        <div class="col2">{{ data.memory }}</div>
      </div>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
  import { ref, onMounted } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoHardwareApi, SysInfoHardwareRes } from '@/api/sysinfo';
  import useLoading from '@/composables/loading';

  const { loading, setLoading } = useLoading(true);
  const data = ref<SysInfoHardwareRes>({
    cpu_count: 0,
    cpu_cores: 0,
    processor: 0,
    module_names: [],
    memory: '',
  });

  // 获取硬件信息数据
  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoHardwareApi();
      data.value = res;
    } catch (err: any) {
      Message.error(err.message || 'Failed to fetch hardware information');
    } finally {
      setLoading(false);
    }
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
</style>
