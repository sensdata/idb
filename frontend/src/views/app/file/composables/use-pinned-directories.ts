import { ref, computed, watch } from 'vue';
import { useLogger } from '@/composables/use-logger';

const STORAGE_KEY = 'idb-pinned-directories';

export interface PinnedDirectory {
  path: string;
  name: string;
  lastSeen?: number;
  exists?: boolean;
}

// Global reactive state
const pinnedDirectories = ref<PinnedDirectory[]>([]);
let isInitialized = false;

const { logWarn } = useLogger('PinnedDirectories');

// Utility functions
const normalizePath = (path: string): string => {
  return path.replace(/\/+$/, '').replace(/^(?!\/)/, '/');
};

const getDisplayName = (path: string): string => {
  return path.split('/').pop() || path;
};

const loadFromStorage = (): PinnedDirectory[] => {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return [];

    const parsed = JSON.parse(stored);
    // Handle legacy format (array of strings) and convert to new format
    if (Array.isArray(parsed) && typeof parsed[0] === 'string') {
      return parsed.map((path) => ({
        path: normalizePath(path),
        name: getDisplayName(path),
        lastSeen: Date.now(),
        exists: true,
      }));
    }

    return Array.isArray(parsed) ? parsed : [];
  } catch (error) {
    logWarn('Failed to load pinned directories from localStorage:', error);
    return [];
  }
};

const saveToStorage = (directories: PinnedDirectory[]): void => {
  try {
    const dataToSave = JSON.stringify(directories);
    localStorage.setItem(STORAGE_KEY, dataToSave);
  } catch (error) {
    logWarn('Failed to save pinned directories to localStorage:', error);
  }
};

const initializeStorage = (): void => {
  if (isInitialized) return;

  pinnedDirectories.value = loadFromStorage();
  isInitialized = true;

  // Watch for changes and persist to localStorage
  watch(
    pinnedDirectories,
    (newDirectories) => {
      saveToStorage(newDirectories);
    },
    { deep: true, immediate: false }
  );
};

export function usePinnedDirectories() {
  initializeStorage();

  const pinnedPaths = computed(() =>
    pinnedDirectories.value.map((dir) => dir.path)
  );

  const isPinned = (path: string): boolean => {
    const normalizedPath = normalizePath(path);
    return pinnedPaths.value.includes(normalizedPath);
  };

  const pinDirectory = (path: string, customName?: string): void => {
    const normalizedPath = normalizePath(path);

    // Don't pin if already pinned
    if (isPinned(normalizedPath)) return;

    const newPinned: PinnedDirectory = {
      path: normalizedPath,
      name: customName || getDisplayName(normalizedPath),
      lastSeen: Date.now(),
      exists: true,
    };

    // Create a new array to ensure reactivity
    const newArray = [...pinnedDirectories.value, newPinned];

    // Sort pinned directories alphabetically by name
    newArray.sort((a, b) =>
      a.name.localeCompare(b.name, undefined, {
        numeric: true,
        caseFirst: 'lower',
      })
    );

    // Replace the entire array to trigger reactivity
    pinnedDirectories.value = newArray;
  };

  const unpinDirectory = (path: string): void => {
    const normalizedPath = normalizePath(path);
    const index = pinnedDirectories.value.findIndex(
      (dir) => dir.path === normalizedPath
    );

    if (index !== -1) {
      // Create a new array without the item to ensure reactivity
      const newArray = pinnedDirectories.value.filter(
        (dir) => dir.path !== normalizedPath
      );
      pinnedDirectories.value = newArray;
    }
  };

  const togglePin = (path: string, customName?: string): void => {
    if (isPinned(path)) {
      unpinDirectory(path);
    } else {
      pinDirectory(path, customName);
    }
  };

  const updateDirectoryExists = (path: string, exists: boolean): void => {
    const normalizedPath = normalizePath(path);
    const directoryIndex = pinnedDirectories.value.findIndex(
      (dir) => dir.path === normalizedPath
    );

    if (directoryIndex !== -1) {
      // Create a new array with updated directory to ensure reactivity
      const newArray = [...pinnedDirectories.value];
      newArray[directoryIndex] = {
        ...newArray[directoryIndex],
        exists,
        lastSeen: Date.now(),
      };
      pinnedDirectories.value = newArray;
    }
  };

  const removeMissingDirectories = (): void => {
    pinnedDirectories.value = pinnedDirectories.value.filter(
      (dir) => dir.exists !== false
    );
  };

  const getPinnedDirectory = (path: string): PinnedDirectory | undefined => {
    const normalizedPath = normalizePath(path);
    return pinnedDirectories.value.find((dir) => dir.path === normalizedPath);
  };

  return {
    pinnedDirectories: computed(() => pinnedDirectories.value),
    pinnedPaths,
    isPinned,
    pinDirectory,
    unpinDirectory,
    togglePin,
    updateDirectoryExists,
    removeMissingDirectories,
    getPinnedDirectory,
  };
}
