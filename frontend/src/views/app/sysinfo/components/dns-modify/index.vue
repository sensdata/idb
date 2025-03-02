<template>
  <a-modal
    v-model:visible="visible"
    :title="$t('app.sysinfo.dnsModify.title')"
    :ok-loading="loading"
    :width="500"
    @before-ok="handleBeforeOk"
  >
    <a-form
      ref="formRef"
      :model="formState"
      :label-col-props="{ span: 5 }"
      :wrapper-col-props="{ span: 18 }"
    >
      <a-form-item
        :label="$t('app.sysinfo.dnsModify.servers')"
        field="servers"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.dnsModify.serversRequired'),
          },
        ]"
      >
        <div class="w-full">
          <div
            v-for="(server, index) in formState.servers"
            :key="index"
            class="dns-server-item"
          >
            <a-input
              v-model="formState.servers[index]"
              :placeholder="$t('app.sysinfo.dnsModify.serverPlaceholder')"
              class="dns-server-input"
            />
            <a-button type="text" status="danger" @click="removeServer(index)">
              <template #icon>
                <icon-delete />
              </template>
            </a-button>
          </div>
          <div class="add-server">
            <a-button type="outline" size="mini" @click="addServer">
              <template #icon>
                <icon-plus />
              </template>
              {{ $t('app.sysinfo.dnsModify.addServer') }}
            </a-button>
          </div>
        </div>
      </a-form-item>
      <a-form-item
        :label="$t('app.sysinfo.dnsModify.timeout')"
        field="timeout"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.dnsModify.timeoutRequired'),
          },
          {
            validator: validateTimeout,
            message: $t('app.sysinfo.dnsModify.timeoutInvalid'),
          },
        ]"
      >
        <a-input-number
          v-model="formState.timeout"
          class="w-40"
          :min="1"
          :max="60"
          :placeholder="$t('app.sysinfo.dnsModify.timeoutPlaceholder')"
        />
        <span class="ml-2">{{ $t('app.sysinfo.dnsModify.seconds') }}</span>
      </a-form-item>
      <a-form-item
        :label="$t('app.sysinfo.dnsModify.retry')"
        field="retry"
        :rules="[
          {
            required: true,
            message: $t('app.sysinfo.dnsModify.retryRequired'),
          },
          {
            validator: validateRetry,
            message: $t('app.sysinfo.dnsModify.retryInvalid'),
          },
        ]"
      >
        <a-input-number
          v-model="formState.retry"
          class="w-40"
          :min="1"
          :max="10"
          :placeholder="$t('app.sysinfo.dnsModify.retryPlaceholder')"
        />
        <span class="ml-2">{{ $t('app.sysinfo.dnsModify.times') }}</span>
      </a-form-item>
    </a-form>
  </a-modal>
</template>

<script lang="ts" setup>
  import { reactive, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import { Message } from '@arco-design/web-vue';
  import type { FormInstance } from '@arco-design/web-vue';
  import useVisible from '@/hooks/visible';
  import useLoading from '@/hooks/loading';
  import { updateDNSApi, UpdateDNSParams } from '@/api/sysinfo';

  const emit = defineEmits(['ok']);
  const { t } = useI18n();
  const { visible, show, hide } = useVisible();
  const { loading, showLoading, hideLoading } = useLoading();

  const formRef = ref<FormInstance>();
  const formState = reactive<UpdateDNSParams>({
    servers: [''],
    timeout: 5,
    retry: 2,
  });

  // 验证超时时间
  const validateTimeout = (
    value: number,
    callback: (error?: string) => void
  ) => {
    if (value < 1 || value > 60) {
      callback(t('app.sysinfo.dnsModify.timeoutInvalid'));
      return;
    }
    callback();
  };

  // 验证重试次数
  const validateRetry = (value: number, callback: (error?: string) => void) => {
    if (value < 1 || value > 10) {
      callback(t('app.sysinfo.dnsModify.retryInvalid'));
      return;
    }
    callback();
  };

  // 添加DNS服务器
  const addServer = () => {
    formState.servers.push('');
  };

  // 删除DNS服务器
  const removeServer = (index: number) => {
    if (formState.servers.length > 1) {
      formState.servers.splice(index, 1);
    } else {
      Message.warning(t('app.sysinfo.dnsModify.atLeastOneServer'));
    }
  };

  // 验证IP地址格式
  const validateIpAddress = (ip: string): boolean => {
    const ipRegex =
      /^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$/;
    return ipRegex.test(ip);
  };

  // 提交前验证
  const handleBeforeOk = async (done: (closed: boolean) => void) => {
    try {
      await formRef.value?.validate();

      // 验证所有DNS服务器格式
      const invalidServers = formState.servers.filter(
        (server) => !validateIpAddress(server)
      );
      if (invalidServers.length > 0) {
        Message.error(t('app.sysinfo.dnsModify.invalidIpFormat'));
        done(false);
        return;
      }

      // 过滤空值
      formState.servers = formState.servers.filter(
        (server) => server.trim() !== ''
      );

      if (formState.servers.length === 0) {
        Message.error(t('app.sysinfo.dnsModify.atLeastOneServer'));
        done(false);
        return;
      }

      showLoading();
      try {
        await updateDNSApi({
          servers: formState.servers,
          timeout: formState.timeout,
          retry: formState.retry,
        });

        Message.success(t('app.sysinfo.dnsModify.updateSuccess'));
        emit('ok');
        done(true);
      } catch (err: any) {
        Message.error(err.message || t('app.sysinfo.dnsModify.updateFailed'));
        done(false);
      } finally {
        hideLoading();
      }
    } catch (err) {
      done(false);
    }
  };

  // 设置DNS数据
  const setDNSData = (data: {
    servers: string[];
    timeout: number;
    retry: number;
  }) => {
    formState.servers = [...data.servers];
    formState.timeout = data.timeout;
    formState.retry = data.retry;
  };

  defineExpose({
    show,
    hide,
    setDNSData,
  });
</script>

<style scoped>
  .dns-server-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }

  .dns-server-input {
    flex: 1;
    margin-right: 8px;
  }

  .add-server {
    margin-top: 8px;
  }
</style>
