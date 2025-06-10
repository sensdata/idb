<template>
  <a-drawer
    v-model:visible="visible"
    :width="600"
    title="IPv6"
    destroy-on-close
    @close="handleClose"
    @before-ok="handleBeforeOk"
  >
    <a-spin :loading="loading" class="w-full">
      <a-form ref="formRef" :model="form" :rules="rules">
        <a-row type="flex" justify="center">
          <a-col :span="22">
            <a-form-item
              field="fixed_cidr_v6"
              :label="$t('app.docker.setting.ipv6.fixed_cidr_v6')"
            >
              <a-input v-model="form.fixed_cidr_v6" />
            </a-form-item>
            <a-form-item>
              <a-checkbox v-model="showMore">
                {{ $t('app.docker.setting.ipv6.advanced') }}
              </a-checkbox>
            </a-form-item>
            <div v-if="showMore">
              <a-form-item field="ip6_tables" label="ip6tables">
                <a-switch v-model="form.ip6_tables"></a-switch>
              </a-form-item>
              <a-form-item field="experimental" label="experimental">
                <a-switch v-model="form.experimental"></a-switch>
              </a-form-item>
            </div>
          </a-col>
        </a-row>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<script lang="ts" setup>
  import { reactive, ref, toRaw } from 'vue';
  import { checkIpV6 } from '@/helper/utils';
  import { Message } from '@arco-design/web-vue';
  import { updateIpv6OptionApi } from '@/api/docker';
  import { useConfirm } from '@/hooks/confirm';
  import { useI18n } from 'vue-i18n';

  const emit = defineEmits(['ok']);

  const loading = ref();
  const visible = ref();
  const formRef = ref();
  const showMore = ref(true);
  const { t } = useI18n();
  const { confirm } = useConfirm();

  const form = reactive({
    fixed_cidr_v6: '',
    ip6_tables: false,
    experimental: false,
  });

  const rules = reactive({
    fixed_cidr_v6: [
      {
        validator(value: any, callback: any) {
          if (!form.fixed_cidr_v6 || form.fixed_cidr_v6.indexOf('/') === -1) {
            callback(
              t('common.message.formatError', {
                field: t('app.docker.setting.ipv6.fixed_cidr_v6'),
              })
            );
            return;
          }
          if (checkIpV6(form.fixed_cidr_v6.split('/')[0])) {
            callback(
              t('common.message.formatError', {
                field: t('app.docker.setting.ipv6.fixed_cidr_v6'),
              })
            );
            return;
          }
          const reg = /^(?:[1-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8])$/;
          if (!reg.test(form.fixed_cidr_v6.split('/')[1])) {
            callback(t('common.message.formatError'));
            return;
          }
          callback();
        },
        trigger: 'blur',
        required: true,
      },
    ],
  });

  const show = () => {
    visible.value = true;
  };

  const hide = () => {
    visible.value = false;
  };

  const setData = (params: any) => {
    form.fixed_cidr_v6 = params.fixed_cidr_v6;
    form.ip6_tables = params.ip6_tables;
    form.experimental = params.experimental;
  };

  const handleClose = () => {
    visible.value = false;
  };

  const handleBeforeOk = async () => {
    const errors = await formRef.value.validate();
    if (errors) {
      return false;
    }
    const params = toRaw(form);
    if (
      await confirm({
        title: t('app.docker.setting.ipv6.confirm.title'),
        content: t('app.docker.setting.ipv6.confirm.content'),
      })
    ) {
      loading.value = true;
      try {
        await updateIpv6OptionApi(params);
        emit('ok');
      } catch (err: any) {
        loading.value = false;
        Message.error(err.message);
        return false;
      }
      return true;
    }
    return false;
  };

  defineExpose({
    show,
    hide,
    setData,
  });
</script>
