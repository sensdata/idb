<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    :title="t('app.docker.network.create.title')"
    :ok-loading="loading"
    unmount-on-close
    @before-ok="onBeforeOk"
    @cancel="onCancel"
  >
    <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical">
      <a-form-item
        field="name"
        :label="t('app.docker.network.create.form.name')"
      >
        <a-input
          v-model="formState.name"
          :placeholder="t('app.docker.network.create.form.name.placeholder')"
        />
      </a-form-item>
      <a-form-item
        field="driver"
        :label="t('app.docker.network.create.form.driver')"
      >
        <a-select
          v-model="formState.driver"
          :placeholder="t('app.docker.network.create.form.driver.placeholder')"
        >
          <a-option value="bridge">bridge</a-option>
          <a-option value="host">host</a-option>
          <a-option value="macvlan">macvlan</a-option>
        </a-select>
      </a-form-item>
      <a-form-item class="mb-1 -mt-2">
        <a-checkbox v-model="formState.ipv4">IPv4</a-checkbox>
      </a-form-item>
      <template v-if="formState.ipv4">
        <div class="grid grid-cols-2 gap-4">
          <a-form-item
            field="subnet"
            :label="t('app.docker.network.create.form.subnet')"
          >
            <a-input
              v-model="formState.subnet"
              :placeholder="
                t('app.docker.network.create.form.subnet.placeholder')
              "
            />
          </a-form-item>
          <a-form-item
            field="gateway"
            :label="t('app.docker.network.create.form.gateway')"
          >
            <a-input
              v-model="formState.gateway"
              :placeholder="
                t('app.docker.network.create.form.gateway.placeholder')
              "
            />
          </a-form-item>
          <a-form-item
            field="ip_range"
            :label="t('app.docker.network.create.form.ip_range')"
          >
            <a-input
              v-model="formState.ip_range"
              :placeholder="
                t('app.docker.network.create.form.ip_range.placeholder')
              "
            />
          </a-form-item>
        </div>
        <a-form-item :label="t('app.docker.network.create.form.exclude_ip')">
          <div class="w-full">
            <div
              v-for="(item, idx) in formState.aux_address"
              :key="idx"
              class="flex items-center gap-2 mt-2"
            >
              <a-input
                v-model="item.key"
                size="small"
                :placeholder="
                  t('app.docker.network.create.form.exclude_ip.label')
                "
                class="flex-1"
              />
              <a-input
                v-model="item.value"
                size="small"
                :placeholder="t('app.docker.network.create.form.exclude_ip.ip')"
                class="flex-[2]"
              />
              <a-button
                type="text"
                size="small"
                status="danger"
                @click="removeExcludeIp('v4', idx)"
                >{{ t('common.delete') }}</a-button
              >
            </div>
            <a-button
              class="mt-2"
              type="dashed"
              size="mini"
              @click="addExcludeIp('v4')"
              >{{
                t('app.docker.network.create.form.exclude_ip.add')
              }}</a-button
            >
          </div>
        </a-form-item>
      </template>
      <a-form-item class="mb-1 -mt-2">
        <a-checkbox v-model="formState.ipv6">IPv6</a-checkbox>
      </a-form-item>
      <template v-if="formState.ipv6">
        <div class="grid grid-cols-2 gap-4">
          <a-form-item
            field="subnet_v6"
            :label="t('app.docker.network.create.form.subnet_v6')"
          >
            <a-input
              v-model="formState.subnet_v6"
              :placeholder="
                t('app.docker.network.create.form.subnet_v6.placeholder')
              "
            />
          </a-form-item>
          <a-form-item
            field="gateway_v6"
            :label="t('app.docker.network.create.form.gateway_v6')"
          >
            <a-input
              v-model="formState.gateway_v6"
              :placeholder="
                t('app.docker.network.create.form.gateway_v6.placeholder')
              "
            />
          </a-form-item>
          <a-form-item
            field="ip_range_v6"
            :label="t('app.docker.network.create.form.ip_range_v6')"
          >
            <a-input
              v-model="formState.ip_range_v6"
              :placeholder="
                t('app.docker.network.create.form.ip_range_v6.placeholder')
              "
            />
          </a-form-item>
        </div>
        <a-form-item :label="t('app.docker.network.create.form.exclude_ip_v6')">
          <div class="w-full">
            <div
              v-for="(item, idx) in formState.aux_address_v6"
              :key="idx"
              class="flex items-center gap-2 mt-2"
            >
              <a-input
                v-model="item.key"
                size="small"
                :placeholder="
                  t('app.docker.network.create.form.exclude_ip.label')
                "
                class="flex-1"
              />
              <a-input
                v-model="item.value"
                size="small"
                :placeholder="t('app.docker.network.create.form.exclude_ip.ip')"
                class="flex-[2]"
              />
              <a-button
                type="text"
                status="danger"
                @click="removeExcludeIp('v6', idx)"
                >{{ t('common.delete') }}</a-button
              >
            </div>
            <a-button
              class="mt-2"
              type="dashed"
              size="mini"
              @click="addExcludeIp('v6')"
              >{{
                t('app.docker.network.create.form.exclude_ip.add')
              }}</a-button
            >
          </div>
        </a-form-item>
      </template>
      <a-form-item
        field="options"
        :label="t('app.docker.network.create.form.options')"
      >
        <a-textarea
          v-model="formState.optionsText"
          :placeholder="t('app.docker.network.create.form.options.placeholder')"
          :auto-size="{
            minRows: 3,
            maxRows: 6,
          }"
        />
      </a-form-item>
      <a-form-item
        field="labels"
        :label="t('app.docker.network.create.form.labels')"
      >
        <a-textarea
          v-model="formState.labelsText"
          :placeholder="t('app.docker.network.create.form.labels.placeholder')"
          :auto-size="{
            minRows: 3,
            maxRows: 6,
          }"
        />
      </a-form-item>
    </a-form>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref, reactive } from 'vue';
  import { useI18n } from 'vue-i18n';
  import type { FormInstance } from '@arco-design/web-vue';
  import { Message } from '@arco-design/web-vue';
  import { createNetworkApi } from '@/api/docker';

  const emit = defineEmits(['success']);
  const { t } = useI18n();
  const visible = ref(false);
  const loading = ref(false);
  const formRef = ref<FormInstance>();
  const formState = reactive({
    name: '',
    driver: '',
    ipv4: true,
    subnet: '',
    gateway: '',
    ip_range: '',
    aux_address: [] as { key: string; value: string }[],
    ipv6: false,
    subnet_v6: '',
    gateway_v6: '',
    ip_range_v6: '',
    aux_address_v6: [] as { key: string; value: string }[],
    optionsText: '',
    labelsText: '',
  });
  const rules = {
    name: [
      {
        required: true,
        message: t('app.docker.network.create.form.name.required'),
      },
    ],
    driver: [
      {
        required: true,
        message: t('app.docker.network.create.form.driver.required'),
      },
    ],
  };
  const addExcludeIp = (type: 'v4' | 'v6') => {
    if (type === 'v4') {
      formState.aux_address.push({ key: '', value: '' });
    } else {
      formState.aux_address_v6.push({ key: '', value: '' });
    }
  };
  const removeExcludeIp = (type: 'v4' | 'v6', idx: number) => {
    if (type === 'v4') {
      formState.aux_address.splice(idx, 1);
    } else {
      formState.aux_address_v6.splice(idx, 1);
    }
  };
  const show = () => {
    visible.value = true;
    formState.name = '';
    formState.driver = '';
    formState.ipv4 = true;
    formState.subnet = '';
    formState.gateway = '';
    formState.ip_range = '';
    formState.aux_address = [];
    formState.ipv6 = false;
    formState.subnet_v6 = '';
    formState.gateway_v6 = '';
    formState.ip_range_v6 = '';
    formState.aux_address_v6 = [];
    formState.optionsText = '';
    formState.labelsText = '';
    formRef.value?.resetFields();
  };
  const hide = () => {
    visible.value = false;
    loading.value = false;
  };
  const onBeforeOk = async () => {
    const errors = await formRef.value?.validate();
    if (errors) {
      return false;
    }
    try {
      loading.value = true;
      await createNetworkApi({
        name: formState.name,
        driver: formState.driver,
        ipv4: formState.ipv4,
        subnet: formState.subnet,
        gateway: formState.gateway,
        ip_range: formState.ip_range,
        aux_address: formState.aux_address,
        ipv6: formState.ipv6,
        subnet_v6: formState.subnet_v6,
        gateway_v6: formState.gateway_v6,
        ip_range_v6: formState.ip_range_v6,
        aux_address_v6: formState.aux_address_v6,
        options: formState.optionsText.split('\n').filter(Boolean),
        labels: formState.labelsText.split('\n').filter(Boolean),
      });
      Message.success(t('app.docker.network.create.success'));
      emit('success');
      hide();
    } catch (e: any) {
      Message.error(e?.message || t('app.docker.network.create.failed'));
    } finally {
      loading.value = false;
    }
    return true;
  };
  const onCancel = hide;
  defineExpose({ show });
</script>

<style scoped></style>
