<template>
  <a-drawer
    :visible="visible"
    height="90vh"
    :title="$t('app.script.logs.title')"
    placement="bottom"
    @cancel="handleClose"
  >
    <idb-table
      ref="gridRef"
      :params="params"
      :columns="columns"
      :fetch="getScriptRecordsApi"
      :expandable="expandable"
      @expand="expand"
    >
      <template #expand-row="{ record }: { record: any }">
        <template v-if="expandData[record.id]">
          <a-spin :loading="expandData[record.id].loading">
            <div
              :ref="
                (el) => {
                  expandData[record.id].el = el;
                }
              "
              style="min-height: 30px; max-height: 400px; overflow: auto"
            >
              <a-empty
                v-if="
                  !expandData[record.id].logs.length &&
                  !expandData[record.id].loading
                "
                :description="$t('app.script.logs.no_logs')"
              />
              <template v-else>
                <logs-view :content="expandData[record.id].logs" />
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
  import { SCRIPT_TYPE } from '@/config/enum';

  const { t } = useI18n();
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  const params = reactive({
    type: SCRIPT_TYPE.Local,
    category: '',
    name: '',
  });

  const columns = [
    {
      dataIndex: 'status',
      title: t('app.script.logs.column.status'),
      slotName: 'status',
    },
    {
      dataIndex: 'start_time',
      title: t('app.script.logs.column.start_time'),
      render: ({ record }: { record: any }) => formatTime(record.start_time),
    },
    {
      dataIndex: 'end_time',
      title: t('app.script.logs.column.end_time'),
      render: ({ record }: { record: any }) => formatTime(record.end_time),
    },
    {
      dataIndex: 'cost_time',
      title: t('app.script.logs.column.cost_time'),
    },
  ];

  const expandable = reactive({
    title: t('app.script.logs.column.logs'),
    width: 80,
    expandedRowKeys: [] as number[],
  });
  const expandData = reactive<{
    [key: number]: {
      logs: string;
      loading: boolean;
      el?: any;
    };
  }>({});
  const loadLogs = async (recordId: number) => {
    if (!expandData[recordId]) {
      expandData[recordId] = {
        logs: '',
        loading: true,
      };
    }
    try {
      const res = await getScriptRunLogApi({
        type: params.type,
        category: params.category,
        name: params.name,
        record_id: recordId,
      });
      if (expandable.expandedRowKeys.includes(recordId)) {
        expandData[recordId].logs = res.logs;
        const { el } = expandData[recordId];
        const isAtBottom =
          el && Math.abs(el.scrollTop - el.scrollHeight + el.clientHeight) < 30;
        if (el && isAtBottom) {
          nextTick(() => {
            el.scrollTop = el.scrollHeight;
          });
        }
      }
    } finally {
      expandData[recordId].loading = false;
    }
  };
  const expand = async (rowKey: number) => {
    if (!expandable.expandedRowKeys.includes(rowKey)) {
      expandable.expandedRowKeys.push(rowKey);
      loadLogs(rowKey);
    } else {
      expandable.expandedRowKeys = expandable.expandedRowKeys.filter(
        (item) => item !== rowKey
      );
      Object.assign(expandData[rowKey], {
        logs: [],
        loading: false,
      });
    }
  };

  const show = (newParams: {
    type: SCRIPT_TYPE;
    category: string;
    name: string;
  }) => {
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
