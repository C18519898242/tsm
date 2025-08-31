# TSM 网络升级方案文档

本文档旨在记录将 TSM (Threshold Signature Module) 应用从本地模拟网络升级为真实网络实现的背景、目标和详细计划。

## 1. 现有场景分析

目前，`tsm` 应用中的多方计算（MPC）协议，包括密钥生成（Key Generation）和签名（Signing），都是在一个单进程内通过 Go channels 模拟的。

- __实现方式__:

  - 所有的参与方（Parties）都作为独立的 Goroutines 在同一个程序实例中运行。
  - 它们之间的通信完全依赖于 Go 的 `chan`（通道）进行消息传递。
  - 在 `keygen.go` 和 `signing.go` 文件中，都有一个“消息路由”（Message Routing）的 Goroutine，它扮演了网络交换机的角色，负责从一个全局的 `outCh` channel 中接收所有参与方发出的消息，并根据消息头中的目标地址，将消息分发给对应的接收方。

- __优点__:

  - __简化开发__: 无需处理真实网络的复杂性，如连接管理、序列化/反序列化、网络错误等。
  - __快速测试__: 可以在本地快速、可靠地运行完整的 MPC 流程，非常适合算法逻辑的开发和调试。

- __局限性__:

  - __非生产环境__: 这种模拟方式无法在真实的分布式环境（多台机器）中部署。
  - __缺少真实世界考量__: 无法测试和处理真实网络中可能遇到的问题，例如网络延迟、丢包、节点掉线等。

## 2. 升级目标

我们的目标是替换掉现有的模拟网络层，实现一个基于真实网络协议（TCP）的通信层。同时，为了便于开发和调试，我们希望能够在一台机器上通过本地回环地址（`127.0.0.1`）和不同的端口（如 `9001`, `9002`, `9003`）来运行多个节点，模拟真实的分布式网络拓扑。

## 3. 实施方案

我们将采用模块化和分层的方式进行重构，确保代码的清晰、可扩展和可维护性。

### 步骤 1: 定义抽象网络接口

为了将业务逻辑与网络实现解耦，我们将首先定义一个通用的网络传输接口 `Transport`。

- __位置__: `tsm/internal/network/transport.go`

- __接口定义__:

  ```go
  package network

  import "github.com/bnb-chain/tss-lib/v2/tss"

  type Transport interface {
      Send(data []byte, to *tss.PartyID) error
      Receive() <-chan tss.Message
      Close() error
  }
  ```

### 步骤 2: 实现基于 TCP 的网络传输

我们将创建一个 `TCPTransport` 结构体，实现上述的 `Transport` 接口。

- __位置__: `tsm/internal/network/tcp_transport.go`

- __核心功能__:

  - __监听__: 每个节点启动时，在指定的 IP 和端口上启动一个 TCP 监听服务。
  - __连接管理__: 维护一个到其他对等节点（Peers）的连接池，以提高消息发送效率。
  - __发送__: `Send` 方法会查找或建立到目标节点的 TCP 连接，并将序列化后的消息（使用 `tss.Message` 的 `WireBytes()` 方法）发送出去。
  - __接收__: `Receive` 方法会启动一个 Goroutine，持续接受新的 TCP 连接。当接收到数据后，会进行反序列化，并将合法的 `tss.Message` 放入一个 channel 中返回。

### 步骤 3: 重构主程序 (`main.go`)

`tsm-cli` 工具需要被重构，从一个执行完整流程的脚本，变为一个可以启动单个 TSS 节点的守护进程。

- __启动方式__: 通过命令行参数来配置节点。

  ```bash
  # 启动第一个节点
  go run ./cmd/tsm-cli -id 0 -port 9001 -peers "127.0.0.1:9001,127.0.0.1:9002,127.0.0.1:9003"

  # 在另一个终端启动第二个节点
  go run ./cmd/tsm-cli -id 1 -port 9002 -peers "127.0.0.1:9001,127.0.0.1:9002,127.0.0.1:9003"
  ```

- __职责__:

  - 解析命令行参数。
  - 初始化并启动 `TCPTransport`。
  - 提供一个简单的交互界面（或 API），用于触发密钥生成和签名操作。

### 步骤 4: 整合网络层到业务逻辑

最后，我们将新的网络层整合到 `keygen` 和 `signing` 的流程中。

- __修改 `Generate...` 和 `Sign...` 函数__:

  - 这些函数将不再创建消息路由的 Goroutine。
  - 它们会接收一个 `Transport` 实例作为参数。
  - 启动一个新的 Goroutine，循环从 `transport.Receive()` channel 中读取消息，并调用 `party.Update()` 方法。
  - 当 `tss.Party` 通过 `outCh` channel 产生消息时，另一个 Goroutine 会读取该消息，并调用 `transport.Send()` 将其发送到网络中。

---

明天我们可以按照这个文档的步骤开始实施。晚安！
