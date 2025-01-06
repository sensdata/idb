<template>
  <idb-table ref="gridRef" :columns="columns" :fetch="getHostListApi">
    <template #leftActions>
      <a-button type="primary" @click="handleCreate">
        <template #icon>
          <icon-plus />
        </template>
        {{ $t('manage.host.list.action.add') }}
      </a-button>
    </template>
    <template #name="{ record }: { record: HostEntity }">
      <div>{{ record.addr }}</div>
      <div>
        <template v-if="record.is_default">
          <span>{{ $t('manage.host.list.name_local') }}</span>
          <span class="green-6"
            >({{ $t('manage.host.list.name_default') }})</span
          >
        </template>
        <template v-else>
          <span>{{ record.name }}</span>
        </template>
      </div>
    </template>
    <template #cpu="{ record }: { record: HostEntity }">
      <a-progress class="inline-progress" :percent="record.cpu_rate" />
    </template>
    <template #memory="{ record }: { record: HostEntity }">
      <a-progress
        class="inline-progress bg-cyan-6"
        :percent="record.memory_rate"
        color="#0FC6C2"
      />
    </template>
    <template #disk="{ record }: { record: HostEntity }">
      <a-progress
        class="inline-progress bg-green-6"
        :percent="record.disk_rate"
        color="#0FC6C2"
      />
    </template>
    <template #operation="{ record }: { record: HostEntity }">
      <idb-dropdown-operation :options="getOperationOptions(record)" />
    </template>
  </idb-table>
  <host-create ref="formRef" @ok="reload"></host-create>
  <ssh-terminal ref="termRef" />
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { useRouter } from 'vue-router';
  import { getHostListApi } from '@/api/host';
  import { HostEntity } from '@/entity/Host';
  import { DEFAULT_APP_ROUTE_NAME } from '@/router/constants';
  import HostCreate from './components/create.vue';
  import SshTerminal from './components/ssh-terminal.vue';

  const { t } = useI18n();
  const router = useRouter();
  const termRef = ref<InstanceType<typeof SshTerminal>>();

  const columns = [
    {
      dataIndex: 'name',
      title: t('manage.host.list.column.name'),
      width: 150,
      slotName: 'name',
    },
    {
      dataIndex: 'group_name',
      title: t('manage.host.list.column.group_name'),
      width: 125,
      render: ({ record }: { record: HostEntity }) => {
        return record.group
          ? record.group?.group_name
          : t('manage.host.list.group_default');
      },
    },
    {
      dataIndex: 'ctrl_end',
      title: t('manage.host.list.column.ctrl_end'),
      width: 120,
      // todo
    },
    {
      dataIndex: 'safe',
      title: t('manage.host.list.column.safe'),
      width: 120,
      // todo
    },
    {
      dataIndex: 'cpu',
      title: t('manage.host.list.column.cpu'),
      width: 110,
      slotName: 'cpu',
    },
    {
      dataIndex: 'memory',
      title: t('manage.host.list.column.memory'),
      width: 110,
      slotName: 'memory',
    },
    {
      dataIndex: 'disk',
      title: t('manage.host.list.column.disk'),
      width: 110,
      slotName: 'disk',
      // todo
    },
    {
      dataIndex: 'network',
      title: t('manage.host.list.column.network'),
      width: 160,
      // todo
    },
    {
      dataIndex: 'operation',
      title: t('common.table.operation'),
      width: 100,
      slotName: 'operation',
    },
  ];

  const getOperationOptions = (record: HostEntity) => {
    return [
      {
        text: t('manage.host.list.operation.goto'),
        click: () => {
          router.push({
            name: DEFAULT_APP_ROUTE_NAME,
            query: { id: record.id },
          });
        },
      },
      {
        text: t('manage.host.list.operation.setting'),
        click: () => {
          console.log('setting', record);
        },
      },
      {
        text: t('manage.host.list.operation.sshTerminal'),
        click: () => {
          termRef.value?.show(record.id);
        },
      },
    ];
  };

  const gridRef = ref();
  const reload = () => {
    gridRef.value?.reload();
  };

  const formRef = ref<InstanceType<typeof HostCreate>>();
  const handleCreate = () => {
    const form = formRef.value;
    form?.reset();
    form?.loadOptions();
    form?.show();
  };
</script>

<style scoped>
  .inline-progress :deep(.arco-progress-line-text, .arco-progress-steps-text) {
    min-width: 0;
    margin-left: 10px;
  }
</style>
