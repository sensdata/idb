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
            checked="rememberPassword"
            :model-value="loginConfig.rememberPassword"
            @change="setRememberPassword as any"
          >
            {{ $t('login.form.rememberPassword') }}
          </a-checkbox>
          <a-link>{{ $t('login.form.forgetPassword') }}</a-link>
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
  import useLoading from '@/hooks/loading';
  import type { LoginDataDo } from '@/api/user';
  import { DEFAULT_ROUTE_NAME } from '@/router/constants';

  function serialize(str: string): string {
    const encoded = btoa(str);
    return encoded.split('').reverse().join('');
  }

  function unserialize(str: string): string {
    try {
      const reversed = str.split('').reverse().join('');
      return atob(reversed);
    } catch (err) {
      return '';
    }
  }

  const router = useRouter();
  const { t } = useI18n();
  const errorMessage = ref('');
  const { loading, setLoading } = useLoading();
  const userStore = useUserStore();

  const loginConfig = useStorage('login-config', {
    name: 'admin',
    password: serialize('admin123'),
    rememberPassword: true,
  });
  const userInfo = reactive({
    name: loginConfig.value.name,
    password: unserialize(loginConfig.value.password),
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
        const { rememberPassword } = loginConfig.value;
        const { name, password } = values;
        loginConfig.value.name = rememberPassword ? name : '';
        loginConfig.value.password = rememberPassword
          ? serialize(password)
          : '';
      } catch (err) {
        errorMessage.value = (err as Error).message;
      } finally {
        setLoading(false);
      }
    }
  };
  const setRememberPassword = (value: boolean) => {
    loginConfig.value.rememberPassword = value;
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
      color: var(--color-text-3);
      font-weight: 500;
      font-size: 24px;
      line-height: 32px;
      text-align: center;
    }
    &-error-msg {
      height: 32px;
      color: rgb(var(--red-6));
      line-height: 32px;
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
