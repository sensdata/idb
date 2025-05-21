export type ConfigMode = 'visual' | 'source';

export interface SSHConfig {
  port: string;
  listenAddress: string;
  permitRootLogin: string;
  passwordAuth: boolean;
  keyAuth: boolean;
  reverseLookup: boolean;
}

export interface LoadingStates {
  port: boolean;
  listen: boolean;
  root: boolean;
  passwordAuth: boolean;
  keyAuth: boolean;
  reverseLookup: boolean;
  sourceConfig: boolean;
}

export interface EditorRefType {
  checkUnsavedChanges: () => boolean;
}
