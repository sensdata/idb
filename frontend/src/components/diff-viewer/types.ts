export interface ParsedDiff {
  historical: string;
  current: string;
}

export interface DiffViewerExpose {
  show: (onRestoreSuccess?: () => void) => Promise<void>;
  executeRestoreSuccessCallback: () => void;
}
