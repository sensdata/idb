<template>
  <div class="simplified-file-tree">
    <ul class="tree-list">
      <li class="tree-section-title">
        {{ t('app.file.sidebar.favorites') }}
      </li>

      <!-- Pinned directories section -->
      <template v-if="pinnedItems.length > 0">
        <li
          v-for="item of pinnedItems"
          :key="`pinned-${item.path}`"
          class="tree-item pinned"
          :class="{ selected: isSelected(item) }"
          @click="handleItemClick(item)"
          @dblclick="handleItemDoubleClick(item)"
        >
          <div class="tree-item-container">
            <!-- Content area -->
            <div class="tree-item-content" :title="item.path">
              <div class="tree-item-icon">
                <folder-icon />
              </div>
              <div class="tree-item-text">
                {{ item.name }}
              </div>
              <div class="tree-item-unpin" @click.stop="handleUnpinClick(item)">
                <icon-close class="unpin-icon" />
              </div>
            </div>
          </div>
        </li>

        <!-- Separator between pinned and regular directories -->
        <li class="tree-separator">
          <div class="separator-line"></div>
        </li>
      </template>
      <li v-else class="tree-section-empty">
        {{ t('app.file.sidebar.noFavorites') }}
      </li>

      <li class="tree-section-title">
        {{ t('app.file.sidebar.directories') }}
      </li>

      <!-- Regular directories section -->
      <li
        v-for="item of regularItems"
        :key="item.path"
        class="tree-item"
        :class="{ selected: isSelected(item) }"
        @click="handleItemClick(item)"
        @dblclick="handleItemDoubleClick(item)"
      >
        <div class="tree-item-container">
          <!-- Content area -->
          <div class="tree-item-content">
            <div class="tree-item-icon">
              <folder-icon />
            </div>
            <div class="tree-item-text">
              {{ item.name }}
            </div>
          </div>
        </div>
      </li>
    </ul>
  </div>
</template>

<script lang="ts" setup>
  import { computed, ref, onMounted, watch } from 'vue';
  import { IconClose } from '@arco-design/web-vue/es/icon';
  import { useI18n } from 'vue-i18n';
  import { SimpleFileInfoEntity } from '@/entity/FileInfo';
  import { getDirectoryInfoApi } from '@/api/file';
  import { useLogger } from '@/composables/use-logger';
  import FolderIcon from '@/assets/icons/color-folder.svg';
  import { usePinnedDirectories } from '../composables/use-pinned-directories';
  import { FileTreeItem } from './file-tree/type';

  const props = defineProps<{
    // Complete file tree
    items: FileTreeItem[];
    // Whether to show hidden files/folders
    showHidden: boolean;
    // Current selected item
    current: SimpleFileInfoEntity | null;
  }>();

  const emit = defineEmits(['itemSelect', 'itemDoubleClick']);
  const { t } = useI18n();

  const {
    pinnedDirectories,
    isPinned,
    unpinDirectory,
    updateDirectoryExists,
    pinnedPathsString,
    loadFavorites,
  } = usePinnedDirectories();

  const { logWarn } = useLogger('SimplifiedFileTree');

  // Get pinned directory items (including nested ones not in root)
  const pinnedDirectoryItems = ref<FileTreeItem[]>([]);

  // Fetch info for pinned directories that are not in the root items
  const fetchPinnedDirectoryInfo = async () => {
    const rootPaths = new Set(props.items.map((item) => item.path));
    // Do NOT clear pinnedDirectoryItems.value here!
    // const missingPinnedDirs: FileTreeItem[] = [];
    // Use Promise.all to avoid await in loop
    const pinnedDirPromises = pinnedDirectories.value
      .filter((pinnedDir) => !rootPaths.has(pinnedDir.path))
      .map(async (pinnedDir) => {
        try {
          const dirInfo = await getDirectoryInfoApi({ path: pinnedDir.path });
          if (dirInfo && dirInfo.is_dir) {
            updateDirectoryExists(pinnedDir.path, true);
            return {
              name: dirInfo.name,
              path: dirInfo.path,
              is_dir: dirInfo.is_dir as boolean,
              extension: dirInfo.extension,
              size: dirInfo.size,
              is_hidden: dirInfo.is_hidden,
            } as FileTreeItem;
          }
          updateDirectoryExists(pinnedDir.path, false);
          return null;
        } catch (error) {
          logWarn(
            `Failed to fetch pinned directory info: ${pinnedDir.path}`,
            error
          );
          updateDirectoryExists(pinnedDir.path, false);
          return null;
        }
      });

    const results = await Promise.all(pinnedDirPromises);
    const validResults = results.filter(
      (item) => item !== null
    ) as FileTreeItem[];
    // Only update after fetch completes
    pinnedDirectoryItems.value = validResults;
  };

  // Watch for changes in pinned directories and refetch when needed
  onMounted(async () => {
    await loadFavorites();
    fetchPinnedDirectoryInfo();
  });

  // Watch only for actual path changes, not metadata updates
  // This prevents the race condition where updateDirectoryExists triggers unnecessary refetches
  watch(
    () => pinnedPathsString.value,
    (newPathsString, oldPathsString) => {
      // Only refetch if the paths actually changed, not just metadata
      if (newPathsString !== oldPathsString) {
        fetchPinnedDirectoryInfo();
      }
    }
  );

  // Pinned directories (combination of root dirs that are pinned + fetched nested dirs)
  const pinnedItems = computed(() => {
    const rootPinnedItems = props.items.filter(
      (item) =>
        item.is_dir &&
        isPinned(item.path) &&
        (props.showHidden || !item.is_hidden)
    );

    // Create a map of existing items for quick lookup
    const existingItemsMap = new Map<string, FileTreeItem>();

    // Add root pinned items
    rootPinnedItems.forEach((item) => {
      existingItemsMap.set(item.path, item);
    });

    // Add fetched pinned directory items
    pinnedDirectoryItems.value.forEach((item) => {
      existingItemsMap.set(item.path, item);
    });

    // Ensure ALL pinned directories are included, even if they're not in current items
    // This prevents disappearing during navigation
    pinnedDirectories.value.forEach((pinnedDir) => {
      if (pinnedDir.exists !== false && !existingItemsMap.has(pinnedDir.path)) {
        // Create a minimal item representation for directories not currently loaded
        const fallbackItem: FileTreeItem = {
          name: pinnedDir.name,
          path: pinnedDir.path,
          is_dir: true,
          extension: '',
          size: 0,
          is_hidden: false,
        };
        existingItemsMap.set(pinnedDir.path, fallbackItem);
      }
    });

    // Convert map values to array and filter by showHidden
    const allPinnedItems = Array.from(existingItemsMap.values()).filter(
      (item) => props.showHidden || !item.is_hidden
    );

    return allPinnedItems.sort((a, b) =>
      a.name.localeCompare(b.name, undefined, {
        numeric: true,
        caseFirst: 'lower',
      })
    );
  });

  // Regular directories (excluding pinned ones)
  const regularItems = computed(() => {
    return props.items
      .filter(
        (item) =>
          item.is_dir &&
          !isPinned(item.path) &&
          (props.showHidden || !item.is_hidden)
      )
      .sort((a, b) =>
        a.name.localeCompare(b.name, undefined, {
          numeric: true,
          caseFirst: 'lower',
        })
      );
  });

  // Check if item is selected (by comparing paths)
  const isSelected = (item: FileTreeItem) => {
    if (!props.current) return false;

    // Get current path and item path
    const currentPath = props.current.path;
    const itemPath = item.path;

    // For exact matches, always highlight
    if (currentPath === itemPath) {
      return true;
    }

    // For parent directories, highlight if current path is within this directory
    // This ensures parent directories stay active when navigating to subdirectories
    if (item.is_dir && currentPath.startsWith(itemPath + '/')) {
      return true;
    }

    return false;
  };

  // Handle click event - emit selection event
  const handleItemClick = (item: FileTreeItem) => {
    emit('itemSelect', item);
  };

  // Handle double-click event - emit double-click event
  const handleItemDoubleClick = (item: FileTreeItem) => {
    emit('itemDoubleClick', item);
  };

  // Handle unpin click event for pinned items
  const handleUnpinClick = async (item: FileTreeItem) => {
    try {
      await unpinDirectory(item.path);
    } catch (error) {
      logWarn(`Failed to unpin directory: ${item.path}`, error);
    }
  };
</script>

<style scoped lang="less">
  .simplified-file-tree {
    position: relative;
    width: 100%;
    padding: 0.5rem 0.375rem 0.5rem 0.75rem;
    margin: 0;
  }

  .tree-list {
    position: relative;
    padding: 0;
    margin: 0;
    list-style: none;
  }

  .tree-section-title {
    padding: 0.25rem 0.5rem;
    margin: 0.5rem 0 0.375rem;
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1.25rem;
    color: var(--color-text-3);
    text-transform: uppercase;
    letter-spacing: 0.02em;
    list-style: none;
  }

  .tree-section-empty {
    padding: 0.25rem 0.5rem 0.5rem;
    font-size: 0.8125rem;
    line-height: 1.125rem;
    color: var(--color-text-4);
    list-style: none;
  }

  .tree-item {
    position: relative;
    width: 100%;
    height: 2.125rem;
    padding: 0 0.5rem;
    margin-bottom: 0.25rem;
    overflow: visible;
    cursor: pointer;
    list-style: none;
    background: transparent;
    border: none;
    border-radius: 0.375rem;
    transition: background-color 0.2s ease, box-shadow 0.2s ease;
  }

  /* Pinned directory styling */
  .tree-item.pinned {
    background-color: var(--color-fill-1);
  }

  /* Hover state */
  .tree-item:hover {
    background-color: var(--color-fill-1);
  }

  .tree-item.pinned:hover {
    background-color: var(--color-fill-2);
  }

  /* Selected state background */
  .tree-item.selected {
    background-color: var(--color-fill-2);
  }

  .tree-item.pinned.selected {
    background-color: rgb(var(--primary-1));
    box-shadow: inset 0 0 0 1px rgb(var(--primary-3));
  }

  /* Selected state indicator */
  .tree-item.selected::before {
    position: absolute;
    top: 15%;
    left: -0.25rem;
    width: 0.2rem;
    height: 70%;
    content: '';
    background-color: rgb(var(--primary-6));
    border-radius: 0.125rem;
  }

  /* Container base style */
  .tree-item-container {
    position: relative;
    display: flex;
    align-items: center;
    width: 100%;
    height: 2rem;
    padding: 0;
    margin: 0;
    background: transparent;
    border: none;
    border-radius: 0;
  }

  /* Content area style */
  .tree-item-content {
    display: flex;
    flex: 1;
    align-items: center;
    width: 100%;
    min-width: 0;
    height: 100%;
    padding: 0.25rem 0.375rem;
    margin: 0;
    background: transparent;
  }

  /* Icon style */
  .tree-item-icon {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 0.875rem;
    height: 0.875rem;
    margin-right: 0.5rem;
  }

  .tree-item-icon :deep(svg) {
    width: 0.875rem;
    height: 0.875rem;
  }

  /* Text style */
  .tree-item-text {
    flex: 1;
    min-width: 0;
    margin-left: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.9375rem;
    line-height: 1.25rem;
    color: var(--color-text-1);
    white-space: nowrap;
  }

  /* Pin icon style */
  .tree-item-pin {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 1rem;
    height: 1rem;
    margin-left: 0.25rem;
    opacity: 0.6;
  }

  .pin-icon {
    width: 0.75rem;
    height: 0.75rem;
    color: var(--color-text-3);
  }

  /* Unpin icon style */
  .tree-item-unpin {
    display: flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    width: 1rem;
    height: 1rem;
    margin-left: 0.1rem;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.2s, color 0.2s;
  }

  .tree-item.pinned:hover .tree-item-unpin,
  .tree-item.pinned.selected .tree-item-unpin {
    opacity: 0.72;
  }

  .tree-item-unpin:hover {
    opacity: 1;
  }

  .unpin-icon {
    font-size: 0.75rem;
    color: var(--color-text-2);
  }

  /* Separator styling */
  .tree-separator {
    position: relative;
    width: 100%;
    height: 0.5rem;
    margin: 0.25rem 0 0.375rem;
    pointer-events: none;
    list-style: none;
  }

  .separator-line {
    position: absolute;
    top: 50%;
    right: 0.25rem;
    left: 0.25rem;
    height: 1px;
    background-color: var(--color-border-2);
    transform: translateY(-50%);
  }
</style>
