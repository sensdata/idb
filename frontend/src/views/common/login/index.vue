<template>
  <div class="login-page">
    <div class="main">
      <div class="intro">
        <img class="logo" alt="logo" src="@/assets/logo.png" />
        <h1 class="title">{{ $t('login.intro.title') }}</h1>
        <p class="slogan">{{ $t('login.intro.slogan') }}</p>
        <div class="features">
          <div
            v-for="feature in features"
            :key="feature.title"
            class="feature-item"
          >
            <component :is="feature.icon" class="icon" />
            <div class="info">
              <h3>{{ feature.title }}</h3>
              <p>{{ feature.desc }}</p>
            </div>
          </div>
        </div>
      </div>
      <div class="content">
        <LoginForm />
      </div>
    </div>
    <Footer />
  </div>
</template>

<script lang="ts" setup>
  import { useI18n } from 'vue-i18n';
  import Feature1Icon from '@/assets/icons/feature1.svg';
  import Feature2Icon from '@/assets/icons/feature2.svg';
  import Feature3Icon from '@/assets/icons/feature3.svg';
  import Feature4Icon from '@/assets/icons/feature4.svg';
  import Footer from '@/components/footer/index.vue';

  import LoginForm from './components/login-form.vue';

  const { t } = useI18n();

  const features = [
    {
      icon: Feature1Icon,
      title: t('login.feature1.title'),
      desc: t('login.feature1.desc'),
    },
    {
      icon: Feature2Icon,
      title: t('login.feature2.title'),
      desc: t('login.feature2.desc'),
    },
    {
      icon: Feature3Icon,
      title: t('login.feature3.title'),
      desc: t('login.feature3.desc'),
    },
    {
      icon: Feature4Icon,
      title: t('login.feature4.title'),
      desc: t('login.feature4.desc'),
    },
  ];
</script>

<style scoped lang="less">
  .login-page {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  .main {
    display: flex;
    flex: 1;
    min-height: 100px;
    .intro {
      width: 720px;
      margin: 30px 0 10px 30px;
      padding: 72px 100px;
      color: var(--idb-brand-text);
      background-color: var(--idb-brand-bg);
      background-image: url('@/assets/background.png');
      background-repeat: no-repeat;
      background-position: left -182px top -133px;
      background-size: 1643px 1209px;
      border-radius: 30px;
      .logo {
        display: block;
        width: 80px;
        height: 80px;
        margin: 0 auto;
      }
      .title {
        margin: 12px 0 0;
        font-weight: 500;
        font-size: 24px;
        line-height: 36px;
        text-align: center;
      }
      .slogan {
        margin-top: 8px;
        margin-bottom: 0;
        font-weight: 400;
        font-size: 20px;
        line-height: 32px;
        text-align: center;
      }
      .features {
        margin-top: 40px;
        .feature-item {
          display: flex;
          flex-direction: row;
          margin-bottom: 32px;
          :last-child {
            margin-bottom: 0;
          }
          .icon {
            margin-right: 40px;
          }
          .info {
            flex: 1;
            min-width: 0;
          }
          h3 {
            margin: 0;
            font-weight: 500;
            font-size: 18px;
            line-height: 20px;
          }
          p {
            margin: 8px 0 0;
            font-weight: 400;
            font-size: 14px;
            line-height: 24px;
          }
        }
      }
    }
    .content {
      position: relative;
      display: flex;
      flex: 1;
      align-items: center;
      justify-content: center;
      padding-bottom: 40px;
    }
  }
</style>
