/**
 * SSH Configuration Response Interface
 */
export interface SSHConfigResponse {
  port: string;
  listen_address: string;
  permit_root_login: string;
  password_authentication: string;
  pubkey_authentication: string;
  use_dns: string;
  auto_start: boolean;
}

/**
 * SSH Configuration Content Response Interface
 */
export interface SSHConfigContentResponse {
  content: string;
}

/**
 * SSH Configuration Interface (for use in the component)
 */
export interface SSHConfig {
  port: string;
  listenAddress: string;
  permitRootLogin: string;
  passwordAuth: boolean;
  keyAuth: boolean;
  reverseLookup: boolean;
  autoStart: boolean;
}
