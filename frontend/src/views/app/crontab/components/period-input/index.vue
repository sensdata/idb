<template>
  <div class="perioid-wrap">
    <div
      v-for="(item, index) of modelValue"
      :key="index"
      class="period-item mb-4"
    >
      <a-select
        v-model="item.type"
        class="w-[120px]"
        :options="periodTypeOptions"
      />
      <a-select
        v-if="showWeek(item)"
        v-model="item.week"
        class="w-[120px] ml-4"
        :options="weekOptions"
      />
      <a-input-number
        v-if="showDay(item)"
        v-model="item.day"
        class="w-[120px] ml-4"
        :min="1"
        :max="31"
      >
        <template #suffix> {{ $t('common.timeUnit.day') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showHour(item)"
        v-model="item.hour"
        class="w-[120px] ml-4"
        :min="0"
        :max="23"
      >
        <template #suffix> {{ $t('common.timeUnit.hour') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showMinute(item)"
        v-model="item.minute"
        class="w-[120px] ml-4"
        :min="0"
        :max="59"
      >
        <template #suffix> {{ $t('common.timeUnit.minute') }} </template>
      </a-input-number>
      <a-input-number
        v-if="showSecond(item)"
        v-model="item.second"
        class="w-[120px] ml-4"
        :min="0"
        :max="59"
      >
        <template #suffix> {{ $t('common.timeUnit.second') }} </template>
      </a-input-number>
      <a-button
        v-if="modelValue.length > 1"
        type="text"
        class="ml-4"
        @click="removePeriod(index)"
      >
        {{ $t('common.delete') }}
      </a-button>
    </div>
    <a-button
      v-if="modelValue.length < 5"
      type="secondary"
      size="small"
      @click="addPeriod"
    >
      {{ $t('common.add') }}
    </a-button>
  </div>
</template>

<script lang="ts" setup>
  import { CRONTAB_PERIOD_TYPE } from '@/config/enum';
  import { PeriodDetailDo } from '@/entity/Crontab';
  import { onMounted } from 'vue';
  import { useI18n } from 'vue-i18n';

  const props = defineProps<{
    modelValue: PeriodDetailDo[];
  }>();
  const emit = defineEmits(['update:modelValue']);

  const { t } = useI18n();

  const addPeriod = () => {
    emit('update:modelValue', [
      ...props.modelValue,
      {
        type: CRONTAB_PERIOD_TYPE.WEEKLY,
        week: 1,
        day: 0,
        hour: 1,
        minute: 0,
        second: 0,
      },
    ]);
  };

  const removePeriod = (index: number) => {
    emit(
      'update:modelValue',
      props.modelValue.filter((_, i) => i !== index)
    );
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
      addPeriod();
    }
  });

  defineExpose({
    addPeriod,
    removePeriod,
  });
</script>
