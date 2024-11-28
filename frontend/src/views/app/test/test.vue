<template>
  <address-bar path="idb-prd/apps/my-sql" :items="items" @goto="handleGoto" />
  <tree-item :items="items" />
</template>

<script lang="ts" setup>
  import { reactive } from 'vue';
  import AddressBar from '@/components/address-bar/index.vue';
  import TreeItem from '@/components/file-tree/index.vue';

  const items: any = reactive([
    {
      path: 'idb-prd/apps/my-sql/aaa',
      name: 'aaa',
      is_dir: true,
      items: [
        {
          name: 'aaa1',
          path: 'idb-prd/apps/my-sql/aaa/aaa1',
          is_dir: true,
          items: [
            {
              name: 'aaa1-1',
              path: 'idb-prd/apps/my-sql/aaa/aaa1/aaa1-1',
            },
            {
              name: 'aaa1-2',
              path: 'idb-prd/apps/my-sql/aaa/aaa1/aaa1-2',
              is_dir: true,
              items: [
                {
                  name: 'aaa1-2-1',
                  path: 'idb-prd/apps/my-sql/aaa/aaa1/aaa1-2/aaa1-2-1',
                  is_dir: true,
                },
                {
                  name: 'aaa1-2-2',
                  path: 'idb-prd/apps/my-sql/aaa/aaa1/aaa1-2/aaa1-2-2',
                },
              ],
            },
          ],
        },
        {
          name: 'aaa2',
          path: 'idb-prd/apps/my-sql/aaa/aaa2',
        },
      ],
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
  ]);

  const handleGoto = (path: string) => {
    console.log(path);
  };
</script>
