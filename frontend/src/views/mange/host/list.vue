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
    <template #cpu="{ record }: { record: HostEntity }">
      <a-progress :percent="record.cpu_rate" />
    </template>
    <template #memory="{ record }: { record: HostEntity }">
      <a-progress :percent="record.memory_rate" color="#0FC6C2" />
    </template>
  </idb-table>
  <host-create ref="formRef" @ok="reload"></host-create>
</template>

<script lang="ts" setup>
  import { ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { getHostListApi } from '@/api/host';
  import { HostEntity } from '@/entity/host';
  import HostCreate from './components/create.vue';

  const { t } = useI18n();

  const columns = [
    {
      dataIndex: 'name',
      title: t('manage.host.list.column.name'),
      width: 175,
    },
    {
      dataIndex: 'group_name',
      title: t('manage.host.list.column.group_name'),
      width: 130,
      render: ({ record }: { record: HostEntity }) => {
        return record.group?.group_name || '';
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
      // todo
    },
    {
      dataIndex: 'disk',
      title: t('manage.host.list.column.disk'),
      width: 110,
      // todo
    },
    {
      dataIndex: 'network',
      title: t('manage.host.list.column.network'),
      width: 160,
      // todo
    },
  ];

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
