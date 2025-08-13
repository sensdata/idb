<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('login.forgotPassword.modal.title')"
    :footer="false"
    :width="500"
    :mask-closable="true"
  >
    <div class="forgot-password-content">
      <a-alert type="info" :show-icon="true" class="mb-4">
        {{ $t('login.forgotPassword.modal.message') }}
      </a-alert>

      <div class="reset-instructions">
        <h4>{{ $t('login.forgotPassword.modal.instructionsTitle') }}</h4>
        <ol class="instructions-list">
          <li>{{ $t('login.forgotPassword.modal.step1') }}</li>
          <li>{{ $t('login.forgotPassword.modal.step2') }}</li>
          <li>{{ $t('login.forgotPassword.modal.step3') }}</li>
        </ol>
      </div>

      <div class="command-example">
        <h4>{{ $t('login.forgotPassword.modal.commandTitle') }}</h4>
        <a-typography-paragraph
          :code="true"
          :copyable="true"
          class="command-text"
        >
          idb rst-pass
        </a-typography-paragraph>
      </div>

      <a-alert type="warning" :show-icon="true" class="mb-4">
        {{ $t('login.forgotPassword.modal.note') }}
      </a-alert>

      <div class="modal-footer">
        <a-button type="primary" @click="handleClose">
          {{ $t('login.forgotPassword.modal.understood') }}
        </a-button>
      </div>
    </div>
  </a-modal>
</template>

<script lang="ts" setup>
  import { computed } from 'vue';

  // Props 定义
  interface Props {
    visible: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    visible: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
  }>();

  // 计算属性
  const visible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  const handleClose = () => {
    visible.value = false;
  };
</script>

<style lang="less" scoped>
  .forgot-password-content {
    .reset-instructions {
      margin-bottom: 1.429rem;

      h4 {
        margin-bottom: 0.857rem;
        font-weight: 600;
        color: var(--color-text-1);
      }

      .instructions-list {
        margin: 0;
        padding-left: 1.429rem;

        li {
          margin-bottom: 0.571rem;
          line-height: 1.5;
          color: var(--color-text-2);
        }
      }
    }

    .command-example {
      margin-bottom: 1.429rem;

      h4 {
        margin-bottom: 0.857rem;
        font-weight: 600;
        color: var(--color-text-1);
      }

      .command-text {
        background-color: var(--color-fill-2);
        border: 1px solid var(--color-border-2);
        border-radius: 0.286rem;
        padding: 0.857rem;
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 1rem;
      }
    }

    .modal-footer {
      text-align: center;
      margin-top: 1.429rem;
    }
  }
</style>
