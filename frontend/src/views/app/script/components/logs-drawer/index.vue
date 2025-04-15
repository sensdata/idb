<template>
  <a-drawer
    :visible="visible"
    height="90vh"
    :title="$t('app.script.logs.title')"
    placement="bottom"
    @cancel="handleClose"
  >
    <idb-table
      v-if="visible"
      ref="gridRef"
      row-key="path"
      :params="params"
      :columns="columns"
      :fetch="getScriptRecordsApi"
      :expandable="expandable"
      @expand="expand"
    >
      <template #expand-row="{ record }: { record: any }">
        <template v-if="expandData[record.path]">
          <a-spin :loading="expandData[record.path].loading">
            <div
              :ref="
                (el) => {
                  expandData[record.path].el = el;
                }
              "
              style="min-height: 30px; max-height: 400px; overflow: auto"
            >
              <a-empty
                v-if="
                  !expandData[record.path].content &&
                  !expandData[record.path].loading
                "
                :description="$t('app.script.logs.no_logs')"
              />
              <template v-else>
                <logs-view :content="expandData[record.path].content" />
              </template>
            </div>
          </a-spin>
        </template>
      </template>
      <template #status="{ record }">
        <a-tag :color="record.status === 'success' ? 'green' : 'red'">
          {{ $t(`app.script.logs.status.${record.status}`) }}
        </a-tag>
      </template>
    </idb-table>
  </a-drawer>
</template>

<script setup lang="ts">
  import { GlobalComponents, nextTick, reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { formatTime } from '@/utils/format';
  import { getScriptRecordsApi, getScriptRunLogApi } from '@/api/script';
  import LogsView from '@/components/logs-view/index.vue';

  const { t } = useI18n();
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  const params = reactive({
    path: '',
  });

  const columns = [
    // {
    //   dataIndex: 'status',
    //   title: t('app.script.logs.column.status'),
    //   slotName: 'status',
    // },
    {
      dataIndex: 'path',
      title: t('app.script.logs.column.path'),
    },
    {
      dataIndex: 'created_at',
      title: t('app.script.logs.column.created_at'),
      render: ({ record }: { record: any }) => formatTime(record.created_at),
    },
  ];

  const expandable = reactive({
    title: t('app.script.logs.column.logs'),
    width: 80,
    expandedRowKeys: [] as string[],
  });
  const expandData = reactive<{
    [key: string]: {
      content: string;
      loading: boolean;
      el?: any;
    };
  }>({});
  const loadLogs = async (path: string) => {
    if (!expandData[path]) {
      expandData[path] = {
        content: '',
        loading: true,
      };
    }
    try {
      const res = await getScriptRunLogApi({
        path,
      });
      if (expandable.expandedRowKeys.includes(path)) {
        expandData[path].content = res.content;
        const { el } = expandData[path];
        const isAtBottom =
          el && Math.abs(el.scrollTop - el.scrollHeight + el.clientHeight) < 30;
        if (el && isAtBottom) {
          nextTick(() => {
            el.scrollTop = el.scrollHeight;
          });
        }
      }
    } finally {
      expandData[path].loading = false;
    }
  };
  const expand = async (path: string) => {
    if (!expandable.expandedRowKeys.includes(path)) {
      expandable.expandedRowKeys.push(path);
      loadLogs(path);
    } else {
      expandable.expandedRowKeys = expandable.expandedRowKeys.filter(
        (item) => item !== path
      );
      Object.assign(expandData[path], {
        content: '',
        loading: false,
      });
    }
  };

  const show = (newParams: { path: string }) => {
    Object.assign(params, newParams);
    visible.value = true;
  };

  const handleClose = () => {
    visible.value = false;
  };

  defineExpose({
    show,
  });
</script>
