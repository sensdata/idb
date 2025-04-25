import { ref, Ref, nextTick } from 'vue';
import type { ComponentPublicInstance } from 'vue';

export interface DropdownOption {
  value: string;
  label: string;
  isDir?: boolean;
  displayValue?: string;
}

interface DropdownContentRef {
  contentRef: HTMLElement | null;
  optionRefs: (ComponentPublicInstance | Element | null)[];
}

export default function useDropdownNavigation(
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>,
  dropdownContentRef: Ref<DropdownContentRef | null>
) {
  const currentSelectedIndex = ref(-1);
  const hoverItem = ref<DropdownOption | null>(null);
  const preloadTimeoutId = ref<ReturnType<typeof setTimeout> | null>(null);

  function ensureSelectedItemVisible() {
    nextTick(() => {
      if (!dropdownContentRef.value?.contentRef) return;

      const container = dropdownContentRef.value.contentRef;
      const selectedOptionEl =
        dropdownContentRef.value.optionRefs[currentSelectedIndex.value];

      if (!selectedOptionEl) return;

      // 获取实际的DOM元素（无论是从ComponentPublicInstance还是直接获取）
      const selectedOption =
        (selectedOptionEl as ComponentPublicInstance)?.$el ||
        (selectedOptionEl as Element);
      const containerTop = container.scrollTop;
      const containerBottom = containerTop + container.clientHeight;
      const elementTop = selectedOption.offsetTop;
      const elementBottom = elementTop + selectedOption.clientHeight;

      if (elementBottom > containerBottom) {
        container.scrollTop = elementBottom - container.clientHeight;
      } else if (elementTop < containerTop) {
        container.scrollTop = elementTop;
      }
    });
  }

  function handleKeyUp() {
    if (!popupVisible.value) {
      popupVisible.value = true;
      return;
    }

    if (currentSelectedIndex.value > 0) {
      currentSelectedIndex.value--;
      ensureSelectedItemVisible();
    } else {
      currentSelectedIndex.value = allOptions.value.length - 1;
      ensureSelectedItemVisible();
    }
  }

  function handleKeyDown() {
    if (!popupVisible.value) {
      popupVisible.value = true;
      return;
    }

    if (currentSelectedIndex.value < allOptions.value.length - 1) {
      currentSelectedIndex.value++;
      ensureSelectedItemVisible();
    } else {
      currentSelectedIndex.value = 0;
      ensureSelectedItemVisible();
    }
  }

  return {
    currentSelectedIndex,
    hoverItem,
    preloadTimeoutId,
    ensureSelectedItemVisible,
    handleKeyUp,
    handleKeyDown,
  };
}
