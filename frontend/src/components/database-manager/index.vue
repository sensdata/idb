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
    <!-- 加载状态 -->
    <div v-if="loading && !composeInfo" class="loading-container">
      <a-spin :loading="true" />
    </div>

    <!-- 主要内容 -->
    <div v-else>
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
    </div>
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
      // 准备所有需要请求的 Promise
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

      // 使用 Promise.allSettled 来处理所有请求，即使部分失败也能显示成功的数据
      const promises: Promise<any>[] = [composesPromise, configPromise];

      // 如果是 MySQL 或 Redis，添加密码和远程访问请求
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

        promises.push(passwordPromise, remoteAccessPromise);
      }

      const results = await Promise.allSettled(promises);

      // 提取错误信息的辅助函数
      const extractErrorMessage = (error: any): string => {
        // 优先使用 response.data.data (后端详细错误)
        const responseData = error?.response?.data;
        if (responseData?.data && typeof responseData.data === 'string') {
          return responseData.data.trim();
        }
        // 其次使用 response.data.message 或 error.message
        return responseData?.message || error?.message || String(error);
      };

      // 处理基本信息（第一个请求）
      const composesResult = results[0];
      if (composesResult.status === 'fulfilled') {
        composeInfo.value =
          composesResult.value.composes.find(
            (c: any) => c.name === composeName.value
          ) || null;

        if (composeInfo.value) {
          portForm.value.port = Number(composeInfo.value.port);
        }
      } else {
        console.error('Failed to load compose info:', composesResult.reason);
        Message.error(extractErrorMessage(composesResult.reason));
      }

      // 处理配置（第二个请求）
      const configResult = results[1];
      if (configResult.status === 'fulfilled') {
        configContent.value = configResult.value.content;
      } else {
        console.error('Failed to load config:', configResult.reason);
        Message.error(extractErrorMessage(configResult.reason));
      }

      // 处理密码和远程访问（第三、四个请求，仅 MySQL/Redis）
      if (databaseType.value === 'mysql' || databaseType.value === 'redis') {
        const pwdResult = results[2];
        const raResult = results[3];

        if (pwdResult.status === 'fulfilled') {
          currentPassword.value = pwdResult.value.password || '******';
        } else {
          console.error('Failed to load password:', pwdResult.reason);
          currentPassword.value = '******';
          // 密码加载失败不显示错误提示，使用默认值即可
        }

        if (raResult.status === 'fulfilled') {
          remoteAccessEnabled.value = raResult.value.remote_access ?? false;
        } else {
          console.error('Failed to load remote access:', raResult.reason);
          remoteAccessEnabled.value = false;
          // 远程访问状态加载失败不显示错误提示，使用默认值即可
        }
      }
    } catch (error) {
      // 这里只会捕获意外的错误，因为 Promise.allSettled 不会抛出错误
      console.error('Unexpected error in loadData:', error);
      Message.error(t('app.store.database.message.loadDataFailed'));
    } finally {
      loading.value = false;
    }
  };

  const show = (
    type: 'mysql' | 'postgresql' | 'redis',
    name: string | (() => Promise<string>)
  ) => {
    // 设置类型并在打开前重置状态，避免出现旧数据和空白闪烁
    databaseType.value = type;

    // 立即进入加载态并清空旧数据
    loading.value = true;
    composeInfo.value = null;
    configContent.value = '';
    currentPassword.value = '******';
    remoteAccessEnabled.value = false;
    startLoading.value = false;
    stopLoading.value = false;
    restartLoading.value = false;
    saveConfigLoading.value = false;
    changePasswordLoading.value = false;
    remoteAccessLoading.value = false;
    changePortLoading.value = false;
    refreshLoading.value = false;

    // 立即显示 drawer
    visible.value = true;

    // 在 drawer 内部加载数据
    (async () => {
      try {
        // 如果 name 是函数，先执行获取实际名称
        if (typeof name === 'function') {
          composeName.value = await name();
        } else {
          composeName.value = name;
        }
        await loadData();
      } catch (error) {
        // 加载失败时显示错误，但 drawer 保持打开状态
        Message.error(t('app.store.database.message.loadDataFailed'));
        console.error('Failed to load database data:', error);
        // 如果在调用 loadData 之前报错（例如获取名称失败），需要手动结束 loading
        loading.value = false;
      }
    })();
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

<style scoped>
  .loading-container {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 400px;
  }
</style>
