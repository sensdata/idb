<template>
  <a-layout-footer class="footer">
    <div class="left">
      <span class="color-primary">Powered by iDB</span>
      <span class="ml-2.5">{{ $t('footer.currentVersion') }}</span>
      <span class="ml-1 color-primary">{{ version }}</span>
      <a-divider direction="vertical" margin="8px" />
      <a-link
        v-if="!isCheckingUpdate"
        class="check-update-link"
        @click="checkUpdate"
      >
        {{ $t('footer.checkUpdate') }}
      </a-link>
      <span v-else class="checking-update">
        <a-spin class="mr-1 scale-75" />
        {{ $t('footer.checkUpdate') }}
      </span>
    </div>
    <div class="right">
      <a-space size="small">
        <template #split>
          <a-divider direction="vertical" margin="0" />
        </template>
        <a-dropdown trigger="click" @select="changeLocale as any">
          <span class="cursor-pointer">
            <language-icon class="language-icon" />
            {{ getLocaleLabel(currentLocale) }}
          </span>
          <template #content>
            <a-doption
              v-for="item in locales"
              :key="item.value"
              :value="item.value"
            >
              <template #icon>
                <icon-check v-show="item.value === currentLocale" />
              </template>
              {{ item.label }}
            </a-doption>
          </template>
        </a-dropdown>
        <span class="color-primary"> {{ $t('footer.licence') }}</span>
        <a-link href="/api/v1/swagger/index.html" target="_blank"> API </a-link>
      </a-space>
    </div>
  </a-layout-footer>
</template>

<script lang="ts" setup>
  import { onMounted, ref } from 'vue';
  import { useI18n } from 'vue-i18n';
  import useLocale from '@/composables/locale';
  import { getLocaleLabel, LOCALE_OPTIONS } from '@/locale';
  import LanguageIcon from '@/assets/icons/language-1.svg';
  import { getPublicVersionApi } from '@/api/public';
  import { isLogin } from '@/helper/auth';
  import { getSettingsAboutApi, upgradeApi } from '@/api/settings';
  import { useConfirm } from '@/composables/confirm';
  import { compareVersion } from '@/helper/utils';
  import { Message } from '@arco-design/web-vue';

  const { t } = useI18n();
  const version = ref('');
  const { changeLocale, currentLocale } = useLocale();
  const locales = [...LOCALE_OPTIONS];
  const { confirm } = useConfirm();
  const count = ref(100);
  const isCheckingUpdate = ref(false);
  const checkUpdate = async () => {
    if (!isLogin()) {
      return;
    }
    if (isCheckingUpdate.value) {
      return;
    }

    isCheckingUpdate.value = true;
    try {
      const data = await getSettingsAboutApi();
      // 版本大小对比
      if (compareVersion(data.new_version, data.version) > 0) {
        if (
          await confirm({
            title: t('footer.checkUpdate'),
            content: t('footer.checkUpdateContent'),
          })
        ) {
          upgradeApi();
          // 倒数100秒，然后刷新
          count.value = 100;
          const countdown = () => {
            count.value--;
            if (count.value <= 0) {
              return;
            }
            Message.info({
              id: 'check-update-countdown',
              content: t('footer.checkUpdateCountdown', {
                count: count.value,
              }),
              duration: 2000,
            });
          };
          countdown();

          const timer = setInterval(() => {
            countdown();
            if (count.value <= 0) {
              clearInterval(timer);
              window.location.reload();
            }
          }, 1000);
        }
      } else {
        // 没有新版本时显示提示
        Message.success(t('footer.checkUpdateNoUpdate'));
      }
    } catch (error) {
      console.error(error);
    } finally {
      isCheckingUpdate.value = false;
    }
  };

  onMounted(async () => {
    const data = await getPublicVersionApi();
    version.value = data.version;
  });
</script>

<style scoped lang="less">
  .footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    height: 38px;
    padding: 0 20px;
    color: var(--color-text-2);
    text-align: center;
    border-top: 1px solid var(--color-border-2);
  }

  .left {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    height: 100%;
  }

  .right {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    height: 100%;
  }

  .language-icon {
    width: 16px;
    height: 16px;
    vertical-align: top;
  }

  .check-update-link {
    font-size: 12px;
    color: var(--color-text-3);
    text-decoration: none;

    &:hover {
      color: var(--color-primary);
    }
  }

  .checking-update {
    font-size: 12px;
    color: var(--color-text-3);
    display: flex;
    align-items: center;
  }
</style>
