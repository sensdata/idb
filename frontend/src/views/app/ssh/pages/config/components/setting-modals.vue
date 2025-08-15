<template>
  <div class="settings-modals-container">
    <a-modal
      :visible="portModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.portModal.save')"
      :cancel-text="$t('app.ssh.portModal.cancel')"
      @cancel="$emit('update:portModalVisible', false)"
      @ok="handlePortSave"
    >
      <template #title>{{ $t('app.ssh.portModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{ $t('app.ssh.portModal.port') }}</span>
          <div class="modal-input-wrapper">
            <a-input v-model="newPort" placeholder="22" />
            <div class="modal-field-description">
              {{ $t('app.ssh.portModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      :visible="listenModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.listenModal.save')"
      :cancel-text="$t('app.ssh.listenModal.cancel')"
      @cancel="$emit('update:listenModalVisible', false)"
      @ok="handleListenSave"
    >
      <template #title>{{ $t('app.ssh.listenModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{
            $t('app.ssh.listenModal.address')
          }}</span>
          <div class="modal-input-wrapper">
            <a-input v-model="newListenAddress" placeholder="0.0.0.0" />
            <div class="modal-field-description">
              {{ $t('app.ssh.listenModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal
      :visible="rootModalVisible"
      :unmount-on-close="true"
      :mask-closable="false"
      :ok-text="$t('app.ssh.rootModal.save')"
      :cancel-text="$t('app.ssh.rootModal.cancel')"
      @cancel="$emit('update:rootModalVisible', false)"
      @ok="handleRootSave"
    >
      <template #title>{{ $t('app.ssh.rootModal.title') }}</template>
      <div class="modal-form-wrapper">
        <div class="modal-form-item">
          <span class="modal-label">{{ $t('app.ssh.rootModal.label') }}</span>
          <div class="modal-input-wrapper">
            <a-switch v-model="newRootEnabled" />
            <div class="modal-field-description">
              {{ $t('app.ssh.rootModal.description') }}
            </div>
          </div>
        </div>
      </div>
    </a-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';

  const props = defineProps<{
    portModalVisible: boolean;
    listenModalVisible: boolean;
    rootModalVisible: boolean;
    port: string;
    listenAddress: string;
    rootEnabled: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:portModalVisible', value: boolean): void;
    (e: 'update:listenModalVisible', value: boolean): void;
    (e: 'update:rootModalVisible', value: boolean): void;
    (e: 'savePort', port: string): void;
    (e: 'saveListen', address: string): void;
    (e: 'saveRoot', enabled: boolean): void;
  }>();

  // 表单值
  const newPort = ref<string>(props.port);
  const newListenAddress = ref<string>(props.listenAddress);
  const newRootEnabled = ref<boolean>(props.rootEnabled);

  // 当属性变化时更新表单值
  watch(
    () => props.port,
    (value) => {
      newPort.value = value;
    }
  );

  watch(
    () => props.listenAddress,
    (value) => {
      newListenAddress.value = value;
    }
  );

  watch(
    () => props.rootEnabled,
    (value) => {
      newRootEnabled.value = value;
    }
  );

  // 弹窗保存处理函数
  const handlePortSave = () => {
    emit('update:portModalVisible', false);
    emit('savePort', newPort.value);
  };

  const handleListenSave = () => {
    emit('update:listenModalVisible', false);
    emit('saveListen', newListenAddress.value);
  };

  const handleRootSave = () => {
    emit('update:rootModalVisible', false);
    emit('saveRoot', newRootEnabled.value);
  };
</script>

<style scoped lang="less">
  .modal-form-wrapper {
    padding: 0 20px;
  }

  .modal-form-item {
    display: flex;
    margin-bottom: 20px;
  }

  .modal-label {
    width: 80px;
    margin-right: 20px;
    color: var(--color-text-1);
    font-weight: 500;
    line-height: 32px;
    text-align: right;
  }

  .modal-input-wrapper {
    display: flex;
    flex: 1;
    flex-direction: column;
  }

  .modal-field-description {
    margin-top: 4px;
    color: var(--color-text-3);
    font-size: 12px;
  }
</style>
