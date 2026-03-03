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
              <template v-if="dnsServers.length">
                <div
                  v-for="(server, index) in dnsServers"
                  :key="`${server}-${index}`"
                >
                  {{ server }}
                </div>
              </template>
              <span v-else class="empty-value">
                {{ $t('app.sysinfo.network.none') }}
              </span>
            </div>
            <div class="col4">
              <a-button type="primary" size="mini" @click="handleModifyDNS">{{
                $t('common.modify')
              }}</a-button>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_source') }}</div>
            <div class="col3">
              {{ dnsSourceText }}
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_timeout') }}</div>
            <div class="col3">
              {{
                data.dns?.timeout
                  ? `${data.dns.timeout}${$t('app.sysinfo.network.seconds')}`
                  : $t('app.sysinfo.network.none')
              }}
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.dns_retry') }}</div>
            <div class="col3">
              {{
                data.dns?.retry
                  ? `${data.dns.retry}${$t('app.sysinfo.network.times')}`
                  : $t('app.sysinfo.network.none')
              }}
            </div>
          </div>
        </div>
      </div>

      <div class="line network-filter-line">
        <div class="col1">{{ $t('app.sysinfo.network.filter') }}</div>
        <div class="colspan">
          <a-tabs
            v-model:active-key="activeFilter"
            type="rounded"
            size="small"
            class="network-filter-tabs"
          >
            <a-tab-pane
              v-for="item in filterTabs"
              :key="item.key"
              :title="item.title"
            />
          </a-tabs>
        </div>
      </div>

      <!-- 网络接口列表 -->
      <div
        v-for="(network, index) in filteredNetworks"
        :key="network.name || network.mac || index"
        class="line"
      >
        <div class="col1">
          {{ $t('app.sysinfo.network.interface') }} {{ index + 1 }}
        </div>
        <div class="colspan">
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.status') }}</div>
            <div class="col3">
              <a-tag :color="getStatusColor(network.status)">
                {{ getStatusText(network.status) }}
              </a-tag>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.name') }}</div>
            <div class="col3">
              {{ network.name || '-' }}
              <span v-if="getNetworkTypeText(network.name)">
                ({{ getNetworkTypeText(network.name) }})
              </span>
              <a-button
                v-if="network.name"
                type="text"
                size="mini"
                class="ml-2"
                @click="copyText(network.name)"
              >
                {{ $t('app.sysinfo.network.copy') }}
              </a-button>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.proto') }}</div>
            <div class="col3">
              {{ getProtoText(network.proto) }}
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.mac') }}</div>
            <div class="col3">
              {{ network.mac || '-' }}
              <a-button
                v-if="network.mac"
                type="text"
                size="mini"
                class="ml-2"
                @click="copyText(network.mac)"
              >
                {{ $t('app.sysinfo.network.copy') }}
              </a-button>
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.oper_state') }}</div>
            <div class="col3">{{ formatOperState(network.oper_state) }}</div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.link_type') }}</div>
            <div class="col3">{{ network.link_type || '-' }}</div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.mtu') }}</div>
            <div class="col3">{{ network.mtu || '-' }}</div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.ip_info') }}</div>
            <div class="col3">
              <a-table
                :columns="ipColumns"
                :data="getIpRows(network.address || [])"
                :pagination="false"
                :row-key="getIpRowKey"
                size="small"
              />
            </div>
          </div>
          <div class="subline">
            <div class="col2">{{ $t('app.sysinfo.network.traffic') }}</div>
            <div class="col3">
              <a-table
                :columns="trafficColumns"
                :data="getTrafficRows(network)"
                :pagination="false"
                :row-key="getTrafficRowKey"
                size="small"
              />
            </div>
          </div>
        </div>
      </div>

      <div v-if="!filteredNetworks.length" class="line">
        <div class="col1"></div>
        <div class="colspan empty-network-text">
          {{ $t('app.sysinfo.network.empty') }}
        </div>
      </div>
    </div>
  </a-spin>
  <dns-modify ref="dnsModifyRef" @ok="handleDNSUpdateSuccess" />
</template>

<script lang="ts" setup>
  import { computed, ref, onMounted, onBeforeUnmount, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import { getSysInfoNetworkApi, SysInfoNetworkRes } from '@/api/sysinfo';
  import useLoading from '@/composables/loading';
  import DnsModify from '@/views/app/sysinfo/components/dns-modify/index.vue';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(true);
  const dnsModifyRef = ref<InstanceType<typeof DnsModify>>();
  const polling = ref(false);
  const NETWORK_FILTER_STORAGE_KEY = 'sysinfo.network.activeFilter';
  const FILTER_KEYS = ['all', 'physical', 'loopback', 'virtual'] as const;

  type NetworkItem = SysInfoNetworkRes['networks'][number];
  type NetworkAddress = NetworkItem['address'][number];
  type IpRow = NetworkAddress & { gateText: string };

  const getDefaultData = (): SysInfoNetworkRes => ({
    dns: {
      retry: 0,
      servers: [],
      source: '',
      timeout: 0,
    },
    networks: [],
  });

  const data = ref<SysInfoNetworkRes>(getDefaultData());
  const dnsServers = computed(() => data.value.dns?.servers || []);
  const dnsSourceText = computed(() => {
    const source = data.value.dns?.source;
    if (source === 'systemd-resolved') {
      return t('app.sysinfo.network.dns_source_systemd');
    }
    if (source === 'resolv.conf') {
      return t('app.sysinfo.network.dns_source_resolv_conf');
    }
    return t('app.sysinfo.network.unknown');
  });
  const networks = computed(() => data.value.networks || []);
  const storedFilter = localStorage.getItem(NETWORK_FILTER_STORAGE_KEY);
  const activeFilter = ref(
    FILTER_KEYS.includes(storedFilter as (typeof FILTER_KEYS)[number])
      ? storedFilter!
      : 'physical'
  );

  const getNetworkKind = (name: string) => {
    if (name === 'lo') return 'loopback';
    if (
      /^(eth|enp|eno|ens|em|wlan|wlp|wwan|bond|team|p\d+p\d+|ib)/i.test(name)
    ) {
      return 'physical';
    }
    if (/^(docker|veth|br-|virbr|vmnet|tun|tap|wg|zt)/i.test(name)) {
      return 'virtual';
    }
    return 'virtual';
  };

  const filterTabs = computed(() => [
    { key: 'all', title: t('app.sysinfo.network.filter_all') },
    { key: 'physical', title: t('app.sysinfo.network.filter_physical') },
    { key: 'loopback', title: t('app.sysinfo.network.filter_loopback') },
    { key: 'virtual', title: t('app.sysinfo.network.filter_virtual') },
  ]);

  const filteredNetworks = computed(() => {
    if (activeFilter.value === 'all') return networks.value;
    return networks.value.filter(
      (item) => getNetworkKind(item.name) === activeFilter.value
    );
  });

  // IP信息表格列定义
  const ipColumns = computed(() => [
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
      dataIndex: 'gateText',
    },
  ]);

  // 流量信息表格列定义
  const trafficColumns = computed(() => [
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
  ]);

  const getStatusColor = (status: string) =>
    status === 'up' ? 'rgb(var(--success-6))' : 'rgb(var(--danger-6))';

  const getStatusText = (status: string) =>
    status === 'up'
      ? t('app.sysinfo.network.status_enabled')
      : t('app.sysinfo.network.status_disabled');

  const getProtoText = (proto: string) => {
    if (proto === 'dhcp') {
      return t('app.sysinfo.network.proto_dhcp');
    }
    if (proto === 'static') {
      return t('app.sysinfo.network.proto_static');
    }
    return t('app.sysinfo.network.unknown');
  };

  const formatOperState = (state?: string) => {
    if (!state) return t('app.sysinfo.network.unknown');
    return state;
  };

  const getNetworkTypeText = (name: string) => {
    if (name === 'eth0') return t('app.sysinfo.network.ethernet');
    if (name === 'lo') return t('app.sysinfo.network.loopback');
    return '';
  };

  const getIpRows = (rows: NetworkAddress[]): IpRow[] =>
    rows.map((row) => ({
      ...row,
      gateText: row.gate?.length ? row.gate.join(', ') : '-',
    }));

  const getIpRowKey = (row: IpRow) => `${row.type}-${row.ip}`;

  const getTrafficRows = (network: NetworkItem) => [
    {
      type: t('app.sysinfo.network.traffic_tx'),
      total: network.traffic?.tx || '-',
      speed: network.traffic?.tx_speed || '-',
    },
    {
      type: t('app.sysinfo.network.traffic_rx'),
      total: network.traffic?.rx || '-',
      speed: network.traffic?.rx_speed || '-',
    },
  ];

  const getTrafficRowKey = (row: { type: string }) => row.type;

  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      Message.success(t('app.sysinfo.network.copy_success'));
    } catch {
      Message.error(t('app.sysinfo.network.copy_failed'));
    }
  };

  // 获取网络信息数据
  const fetchData = async (silent = false) => {
    if (silent && polling.value) return;

    try {
      polling.value = true;
      if (!silent) {
        setLoading(true);
      }
      const res = await getSysInfoNetworkApi();
      data.value = res;
    } catch (err: any) {
      if (!silent) {
        Message.error(err.message || 'Failed to fetch network information');
      }
    } finally {
      if (!silent) {
        setLoading(false);
      }
      polling.value = false;
    }
  };

  // 定时刷新数据
  let timer: ReturnType<typeof setInterval> | null = null;
  const startTimer = () => {
    if (timer) return;

    timer = setInterval(() => {
      if (document.visibilityState !== 'visible') return;
      fetchData(true);
    }, 5000);
  };

  const stopTimer = () => {
    if (timer !== null) {
      clearInterval(timer);
      timer = null;
    }
  };

  const handleVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      fetchData(true);
      startTimer();
      return;
    }

    stopTimer();
  };

  // 处理DNS修改
  const handleModifyDNS = () => {
    if (dnsModifyRef.value && data.value.dns) {
      dnsModifyRef.value.setDNSData({
        servers: [...data.value.dns.servers],
        timeout: data.value.dns.timeout,
        retry: data.value.dns.retry,
      });
      dnsModifyRef.value.show();
    }
  };

  // DNS更新成功后的处理
  const handleDNSUpdateSuccess = () => {
    // 立即刷新数据
    fetchData();
  };

  watch(
    activeFilter,
    (value) => {
      localStorage.setItem(NETWORK_FILTER_STORAGE_KEY, value);
    },
    { immediate: true }
  );

  onMounted(() => {
    fetchData();
    startTimer();
    document.addEventListener('visibilitychange', handleVisibilityChange);
  });

  onBeforeUnmount(() => {
    stopTimer();
    document.removeEventListener('visibilitychange', handleVisibilityChange);
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
    align-items: flex-start;
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
    width: 120px;
    margin-right: 40px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .col3 {
    flex: 1;
    min-width: 200px;
    font-size: 14px;
    color: var(--color-text-1);
  }

  .col4 {
    width: 50px;
    margin-left: 30px;
  }

  .network-filter-line {
    align-items: center;
    padding: 12px 0;
  }

  .network-filter-line .col1 {
    line-height: 28px;
  }

  .network-filter-tabs {
    :deep(.arco-tabs-nav) {
      margin: 0;
    }
    :deep(.arco-tabs-nav::before) {
      display: none;
    }
    :deep(.arco-tabs-nav-ink) {
      display: none;
    }
    :deep(.arco-tabs-content) {
      display: none;
    }
  }

  .empty-network-text {
    font-size: 14px;
    color: var(--color-text-3);
  }

  .empty-value {
    color: var(--color-text-3);
  }
</style>
