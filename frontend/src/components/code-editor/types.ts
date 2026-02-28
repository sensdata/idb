export interface FileInfo {
  name: string;
  path: string;
  size?: number;
}

export interface EditorProps {
  modelValue: string;
  loading?: boolean;
  readOnly?: boolean;
  autofocus?: boolean;
  indentWithTab?: boolean;
  tabSize?: number;
  file?: FileInfo | null;
  extensions?: any[];
  isPartialView?: boolean;
  loadingText?: string;
}

export interface EditorEmits {
  (e: 'update:modelValue', value: string): void;
  (e: 'editorReady', payload: { view: any }): void;
  (e: 'contentDoubleClick'): void;
}

export interface EditorInstance {
  editorView: any;
  focus: () => void;
  scrollToTop: () => void;
  scrollToBottom: () => void;
}
