<template>
  <a-modal
    :visible="visible"
    :title="$t('app.ssh.listenModal.title')"
    @update:visible="emits('update:visible', $event)"
    @ok="handleSave"
    @cancel="handleCancel"
  >
    <div class="modal-form-wrapper">
      <div class="modal-form-item">
        <div class="modal-label">{{ $t('app.ssh.listen.label') }}</div>
        <div class="modal-input-wrapper">
          <a-input v-model="addressValue" placeholder="0.0.0.0" />
          <div class="modal-field-description">
            {{ $t('app.ssh.listen.description') }}
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { ref, defineProps, defineEmits, watch } from 'vue';

  const props = defineProps({
    visible: {
      type: Boolean,
      required: true,
    },
    address: {
      type: String,
      required: true,
    },
  });

  const emits = defineEmits(['update:visible', 'save']);

  // Local copy of the address for editing
  const addressValue = ref(props.address);

  // Update local value when prop changes
  watch(
    () => props.address,
    (newValue) => {
      addressValue.value = newValue;
    }
  );

  // Ensure modal is properly updated
  watch(
    () => props.visible,
    (newValue) => {
      if (newValue) {
        // Modal opened, reset form values
        addressValue.value = props.address;
      }
    }
  );

  // Handle save button click
  const handleSave = () => {
    emits('save', addressValue.value);
  };

  // Handle cancel button click
  const handleCancel = () => {
    emits('update:visible', false);
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
