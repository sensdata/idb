import { computed } from 'vue';
import { useHostStore } from '@/store';
import { useRoute, useRouter } from 'vue-router';

export default function usetCurrentHost() {
  const route = useRoute();
  const router = useRouter();
  const hostStore = useHostStore();
  const currentHostId = computed(() => {
    return +(route?.query?.id || hostStore.currentId || '') || undefined;
  });
  if (currentHostId.value !== hostStore.currentId) {
    hostStore.setCurrentId(currentHostId.value);
  }
  return {
    currentHostId,
    switchHost: (id: number, redirect = false) => {
      hostStore.setCurrentId(id);
      if (redirect) {
        router.replace({
          query: {
            ...route?.query,
            id,
          },
        });
      }
    },
  };
}
