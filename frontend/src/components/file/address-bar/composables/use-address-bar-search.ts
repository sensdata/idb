import { ref, Ref, computed, onUnmounted } from 'vue';
import { debounce } from 'lodash';
import { searchFileListApi } from '@/api/file';
import { FileInfoEntity } from '@/entity/FileInfo';
import { useLogger } from '@/composables/use-logger';
import { getSearchPath, getSearchTerm } from '../utils';
import { DropdownOption } from './use-dropdown-navigation';

// Constants
const DEBOUNCE_DELAY = 150;
const SCROLL_DEBOUNCE_DELAY = 200;
const PAGE_SIZE = 100;
const SCROLL_THRESHOLD = 50;

// Types - 先定义所有接口
interface SearchParams {
  path: string;
  word: string;
}

interface SearchEmitFunction {
  (event: 'search', params: SearchParams): void;
}

interface UseAddressBarSearchParams {
  value: Ref<string>;
  emit: SearchEmitFunction;
  allOptions: Ref<DropdownOption[]>;
  popupVisible: Ref<boolean>;
}

export default function useAddressBarSearch({
  value,
  emit,
  allOptions,
  popupVisible,
}: UseAddressBarSearchParams) {
  const { logDebug, logError } = useLogger('AddressBarSearch');
  const currentPage = ref(1);
  const hasMore = ref(true);
  const isLoading = ref(false);
  const isSearching = ref(false);
  const searchWord = ref('');
  const triggerByTab = ref(false);
  const lastInputTime = ref(0);

  const computedPopupVisible = computed(
    () =>
      (popupVisible.value && allOptions.value.length > 0) || isSearching.value
  );

  const performSearch = (searchTerm: string, searchPath: string) => {
    const searchParams: SearchParams = {
      path: searchPath,
      word: searchTerm,
    };

    isSearching.value = !!searchTerm;
    popupVisible.value = !!searchTerm;
    emit('search', searchParams);
  };

  const handleInputValueChange = debounce(() => {
    lastInputTime.value = Date.now();

    const searchPath = getSearchPath(value.value);
    const searchTerm = getSearchTerm(value.value);

    searchWord.value = searchTerm;
    currentPage.value = 1;
    hasMore.value = true;

    performSearch(searchTerm, searchPath);
    triggerByTab.value = false;
  }, DEBOUNCE_DELAY);

  const handleSearchComplete = () => {
    logDebug('🔍 handleSearchComplete called, resetting isSearching to false');
    isSearching.value = false;
  };

  const loadMoreItems = async (): Promise<void> => {
    if (isLoading.value || !hasMore.value || !searchWord.value) return;

    isLoading.value = true;
    try {
      currentPage.value++;

      const searchPath = getSearchPath(value.value);

      const data = await searchFileListApi({
        page: currentPage.value,
        page_size: PAGE_SIZE,
        show_hidden: true,
        path: searchPath,
        dir: true,
        search: searchWord.value,
      });

      if (data.items?.length > 0) {
        const newOptions: DropdownOption[] = data.items.map(
          (item: FileInfoEntity) => ({
            value: item.name,
            label: item.name,
            isDir: item.is_dir,
            displayValue: item.is_dir ? `${item.name}/` : item.name,
          })
        );

        const searchTerm = searchWord.value.toLowerCase();
        const filteredOptions = newOptions.filter((option) =>
          option.value.toLowerCase().startsWith(searchTerm)
        );

        allOptions.value = [...allOptions.value, ...filteredOptions];
      } else {
        hasMore.value = false;
      }
    } catch (error) {
      logError('Failed to load more items:', error);
      hasMore.value = false;
    } finally {
      isLoading.value = false;
    }
  };

  const handleScroll = debounce(async (element: HTMLElement) => {
    if (!element) return;

    const { scrollTop, scrollHeight, clientHeight } = element;
    const isNearBottom =
      scrollHeight - scrollTop - clientHeight < SCROLL_THRESHOLD;

    if (isNearBottom && !isLoading.value && hasMore.value) {
      await loadMoreItems();
    }
  }, SCROLL_DEBOUNCE_DELAY);

  // Cleanup debounced functions on unmount
  onUnmounted(() => {
    handleInputValueChange.cancel();
    handleScroll.cancel();
  });

  return {
    // State
    currentPage,
    hasMore,
    isLoading,
    isSearching,
    searchWord,
    triggerByTab,

    // Computed
    computedPopupVisible,

    // Methods
    handleInputValueChange,
    handleSearchComplete,
    loadMoreItems,
    handleScroll,
  } as const;
}
