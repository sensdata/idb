<template>
  <a-drawer
    v-model:visible="visible"
    :title="
      $t('app.store.database.manager.title', {
        type: databaseType.toUpperCase(),
        name: composeName,
      })
    "
    width="800px"
    :footer="false"
    unmount-on-close
  >
    <a-spin :loading="loading">
      <a-tabs v-if="composeInfo" default-active-key="info">
        <a-tab-pane key="info" :title="$t('app.store.database.tab.info')">
          <a-descriptions :column="2" bordered>
            <a-descriptions-item :label="$t('app.store.database.info.name')">
              {{ composeInfo.name }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.store.database.info.version')">
              {{ composeInfo.version }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.store.database.info.port')">
              {{ composeInfo.port }}
            </a-descriptions-item>
            <a-descriptions-item :label="$t('app.store.database.info.status')">
              <a-space>
                <a-tag
                  :color="
                    composeInfo.status === 'running'
                      ? 'green'
                      : composeInfo.status === 'exited'
                      ? 'red'
                      : 'gray'
                  "
                >
                  {{ composeInfo.status }}
                </a-tag>
                <a-button
                  size="mini"
                  :loading="refreshLoading"
                  @click="handleRefreshStatus"
                >
                  {{ $t('app.store.database.button.refreshStatus') }}
                </a-button>
              </a-space>
            </a-descriptions-item>
          </a-descriptions>

          <a-divider />

          <a-space>
            <a-button
              v-if="composeInfo.status !== 'running'"
              type="primary"
              :loading="startLoading"
              @click="handleOperation('start')"
            >
              {{ $t('app.store.database.button.start') }}
            </a-button>
            <a-button
              v-if="composeInfo.status === 'running'"
              status="warning"
              :loading="stopLoading"
              @click="handleOperation('stop')"
            >
              {{ $t('app.store.database.button.stop') }}
            </a-button>
            <a-button
              v-if="composeInfo.status === 'running'"
              :loading="restartLoading"
              @click="handleOperation('restart')"
            >
              {{ $t('app.store.database.button.restart') }}
            </a-button>
          </a-space>
        </a-tab-pane>

        <a-tab-pane key="config" :title="$t('app.store.database.tab.config')">
          <a-textarea
            v-model="configContent"
            :auto-size="{ minRows: 15, maxRows: 25 }"
            :placeholder="$t('app.store.database.config.placeholder')"
          />
          <a-space class="mt-4">
            <a-button
              type="primary"
              :loading="saveConfigLoading"
              @click="handleSaveConfig"
            >
              {{ $t('app.store.database.button.save') }}
            </a-button>
            <a-button @click="loadConfig">{{
              $t('app.store.database.button.refresh')
            }}</a-button>
          </a-space>
        </a-tab-pane>

        <a-tab-pane
          v-if="databaseType === 'mysql' || databaseType === 'redis'"
          key="password"
          :title="$t('app.store.database.tab.password')"
        >
          <a-form :model="passwordForm" layout="vertical">
            <a-form-item :label="$t('app.store.database.password.current')">
              <a-input
                v-model="currentPassword"
                disabled
                placeholder="******"
              />
            </a-form-item>
            <a-form-item :label="$t('app.store.database.password.new')">
              <a-input-password
                v-model="passwordForm.newPassword"
                :placeholder="$t('app.store.database.password.newPlaceholder')"
              />
            </a-form-item>
            <a-form-item>
              <a-button
                type="primary"
                :loading="changePasswordLoading"
                @click="handleChangePassword"
              >
                {{ $t('app.store.database.button.changePassword') }}
              </a-button>
            </a-form-item>
          </a-form>
        </a-tab-pane>

        <a-tab-pane
          v-if="databaseType === 'mysql' || databaseType === 'redis'"
          key="remote"
          :title="$t('app.store.database.tab.remote')"
        >
          <a-form :model="{ remoteAccessEnabled }" layout="vertical">
            <a-form-item :label="$t('app.store.database.remote.status')">
              <a-switch
                v-model="remoteAccessEnabled"
                :loading="remoteAccessLoading"
                @change="handleToggleRemoteAccess"
              >
                <template #checked>{{
                  $t('app.store.database.remote.enabled')
                }}</template>
                <template #unchecked>{{
                  $t('app.store.database.remote.disabled')
                }}</template>
              </a-switch>
            </a-form-item>
            <a-alert v-if="remoteAccessEnabled" type="warning">
              {{ $t('app.store.database.remote.warning') }}
            </a-alert>
          </a-form>
        </a-tab-pane>

        <a-tab-pane
          v-if="databaseType === 'postgresql'"
          key="port"
          :title="$t('app.store.database.tab.port')"
        >
          <a-form :model="portForm" layout="vertical">
            <a-form-item :label="$t('app.store.database.port.label')">
              <a-input-number
                v-model="portForm.port"
                :min="1024"
                :max="65535"
                :placeholder="$t('app.store.database.port.placeholder')"
              />
            </a-form-item>
            <a-form-item>
              <a-button
                type="primary"
                :loading="changePortLoading"
                @click="handleChangePort"
              >
                {{ $t('app.store.database.button.changePort') }}
              </a-button>
            </a-form-item>
          </a-form>
        </a-tab-pane>
      </a-tabs>
    </a-spin>
  </a-drawer>
</template>

<script setup lang="ts">
  import { ref } from 'vue';
  import { Message } from '@arco-design/web-vue';
  import { useI18n } from 'vue-i18n';
  import {
    getMysqlComposesApi,
    mysqlOperationApi,
    getMysqlConfApi,
    setMysqlConfApi,
    getMysqlPasswordApi,
    setMysqlPasswordApi,
    getMysqlRemoteAccessApi,
    setMysqlRemoteAccessApi,
    getPostgreSqlComposesApi,
    postgreSqlOperationApi,
    getPostgreSqlConfApi,
    setPostgreSqlConfApi,
    postgreSqlSetPortApi,
    getRedisComposesApi,
    redisOperationApi,
    getRedisConfApi,
    setRedisConfApi,
    getRedisPasswordApi,
    setRedisPasswordApi,
    getRedisRemoteAccessApi,
    setRedisRemoteAccessApi,
  } from '@/api/database';
  import { DatabaseComposeInfo } from '@/entity/Database';

  const { t } = useI18n();

  const visible = ref(false);
  const loading = ref(false);
  const databaseType = ref<'mysql' | 'postgresql' | 'redis'>('mysql');
  const composeName = ref('');
  const composeInfo = ref<DatabaseComposeInfo | null>(null);
  const configContent = ref('');
  const currentPassword = ref('');
  const remoteAccessEnabled = ref(false);

  // 各个操作的 loading 状态
  const startLoading = ref(false);
  const stopLoading = ref(false);
  const restartLoading = ref(false);
  const saveConfigLoading = ref(false);
  const changePasswordLoading = ref(false);
  const remoteAccessLoading = ref(false);
  const changePortLoading = ref(false);
  const refreshLoading = ref(false);

  const passwordForm = ref({
    newPassword: '',
  });

  const portForm = ref({
    port: 5432,
  });

  // 定义所有加载函数
  const loadConfig = async () => {
    try {
      let response;
      if (databaseType.value === 'mysql') {
        response = await getMysqlConfApi({ name: composeName.value });
      } else if (databaseType.value === 'postgresql') {
        response = await getPostgreSqlConfApi({ name: composeName.value });
      } else {
        response = await getRedisConfApi({ name: composeName.value });
      }
      configContent.value = response.content;
    } catch (error) {
      Message.error(t('app.store.database.message.loadConfigFailed'));
    }
  };

  const loadPassword = async () => {
    try {
      let response;
      if (databaseType.value === 'mysql') {
        response = await getMysqlPasswordApi({ name: composeName.value });
      } else {
        response = await getRedisPasswordApi({ name: composeName.value });
      }
      currentPassword.value = response.password || '******';
    } catch (error) {
      console.error(t('app.store.database.message.loadPasswordFailed'), error);
    }
  };

  const loadData = async (options: { delayBeforeLoad?: boolean } = {}) => {
    loading.value = true;

    // 等待 3 秒后再请求数据，确保后端状态已经更新
    // 因为数据库操作（启动/停止/重启）需要时间生效，立即请求可能获取到旧状态
    if (options.delayBeforeLoad) {
      await new Promise((resolve) => {
        setTimeout(resolve, 3000);
      });
    }

    try {
      // 1) 基本信息和配置是必需的：用 Promise.all
      let composesPromise: Promise<any>;
      if (databaseType.value === 'mysql') {
        composesPromise = getMysqlComposesApi({ page: 1, page_size: 100 });
      } else if (databaseType.value === 'postgresql') {
        composesPromise = getPostgreSqlComposesApi({ page: 1, page_size: 100 });
      } else {
        composesPromise = getRedisComposesApi({ page: 1, page_size: 100 });
      }

      let configPromise: Promise<any>;
      if (databaseType.value === 'mysql') {
        configPromise = getMysqlConfApi({ name: composeName.value });
      } else if (databaseType.value === 'postgresql') {
        configPromise = getPostgreSqlConfApi({ name: composeName.value });
      } else {
        configPromise = getRedisConfApi({ name: composeName.value });
      }

      const [composesResponse, configResponse] = await Promise.all([
        composesPromise,
        configPromise,
      ]);

      // 处理基本信息
      composeInfo.value =
        composesResponse.composes.find(
          (c: any) => c.name === composeName.value
        ) || null;

      if (composeInfo.value) {
        portForm.value.port = Number(composeInfo.value.port);
      }

      // 处理配置
      configContent.value = configResponse.content;

      // 2) 密码和远程访问（仅 MySQL/Redis）：用 Promise.allSettled，以免单个失败导致整体失败
      if (databaseType.value === 'mysql' || databaseType.value === 'redis') {
        let passwordPromise: Promise<any>;
        let remoteAccessPromise: Promise<any>;

        if (databaseType.value === 'mysql') {
          passwordPromise = getMysqlPasswordApi({
            name: composeName.value,
          });
          remoteAccessPromise = getMysqlRemoteAccessApi({
            name: composeName.value,
          });
        } else {
          passwordPromise = getRedisPasswordApi({
            name: composeName.value,
          });
          remoteAccessPromise = getRedisRemoteAccessApi({
            name: composeName.value,
          });
        }

        const [pwdResult, raResult] = await Promise.allSettled([
          passwordPromise,
          remoteAccessPromise,
        ]);

        if (pwdResult.status === 'fulfilled') {
          currentPassword.value = pwdResult.value.password || '******';
        } else {
          currentPassword.value = '******';
        }
        if (raResult.status === 'fulfilled') {
          remoteAccessEnabled.value = raResult.value.remote_access ?? false;
        } else {
          remoteAccessEnabled.value = false;
        }
      }
    } catch (error) {
      Message.error(t('app.store.database.message.loadDataFailed'));
      throw error; // 重新抛出错误，以便 show 函数知道加载失败
    } finally {
      loading.value = false;
    }
  };

  const show = async (type: 'mysql' | 'postgresql' | 'redis', name: string) => {
    databaseType.value = type;
    composeName.value = name;
    // 先加载数据，加载成功后再显示 drawer
    try {
      await loadData();
      // 只有在数据加载成功后才显示 drawer
      if (composeInfo.value) {
        visible.value = true;
      }
    } catch (error) {
      // 加载失败时不显示 drawer
    }
  };

  const handleOperation = async (operation: 'start' | 'stop' | 'restart') => {
    // 根据操作类型设置对应的 loading 状态
    let loadingRef;
    if (operation === 'start') {
      loadingRef = startLoading;
    } else if (operation === 'stop') {
      loadingRef = stopLoading;
    } else {
      loadingRef = restartLoading;
    }
    loadingRef.value = true;
    try {
      if (databaseType.value === 'mysql') {
        await mysqlOperationApi({ name: composeName.value, operation });
      } else if (databaseType.value === 'postgresql') {
        await postgreSqlOperationApi({ name: composeName.value, operation });
      } else {
        await redisOperationApi({ name: composeName.value, operation });
      }
      Message.success(
        t('app.store.database.message.operationSuccess', { operation })
      );
      await loadData({ delayBeforeLoad: true });
    } catch (error) {
      Message.error(
        t('app.store.database.message.operationFailed', { operation })
      );
    } finally {
      loadingRef.value = false;
    }
  };

  const handleSaveConfig = async () => {
    saveConfigLoading.value = true;
    try {
      if (databaseType.value === 'mysql') {
        await setMysqlConfApi({
          name: composeName.value,
          content: configContent.value,
        });
      } else if (databaseType.value === 'postgresql') {
        await setPostgreSqlConfApi({
          name: composeName.value,
          content: configContent.value,
        });
      } else {
        await setRedisConfApi({
          name: composeName.value,
          content: configContent.value,
        });
      }
      Message.success(t('app.store.database.message.configSaveSuccess'));
    } catch (error) {
      Message.error(t('app.store.database.message.configSaveFailed'));
    } finally {
      saveConfigLoading.value = false;
    }
  };

  const handleChangePassword = async () => {
    if (!passwordForm.value.newPassword) {
      Message.warning(t('app.store.database.message.passwordRequired'));
      return;
    }

    changePasswordLoading.value = true;
    try {
      if (databaseType.value === 'mysql') {
        await setMysqlPasswordApi({
          name: composeName.value,
          new_pass: passwordForm.value.newPassword,
        });
      } else {
        await setRedisPasswordApi({
          name: composeName.value,
          new_pass: passwordForm.value.newPassword,
        });
      }
      Message.success(t('app.store.database.message.passwordChangeSuccess'));
      passwordForm.value.newPassword = '';
      await loadPassword();
    } catch (error) {
      Message.error(t('app.store.database.message.passwordChangeFailed'));
    } finally {
      changePasswordLoading.value = false;
    }
  };

  const handleToggleRemoteAccess = async (value: boolean | string | number) => {
    const enabled = Boolean(value);
    remoteAccessLoading.value = true;
    try {
      if (databaseType.value === 'mysql') {
        await setMysqlRemoteAccessApi({
          name: composeName.value,
          remote_access: enabled,
        });
      } else {
        await setRedisRemoteAccessApi({
          name: composeName.value,
          remote_access: enabled,
        });
      }
      const status = enabled
        ? t('app.store.database.remoteStatus.enabled')
        : t('app.store.database.remoteStatus.disabled');
      Message.success(
        t('app.store.database.message.remoteAccessSuccess', { status })
      );
    } catch (error) {
      Message.error(t('app.store.database.message.remoteAccessFailed'));
      remoteAccessEnabled.value = !enabled;
    } finally {
      remoteAccessLoading.value = false;
    }
  };

  const handleChangePort = async () => {
    changePortLoading.value = true;
    try {
      await postgreSqlSetPortApi({
        name: composeName.value,
        port: String(portForm.value.port),
      });
      Message.success(t('app.store.database.message.portChangeSuccess'));
      await loadData();
    } catch (error) {
      Message.error(t('app.store.database.message.portChangeFailed'));
    } finally {
      changePortLoading.value = false;
    }
  };

  const handleRefreshStatus = async () => {
    refreshLoading.value = true;
    try {
      await loadData();
      Message.success(t('app.store.database.message.refreshSuccess'));
    } catch (error) {
      Message.error(t('app.store.database.message.refreshFailed'));
    } finally {
      refreshLoading.value = false;
    }
  };

  defineExpose({
    show,
  });
</script>
