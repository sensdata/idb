import { ref, Ref, computed } from 'vue';
import { debounce } from 'lodash';
import { searchFileListApi } from '@/api/file';
import { FileInfoEntity } from '@/entity/FileInfo';
import { getSearchPath, getSearchTerm } from '../utils';
import { DropdownOption } from './use-dropdown-navigation';

export default function useAddressBarSearch(
  value: Ref<string>,
  emit: any,
  allOptions: Ref<DropdownOption[]>,
  popupVisible: Ref<boolean>,
  dropdownContentRef: Ref<HTMLElement | null>
) {
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

  const handleInputValueChange = debounce(() => {
    lastInputTime.value = Date.now();

    const searchPath = getSearchPath(value.value);
    const searchTerm = getSearchTerm(value.value);

    searchWord.value = searchTerm;
    currentPage.value = 1;
    hasMore.value = true;

    const searchParams = {
      path: searchPath,
      word: searchTerm,
    };

    if (searchTerm) {
      isSearching.value = true;
      popupVisible.value = true;
    } else {
      isSearching.value = false;
      popupVisible.value = false;
    }

    triggerByTab.value = false;
    emit('search', searchParams);
  }, 150);

  const loadMoreItems = async () => {
    if (isLoading.value || !hasMore.value || !searchWord.value) return;

    isLoading.value = true;
    try {
      currentPage.value++;

      const searchPath = getSearchPath(value.value);

      const data = await searchFileListApi({
        page: currentPage.value,
        page_size: 100,
        show_hidden: true,
        path: searchPath,
        dir: true,
        search: searchWord.value,
      });

      if (data.items && data.items.length > 0) {
        const newOptions = data.items.map((item: FileInfoEntity) => ({
          value: item.name,
          label: item.name,
          isDir: item.is_dir,
          displayValue: item.is_dir ? `${item.name}/` : item.name,
        }));

        const searchTerm = searchWord.value.toLowerCase();
        const filteredOptions = newOptions.filter((option) =>
          option.value.toLowerCase().startsWith(searchTerm)
        );

        allOptions.value = [...allOptions.value, ...filteredOptions];
      } else {
        hasMore.value = false;
      }
    } catch (error) {
      console.error('Failed to load more items:', error);
      hasMore.value = false;
    } finally {
      isLoading.value = false;
    }
  };

  const handleScroll = debounce(async () => {
    if (!dropdownContentRef.value) return;

    const { scrollTop, scrollHeight, clientHeight } = dropdownContentRef.value;
    if (
      scrollHeight - scrollTop - clientHeight < 50 &&
      !isLoading.value &&
      hasMore.value
    ) {
      await loadMoreItems();
    }
  }, 200);

  return {
    currentPage,
    hasMore,
    isLoading,
    isSearching,
    searchWord,
    triggerByTab,
    computedPopupVisible,
    handleInputValueChange,
    loadMoreItems,
    handleScroll,
  };
}
