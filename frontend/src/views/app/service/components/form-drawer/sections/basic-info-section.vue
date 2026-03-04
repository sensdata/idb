<template>
  <div :class="styles.form_section">
    <h3 :class="styles.section_title">{{
      $t('app.service.form.section.basic')
    }}</h3>

    <div :class="styles.form_row">
      <a-form-item
        field="name"
        :label="$t('app.service.form.field.name')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.name.required'),
          },
        ]"
        :class="styles.form_item_half"
      >
        <a-input
          :model-value="formModel.name"
          :placeholder="$t('app.service.form.field.name')"
          allow-clear
          :disabled="readonly"
          @update:model-value="updateField('name', $event)"
        >
          <template #prefix>
            <icon-tag />
          </template>
        </a-input>
      </a-form-item>

      <a-form-item
        field="serviceType"
        :label="$t('app.service.form.field.service_type')"
        :rules="[
          {
            required: true,
            message: $t('app.service.form.validate.service_type.required'),
          },
        ]"
        :class="styles.form_item_half"
      >
        <a-select
          :model-value="formModel.serviceType"
          :disabled="readonly"
          @update:model-value="updateField('serviceType', $event)"
        >
          <a-option
            v-for="serviceType in serviceTypes"
            :key="serviceType.value"
            :value="serviceType.value"
          >
            {{ serviceType.label }}
          </a-option>
        </a-select>
      </a-form-item>
    </div>

    <a-form-item
      field="description"
      :label="$t('app.service.form.field.description')"
      :rules="[
        {
          required: true,
          message: $t('app.service.form.validate.description.required'),
        },
      ]"
    >
      <a-textarea
        :model-value="formModel.description"
        :placeholder="$t('app.service.form.field.description')"
        :auto-size="{ minRows: 2, maxRows: 3 }"
        :disabled="readonly"
        @update:model-value="updateField('description', $event)"
      />
    </a-form-item>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue';
  import { useI18n } from 'vue-i18n';

  const props = defineProps<{
    readonly?: boolean;
    styles: any;
  }>();

  const formModel = defineModel<any>('formModel');

  const { t } = useI18n();

  const readonly = computed(() => Boolean(props.readonly));

  const serviceTypes = computed(() => [
    {
      value: 'simple',
      label: `Simple (${t('app.service.form.service_type.simple')})`,
    },
    {
      value: 'forking',
      label: `Forking (${t('app.service.form.service_type.forking')})`,
    },
    {
      value: 'oneshot',
      label: `Oneshot (${t('app.service.form.service_type.oneshot')})`,
    },
    {
      value: 'notify',
      label: `Notify (${t('app.service.form.service_type.notify')})`,
    },
  ]);

  const updateField = (field: string, value: any) => {
    formModel.value = {
      ...formModel.value,
      [field]: value,
    };
  };
</script>
