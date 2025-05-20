<template>
  <a-modal
    :visible="visible"
    :title="$t('app.ssh.rootModal.title')"
    @update:visible="emits('update:visible', $event)"
    @ok="handleSave"
    @cancel="handleCancel"
  >
    <div class="modal-form-wrapper">
      <div class="modal-form-item">
        <div class="modal-label">{{ $t('app.ssh.root.label') }}</div>
        <div class="modal-input-wrapper">
          <a-radio-group v-model="enabledValue">
            <a-radio :value="true">{{ $t('app.ssh.rootModal.allow') }}</a-radio>
            <a-radio :value="false">{{ $t('app.ssh.rootModal.deny') }}</a-radio>
          </a-radio-group>
          <div class="modal-field-description">
            {{ $t('app.ssh.root.description') }}
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
    enabled: {
      type: Boolean,
      required: true,
    },
  });

  const emits = defineEmits(['update:visible', 'save']);

  // Local copy of the enabled state for editing
  const enabledValue = ref(props.enabled);

  // Update local value when prop changes
  watch(
    () => props.enabled,
    (newValue) => {
      enabledValue.value = newValue;
    }
  );

  // Ensure modal is properly updated
  watch(
    () => props.visible,
    (newValue) => {
      if (newValue) {
        // Modal opened, reset form values
        enabledValue.value = props.enabled;
      }
    }
  );

  // Handle save button click
  const handleSave = () => {
    emits('save', enabledValue.value);
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
