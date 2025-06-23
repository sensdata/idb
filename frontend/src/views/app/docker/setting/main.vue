<template>
  <a-spin :loading="loading" class="w-full">
    <div class="docker-setting">
      <a-card hoverable size="small">
        <div class="flex w-full flex-col gap-4 md:flex-row">
          <div class="flex flex-wrap gap-4">
            <a-tag color="#67c23a">Docker</a-tag>
            <a-tag
              v-if="formState.status === 'Running'"
              color="lime"
              bordered
              class="rounded-file"
            >
              {{ $t('app.docker.setting.status.running') }}
            </a-tag>
            <a-tag v-if="formState.status === 'stopped'" color="gray" bordered>
              {{ $t('app.docker.setting.status.stopped') }}
            </a-tag>
            <a-tag color="blue">
              {{ $t('app.docker.setting.version') }}: {{ formState.version }}
            </a-tag>
          </div>
          <div class="flex items-center">
            <a-button
              v-if="formState.status === 'Running'"
              type="primary"
              size="mini"
              @click="handleOperator('stop')"
            >
              {{ $t('app.docker.setting.status.stop') }}
            </a-button>
            <a-button
              v-if="formState.status === 'Stopped'"
              type="primary"
              size="mini"
              @click="handleOperator('start')"
            >
              {{ $t('app.docker.setting.status.start') }}
            </a-button>
            <a-divider direction="vertical" />
            <a-button
              type="primary"
              size="mini"
              @click="handleOperator('restart')"
            >
              {{ $t('app.docker.setting.status.restart') }}
            </a-button>
          </div>
        </div>
      </a-card>
      <a-card hoverable class="mt-4">
        <div class="form-container">
          <a-radio-group v-model="mode" type="button">
            <a-radio value="form">
              {{ $t('app.docker.setting.mode.form') }}
            </a-radio>
            <a-radio value="file">
              {{ $t('app.docker.setting.mode.file') }}
            </a-radio>
          </a-radio-group>
        </div>
        <a-form
          v-if="mode === 'form'"
          :model="formState"
          class="w-[600px] mt-4"
        >
          <a-form-item
            :label="$t('app.docker.setting.mirror.mirrors')"
            field="registry_mirrors"
          >
            <a-textarea
              v-if="!!formState.registry_mirrors"
              :model-value="formState.registry_mirrors"
              :rows="5"
              disabled
            />
            <a-input
              v-else
              :model-value="$t('app.docker.setting.mirror.empty')"
              disabled
            />
            <a-button
              type="text"
              shape="round"
              class="ml-2"
              @click="handleMirrors"
            >
              <template #icon>
                <icon-settings />
              </template>
            </a-button>
          </a-form-item>
          <a-form-item
            :label="$t('app.docker.setting.registry.registries')"
            field="insecure_registries"
          >
            <a-textarea
              v-if="!!formState.insecure_registries"
              :model-value="formState.insecure_registries"
              :rows="5"
              disabled
            />
            <a-input
              v-else
              :model-value="$t('app.docker.setting.registry.empty')"
              disabled
            />
            <a-button
              type="text"
              shape="round"
              class="ml-2"
              @click="handleRegistries"
            >
              <template #icon>
                <icon-settings />
              </template>
            </a-button>
          </a-form-item>
          <a-form-item label="IPv6" field="ipv6">
            <a-switch
              v-model="formState.ipv6"
              :before-change="handleIPv6"
            ></a-switch>
            <div v-if="formState.ipv6">
              <!-- todo -->
            </div>
          </a-form-item>
          <a-form-item
            :label="$t('app.docker.setting.log.title')"
            field="hasLogOption"
          >
            <a-switch
              v-model="formState.log_option_show"
              :before-change="handleLogOption"
            ></a-switch>
            <a-button
              type="text"
              shape="round"
              class="ml-2"
              @click="handleLogOption"
            >
              <template #icon>
                <icon-settings />
              </template>
            </a-button>
            <div v-if="formState.log_option_show">
              <a-tag>
                {{ $t('app.docker.setting.log.max_size') }}:
                {{ formState.log_max_size }}
              </a-tag>
              <a-tag style="margin-left: 5px">
                {{ $t('app.docker.setting.log.max_file') }}:
                {{ formState.log_max_file }}
              </a-tag>
            </div>
          </a-form-item>
          <a-form-item label="iptables" field="iptables">
            <a-switch
              v-model="formState.iptables"
              :before-change="($event) => handleSaveField('iptables', $event)"
            ></a-switch>
          </a-form-item>
          <a-form-item label="Live restore" field="live_restore">
            <a-switch
              v-model="formState.live_restore"
              :disabled="formState.is_swarm"
              :before-change="
                ($event) => handleSaveField('live_restore', $event)
              "
            ></a-switch>
          </a-form-item>
          <a-form-item label="cgroup driver" field="cgroup_driver">
            <a-radio-group
              :model-value="formState.cgroup_driver"
              @change="handleSaveField('cgroup_driver', $event, true)"
            >
              <a-radio value="cgroupfs">cgroupfs</a-radio>
              <a-radio value="systemd">systemd</a-radio>
            </a-radio-group>
          </a-form-item>
          <a-form-item
            :label="$t('app.docker.setting.socketPath.socket_path')"
            field="dockerSocketPath"
          >
            <a-input v-model="formState.dockerSocketPath" disabled />
            <a-button
              type="text"
              shape="round"
              class="ml-2"
              @click="handleChangeSocketPath"
            >
              <template #icon>
                <icon-settings />
              </template>
            </a-button>
          </a-form-item>
        </a-form>
        <div v-else class="mt-4 w-[600px]">
          <codemirror
            v-model="daemonJsonContent"
            :style="{ width: '100%', height: '400px' }"
            theme="cobalt"
            :tabSize="4"
            :extensions="extensions"
            autofocus
            indent-with-tab
            line-wrapping
            match-brackets
            style-active-line
          />
          <div class="mt-4">
            <a-button type="primary" @click="handleSaveJson">
              {{ $t('common.save') }}
            </a-button>
          </div>
        </div>
      </a-card>
    </div>
    <LogDrawer ref="logDrawerRef" @ok="onDrawerOk" />
    <MirrorDrawer ref="mirrorDrawerRef" @ok="onDrawerOk" />
    <RegistryDrawer ref="registryDrawerRef" @ok="onDrawerOk" />
    <Ipv6Drawer ref="ipv6DrawerRef" @ok="onDrawerOk" />
    <SocketPathDrawer ref="socketPathDrawerRef" @ok="onDrawerOk" />
  </a-spin>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n';
  import {
    dockerOperationApi,
    getDockerConfApi,
    getDockerConfRawApi,
    updateDockerConfApi,
    updateDockerConfRawApi,
  } from '@/api/docker';
  import { onMounted, reactive, ref, watch } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useConfirm } from '@/composables/confirm';
  import { Codemirror } from 'vue-codemirror';
  import { json } from '@codemirror/lang-json';
  import { oneDark } from '@codemirror/theme-one-dark';
  import LogDrawer from './components/log.vue';
  import MirrorDrawer from './components/mirror.vue';
  import RegistryDrawer from './components/registry.vue';
  import Ipv6Drawer from './components/ipv6.vue';
  import SocketPathDrawer from './components/socket-path.vue';

  const loading = ref(false);
  const formState = reactive({
    status: '',
    version: '',
    is_swarm: false,
    registry_mirrors: '',
    insecure_registries: '',
    live_restore: false,
    iptables: true,
    cgroup_driver: '',
    ipv6: false,
    fixed_cidr_v6: '',
    ip6_tables: false,
    experimental: false,
    log_option_show: false,
    log_max_size: '',
    log_max_file: '',
    dockerSocketPath: '',
  });
  const mode = ref('form');
  const daemonJsonContent = ref('');
  const extensions = [json(), oneDark];
  const { confirm } = useConfirm();
  const { t } = useI18n();

  // drawer refs
  const logDrawerRef = ref();
  const mirrorDrawerRef = ref();
  const registryDrawerRef = ref();
  const ipv6DrawerRef = ref();
  const socketPathDrawerRef = ref();

  const load = async () => {
    if (mode.value === 'form') {
      loading.value = true;
      try {
        const res = await getDockerConfApi();
        formState.status = res.status;
        formState.version = res.version;
        formState.is_swarm = res.is_swarm;
        formState.registry_mirrors = (res.registry_mirrors || []).join('\r\n');
        formState.insecure_registries = (res.insecure_registries || []).join(
          '\r\n'
        );
        formState.live_restore = res.live_restore;
        formState.iptables = res.ip_tables;
        formState.cgroup_driver = res.cgroup_driver;
        formState.ipv6 = res.ipv6;
        formState.fixed_cidr_v6 = res.fixed_cidr_v6;
        formState.ip6_tables = res.ip6_tables;
        formState.experimental = res.experimental;
        formState.log_option_show = !!(res.log_max_file || res.log_max_size);
        formState.log_max_size = res.log_max_size;
        formState.log_max_file = res.log_max_file;
      } catch (err) {
        console.error(err);
      } finally {
        loading.value = false;
      }
    } else {
      loading.value = true;
      try {
        const res = await getDockerConfRawApi();
        daemonJsonContent.value = res.content;
      } catch (err) {
        console.error(err);
      } finally {
        loading.value = false;
      }
    }
  };

  watch(mode, () => {
    load();
  });

  const handleOperator = async (operation: string) => {
    loading.value = true;
    try {
      await dockerOperationApi({ operation: operation as any });
      load();
      Message.success(t('common.message.operationSuccess'));
    } catch (err: any) {
      Message.error(err.message);
    } finally {
      loading.value = false;
    }
  };

  const handleMirrors = () => {
    mirrorDrawerRef.value?.setData({
      registry_mirrors: formState.registry_mirrors,
    });
    mirrorDrawerRef.value?.show();
  };

  const handleRegistries = () => {
    registryDrawerRef.value?.setData({
      insecure_registries: formState.insecure_registries,
    });
    registryDrawerRef.value?.show();
  };

  const handleIPv6 = () => {
    ipv6DrawerRef.value?.setData({
      fixed_cidr_v6: formState.fixed_cidr_v6,
      ip6_tables: formState.ip6_tables,
      experimental: formState.experimental,
    });
    ipv6DrawerRef.value?.show();
    return false;
  };

  const handleLogOption = () => {
    logDrawerRef.value?.setData({
      log_max_size: formState.log_max_size,
      log_max_file: formState.log_max_file,
    });
    logDrawerRef.value?.show();
    return false;
  };

  const handleChangeSocketPath = () => {
    socketPathDrawerRef.value?.setData({
      socket_path: formState.dockerSocketPath,
    });
    socketPathDrawerRef.value?.show();
  };

  const handleSaveField = async (
    key: keyof typeof formState,
    value: any,
    changeEvent?: boolean
  ) => {
    if (
      await confirm({
        title: t('app.docker.setting.log.confirm.title'),
        content: t('app.docker.setting.log.confirm.content'),
      })
    ) {
      try {
        loading.value = true;
        updateDockerConfApi({
          key,
          value,
        });
        Message.success(t('common.message.saveSuccess'));
        if (changeEvent) {
          Object.assign(formState, {
            [key]: value,
          });
        }
        return true;
      } catch (err: any) {
        Message.error(err.message);
      } finally {
        loading.value = false;
      }
    }
    return false;
  };

  function onDrawerOk() {
    load();
  }

  async function handleSaveJson() {
    if (
      await confirm({
        title: t('app.docker.setting.log.confirm.title'),
        content: t('app.docker.setting.log.confirm.content'),
      })
    ) {
      try {
        loading.value = true;
        updateDockerConfRawApi({
          content: daemonJsonContent.value,
        });
        Message.success(t('common.message.saveSuccess'));
      } catch (err: any) {
        Message.error(err.message);
      } finally {
        loading.value = false;
      }
    }
  }

  onMounted(() => {
    load();
  });
</script>
