<template>
  <div :class="styles.form_section">
    <h3 :class="styles.section_title">{{
      $t('app.service.form.section.execution')
    }}</h3>

    <div :class="styles.form_item_with_help">
      <a-form-item
        field="execStart"
        :label="$t('app.service.form.field.exec_start')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.exec_start.required'),
          },
        ]"
        :class="styles.command_form_item"
      >
        <a-input
          v-model="formData.execStart"
          :placeholder="$t('app.service.form.placeholder.exec_start')"
          allow-clear
        >
          <template #prefix>
            <icon-command />
          </template>
        </a-input>
      </a-form-item>
      <div :class="styles.form_help">
        {{ $t('app.service.form.help.exec_start') }}
      </div>
    </div>

    <div :class="styles.form_row">
      <a-form-item
        field="workingDirectory"
        :label="$t('app.service.form.field.working_directory')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.working_directory.required'),
          },
        ]"
        :class="styles.form_item_half"
      >
        <a-input
          v-model="formData.workingDirectory"
          :placeholder="$t('app.service.form.placeholder.working_directory')"
          allow-clear
        >
          <template #prefix>
            <icon-folder-opened />
          </template>
        </a-input>
      </a-form-item>

      <div :class="[styles.form_item_with_help, styles.form_item_half]">
        <a-form-item
          field="environment"
          :label="$t('app.service.form.field.environment')"
        >
          <a-input-group>
            <a-input
              v-model="formData.environment"
              :placeholder="$t('app.service.form.placeholder.environment')"
              readonly
              @click="openEnvironmentEditor"
            >
              <template #prefix>
                <icon-apps />
              </template>
            </a-input>
            <a-button type="outline" @click="openEnvironmentEditor">
              <template #icon>
                <icon-settings />
              </template>
            </a-button>
          </a-input-group>
        </a-form-item>
        <div :class="styles.form_help">
          {{ $t('app.service.form.help.environment') }}
        </div>
      </div>
    </div>

    <div :class="styles.form_row">
      <a-form-item
        field="user"
        :label="$t('app.service.form.field.user')"
        :class="styles.form_item_half"
      >
        <a-input
          v-model="formData.user"
          :placeholder="$t('app.service.form.placeholder.user')"
          allow-clear
        >
          <template #prefix>
            <icon-user />
          </template>
        </a-input>
      </a-form-item>

      <a-form-item
        field="group"
        :label="$t('app.service.form.field.group')"
        :class="styles.form_item_half"
      >
        <a-input
          v-model="formData.group"
          :placeholder="$t('app.service.form.placeholder.group')"
          allow-clear
        >
          <template #prefix>
            <icon-team />
          </template>
        </a-input>
      </a-form-item>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';

  interface ServiceFormStyles {
    form_section: string;
    section_title: string;
    form_item_with_help: string;
    command_form_item: string;
    form_help: string;
    form_row: string;
    form_item_half: string;
  }

  interface ServiceExecutionForm {
    execStart: string;
    workingDirectory: string;
    environment: string;
    user: string;
    group: string;
  }

  interface Props {
    formModel: ServiceExecutionForm;
    styles: ServiceFormStyles;
  }

  const props = defineProps<Props>();

  const emit = defineEmits<{
    'openEnvironmentEditor': [];
    'update:formModel': [value: ServiceExecutionForm];
  }>();

  // Create a computed proxy for v-model binding
  const formData = computed({
    get: () => props.formModel,
    set: (newValue) => emit('update:formModel', newValue),
  });

  const openEnvironmentEditor = () => {
    emit('openEnvironmentEditor');
  };
</script>

<style scoped>
  .form-section {
    margin-bottom: 24px;
  }

  .section-title {
    margin-bottom: 16px;
    font-size: 16px;
    font-weight: 500;
  }

  .form-row {
    display: flex;
    gap: 16px;
  }

  .form-item-half {
    flex: 1;
    min-width: 0;
  }

  .form-item-with-help {
    margin-bottom: 8px;
  }

  .form-help {
    margin-top: -12px;
    font-size: 12px;
    color: var(--color-text-3);
  }
</style>
