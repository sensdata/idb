import { ref, computed, watch } from 'vue';
import {
  FavoriteFileEntity,
  favoriteFileApi,
  getFavoriteFilesApi,
  unFavoriteFileApi,
} from '@/api/file';
import { useLogger } from '@/composables/use-logger';
import useCurrentHost from '@/composables/current-host';

export interface PinnedDirectory {
  id?: number;
  path: string;
  name: string;
  lastSeen?: number;
  exists?: boolean;
}

// Global reactive state
const pinnedDirectories = ref<PinnedDirectory[]>([]);
const loadingFavorites = ref(false);
let loadedHostId: number | undefined;
let loadRequestId = 0;

const { logWarn } = useLogger('PinnedDirectories');

// Utility functions
const normalizePath = (path: string): string => {
  return path.replace(/\/+$/, '').replace(/^(?!\/)/, '/');
};

const getDisplayName = (path: string): string => {
  return path.split('/').pop() || path;
};

const sortDirectories = (directories: PinnedDirectory[]): PinnedDirectory[] => {
  return [...directories].sort((a, b) =>
    a.name.localeCompare(b.name, undefined, {
      numeric: true,
      caseFirst: 'lower',
    })
  );
};

const mapFavoriteToPinned = (favorite: FavoriteFileEntity): PinnedDirectory => {
  const path = normalizePath(String(favorite.source || '/'));
  return {
    id: favorite.id,
    path,
    name: String(favorite.name || getDisplayName(path)),
    lastSeen: Date.now(),
    exists: true,
  };
};

export function usePinnedDirectories() {
  const { currentHostId } = useCurrentHost();

  // Add a computed property for watching path changes only
  const pinnedPathsString = computed(() =>
    pinnedDirectories.value
      .map((dir) => dir.path)
      .sort()
      .join('|')
  );

  const isPinned = (path: string): boolean => {
    const normalizedPath = normalizePath(path);
    return pinnedDirectories.value.some((dir) => dir.path === normalizedPath);
  };

  const loadFavorites = async (force = false): Promise<void> => {
    const hostId = currentHostId.value;
    const requestId = ++loadRequestId;
    if (!hostId) {
      pinnedDirectories.value = [];
      loadedHostId = undefined;
      loadingFavorites.value = false;
      return;
    }

    if (!force && loadedHostId === hostId) {
      return;
    }

    loadingFavorites.value = true;
    try {
      const res = await getFavoriteFilesApi({ page: 1, page_size: 1000 });
      // 忽略过期请求，避免快速切主机造成数据串台
      if (requestId !== loadRequestId || currentHostId.value !== hostId) {
        return;
      }
      const items = Array.isArray(res?.items) ? res.items : [];
      pinnedDirectories.value = sortDirectories(items.map(mapFavoriteToPinned));
      loadedHostId = hostId;
    } catch (error) {
      if (requestId !== loadRequestId) {
        return;
      }
      logWarn('Failed to load favorite directories:', error);
      pinnedDirectories.value = [];
    } finally {
      if (requestId === loadRequestId) {
        loadingFavorites.value = false;
      }
    }
  };

  watch(
    currentHostId,
    () => {
      loadFavorites(true).catch((error) => {
        logWarn(
          'Failed to refresh favorite directories after host change:',
          error
        );
      });
    },
    { immediate: false }
  );

  const pinDirectory = async (
    path: string,
    customName?: string
  ): Promise<void> => {
    const normalizedPath = normalizePath(path);

    // Don't pin if already pinned
    if (isPinned(normalizedPath)) return;

    const favorite = await favoriteFileApi({ source: normalizedPath });
    const newPinned: PinnedDirectory = favorite?.id
      ? mapFavoriteToPinned(favorite)
      : {
          path: normalizedPath,
          name: customName || getDisplayName(normalizedPath),
          lastSeen: Date.now(),
          exists: true,
        };

    pinnedDirectories.value = sortDirectories([
      ...pinnedDirectories.value,
      newPinned,
    ]);
  };

  const unpinDirectory = async (path: string): Promise<void> => {
    const normalizedPath = normalizePath(path);
    const target = pinnedDirectories.value.find(
      (dir) => dir.path === normalizedPath
    );

    if (!target) return;
    if (target.id) {
      await unFavoriteFileApi({ id: target.id });
      pinnedDirectories.value = pinnedDirectories.value.filter(
        (dir) => dir.path !== normalizedPath
      );
      return;
    }

    await loadFavorites(true);
    const refreshed = pinnedDirectories.value.find(
      (dir) => dir.path === normalizedPath
    );
    if (refreshed?.id) {
      await unFavoriteFileApi({ id: refreshed.id });
    }
    pinnedDirectories.value = pinnedDirectories.value.filter(
      (dir) => dir.path !== normalizedPath
    );
  };

  const updateDirectoryExists = (path: string, exists: boolean): void => {
    const normalizedPath = normalizePath(path);
    const directoryIndex = pinnedDirectories.value.findIndex(
      (dir) => dir.path === normalizedPath
    );

    if (directoryIndex !== -1) {
      const currentDir = pinnedDirectories.value[directoryIndex];

      // Only update if the exists status actually changed
      if (currentDir.exists !== exists) {
        // Create a new array with updated directory to ensure reactivity
        const newArray = [...pinnedDirectories.value];
        newArray[directoryIndex] = {
          ...newArray[directoryIndex],
          exists,
          lastSeen: Date.now(),
        };
        pinnedDirectories.value = newArray;
      }
    }
  };

  return {
    pinnedDirectories: computed(() => pinnedDirectories.value),
    loadingFavorites: computed(() => loadingFavorites.value),
    pinnedPathsString, // For efficient watching of path changes only
    loadFavorites,
    isPinned,
    pinDirectory,
    unpinDirectory,
    updateDirectoryExists,
  };
}
