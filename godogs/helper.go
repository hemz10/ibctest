package godog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/client"
	"github.com/strangelove-ventures/ibctest/v6"
	"github.com/strangelove-ventures/ibctest/v6/ibc"
	"github.com/strangelove-ventures/ibctest/v6/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var eRep *testreporter.RelayerExecReporter

const ibcPath = "gaia-osmo-demo"

type chain struct {
	source, dest ibc.Chain
	t            *testing.T
	r            ibc.Relayer
	ic           *ibctest.Interchain
	network      string
	client       *client.Client
	ctx          context.Context
}

var (
	chains = []*ibctest.ChainSpec{
		// Source chain
		{Name: "gaia", Version: "v7.0.0", ChainConfig: ibc.ChainConfig{
			GasPrices: "0.0uatom",
		}},
		// Destination chain
		{Name: "osmosis", Version: "v11.0.0"},
	}

	// Amount to fund user wallet
	fundAmount = int64(100_000_000)

	// Amount to transfer from source chain to destination chain
	amountToSend = int64(1_000_000)

	// Other variables for test cases
	sourceUser, destUser                   ibc.Wallet
	sourceChannelID, destChannelID, codeId string
	sourceUserBalInitial                   int64
	transactionStatus                      bool
	rep                                    *testreporter.Reporter
	sourceClientOutputs, destClientOutputs ibc.ClientOutputs
	cmdOutput                              ibc.RelayerExecResult
)

func (c *chain) createChain() error {
	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(c.t), chains)
	if cf == nil {
		return fmt.Errorf("chain factory failed")
	}
	chains, _ := cf.Chains(c.t.Name())
	c.source, c.dest = chains[0], chains[1]
	c.client, c.network = ibctest.DockerSetup(c.t)
	c.r = ibctest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(c.t)).Build(
		c.t, c.client, c.network)
	c.ic = ibctest.NewInterchain().
		AddChain(c.source).
		AddChain(c.dest).
		AddRelayer(c.r, "relayer").
		AddLink(ibctest.InterchainLink{
			Chain1:  c.source,
			Chain2:  c.dest,
			Relayer: c.r,
			Path:    ibcPath,
		})

	// Log location
	f, err := ibctest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(c.t, err)
	// Reporter/logs
	rep = testreporter.NewReporter(f)
	eRep = rep.RelayerExecReporter(c.t)

	// Build interchain
	require.NoError(c.t, c.ic.Build(c.ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          c.t.Name(),
		Client:            c.client,
		NetworkID:         c.network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false},
	),
	)
	return nil
}

func (c *chain) relaySetup() error {
	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(c.t), chains)
	if cf == nil {
		return fmt.Errorf("chain factory failed")
	}
	chains, _ := cf.Chains(c.t.Name())
	c.source, c.dest = chains[0], chains[1]
	c.client, c.network = ibctest.DockerSetup(c.t)
	c.r = ibctest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(c.t)).Build(
		c.t, c.client, c.network)
	c.ic = ibctest.NewInterchain().
		AddChain(c.source).
		AddChain(c.dest).
		AddRelayer(c.r, "relayer").
		AddLink(ibctest.InterchainLink{
			Chain1:  c.source,
			Chain2:  c.dest,
			Relayer: c.r,
			Path:    ibcPath,
		})

	// Log location
	f, err := ibctest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(c.t, err)
	// Reporter/logs
	rep = testreporter.NewReporter(f)
	eRep = rep.RelayerExecReporter(c.t)

	// Build interchain
	require.NoError(c.t, c.ic.Build(c.ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          c.t.Name(),
		Client:            c.client,
		NetworkID:         c.network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: true},
	),
	)
	return nil
}

func parseConsensusState(stdout, stderr string) (ibc.CLientConsensusState, error) {
	var consensusState ibc.CLientConsensusState
	for _, consensus := range strings.Split(stdout, "\n") {
		if strings.TrimSpace(consensus) == "" {
			continue
		}
		if err := json.Unmarshal([]byte(consensus), &consensusState); err != nil {
			fmt.Errorf("unmarshal consensus error %s: %v", consensus, err)
			continue
		}
	}

	return consensusState, nil
}
