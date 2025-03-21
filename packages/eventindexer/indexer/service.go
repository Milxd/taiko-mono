package indexer

import (
	"time"

	"github.com/cyberhorsey/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/taikoxyz/taiko-mono/packages/eventindexer"
	"github.com/taikoxyz/taiko-mono/packages/eventindexer/contracts/bridge"
	"github.com/taikoxyz/taiko-mono/packages/eventindexer/contracts/proverpool"
	"github.com/taikoxyz/taiko-mono/packages/eventindexer/contracts/swap"
	"github.com/taikoxyz/taiko-mono/packages/eventindexer/contracts/taikol1"
)

var (
	ZeroAddress = common.HexToAddress("0x0000000000000000000000000000000000000000")
)

type Service struct {
	eventRepo      eventindexer.EventRepository
	blockRepo      eventindexer.BlockRepository
	statRepo       eventindexer.StatRepository
	nftBalanceRepo eventindexer.NFTBalanceRepository
	ethClient      *ethclient.Client

	processingBlockHeight uint64

	blockBatchSize      uint64
	subscriptionBackoff time.Duration

	taikol1    *taikol1.TaikoL1
	proverPool *proverpool.ProverPool
	bridge     *bridge.Bridge
	swaps      []*swap.Swap

	indexNfts bool
}

type NewServiceOpts struct {
	EventRepo           eventindexer.EventRepository
	BlockRepo           eventindexer.BlockRepository
	StatRepo            eventindexer.StatRepository
	NFTBalanceRepo      eventindexer.NFTBalanceRepository
	EthClient           *ethclient.Client
	RPCClient           *rpc.Client
	SrcTaikoAddress     common.Address
	ProverPoolAddress   common.Address
	SrcBridgeAddress    common.Address
	SrcSwapAddresses    []common.Address
	BlockBatchSize      uint64
	SubscriptionBackoff time.Duration
	IndexNFTs           bool
}

func NewService(opts NewServiceOpts) (*Service, error) {
	if opts.EventRepo == nil {
		return nil, eventindexer.ErrNoEventRepository
	}

	if opts.IndexNFTs && opts.NFTBalanceRepo == nil {
		return nil, eventindexer.ErrNoNFTBalanceRepository
	}

	if opts.EthClient == nil {
		return nil, eventindexer.ErrNoEthClient
	}

	if opts.RPCClient == nil {
		return nil, eventindexer.ErrNoRPCClient
	}

	var taikoL1 *taikol1.TaikoL1

	var err error

	if opts.SrcTaikoAddress.Hex() != ZeroAddress.Hex() {
		taikoL1, err = taikol1.NewTaikoL1(opts.SrcTaikoAddress, opts.EthClient)
		if err != nil {
			return nil, errors.Wrap(err, "contracts.NewTaikoL1")
		}
	}

	var bridgeContract *bridge.Bridge

	if opts.SrcBridgeAddress.Hex() != ZeroAddress.Hex() {
		bridgeContract, err = bridge.NewBridge(opts.SrcBridgeAddress, opts.EthClient)
		if err != nil {
			return nil, errors.Wrap(err, "contracts.NewBridge")
		}
	}

	var swapContracts []*swap.Swap

	if opts.SrcSwapAddresses != nil && len(opts.SrcSwapAddresses) > 0 {
		for _, v := range opts.SrcSwapAddresses {
			swapContract, err := swap.NewSwap(v, opts.EthClient)
			if err != nil {
				return nil, errors.Wrap(err, "contracts.NewBridge")
			}

			swapContracts = append(swapContracts, swapContract)
		}
	}

	var proverPool *proverpool.ProverPool
	if opts.ProverPoolAddress.Hex() != ZeroAddress.Hex() {
		proverPool, err = proverpool.NewProverPool(opts.ProverPoolAddress, opts.EthClient)
		if err != nil {
			return nil, errors.Wrap(err, "proverpool.NewProverPool")
		}
	}

	return &Service{
		eventRepo:      opts.EventRepo,
		blockRepo:      opts.BlockRepo,
		statRepo:       opts.StatRepo,
		nftBalanceRepo: opts.NFTBalanceRepo,
		ethClient:      opts.EthClient,
		taikol1:        taikoL1,
		bridge:         bridgeContract,
		proverPool:     proverPool,
		swaps:          swapContracts,

		blockBatchSize:      opts.BlockBatchSize,
		subscriptionBackoff: opts.SubscriptionBackoff,

		indexNfts: opts.IndexNFTs,
	}, nil
}
