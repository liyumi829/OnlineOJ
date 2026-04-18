# 在线OJ

#### 介绍

仿 LeetCode 的网页在线 OJ 项目。

#### 软件架构

* Web 前端：CSS + Javascript
* 后端服务：

  1. MVC服务：Gin框架 + GORM对象映射 + 采用 RPC 调用 + MySQL
  2. 后台判题服务节点：RPC 服务端 + Redis


#### 安装教程

* 在 Linux 环境下（本项目未使用 docker 实现沙盒执行环境，后期迭代实现）

    利用 makefile 对项目进行整体编译和服务，具体详情：`make help`

#### 使用说明

1.  本项目的中间件配置主要从配置文件中读取（`pkg/config` 实现了 yaml 合并配置文件信息）：

    * `OnlineOj/gateway/config` 其中 `config.local.yaml` 存放敏感信息。例如：MySQL 等密码信息等。
    * `OnlineOj/judge/config` 其中 `config.local.yaml` 存放铭感信息


2.  本项目上层管理下层判题节点采用 RPC 调用，实现了**节点管理**：心跳机制、熔断、限流、负载均衡等策略。但是更好地方法安排是：利用 RabbitMQ 来实现任务的发布实现异步效果。但是本项目还没有实现。

    * 在进行正式使用的时候，需要进行测试：上层并发开多少下层的节点接收能力比较合适。控制 gateway 的 `WORKER_NUMBER` 的数量。

        例如：在我本人的单机（4核 4G）测试下更建议 `WORKER_NUMBER` 的数量为：`10` 是最合适的
