<template>
  <div class="p-4">
    <address-bar path="idb-prd/apps/my-sql" :items="items" @goto="handleGoto" />
  </div>
</template>

<script lang="ts" setup>
  import AddressBar from '@/components/address-bar/index.vue';

  const items = [
    {
      path: 'idb-prd/apps/my-sql/aaa',
      name: 'aaa',
    },
    {
      path: 'idb-prd/apps/my-sql/aab',
      name: 'aab',
    },
    {
      path: 'idb-prd/apps/my-sql/aac',
      name: 'aac',
    },
    {
      path: 'idb-prd/apps/my-sql/abb',
      name: 'abb',
    },
    {
      path: 'idb-prd/apps/my-sql/abc',
      name: 'abc',
    },
    {
      path: 'idb-prd/apps/my-sql/acc',
      name: 'acc',
    },
    {
      path: 'idb-prd/apps/my-sql/bbc',
      name: 'bbc',
    },
  ];

  const handleGoto = (path: string) => {
    console.log(path);
  };
</script>
