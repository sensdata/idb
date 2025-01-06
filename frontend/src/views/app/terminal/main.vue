<script setup lang="ts">
  import { inject, onMounted, onUnmounted } from 'vue';

  const openTerminal = inject<() => void>('openTerminal');
  const closeTerminal = inject<() => void>('closeTerminal');

  onMounted(() => {
    openTerminal?.();
  });
  onUnmounted(() => {
    closeTerminal?.();
  });
</script>
