<template>
  <a-spin :loading="loading">
    <div class="box">
      <!-- DNS信息 -->
      <div class="line">
        <div class="col1">{{ $t('app.sysinfo.network.dns') }}</div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_servers') }}</div>
            <div class="col3">
              <div v-for="(server, index) in data.dns?.servers" :key="index">
                {{ server }}
              </div>
            </div>
            <div class="col4">
              <a-button type="primary" size="mini" @click="handleModifyDNS">{{
                $t('common.modify')
              }}</a-button>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_timeout') }}</div>
            <div class="col3"
              >{{ data.dns?.timeout
              }}{{ $t('app.sysinfo.network.seconds') }}</div
            >
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_retry') }}</div>
            <div class="col3"
              >{{ data.dns?.retry }}{{ $t('app.sysinfo.network.times') }}</div
            >
          </div>
        </div>
      </div>

      <!-- 网络接口列表 -->
      <div v-for="(network, index) in data.networks" :key="index" class="line">
        <div class="col1">
          {{ $t('app.sysinfo.network.interface') }} {{ index + 1 }}
        </div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.status') }}</div>
            <div class="col3">
              <a-tag :color="network.status === 'up' ? 'green' : 'red'">
                {{
                  network.status === 'up'
                    ? $t('app.sysinfo.network.status_enabled')
                    : $t('app.sysinfo.network.status_disabled')
                }}
              </a-tag>
            </div>
            <div v-if="index === 0" class="col4">
              <a-button
                type="primary"
                size="mini"
                @click="handleModifyNetwork"
                >{{ $t('common.modify') }}</a-button
              >
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.name') }}</div>
            <div class="col3">
              {{ network.name }}
              <span v-if="network.name === 'eth0'">
                ({{ $t('app.sysinfo.network.ethernet') }})
              </span>
              <span v-else-if="network.name === 'lo'">
                ({{ $t('app.sysinfo.network.loopback') }})
              </span>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.proto') }}</div>
            <div class="col3">
              {{
                network.proto === 'dhcp'
                  ? $t('app.sysinfo.network.proto_dhcp')
                  : $t('app.sysinfo.network.proto_static')
              }}
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.mac') }}</div>
            <div class="col3">{{ network.mac }}</div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.ip_info') }}</div>
            <div class="col3">
              <a-table
                :columns="ipColumns"
                :data="[network.address]"
                :pagination="false"
                size="small"
              />
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.traffic') }}</div>
            <div class="col3">
              <a-table
                :columns="trafficColumns"
                :data="[
                  {
                    type: $t('app.sysinfo.network.traffic_tx'),
                    total: network.traffic.tx,
                    speed: network.traffic.tx_speed,
                  },
                  {
                    type: $t('app.sysinfo.network.traffic_rx'),
                    total: network.traffic.rx,
                    speed: network.traffic.rx_speed,
                  },
                ]"
                :pagination="false"
                size="small"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-spin>
</template>

<script lang="ts" setup>
  import { ref, onMounted, onBeforeUnmount } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoNetworkApi, SysInfoNetworkRes } from '@/api/sysinfo';
  import useLoading from '@/hooks/loading';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const data = ref<SysInfoNetworkRes>({
    dns: {
      retry: 0,
      servers: [''],
      timeout: 0,
    },
    networks: [
      {
        address: {
          gate: '',
          ip: '',
          mask: '',
          type: '',
        },
        mac: '',
        name: '',
        proto: '',
        status: '',
        traffic: {
          rx: '',
          rx_bytes: 0,
          rx_speed: '',
          tx: '',
          tx_bytes: 0,
          tx_speed: '',
        },
      },
    ],
  });

  // IP信息表格列定义
  const ipColumns = [
    {
      title: t('app.sysinfo.network.ip_type'),
      dataIndex: 'type',
    },
    {
      title: t('app.sysinfo.network.ip_address'),
      dataIndex: 'ip',
    },
    {
      title: t('app.sysinfo.network.ip_mask'),
      dataIndex: 'mask',
    },
    {
      title: t('app.sysinfo.network.ip_gate'),
      dataIndex: 'gate',
    },
  ];

  // 流量信息表格列定义
  const trafficColumns = [
    {
      title: t('app.sysinfo.network.traffic_type'),
      dataIndex: 'type',
    },
    {
      title: t('app.sysinfo.network.traffic_total'),
      dataIndex: 'total',
    },
    {
      title: t('app.sysinfo.network.traffic_speed'),
      dataIndex: 'speed',
    },
  ];

  // 获取网络信息数据
  const fetchData = async () => {
    try {
      setLoading(true);
      const res = await getSysInfoNetworkApi();
      data.value = res;
    } catch (err: any) {
      Message.error(err.message || 'Failed to fetch network information');
    } finally {
      setLoading(false);
    }
  };

  // 定时刷新数据
  let timer: number | null = null;
  const startTimer = () => {
    timer = window.setInterval(() => {
      fetchData();
    }, 5000);
  };

  const stopTimer = () => {
    if (timer) {
      clearInterval(timer);
      timer = null;
    }
  };

  const handleModifyDNS = () => {
    console.log('handleModifyDNS');
  };

  const handleModifyNetwork = () => {
    console.log('handleModifyNetwork');
  };

  onMounted(() => {
    fetchData();
    startTimer();
  });

  onBeforeUnmount(() => {
    stopTimer();
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
    width: 120px;
    margin-right: 40px;
    color: var(--color-text-1);
    font-size: 14px;
  }

  .col3 {
    flex: 1;
    min-width: 200px;
    color: var(--color-text-1);
    font-size: 14px;
  }

  .col4 {
    width: 50px;
    margin-left: 30px;
  }
</style>
