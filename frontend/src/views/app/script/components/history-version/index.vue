<template>
  <a-drawer
    :visible="visible"
    width="500px"
    :title="$t('app.script.history_version.title')"
    placement="right"
    :footer="false"
    @cancel="handleClose"
  >
    <idb-table
      v-if="visible"
      ref="gridRef"
      :columns="columns"
      :params="params"
      :fetch="getScriptVersionListApi"
    >
      <template #operation="{ record }">
        <div class="operation">
          <a-button type="text" size="small" @click="handleRestore(record)">
            {{ $t('app.script.history_version.restore') }}
          </a-button>
        </div>
      </template>
    </idb-table>
  </a-drawer>
</template>

<script setup lang="ts">
  import { GlobalComponents, reactive, ref, toRaw } from 'vue';
  import { useI18n } from 'vue-i18n';
  import {
    restoreScriptVersionApi,
    getScriptVersionListApi,
  } from '@/api/script';
  import { SCRIPT_TYPE } from '@/config/enum';
  import { useConfirm } from '@/hooks/confirm';
  import { Message } from '@arco-design/web-vue';
  import { formatTime } from '@/utils/format';

  const props = defineProps<{
    type: SCRIPT_TYPE;
  }>();

  const emit = defineEmits(['ok']);

  const { t } = useI18n();
  const visible = ref(false);
  const gridRef = ref<InstanceType<GlobalComponents['IdbTable']>>();

  const params = reactive({
    type: props.type,
    name: '',
    category: '',
  });

  const columns = [
    {
      dataIndex: 'commit_hash',
      title: t('app.script.history_version.commit_hash'),
      render: ({ record }: { record: any }) => record.commit_hash.slice(0, 7),
    },
    {
      dataIndex: 'date',
      title: t('app.script.history_version.date'),
      render: ({ record }: { record: any }) => formatTime(record.date),
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      align: 'center' as const,
      slotName: 'operation',
    },
  ];

  const reload = () => {
    gridRef.value?.reload();
  };

  const { confirm } = useConfirm();

  const handleRestore = async (record: any) => {
    if (await confirm('app.script.history_version.restore_confirm')) {
      await restoreScriptVersionApi({
        ...toRaw(params),
        commit_hash: record.commit_hash,
      });
      Message.success('common.message.operationSuccess');
      reload();
    }
  };

  const setParams = (p: { name: string; category: string }) => {
    Object.assign(params, p);
  };

  const show = () => {
    visible.value = true;
  };

  const handleClose = () => {
    visible.value = false;
    emit('ok');
  };

  defineExpose({
    show,
    setParams,
  });
</script>
