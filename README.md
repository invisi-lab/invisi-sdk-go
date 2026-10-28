# Invisi Go SDK

Go SDK 对齐当前 InvisiProtocol 的 Solidity / Noir 实现，模块为 github.com/invisi-lab/invisi-sdk-go。入金使用 shield；资金凭证花费、私密转账与提现使用 transact；授信使用 createCredit、creditDraw、cancelCredit 和 topUpCredit。

## 开发与验证

需要 Go 1.26.1+。安装模块后可直接使用已生成的合约绑定，无需兄弟仓库。

~~~sh
go test ./...
go test -race ./...
go vet ./...
~~~

重新生成 ABI 与 Go 绑定时，提供已构建的协议仓库和 abigen：

~~~sh
INVISI_PROTOCOL_ROOT=/absolute/path/to/invisi-protocol go generate ./contracts
~~~

加密与交易测试使用协议实现生成的独立向量：

~~~sh
node core/testdata/generate_vectors.mjs /absolute/path/to/invisi-protocol
node internal/testvectors/generate.mjs /absolute/path/to/invisi-protocol
~~~

## 包与职责

| 包 | 内容 |
|---|---|
| core | Poseidon2、凭证承诺、消费标记、随机 nonce / 共享密钥、凭证加解密与授信剩余恢复 |
| contracts | 当前 Protocol、Factory、普通 / 授信验证路由和 ERC20 的生成绑定 |
| payer | Shield、Transact、聚合、授信创建的 calldata、公开输入和授权摘要 |
| receiver | 发票、支付请求、无签名全额提现及授信扣款、取消和追加 |
| tree | 资金凭证树及历史树的同步、滚树发现、Merkle 路径 |
| ipp | IPP 请求与 Voucher、HTTP 客户端、短链解析和中继 |
| deployments | 网络元数据、Factory 实例查询 |

调用方提供真实 chainID、protocolAddress、tokenAddress。网络注册表只提供网络元数据；实例地址从已验证的 Factory 查询或调用方配置获取。构建 calldata、公开输入和签名摘要不会广播交易。

## 资金凭证

普通凭证使用 Note，私密字段为 NoteSecret。ComputeNoteNullifier 同时绑定凭证密钥和 ReceivingAddress；生成身份承诺、重算凭证、消费标记使用同一组原始数据。

GenerateNonce 均匀生成随机 Field。DeriveNoteSecret 是可选的发票密钥派生策略。每张输出独立提供 nonce，给自己的输出采用相同的凭证结构。实际输出树与叶位置从插入事件获取。

BuildShieldCalldata 接收 1–10 个输出原像，并核对承诺、密文和入金总额。BuildTransactCalldata 始终使用含 Sources、Nullifiers、OutputCommitments、WithdrawalAddress、WithdrawalAmount、Deadline 的统一参数。Sources 在纯入金时为空，单树时一项，跨树时多项；来源树编号必须递增。合约 tuple 直接使用生成类型，字段名如 InputSource.TreeId。

普通电路按输入与输出总数选择 14 / 56 / 112 / 168 档。只有 14 档支持 4 / 8 / 12 个成员的聚合。BuildTransactAggregateCalldata 使用统一请求数组，每项含 Params、Ciphertexts、Funding。

有输入的私密输出、部分提现或更换提现地址，需要绑定钱包的所有者授权。BuildUnsignedWithdrawalCalldata 只允许将所有输入余额全额提现至共同绑定地址，不允许私密输出或入金。无输入的私密拆分至少生成两张凭证。

TransactPublicInputs 返回业务公开语句，TransactUserPublicInputs 包含完整用户电路语句及 EIP-712 授权摘要两个 limb。AggregatePublicInputs 返回路由适配器的业务字段，每个成员 56 个 Field。

## 授信

CreditSharedKey 是独立随机非零 Field，通过 GenerateCreditSharedKey 生成。ComputeCreditSharedKeyHash 计算其 Tag 7 哈希。消费方身份、所有者身份和出资方分别独立；知道共享密钥不能替代消费方的凭证密钥或所有者钱包签名。VerifyCreditCommitment 从收到的共享密钥本地计算哈希并核对完整承诺。

授信扣款档位为 1 / 12 / 50 / 120。部分扣款保留消费方身份、所有者身份、有效期及共享密钥。RecoverCreditRemainder 根据原凭证与已认证的 draw 事件恢复剩余金额和 nonce，并核对消费标记与输出承诺。

BuildCancelCreditCalldata 取消整张授信，将余额退回绑定所有者；到期后或协议暂停时仍可取消。BuildTopUpCreditCalldata 只允许在原有效期内追加，增加余额并更换 nonce，新有效期可保持或延长。PreviousExpiry 绑定输入凭证，Creation.Expiry 绑定输出凭证；TopUpCreditPublicInputs 有 48 个公开字段。

创建、取消和追加必须提供所有者钱包授权。公开入金另有出资方授权；同一钱包同时担任所有者和签名出资方时，按顺序消耗相邻 nonce，使用 CreateCreditNonceGuide / TopUpCreditNonceGuide 计算。无入金使用零值 FundingParams，由构建器展开为全零合约 tuple。

## 加密与证明

ViewingPublicKey、EncryptPrivateTransferNote / DecryptPrivateTransferNote、EncryptCreditNote / DecryptCreditNote 与协议的 secp256k1 ECDH、HKDF、AES-GCM 格式互通，普通凭证为 178 字节、授信为 254 字节。解密会验证上下文及承诺。

Go 构建器接收当前电路生成的原始证明。普通 Transact 使用原始用户证明；授信构建器使用 EncodeCreditProof 编码 direct 模式。证明生成与密码学验证由调用方的证明后端完成。公开参数检查不能证明私密所有权或金额守恒；签名、私密见证与证明必须使用同一实例和完整公开语句。

## 树同步与 IPP

树同步仅以 CommitmentsInserted / BatchOutputs 为插入依据，处理滚树、重复日志与数据缺口，保留历史树路径。重组、冲突和缺失历史通过 Error / OnError 报告；失败批次不会部分修改树。没有区块头接口时，轮询通过回放历史核对已保存叶子。

IPP Voucher 使用 NoteSecret，wire 为 note_secret。公开查询不携带凭证密钥。AuthorizeRequest 明确 SpenderIdentityCommitment 和 OwnerIdentityCommitment；授权结果包含 CreditSharedKeyHash。Propose / Authorize 提供承诺材料，输出位置由实际插入事件决定。

Withdraw 只查询指定实例与树的公共 Merkle 数据，路径深度固定 16；服务端金额只供参考，本地按完整凭证计算余额。授信生命周期使用本地 payer / receiver 构建器。

Relay 使用 shield、transact、create-credit、credit-draw、cancel-credit、top-up-credit。calldata 模式必须指定 ProtocolAddress，与构造签名和证明的实例一致。witness / prove 模式只接受当前电路及操作组合，会把私密见证交付给所选证明服务。接入的 IPP 服务须实现 SDK 定义的字段和响应结构。

协议定义和基准结果位于兄弟仓库 protocol-docs。
