package main

import (
	"context"      // 上下文包，用于控制 goroutine 生命周期
	"crypto/ecdsa" // ECDSA 椭圆曲线数字签名算法
	"fmt"          // 格式化输出包
	"log"          // 日志包
	"math/big"     // 大整数包，用于处理以太坊的 uint256
	"os"           // 操作系统功能包，读取环境变量
	"os/signal"    // 信号处理包，用于监听系统信号
	"syscall"      // 系统调用包，用于定义信号常量
	"time"         // 时间处理包

	"github.com/ethereum/go-ethereum/common"     // 以太坊通用类型（地址、哈希等）
	"github.com/ethereum/go-ethereum/core/types" // 核心类型（区块、交易等）
	"github.com/ethereum/go-ethereum/crypto"     // 加密功能（私钥、签名等）
	"github.com/ethereum/go-ethereum/ethclient"  // 以太坊客户端包
)

func main() {

	alchemyKey := os.Getenv("ALCHEMY_KEY")
	if alchemyKey == "" {
		log.Fatal("未获取到 ALCHEMY_KEY ")
	}

	// 构造 Sepolia 测试网络的 RPC URL
	rpcURL := fmt.Sprintf("https://eth-sepolia.g.alchemy.com/v2/%s", alchemyKey)

	// 创建可取消的上下文，用于控制程序生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 连接到 Sepolia 测试网络
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("连接 Sepolia 网络失败: %v", err)
	}
	defer client.Close()

	// 验证连接成功，获取网络 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		log.Fatalf("获取 Chain ID 失败: %v", err)
	}
	fmt.Printf("已连接到 Sepolia 测试网络 (Chain ID: %s)\n\n", chainID.String())

	fmt.Println("查询区块信息======================start")

	// 获取最新区块号
	latestBlockNum, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}
	fmt.Printf("最新区块号: %d\n", latestBlockNum)

	// 查询指定区块的详细信息
	blockNumber := big.NewInt(int64(latestBlockNum))
	block, err := client.BlockByNumber(ctx, blockNumber)
	if err != nil {
		log.Fatalf("获取区块信息失败: %v", err)
	}

	// 输出区块详细信息
	printBlockInfo(block)
	fmt.Println("查询区块信息======================end")

	fmt.Println("发送交易======================start")

	// 从环境变量读取发送方私钥
	privateKeyHex := os.Getenv("SENDER_PRIVATE_KEY")
	if privateKeyHex == "" {
		fmt.Println("发送方私钥不存在 请退出")

		// 等待用户按 Ctrl+C 退出
		waitForExit()
		return
	}

	// 发送交易
	txHash, err := sendTransaction(ctx, client, privateKeyHex)
	if err != nil {
		log.Printf("发送交易失败: %v", err)
		waitForExit()
		return
	}
	fmt.Printf("交易成功，交易哈希: %s\n", txHash)

	// 等待用户按 Ctrl+C 退出
	waitForExit()
}

// printBlockInfo 打印区块详细信息
func printBlockInfo(block *types.Block) {
	fmt.Println("=========区块信息=========")

	fmt.Printf("区块哈希:     %s\n", block.Hash().Hex())
	fmt.Printf("区块号:       %d\n", block.Number().Uint64())

	timestamp := time.Unix(int64(block.Time()), 0)
	fmt.Printf("时间戳:       %d (%s)\n", block.Time(), timestamp.Format("2006-01-02 15:04:05"))

	fmt.Printf("父区块哈希:   %s\n", block.ParentHash().Hex())
	fmt.Printf("Gas 限制:     %d\n", block.GasLimit())
	fmt.Printf("Gas 使用量:   %d\n", block.GasUsed())
	fmt.Printf("交易数量:     %d\n", len(block.Transactions()))
	fmt.Printf("矿工地址:     %s\n", block.Coinbase().Hex())
	fmt.Printf("难度:         %s\n", block.Difficulty().String())
	fmt.Printf("区块大小:     %d 字节\n", block.Size())

	if block.BaseFee() != nil {
		fmt.Printf("基础费用:     %s Wei\n", block.BaseFee().String())
	}

	fmt.Println("=========区块信息=========")
}

// sendTransaction 发送一笔以太币转账交易
func sendTransaction(ctx context.Context, client *ethclient.Client, privateKeyHex string) (string, error) {
	// 解析私钥
	if len(privateKeyHex) >= 2 && privateKeyHex[0:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("私钥解析失败: %w", err)
	}

	// 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("公钥类型转换失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("发送方地址: %s\n", fromAddress.Hex())

	// 接收方地址
	toAddress := common.HexToAddress("0x70997970C51812dc3A010C7d01b50e0d17dc79C8")
	fmt.Printf("接收方地址: %s\n", toAddress.Hex())

	// 获取 nonce
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("获取 nonce 失败: %w", err)
	}
	fmt.Printf("当前 Nonce: %d\n", nonce)

	// 获取 Gas 价格建议
	gasTipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Gas Tip Cap 失败: %w", err)
	}

	// 获取当前区块的 base fee
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("获取区块头失败: %w", err)
	}

	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(header.BaseFee, big.NewInt(2)),
		gasTipCap,
	)

	// 设置转账金额（0.001 ETH）
	value := big.NewInt(1000000000000000)
	fmt.Printf("转账金额: %s Wei (0.001 ETH)\n", value.String())

	gasLimit := uint64(21000)

	// 获取链 ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return "", fmt.Errorf("获取 Chain ID 失败: %w", err)
	}

	// 构造 EIP-1559 交易
	txData := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      nil,
	}
	tx := types.NewTx(txData)

	// 签名交易
	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 发送交易
	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("发送交易失败: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// 等待用户按ctrl+C退出
func waitForExit() {

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
}
