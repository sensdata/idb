<template>
  <div class="perioid-wrap">
    <div class="period-item mb-4">
      <a-select
        v-model="localPeriod.type"
        class="w-[120px]"
        :options="periodTypeOptions"
        @change="handlePeriodChange"
      />
      <a-select
        v-if="showWeek(localPeriod)"
        v-model="localPeriod.week"
        class="w-[120px] ml-4"
        :options="weekOptions"
        @change="handlePeriodChange"
      />
      <a-input-number
        v-if="showDay(localPeriod)"
        v-model="localPeriod.day"
        class="w-[120px] ml-4"
        :min="1"
        :max="31"
        @change="handlePeriodChange"
      >
        <template #suffix> {{ $t('common.timeUnit.day') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showHour(localPeriod)"
        v-model="localPeriod.hour"
        class="w-[120px] ml-4"
        :min="0"
        :max="23"
        @change="handlePeriodChange"
      >
        <template #suffix> {{ $t('common.timeUnit.hour') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showMinute(localPeriod)"
        v-model="localPeriod.minute"
        class="w-[120px] ml-4"
        :min="0"
        :max="59"
        @change="handlePeriodChange"
      >
        <template #suffix> {{ $t('common.timeUnit.minute') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showSecond(localPeriod)"
        v-model="localPeriod.second"
        class="w-[120px] ml-4"
        :min="0"
        :max="59"
        @change="handlePeriodChange"
      >
        <template #suffix> {{ $t('common.timeUnit.second') }} </template>
      </a-input-number>
    </div>
  </div>
</template>

<script lang="ts" setup>
  import { CRONTAB_PERIOD_TYPE } from '@/config/enum';
  import { PeriodDetailDo } from '@/entity/Crontab';
  import { onMounted, ref, watch } from 'vue';
  import { useI18n } from 'vue-i18n';

  const props = defineProps<{
    modelValue: PeriodDetailDo[];
  }>();
  const emit = defineEmits(['update:modelValue']);

  const { t } = useI18n();

  // 创建本地副本以避免直接修改props
  const localPeriod = ref<PeriodDetailDo>({
    type: CRONTAB_PERIOD_TYPE.WEEKLY,
    week: 1,
    day: 0,
    hour: 1,
    minute: 0,
    second: 0,
  });

  watch(
    () => props.modelValue,
    (newValue) => {
      if (newValue && newValue.length > 0) {
        localPeriod.value = { ...newValue[0] };
      }
    },
    { immediate: true, deep: true }
  );

  const handlePeriodChange = () => {
    const updatedPeriod = { ...localPeriod.value };

    // 根据周期类型设置默认值
    if (
      updatedPeriod.type === CRONTAB_PERIOD_TYPE.WEEKLY &&
      !updatedPeriod.week
    ) {
      updatedPeriod.week = 1;
    }
    if (
      (updatedPeriod.type === CRONTAB_PERIOD_TYPE.MONTHLY ||
        updatedPeriod.type === CRONTAB_PERIOD_TYPE.EVERY_N_DAYS) &&
      !updatedPeriod.day
    ) {
      updatedPeriod.day = 1;
    }

    // 确保时、分、秒字段有有效值
    if (
      ![
        CRONTAB_PERIOD_TYPE.HOURLY,
        CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES,
        CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS,
      ].includes(updatedPeriod.type) &&
      updatedPeriod.hour === undefined
    ) {
      updatedPeriod.hour = 0;
    }

    if (
      ![CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS].includes(updatedPeriod.type) &&
      updatedPeriod.minute === undefined
    ) {
      updatedPeriod.minute = 0;
    }

    if (
      updatedPeriod.type === CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS &&
      updatedPeriod.second === undefined
    ) {
      updatedPeriod.second = 0;
    }

    localPeriod.value = updatedPeriod;
    emit('update:modelValue', [updatedPeriod]);
  };

  const periodTypeOptions = [
    {
      value: CRONTAB_PERIOD_TYPE.MONTHLY,
      label: t('app.crontab.enum.periodType.monthly'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.WEEKLY,
      label: t('app.crontab.enum.periodType.weekly'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.DAILY,
      label: t('app.crontab.enum.periodType.daily'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.HOURLY,
      label: t('app.crontab.enum.periodType.hourly'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.EVERY_N_DAYS,
      label: t('app.crontab.enum.periodType.every_n_days'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.EVERY_N_HOURS,
      label: t('app.crontab.enum.periodType.every_n_hours'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES,
      label: t('app.crontab.enum.periodType.every_n_minutes'),
    },
    {
      value: CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS,
      label: t('app.crontab.enum.periodType.every_n_seconds'),
    },
  ];

  const weekOptions = [
    {
      value: 1,
      label: t('app.crontab.enum.week.monday'),
    },
    {
      value: 2,
      label: t('app.crontab.enum.week.tuesday'),
    },
    {
      value: 3,
      label: t('app.crontab.enum.week.wednesday'),
    },
    {
      value: 4,
      label: t('app.crontab.enum.week.thursday'),
    },
    {
      value: 5,
      label: t('app.crontab.enum.week.friday'),
    },
    {
      value: 6,
      label: t('app.crontab.enum.week.saturday'),
    },
    {
      value: 7,
      label: t('app.crontab.enum.week.sunday'),
    },
  ];

  // 控制不同时间单位的显示逻辑
  const showWeek = (item: PeriodDetailDo) => {
    return item.type === CRONTAB_PERIOD_TYPE.WEEKLY;
  };
  const showDay = (item: PeriodDetailDo) => {
    return [
      CRONTAB_PERIOD_TYPE.MONTHLY,
      CRONTAB_PERIOD_TYPE.EVERY_N_DAYS,
    ].includes(item.type);
  };
  const showHour = (item: PeriodDetailDo) => {
    return ![
      CRONTAB_PERIOD_TYPE.HOURLY,
      CRONTAB_PERIOD_TYPE.EVERY_N_MINUTES,
      CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS,
    ].includes(item.type);
  };
  const showMinute = (item: PeriodDetailDo) => {
    return ![CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS].includes(item.type);
  };
  const showSecond = (item: PeriodDetailDo) => {
    return item.type === CRONTAB_PERIOD_TYPE.EVERY_N_SECONDS;
  };

  onMounted(() => {
    if (props.modelValue.length === 0) {
      // 使用默认周期初始化
      emit('update:modelValue', [
        {
          type: CRONTAB_PERIOD_TYPE.WEEKLY,
          week: 1,
          day: 0,
          hour: 1,
          minute: 0,
          second: 0,
        },
      ]);
    }
  });
</script>
