<template>
  <div class="max-w-2xl">
    <a-form :model="form" :rules="rules" @submit="handleSubmit">
      <a-form-item :label="t('manage.settings.listenIp')" field="bind_ip">
        <a-select
          v-model="form.bind_ip"
          :options="ipOptions"
          :placeholder="t('manage.settings.selectIp')"
          :loading="isIpLoading"
          class="w-40"
        />
      </a-form-item>
      <a-form-item :label="t('manage.settings.listenPort')" field="bind_port">
        <a-input-number
          v-model="form.bind_port"
          :placeholder="t('manage.settings.enterPort')"
          class="w-40"
          hide-button
          :min="1"
          :max="65535"
        />
      </a-form-item>
      <a-form-item :label="t('manage.settings.domain')" field="bind_domain">
        <a-space>
          <a-input
            v-model="form.bind_domain"
            :placeholder="t('manage.settings.enterDomain')"
            class="w-80"
          />
          <a-tooltip :content="t('manage.settings.domainTip')">
            <icon-question-circle />
          </a-tooltip>
        </a-space>
      </a-form-item>
      <a-form-item :label="t('manage.settings.https')" field="https">
        <a-switch
          v-model="isHttpsEnabled"
          :checked-value="'yes'"
          :unchecked-value="'no'"
        />
      </a-form-item>
      <template v-if="isHttpsEnabled === 'yes'">
        <a-form-item
          :label="t('manage.settings.certType')"
          field="https_cert_type"
        >
          <a-radio-group v-model="form.https_cert_type" type="button">
            <a-radio
              v-for="option in certTypeOptions"
              :key="option.value"
              :value="option.value"
            >
              {{ option.label }}
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <template v-if="form.https_cert_type === 'custom'">
          <a-form-item
            :label="t('manage.settings.certPath')"
            field="https_cert_path"
          >
            <FileSelector v-model="form.https_cert_path" />
          </a-form-item>

          <a-form-item
            :label="t('manage.settings.keyPath')"
            field="https_key_path"
          >
            <FileSelector v-model="form.https_key_path" />
          </a-form-item>
        </template>
      </template>
      <a-form-item>
        <a-button type="primary" html-type="submit" :loading="isSubmitting">
          {{ t('manage.settings.save') }}
        </a-button>
      </a-form-item>
    </a-form>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { SelectOptionData } from '@arco-design/web-vue/es/select/interface';
  import FileSelector from '@/components/file/file-selector/index.vue';
  import {
    getSettingsApi,
    updateSettingsApi,
    getAvailableIpsApi,
    type SettingsForm,
  } from '@/api/settings';

  const { t } = useI18n();

  const form = ref<SettingsForm>({
    bind_ip: '',
    bind_port: 80,
    bind_domain: '',
    https: 'no',
    https_cert_type: 'default',
    https_cert_path: '',
    https_key_path: '',
  });

  const rules = computed(() => ({
    bind_ip: [{ required: true, message: t('manage.settings.ipRequired') }],
    bind_port: [
      { required: true, message: t('manage.settings.portRequired') },
      {
        validator: (value: number, callback: (error?: string) => void) => {
          if (value < 1 || value > 65535) {
            callback(t('manage.settings.portInvalid'));
            return;
          }
          callback();
        },
      },
    ],
    https: [{ required: true, message: t('manage.settings.httpsRequired') }],
    https_cert_type: [
      {
        required: form.value.https === 'yes',
        message: t('manage.settings.certTypeRequired'),
      },
    ],
    https_cert_path: [
      {
        required:
          form.value.https === 'yes' && form.value.https_cert_type === 'custom',
        message: t('manage.settings.certPathRequired'),
      },
    ],
    https_key_path: [
      {
        required:
          form.value.https === 'yes' && form.value.https_cert_type === 'custom',
        message: t('manage.settings.keyPathRequired'),
      },
    ],
  }));

  const isHttpsEnabled = computed({
    get: () => form.value.https,
    set: (val) => {
      form.value.https = val;
    },
  });

  const ipOptions = ref<SelectOptionData[]>([
    {
      label: t('manage.settings.allIp'),
      value: '0.0.0.0',
    },
  ]);
  const isIpLoading = ref(false);
  const isSubmitting = ref(false);

  const certTypeOptions = computed(() => [
    { label: t('manage.settings.defaultCert'), value: 'default' },
    { label: t('manage.settings.customCert'), value: 'custom' },
  ]);

  const fetchSettings = async () => {
    try {
      const data = await getSettingsApi();
      form.value = data;
    } catch (error) {
      Message.error(t('manage.settings.loadFailed'));
    }
  };

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const fetchIpOptions = async () => {
    try {
      isIpLoading.value = true;
      const data = await getAvailableIpsApi();
      ipOptions.value = data.ips.map((item) => ({
        label: item.ip === '0.0.0.0' ? t('manage.settings.allIp') : item.ip,
        value: item.ip,
      }));
    } catch (error) {
      Message.error(t('manage.settings.loadIpsFailed'));
    } finally {
      isIpLoading.value = false;
    }
  };

  const handleSubmit = async () => {
    try {
      isSubmitting.value = true;
      const ret = await updateSettingsApi(form.value);
      if (form.value.bind_domain) {
        Message.success(t('manage.settings.successRedirect'));
        setTimeout(() => {
          window.location.href = ret.redirect_url;
        }, 3e3);
      } else {
        Message.success(t('manage.settings.saveSuccess'));
      }
    } catch (error) {
      Message.error(t('manage.settings.saveFailed'));
    } finally {
      isSubmitting.value = false;
    }
  };

  onMounted(() => {
    fetchIpOptions();
    fetchSettings();
  });
</script>
