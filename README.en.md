# Online OJ
## Introduction
A web-based online OJ project imitating LeetCode.

## Software Architecture
- Web Frontend: CSS + JavaScript
- Backend Services:
  1. MVC Service: Gin Framework + GORM ORM + RPC Invocation + MySQL
  2. Background Judge Service Node: RPC Server + Redis

## Installation Guide
- Linux environment required (Docker sandbox execution environment is not implemented in this project and will be added in later iterations).

The entire project is compiled and deployed via Makefile. See details: `make help`

## Instructions for Use
1. Middleware configurations of this project are mainly read from configuration files (yaml configuration merging is implemented in `pkg/config`):
   - `OnlineOj/gateway/config`: Sensitive information such as MySQL passwords are stored in `config.local.yaml`.
   - `OnlineOj/judge/config`: Sensitive information are stored in `config.local.yaml`.

2. The upper-layer service invokes lower-layer judge nodes through RPC, implementing **node management**: heartbeat mechanism, circuit breaking, rate limiting, load balancing and other strategies.

However, a better approach is to use RabbitMQ for task publishing to implement asynchronous processing, which has not been completed in this project yet.

Before official use, tests are required to find the appropriate number of concurrent upper-layer requests matching the receiving capacity of lower-layer nodes. Adjust the `WORKER_NUMBER` configuration of the gateway.

Example: On a standalone 4-core 4GB server, the recommended optimal value for `WORKER_NUMBER` is **10**.