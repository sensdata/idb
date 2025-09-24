# Docker 安装检测组件

这是一个用于检测和安装 Docker 环境的 Vue 组件。

## 功能特性

- 🔍 自动检测 Docker 安装状态
- 📦 一键安装 Docker 环境
- 🔄 实时显示安装进度
- 🌐 支持国际化
- 📱 响应式设计

## 使用方法

### 基本使用

```vue
<template>
  <docker-install-guide 
    @status-change="handleStatusChange"
    @install-complete="handleInstallComplete"
  />
</template>

<script setup>
import DockerInstallGuide from '@/components/docker-install-guide/index.vue';

const handleStatusChange = (status) => {
  console.log('Docker 状态:', status);
};

const handleInstallComplete = () => {
  console.log('Docker 安装完成');
};
</script>
```

### 属性 (Props)

| 属性名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `showStatus` | `boolean` | `true` | 是否显示状态卡片 |
| `autoCheck` | `boolean` | `true` | 是否自动检查 Docker 状态 |

### 事件 (Events)

| 事件名 | 参数 | 说明 |
|--------|------|------|
| `status-change` | `status: string` | Docker 状态变化时触发 |
| `install-complete` | - | Docker 安装完成时触发 |

### 暴露的方法

| 方法名 | 说明 |
|--------|------|
| `checkDockerStatus()` | 手动检查 Docker 状态 |
| `dockerStatus` | 获取当前 Docker 状态 |

## API 接口

组件使用以下 API 接口：

- `getDockerInstallStatusApi()` - 获取 Docker 安装状态
- `dockerInstallApi()` - 安装 Docker

## 状态说明

- `installed` - Docker 已安装
- `not installed` - Docker 未安装
- `checking` - 正在检查状态

## 样式定制

组件使用 Arco Design 的样式系统，可以通过 CSS 变量进行定制：

```css
.docker-install-guide {
  /* 自定义样式 */
}
```

## 国际化

组件支持中英文国际化，相关翻译文件位于：

- `locale/zh-CN.ts` - 中文翻译
- `locale/en-US.ts` - 英文翻译

## 使用场景

1. **应用商店页面** - 在安装应用前检查 Docker 环境
2. **容器管理页面** - 确保 Docker 服务可用
3. **系统设置页面** - 提供 Docker 环境管理功能

## 注意事项

- 安装 Docker 需要管理员权限
- 安装过程可能需要几分钟时间
- 建议在安装完成后重启相关服务
