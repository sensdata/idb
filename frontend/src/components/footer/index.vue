<template>
  <a-layout-footer class="footer">
    <div class="left">
      <span class="color-primary">Powered by iDB</span>
      <span class="ml-2.5">{{ $t('footer.currentVersion', { version }) }}</span>
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
        <span class="color-primary">API</span>
      </a-space>
    </div>
  </a-layout-footer>
</template>

<script lang="ts" setup>
  import LanguageIcon from '@/assets/icons/language-1.svg';
  import useLocale from '@/hooks/locale';
  import { getLocaleLabel, LOCALE_OPTIONS } from '@/locale';

  const version = import.meta.env.VITE_APP_VERSION as string;
  const { changeLocale, currentLocale } = useLocale();
  const locales = [...LOCALE_OPTIONS];
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
    align-items: flex-start;
    justify-content: center;
  }

  .right {
    display: flex;
    align-items: flex-end;
    justify-content: center;
  }

  .language-icon {
    width: 16px;
    height: 16px;
    vertical-align: top;
  }
</style>
