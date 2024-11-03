<template>
  <div class="box">
    <div class="line">
      <div class="col1">服务器时间</div>
      <div class="col2">2024-01-02 14:12:15</div>
      <div class="col3"></div>
      <div class="col4">
        <a-space>
          <a-button type="primary" size="mini">修改</a-button>
          <a-button type="primary" size="mini">同步时间</a-button>
        </a-space>
      </div>
    </div>
    <div class="line">
      <div class="col1">当前时区</div>
      <div class="col2">Asia/Shanghai</div>
      <div class="col3"></div>
      <div class="col4">
        <a-button type="primary" size="mini">修改</a-button>
      </div>
    </div>
    <div class="line">
      <div class="col1">启动时间</div>
      <div class="col2">2024-01-02 14:12:15</div>
      <div class="col3"></div>
      <div class="col4">
        <a-tag color="blue">繁忙</a-tag>
      </div>
    </div>
    <div class="line">
      <div class="col1">运行时间</div>
      <div class="col2">965天23小时11分12秒</div>
      <div class="col3"></div>
      <div class="col4"></div>
    </div>
    <div class="line">
      <div class="col1">空闲时间</div>
      <div class="col2">965天23小时11分12秒</div>
      <div class="col3"></div>
      <div class="col4">
        <a-tag color="green">92.41空闲</a-tag>
      </div>
    </div>
    <div class="line">
      <div class="col1">CPU使用率</div>
      <div class="col2">53.88%</div>
      <div class="col3"></div>
      <div class="col4"></div>
    </div>
    <div class="line">
      <div class="col1">当前负载</div>
      <div class="colspan">
        <div class="subline">
          <div class="col2">1分钟进程数目：</div>
          <div class="col3">0.00</div>
          <div class="col4">
            <a-tag color="blue">繁忙</a-tag>
          </div>
        </div>
        <div class="subline">
          <div class="col2">5分钟进程数目：</div>
          <div class="col3">0.23</div>
          <div class="col4">
            <a-tag color="cyan">正常</a-tag>
          </div>
        </div>
        <div class="subline">
          <div class="col2">15分钟进程数目：</div>
          <div class="col3">0.45</div>
          <div class="col4">
            <a-tag color="cyan">正常</a-tag>
          </div>
        </div>
      </div>
    </div>
    <div class="line">
      <div class="col1">内存使用</div>
      <div class="colspan">
        <div class="subline">
          <div class="col2">
            总可用：
            <a-tooltip content="todo">
              <icon-question-circle-fill class="color-primary cursor-pointer" />
            </a-tooltip>
          </div>
          <div class="col3">0.00</div>
          <div class="col4"></div>
        </div>
        <div class="subline">
          <div class="col2">剩余可用：</div>
          <div class="col3">0.00</div>
          <div class="col4">
            <a-tag color="gold">34.2%闲置</a-tag>
          </div>
        </div>
        <div class="subline">
          <div class="col2">
            已使用：
            <a-tooltip content="todo">
              <icon-question-circle-fill class="color-primary cursor-pointer" />
            </a-tooltip>
          </div>
          <div class="col3">0.00</div>
          <div class="col4"></div>
        </div>
        <div class="subline">
          <div class="col2">进程占用：</div>
          <div class="col3">0.00</div>
          <div class="col4">
            <a-link class="text-sm">查看内存使用情况</a-link>
          </div>
        </div>
        <div class="subline">
          <div class="col2">
            缓冲区：
            <a-tooltip content="todo">
              <icon-question-circle-fill class="color-primary cursor-pointer" />
            </a-tooltip>
          </div>
          <div class="col3">0.00</div>
          <div class="col4"></div>
        </div>
        <div class="subline">
          <div class="col2">
            缓存区：
            <a-tooltip content="todo">
              <icon-question-circle-fill class="color-primary cursor-pointer" />
            </a-tooltip>
          </div>
          <div class="col3">0.00</div>
          <div class="col4">
            <a-space>
              <a-button type="primary" size="mini">清理缓存</a-button>
              <a-button type="primary" size="mini">自动清理设置</a-button>
            </a-space>
          </div>
        </div>
      </div>
    </div>
    <div class="line no-border">
      <div class="col1">虚拟内存</div>
      <div class="col2">未监测到虚拟内容</div>
      <div class="col3">挂载点</div>
      <div class="col4">
        <a-button type="primary" size="mini">立即创建虚拟内容</a-button>
      </div>
    </div>
    <div class="line">
      <div class="col1">(交换空间)</div>
      <div class="col2">/</div>
      <div class="col3">40G</div>
      <div class="col4"></div>
    </div>
    <div class="line">
      <div class="col1">存储空间</div>
      <div class="colspan mb-6">
        <a-table :columns="mountColumns" :data="mountData" :pagination="false">
          <template #rate="{ record }">
            <a-tag :color="record.rate >= 80 ? 'orangered' : 'green'">
              {{ record.rate }} %
            </a-tag>
          </template>
        </a-table>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
  const mountColumns = [
    {
      title: '挂载点',
      dataIndex: 'mount',
      width: 120,
    },
    {
      title: '总大小',
      dataIndex: 'total',
      width: 120,
    },
    {
      title: '已使用',
      dataIndex: 'used',
      width: 120,
    },
    {
      title: '剩余可用',
      dataIndex: 'free',
      width: 120,
    },
    {
      title: '使用率',
      dataIndex: 'rate',
      slotName: 'rate',
      width: 120,
    },
  ];
  const mountData = [
    {
      mount: '/',
      total: '40G',
      used: '5.2G',
      free: '33G',
      rate: 13.12,
    },
  ];
</script>

<style lang="less" scoped>
  .box {
    border: 1px solid var(--color-border-2);
    padding: 0 16px;
    width: 940px;
    margin: 0 auto;
  }
  .line {
    display: flex;
    justify-content: flex-start;
    align-items: flex-start;
    padding: 12px 40px;
    line-height: 24px;
    border-bottom: 1px solid var(--color-border-2);
    &:last-child {
      border-bottom: none;
    }
  }
  .no-border {
    border-bottom: none;
  }
  .colspan {
    flex: 1;
  }
  .subline {
    width: 100%;
    display: flex;
    justify-content: flex-start;
    align-items: top;
    margin-bottom: 14px;
    &:last-child {
      margin-bottom: 0;
    }
  }
  .col1 {
    width: 160px;
    text-align: right;
    color: var(--color-text-2);
    font-size: 14px;
    margin-right: 40px;
  }
  .col2 {
    width: 160px;
    margin-right: 40px;
    color: var(--color-text-1);
    font-size: 14px;
  }
  .col3 {
    width: 50px;
    margin-right: 30px;
    color: var(--color-text-1);
    font-size: 14px;
  }
  .col4 {
    min-width: 160px;
  }
</style>
