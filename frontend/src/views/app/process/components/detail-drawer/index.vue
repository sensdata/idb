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
            :data="infoRef?.fs"
            :pagination="false"
            size="small"
          />
        </a-tab-pane>
        <a-tab-pane key="env" :title="$t('app.process.detailDrawer.envInfo')">
          <div class="editor-container">
            <CodeEditor
              :model-value="infoRef?.env || ''"
              :file="{ name: 'env', path: '/tmp/env' }"
              :extensions="lightThemeExtensions"
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
            :data="infoRef?.network"
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
  import { getProcessDetailApi } from '@/api/process';
  import CodeEditor from '@/components/code-editor/index.vue';
  import { githubLight } from '@fsegurai/codemirror-theme-github-light';

  const { t } = useI18n();
  const { loading, setLoading } = useLoading(false);

  const infoRef = ref();

  // 添加浅色主题扩展，提供语法高亮颜色
  const lightThemeExtensions = computed(() => [githubLight]);

  const baseData = computed(() => {
    const info = infoRef.value;
    if (!info) {
      return [];
    }

    return [
      {
        label: t('app.process.detailDrawer.name'),
        value: info.name,
      },
      {
        label: t('app.process.detailDrawer.status'),
        value: info.status,
      },
      {
        label: t('app.process.detailDrawer.pid'),
        value: info.pid,
      },
      {
        label: t('app.process.detailDrawer.pppid'),
        value: info.pppid,
      },
      {
        label: t('app.process.detailDrawer.threads'),
        value: info.threads,
      },
      {
        label: t('app.process.detailDrawer.connections'),
        value: info.connections,
      },
      {
        label: t('app.process.detailDrawer.diskRead'),
        value: info.disk_read,
      },
      {
        label: t('app.process.detailDrawer.diskWrite'),
        value: info.disk_write,
      },
      {
        label: t('app.process.detailDrawer.user'),
        value: info.user,
      },
      {
        label: t('app.process.detailDrawer.startTime'),
        value: info.start_time,
      },
      {
        label: t('app.process.detailDrawer.startCommand'),
        value: info.start_command,
      },
    ];
  });

  const memoryData = computed(() => {
    const memory = infoRef.value?.memory;
    if (!memory) {
      return [];
    }

    return [
      {
        label: t('app.process.detailDrawer.rss'),
        value: memory.rss,
      },
      {
        label: t('app.process.detailDrawer.swap'),
        value: memory.swap,
      },
      {
        label: t('app.process.detailDrawer.vms'),
        value: memory.vms,
      },
      {
        label: t('app.process.detailDrawer.hwm'),
        value: memory.hwm,
      },
      {
        label: t('app.process.detailDrawer.data'),
        value: memory.data,
      },
      {
        label: t('app.process.detailDrawer.stack'),
        value: memory.stack,
      },
      {
        label: t('app.process.detailDrawer.locked'),
        value: memory.locked,
      },
    ];
  });

  const fsColumns = [
    {
      dataIndex: 'file',
      title: t('app.process.detailDrawer.fs.file'),
      width: 500,
    },
    {
      dataIndex: 'fd',
      title: t('app.process.detailDrawer.fs.fd'),
      width: 100,
    },
  ];

  const networkColumns = [
    {
      dataIndex: 'localAddres',
      title: t('app.process.detailDrawer.network.localAddress'),
      width: 250,
    },
    {
      dataIndex: 'remoteAddress',
      title: t('app.process.detailDrawer.network.remoteAddress'),
      width: 250,
    },
    {
      dataIndex: 'state',
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
    box-shadow: 0 2px 8px rgb(0 0 0 / 10%);
  }
</style>
