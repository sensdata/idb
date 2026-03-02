<template>
  <div class="login-form-wrapper">
    <div class="login-form-logo">
      <img alt="logo" src="@/assets/logo-wide.png" />
    </div>
    <div class="login-form-title">{{ $t('login.form.title') }}</div>
    <div class="login-form-error-msg">{{ errorMessage }}</div>
    <a-form
      ref="loginForm"
      :model="userInfo"
      class="login-form"
      layout="vertical"
      @submit="handleSubmit"
    >
      <a-form-item
        field="name"
        :rules="[{ required: true, message: $t('login.form.userName.errMsg') }]"
        :validate-trigger="['change', 'blur']"
        hide-label
      >
        <a-input
          v-model="userInfo.name"
          :placeholder="$t('login.form.userName.placeholder')"
          size="large"
        >
          <template #prefix>
            <icon-user />
          </template>
        </a-input>
      </a-form-item>
      <a-form-item
        field="password"
        :rules="[{ required: true, message: $t('login.form.password.errMsg') }]"
        :validate-trigger="['change', 'blur']"
        hide-label
      >
        <a-input-password
          v-model="userInfo.password"
          :placeholder="$t('login.form.password.placeholder')"
          size="large"
          allow-clear
        >
          <template #prefix>
            <icon-lock />
          </template>
        </a-input-password>
      </a-form-item>
      <a-space :size="16" direction="vertical">
        <div class="login-form-password-actions">
          <a-checkbox
            :model-value="loginConfig.rememberAccount"
            @change="setRememberAccount as any"
          >
            {{ $t('login.form.rememberAccount') }}
          </a-checkbox>
          <a-link @click="showForgotPasswordModal">
            {{ $t('login.form.forgetPassword') }}
          </a-link>
        </div>
        <a-button
          type="primary"
          html-type="submit"
          long
          :loading="loading"
          size="large"
        >
          {{ $t('login.form.login') }}
        </a-button>
        <!-- <a-button type="text" long class="login-form-register-btn">
          {{ $t('login.form.register') }}
        </a-button> -->
      </a-space>
    </a-form>

    <!-- 忘记密码模态框 -->
    <ForgotPasswordModal v-model:visible="forgotPasswordModalVisible" />
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive } from 'vue';
  import { useRouter } from 'vue-router';
  import { Message } from '@arco-design/web-vue';
  import { ValidatedError } from '@arco-design/web-vue/es/form/interface';
  import { useI18n } from 'vue-i18n';
  import { useStorage } from '@vueuse/core';
  import { useUserStore } from '@/store';
  import useLoading from '@/composables/loading';
  import type { LoginDataDo } from '@/api/user';
  import { DEFAULT_ROUTE_NAME } from '@/router/constants';
  import ForgotPasswordModal from './forgot-password-modal.vue';

  const router = useRouter();
  const { t } = useI18n();
  const errorMessage = ref('');
  const { loading, setLoading } = useLoading();
  const userStore = useUserStore();

  // 忘记密码模态框状态
  const forgotPasswordModalVisible = ref(false);

  const loginConfig = useStorage('login-config', {
    name: '',
    rememberAccount: true,
  });
  if (
    loginConfig.value.rememberAccount === undefined &&
    'rememberPassword' in loginConfig.value
  ) {
    loginConfig.value.rememberAccount = Boolean(
      (loginConfig.value as any).rememberPassword
    );
  }
  const userInfo = reactive({
    name: loginConfig.value.rememberAccount ? loginConfig.value.name : '',
    password: '',
  });

  const handleSubmit = async ({
    errors,
    values,
  }: {
    errors: Record<string, ValidatedError> | undefined;
    values: Record<string, any>;
  }) => {
    if (loading.value) return;
    if (!errors) {
      setLoading(true);
      try {
        await userStore.login(values as LoginDataDo);
        const { redirect, ...othersQuery } = router.currentRoute.value.query;

        if (redirect) {
          if (typeof redirect === 'string' && redirect.includes('/')) {
            router.push(redirect);
          } else {
            router.push({
              name: redirect as string,
              query: {
                ...othersQuery,
              },
            });
          }
        } else {
          router.push({
            name: DEFAULT_ROUTE_NAME,
          });
        }

        Message.success(t('login.form.login.success'));
        const { rememberAccount } = loginConfig.value;
        const { name } = values;
        loginConfig.value.name = rememberAccount ? name : '';
        (loginConfig.value as any).password = '';
        (loginConfig.value as any).rememberPassword = undefined;
      } catch (err) {
        errorMessage.value = (err as Error).message;
        Message.error(errorMessage.value || 'Login failed');
      } finally {
        setLoading(false);
      }
    }
  };
  const setRememberAccount = (value: boolean) => {
    loginConfig.value.rememberAccount = value;
  };

  // 显示忘记密码模态框
  const showForgotPasswordModal = () => {
    forgotPasswordModalVisible.value = true;
  };
</script>

<style scoped lang="less">
  .login-form {
    &-wrapper {
      width: 320px;
    }
    &-logo {
      width: 128px;
      margin: 0 auto;
      img {
        width: 100%;
      }
    }
    &-title {
      margin-top: 32px;
      margin-bottom: 8px;
      font-size: 24px;
      font-weight: 500;
      line-height: 32px;
      color: var(--color-text-3);
      text-align: center;
    }
    &-error-msg {
      height: 32px;
      line-height: 32px;
      color: rgb(var(--red-6));
    }
    &-password-actions {
      display: flex;
      justify-content: space-between;
    }
    &-register-btn {
      color: var(--color-text-3) !important;
    }
  }
</style>
