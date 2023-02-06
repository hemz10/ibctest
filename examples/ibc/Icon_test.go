package ibc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ibctest "github.com/strangelove-ventures/ibctest/v6"
	"github.com/strangelove-ventures/ibctest/v6/ibc"
	"github.com/strangelove-ventures/ibctest/v6/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestIcon(t *testing.T) {
	ctx := context.Background()

	// Chain Factory
	cf := ibctest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*ibctest.ChainSpec{
		{Name: "icon", ChainConfig: ibc.ChainConfig{
			Type:    "icon",
			Name:    "icon",
			ChainID: "icon-1",
			Images: []ibc.DockerImage{
				{
					Repository: "hemz1012/goloop", // FOR LOCAL IMAGE USE: Docker Image Name
					Version:    "latest",          // FOR LOCAL IMAGE USE: Docker Image Tag
				},
			},
			Bin:            "goloop",
			Bech32Prefix:   "icon",
			Denom:          "icx",
			GasPrices:      "0.00icx",
			GasAdjustment:  1.3,
			TrustingPeriod: "508h",
			NoHostMount:    false},
		},
		// {Name: "osmosis", Version: "v11.0.0"},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	icon := chains[0]
	client, network := ibctest.DockerSetup(t)
	ic := ibctest.NewInterchain().
		AddChain(icon)
	// Log location
	f, err := ibctest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(t)

	// Build interchain
	require.NoError(t, ic.Build(ctx, eRep, ibctest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: ibctest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false},
	),
	)
}
