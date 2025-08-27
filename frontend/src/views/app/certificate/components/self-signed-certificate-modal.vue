<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.generateSelfSigned')"
    :width="700"
    :footer="true"
    @cancel="handleCancel"
  >
    <template #footer>
      <div class="drawer-footer">
        <a-button @click="handleCancel">
          {{ $t('common.cancel') }}
        </a-button>
        <a-button
          type="primary"
          :loading="loading"
          :disabled="!isFormValid"
          @click="handleSubmit"
        >
          {{ $t('common.confirm') }}
        </a-button>
      </div>
    </template>
    <div class="drawer-content">
      <a-form
        ref="formRef"
        :model="form"
        :rules="rules"
        layout="vertical"
        @submit="handleSubmit"
      >
        <a-form-item
          field="alias"
          :label="$t('app.certificate.form.alias')"
          required
        >
          <a-input
            v-model="form.alias"
            :placeholder="$t('app.certificate.form.aliasPlaceholder')"
            disabled
          />
        </a-form-item>

        <a-row :gutter="16">
          <a-col :span="12">
            <a-form-item
              field="expire_value"
              :label="$t('app.certificate.form.expireValue')"
              required
            >
              <a-input-number
                v-model="form.expire_value"
                :min="1"
                :max="9999"
                :placeholder="$t('app.certificate.form.expireValuePlaceholder')"
              />
            </a-form-item>
          </a-col>
          <a-col :span="12">
            <a-form-item
              field="expire_unit"
              :label="$t('app.certificate.form.expireUnit')"
              required
            >
              <a-select
                v-model="form.expire_unit"
                :placeholder="$t('app.certificate.form.expireUnitPlaceholder')"
              >
                <a-option value="day">{{
                  $t('app.certificate.form.days')
                }}</a-option>
                <a-option value="year">{{
                  $t('app.certificate.form.years')
                }}</a-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>

        <a-form-item
          field="alt_domains"
          :label="$t('app.certificate.form.altDomains')"
        >
          <a-textarea
            v-model="form.alt_domains"
            :placeholder="$t('app.certificate.form.altDomainsPlaceholder')"
            :rows="4"
          />
          <template #help>
            {{ $t('app.certificate.form.altDomainsHelp') }}
          </template>
        </a-form-item>

        <a-form-item field="alt_ips" :label="$t('app.certificate.form.altIPs')">
          <a-textarea
            v-model="form.alt_ips"
            :placeholder="$t('app.certificate.form.altIPsPlaceholder')"
            :rows="4"
          />
          <template #help>
            {{ $t('app.certificate.form.altIPsHelp') }}
          </template>
        </a-form-item>

        <a-form-item field="is_ca" :label="$t('app.certificate.form.isCA')">
          <a-switch v-model="form.is_ca" />
          <template #help>
            {{ $t('app.certificate.form.isCAHelp') }}
          </template>
        </a-form-item>
      </a-form>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { SelfSignedRequest } from '@/api/certificate';

  // Props 定义
  interface Props {
    visible: boolean;
    alias: string;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'ok', form: SelfSignedRequest): void;
  }>();

  const { t } = useI18n();

  // 响应式数据
  const formRef = ref<FormInstance>();
  const form = ref<SelfSignedRequest>({
    alias: '',
    expire_unit: 'year',
    expire_value: 1,
    alt_domains: '',
    alt_ips: '',
    is_ca: false,
  });

  // 计算属性
  const drawerVisible = computed({
    get: () => props.visible,
    set: (value) => emit('update:visible', value),
  });

  // 表单验证状态
  const isFormValid = computed(() => {
    return (
      form.value.alias.trim() !== '' &&
      form.value.expire_value > 0 &&
      (form.value.expire_unit === 'day' || form.value.expire_unit === 'year') &&
      /^[a-zA-Z0-9_-]+$/.test(form.value.alias)
    );
  });

  // 表单验证规则
  const rules = {
    alias: [
      {
        required: true,
        message: t('app.certificate.form.aliasRequired'),
      },
    ],
    expire_value: [
      {
        required: true,
        message: t('app.certificate.form.expireValueRequired'),
      },
      {
        type: 'number' as const,
        min: 1,
        max: 9999,
        message: t('app.certificate.form.expireValueRange'),
      },
    ],
    expire_unit: [
      {
        required: true,
        message: t('app.certificate.form.expireUnitRequired'),
      },
    ],
    alt_domains: [
      {
        validator: (value: string, callback: any) => {
          if (value) {
            const domains = value.split('\n').filter((d) => d.trim());
            const domainRegex =
              /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;
            const wildcardDomainRegex =
              /^\*\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$/;

            for (const domain of domains) {
              if (
                !domainRegex.test(domain) &&
                !wildcardDomainRegex.test(domain)
              ) {
                callback(
                  t('app.certificate.form.altDomainsInvalid', { domain })
                );
                return;
              }
            }
          }
          callback();
        },
      },
    ],
    alt_ips: [
      {
        validator: (value: string, callback: any) => {
          if (value) {
            const ips = value.split('\n').filter((ip) => ip.trim());
            const ipv4Regex =
              /^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
            const ipv6Regex = /^(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}$/;

            for (const ip of ips) {
              if (!ipv4Regex.test(ip) && !ipv6Regex.test(ip)) {
                callback(t('app.certificate.form.altIPsInvalid', { ip }));
                return;
              }
            }
          }
          callback();
        },
      },
    ],
  };

  // 重置表单
  const resetForm = () => {
    form.value = {
      alias: props.alias,
      expire_unit: 'year',
      expire_value: 1,
      alt_domains: '',
      alt_ips: '',
      is_ca: false,
    };
    formRef.value?.resetFields();
  };

  // 处理提交
  const handleSubmit = async () => {
    try {
      const errors = await formRef.value?.validate();
      if (!errors) {
        // 验证通过，没有错误
        emit('ok', { ...form.value });
      }
    } catch (error) {
      console.error('Form validation failed:', error);
    }
  };

  // 处理取消
  const handleCancel = () => {
    drawerVisible.value = false;
  };

  // 监听弹窗显示状态和alias变化
  watch([() => props.visible, () => props.alias], ([visible, alias]) => {
    if (visible) {
      resetForm();
      form.value.alias = alias;
    }
  });
</script>

<style scoped>
  .drawer-content {
    height: 100%;
    padding: 1.67rem;
    overflow-y: auto;
  }

  .drawer-footer {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    padding: 1.33rem 0;
  }

  :deep(.arco-form-item) {
    margin-bottom: 2rem;
  }

  :deep(.arco-form-item-label) {
    margin-bottom: 0.67rem;
    font-weight: 500;
  }

  :deep(.arco-form-item-help) {
    margin-top: 0.33rem;
    font-size: 1rem;
    color: var(--color-text-3);
  }

  :deep(.arco-input) {
    height: 3.33rem;
  }

  :deep(.arco-input-number) {
    height: 3.33rem;
  }

  :deep(.arco-select) {
    height: 3.33rem;
  }

  :deep(.arco-select-view-single) {
    height: 3.33rem;
  }

  :deep(.arco-textarea) {
    min-height: 8.33rem;
  }
</style>
