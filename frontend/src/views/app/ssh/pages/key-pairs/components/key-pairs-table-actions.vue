<template>
  <div class="operation">
    <a-button
      type="text"
      size="small"
      :disabled="record.status !== status.ENABLED"
      @click="handleDownload(record)"
    >
      {{ $t('app.ssh.keyPairs.download') }}
    </a-button>
    <a-button
      type="text"
      size="small"
      @click="handleToggle(record, record.status !== status.ENABLED)"
    >
      {{
        record.status === status.ENABLED
          ? $t('app.ssh.keyPairs.disable')
          : $t('app.ssh.keyPairs.enable')
      }}
    </a-button>
    <a-button
      type="text"
      size="small"
      status="danger"
      @click="handleDelete(record)"
    >
      {{ $t('app.ssh.keyPairs.delete') }}
    </a-button>
  </div>
</template>

<script setup lang="ts">
  import { SSHKeyRecord, SSHKeyStatus } from '@/views/app/ssh/types';

  defineProps<{
    record: SSHKeyRecord;
    status: typeof SSHKeyStatus;
  }>();

  const emit = defineEmits<{
    (e: 'download', record: SSHKeyRecord): void;
    (e: 'toggle', record: SSHKeyRecord, enable: boolean): void;
    (e: 'delete', record: SSHKeyRecord): void;
  }>();

  const handleDownload = (record: SSHKeyRecord) => {
    emit('download', record);
  };

  const handleToggle = (record: SSHKeyRecord, enable: boolean) => {
    emit('toggle', record, enable);
  };

  const handleDelete = (record: SSHKeyRecord) => {
    emit('delete', record);
  };
</script>

<style scoped lang="less">
  .operation {
    display: flex;
    justify-content: center;

    :deep(.arco-btn-size-small) {
      padding-right: 4px;
      padding-left: 4px;
    }
  }
</style>
