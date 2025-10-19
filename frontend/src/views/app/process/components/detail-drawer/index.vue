<template>
  <a-drawer
    :width="800"
    :visible="visible"
    :title="$t('app.process.detailDrawer.title', { name: infoRef?.name })"
    unmountOnClose
    :footer="false"
    @cancel="handleCancel"
  >
    <a-spin :loading="loading" style="width: 100%">
      <a-tabs>
        <a-tab-pane key="base" :title="$t('app.process.detailDrawer.baseInfo')">
          <a-descriptions :data="baseData" :column="2" size="large" bordered />
        </a-tab-pane>
        <a-tab-pane
          key="memory"
          :title="$t('app.process.detailDrawer.memoryInfo')"
        >
          <a-descriptions
            :data="memoryData"
            :column="2"
            size="large"
            bordered
          />
        </a-tab-pane>
        <a-tab-pane key="fs" :title="$t('app.process.detailDrawer.fsInfo')">
          <a-table
            :columns="fsColumns"
            :data="infoRef?.open_files"
            :pagination="false"
            size="small"
          />
        </a-tab-pane>
        <a-tab-pane key="env" :title="$t('app.process.detailDrawer.envInfo')">
          <div class="editor-container">
            <CodeEditor
              :model-value="(infoRef?.envs || []).join('\n')"
              :file="{ name: 'env', path: '/tmp/env' }"
              :readonly="true"
            />
          </div>
        </a-tab-pane>
        <a-tab-pane
          key="network"
          :title="$t('app.process.detailDrawer.networkInfo')"
        >
          <a-table
            :columns="networkColumns"
            :data="infoRef?.net_conns"
            :pagination="false"
            size="small"
          />
        </a-tab-pane>
      </a-tabs>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { computed, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLoading from '@/composables/loading';
  import { formatTime } from '@/utils/format';
  import { getProcessDetailApi } from '@/api/process';
  import CodeEditor from '@/components/code-editor/index.vue';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const infoRef = ref();

  const baseData = computed(() => {
    const info = infoRef.value;
    if (!info || !info.basic) {
      return [];
    }

    const basic = info.basic;
    return [
      {
        label: t('app.process.detailDrawer.name'),
        value: basic.name,
      },
      {
        label: t('app.process.detailDrawer.status'),
        value: basic.status,
      },
      {
        label: t('app.process.detailDrawer.pid'),
        value: basic.pid,
      },
      {
        label: t('app.process.detailDrawer.pppid'),
        value: basic.ppid,
      },
      {
        label: t('app.process.detailDrawer.threads'),
        value: basic.threads,
      },
      {
        label: t('app.process.detailDrawer.connections'),
        value: basic.connections,
      },
      {
        label: t('app.process.detailDrawer.diskRead'),
        value: `${(basic.disk_read / 1024 / 1024).toFixed(2)} MB`,
      },
      {
        label: t('app.process.detailDrawer.diskWrite'),
        value: `${(basic.disk_write / 1024 / 1024).toFixed(2)} MB`,
      },
      {
        label: t('app.process.detailDrawer.user'),
        value: basic.user,
      },
      {
        label: t('app.process.detailDrawer.startTime'),
        value: formatTime(basic.create_time * 1000),
      },
      {
        label: t('app.process.detailDrawer.startCommand'),
        value: basic.cmdline,
      },
    ];
  });

  const formatBytes = (bytes: number) => {
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
  };

  const memoryData = computed(() => {
    const memory = infoRef.value?.memory;
    if (!memory) {
      return [];
    }

    return [
      {
        label: t('app.process.detailDrawer.rss'),
        value: formatBytes(memory.rss),
      },
      {
        label: t('app.process.detailDrawer.swap'),
        value: formatBytes(memory.swap),
      },
      {
        label: t('app.process.detailDrawer.vms'),
        value: formatBytes(memory.vms),
      },
      {
        label: t('app.process.detailDrawer.hwm'),
        value: formatBytes(memory.hwm),
      },
      {
        label: t('app.process.detailDrawer.data'),
        value: formatBytes(memory.data),
      },
      {
        label: t('app.process.detailDrawer.stack'),
        value: formatBytes(memory.stack),
      },
      {
        label: t('app.process.detailDrawer.locked'),
        value: formatBytes(memory.locked),
      },
    ];
  });

  const fsColumns = [
    {
      dataIndex: 'path',
      title: t('app.process.detailDrawer.fs.file'),
      width: 600,
    },
  ];

  const networkColumns = [
    {
      dataIndex: 'protocol',
      title: t('app.process.detailDrawer.network.protocol'),
      width: 80,
    },
    {
      dataIndex: 'local_addr',
      title: t('app.process.detailDrawer.network.localAddress'),
      width: 150,
      render: ({ record }: { record: any }) =>
        `${record.local_addr}:${record.local_port}`,
    },
    {
      dataIndex: 'remote_addr',
      title: t('app.process.detailDrawer.network.remoteAddress'),
      width: 150,
      render: ({ record }: { record: any }) =>
        `${record.remote_addr}:${record.remote_port}`,
    },
    {
      dataIndex: 'status',
      title: t('app.process.detailDrawer.network.state'),
      width: 100,
    },
  ];

  const params = ref();
  const setParams = (p: { pid: string }) => {
    params.value = p;
  };

  const load = async () => {
    setLoading(true);
    try {
      const res = await getProcessDetailApi(params.value);
      infoRef.value = res;
    } finally {
      setLoading(false);
    }
  };

  const visible = ref(false);
  const show = () => {
    visible.value = true;
  };
  const hide = () => {
    visible.value = false;
  };

  const handleCancel = () => {
    hide();
  };

  defineExpose({
    show,
    hide,
    setParams,
    load,
  });
</script>

<style scoped>
  .editor-container {
    position: relative;
    display: flex;
    width: 100%;
    height: 200px;
    overflow: hidden;
    border: 1px solid var(--color-border-2);
    border-radius: 8px;
    box-shadow: 0 2px 8px var(--idb-shadow-light);
  }
</style>
