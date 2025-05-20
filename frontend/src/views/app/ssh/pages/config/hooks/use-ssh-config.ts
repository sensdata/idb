import { ref } from 'vue';
import { Message } from '@arco-design/web-vue';
import { useI18n } from 'vue-i18n';
import { getSSHConfig, updateSSHConfigContent } from '@/api/ssh';
import { SSHConfig } from '../../../types';

export function useSSHConfig(hostId: number) {
  const { t } = useI18n();

  // Create a reactive SSH config object with defaults
  const sshConfig = ref<SSHConfig>({
    port: '22',
    listenAddress: '0.0.0.0',
    permitRootLogin: 'yes',
    passwordAuth: true,
    keyAuth: true,
    reverseLookup: true,
    autoStart: true,
  });

  /**
   * Fetch SSH configuration from the server
   */
  const fetchConfig = async () => {
    try {
      if (!hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      const response = await getSSHConfig(hostId);

      // Map API response to the local config format
      sshConfig.value = {
        port: response.port || '22',
        listenAddress: response.listen_address || '0.0.0.0',
        permitRootLogin: response.permit_root_login || 'yes',
        passwordAuth: response.password_authentication === 'yes',
        keyAuth: response.pubkey_authentication === 'yes',
        reverseLookup: response.use_dns === 'yes',
        autoStart: response.auto_start,
      };
    } catch (error) {
      console.error('Error fetching SSH config:', error);
      Message.error(t('app.ssh.error.fetchFailed'));
    }
  };

  /**
   * Update SSH configuration to the server
   */
  const updateConfig = async () => {
    try {
      if (!hostId) {
        Message.error(t('app.ssh.error.noHost'));
        return;
      }

      // Create the config file content from current state
      const configContent = `# /etc/ssh/sshd_config
# SSH服务器配置文件

# 端口设置
Port ${sshConfig.value.port}

# 监听地址
ListenAddress ${sshConfig.value.listenAddress}

# Root登录设置
PermitRootLogin ${sshConfig.value.permitRootLogin}

# 认证设置
PasswordAuthentication ${sshConfig.value.passwordAuth ? 'yes' : 'no'}
PubkeyAuthentication ${sshConfig.value.keyAuth ? 'yes' : 'no'}

# DNS反向解析
UseDNS ${sshConfig.value.reverseLookup ? 'yes' : 'no'}

# 其他设置
X11Forwarding yes
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server
`;

      await updateSSHConfigContent(hostId, configContent);
    } catch (error) {
      console.error('Error updating SSH config:', error);
      throw error;
    }
  };

  return {
    sshConfig,
    fetchConfig,
    updateConfig,
  };
}
