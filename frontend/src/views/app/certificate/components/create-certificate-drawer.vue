<template>
  <a-drawer
    v-model:visible="drawerVisible"
    :title="$t('app.certificate.createGroup')"
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
        <!-- 基本信息 -->
        <div class="form-section">
          <h4 class="section-title">{{ $t('app.certificate.basicInfo') }}</h4>
          <a-row :gutter="16">
            <a-col :span="24">
              <a-form-item
                field="alias"
                :label="$t('app.certificate.form.alias')"
                required
              >
                <a-input
                  v-model="form.alias"
                  :placeholder="$t('app.certificate.form.aliasPlaceholder')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="domain_name"
                :label="$t('app.certificate.form.domainName')"
                required
              >
                <a-input
                  v-model="form.domain_name"
                  :placeholder="
                    $t('app.certificate.form.domainNamePlaceholder')
                  "
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <!-- 证书配置 -->
        <div class="form-section">
          <h4 class="section-title">{{
            $t('app.certificate.certificateConfig')
          }}</h4>
          <a-row :gutter="16">
            <a-col :span="24">
              <a-form-item
                field="email"
                :label="$t('app.certificate.form.email')"
                required
              >
                <a-input
                  v-model="form.email"
                  :placeholder="$t('app.certificate.form.emailPlaceholder')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="key_algorithm"
                :label="$t('app.certificate.form.keyAlgorithm')"
                required
              >
                <a-select
                  v-model="form.key_algorithm"
                  :placeholder="
                    $t('app.certificate.form.keyAlgorithmPlaceholder')
                  "
                >
                  <a-option
                    v-for="option in KEY_ALGORITHM_OPTIONS"
                    :key="option.value"
                    :value="option.value"
                  >
                    {{ option.label }}
                  </a-option>
                </a-select>
              </a-form-item>
            </a-col>
          </a-row>
        </div>

        <!-- 组织信息 -->
        <div class="form-section">
          <h4 class="section-title">{{
            $t('app.certificate.organizationInfo')
          }}</h4>
          <a-row :gutter="16">
            <a-col :span="24">
              <a-form-item
                field="organization"
                :label="$t('app.certificate.form.organization')"
              >
                <a-input
                  v-model="form.organization"
                  :placeholder="
                    $t('app.certificate.form.organizationPlaceholder')
                  "
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="organization_unit"
                :label="$t('app.certificate.form.organizationUnit')"
              >
                <a-input
                  v-model="form.organization_unit"
                  :placeholder="
                    $t('app.certificate.form.organizationUnitPlaceholder')
                  "
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="country"
                :label="$t('app.certificate.form.country')"
              >
                <a-input
                  v-model="form.country"
                  :placeholder="$t('app.certificate.form.countryPlaceholder')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="province"
                :label="$t('app.certificate.form.province')"
              >
                <a-input
                  v-model="form.province"
                  :placeholder="$t('app.certificate.form.provincePlaceholder')"
                />
              </a-form-item>
            </a-col>
            <a-col :span="24">
              <a-form-item
                field="city"
                :label="$t('app.certificate.form.city')"
              >
                <a-input
                  v-model="form.city"
                  :placeholder="$t('app.certificate.form.cityPlaceholder')"
                />
              </a-form-item>
            </a-col>
          </a-row>
        </div>
      </a-form>
    </div>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { ref, computed, watch } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue/es/form';
  import type { CreateGroupRequest } from '@/api/certificate';
  import { KEY_ALGORITHM_OPTIONS, DEFAULT_KEY_ALGORITHM } from '../constants';

  // Props 定义
  interface Props {
    visible: boolean;
    loading?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    loading: false,
  });

  // 事件定义
  const emit = defineEmits<{
    (e: 'update:visible', visible: boolean): void;
    (e: 'ok', form: CreateGroupRequest): void;
  }>();

  const { t } = useI18n();

  // 响应式数据
  const formRef = ref<FormInstance>();

  // 表单数据
  const form = ref<CreateGroupRequest>({
    alias: '',
    domain_name: '',
    email: '',
    organization: '',
    organization_unit: '',
    country: '',
    province: '',
    city: '',
    key_algorithm: DEFAULT_KEY_ALGORITHM,
  });

  // 计算属性
  const drawerVisible = computed({
    get: () => {
      return props.visible;
    },
    set: (value) => {
      emit('update:visible', value);
    },
  });

  // 表单验证状态
  const isFormValid = computed(() => {
    return (
      form.value.alias.trim() !== '' &&
      form.value.domain_name.trim() !== '' &&
      form.value.email.trim() !== '' &&
      form.value.key_algorithm.trim() !== ''
    );
  });

  // 表单验证规则
  const rules = computed(() => ({
    alias: [
      {
        required: true,
        message: t('app.certificate.form.aliasRequired'),
      },
    ],
    domain_name: [
      {
        required: true,
        message: t('app.certificate.form.domainNameRequired'),
      },
    ],
    email: [
      {
        required: true,
        message: t('app.certificate.form.emailRequired'),
      },
      {
        type: 'email' as const,
        message: t('app.certificate.form.emailInvalid'),
      },
    ],
    key_algorithm: [
      {
        required: true,
        message: t('app.certificate.form.keyAlgorithmRequired'),
      },
    ],
  }));

  // 重置表单
  const resetForm = () => {
    form.value = {
      alias: '',
      domain_name: '',
      email: '',
      organization: '',
      organization_unit: '',
      country: '',
      province: '',
      city: '',
      key_algorithm: DEFAULT_KEY_ALGORITHM,
    };
    formRef.value?.resetFields();
  };

  // 处理提交
  const handleSubmit = async () => {
    if (!formRef.value) return;

    try {
      const errors = await formRef.value.validate();
      if (!errors) {
        emit('ok', form.value);
      }
    } catch (error) {
      // 验证失败，不做处理
    }
  };

  // 处理取消
  const handleCancel = () => {
    drawerVisible.value = false;
  };

  // 监听弹窗显示状态
  watch(
    () => props.visible,
    (visible) => {
      if (visible) {
        resetForm();
      }
    }
  );
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

  .form-section {
    margin-bottom: 2.67rem;
  }

  .form-section:last-child {
    margin-bottom: 0;
  }

  .section-title {
    padding-bottom: 0.67rem;
    margin: 0 0 1.33rem 0;
    font-size: 1.33rem;
    font-weight: 600;
    color: var(--color-text-1);
    border-bottom: 1px solid var(--color-border-2);
  }

  :deep(.arco-form-item) {
    margin-bottom: 1.67rem;
  }

  :deep(.arco-form-item-label) {
    margin-bottom: 0.67rem;
    font-weight: 500;
  }

  :deep(.arco-input) {
    height: 3.33rem;
  }

  :deep(.arco-select) {
    height: 3.33rem;
  }

  :deep(.arco-select-view-single) {
    height: 3.33rem;
  }
</style>
