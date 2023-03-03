package ibc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	iconchain "interchaintest/chain/icon"

	ibctest "github.com/strangelove-ventures/interchaintest/v6"
	"github.com/strangelove-ventures/interchaintest/v6/ibc"
	"github.com/strangelove-ventures/interchaintest/v6/testreporter"
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
			GasPrices:      "0.001icx",
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
	icon.ExportState(ctx, 1)
	time.Sleep(3 * time.Second)
	height, _ := icon.Height(ctx)
	fmt.Println("Latest height: ", height)

	bal, _ := icon.GetBalance(ctx, "hxc088a2e09809ba05b75e06ed247935020a2bc0c5", "icx")
	fmt.Println(bal, " icx")
	fmt.Println(icon.GetGasFeesInNativeDenom(100000000000000))

	iconChain := icon.(*iconchain.IconChain)
	scoreAddress, _ := iconChain.DeployContract(ctx, "/home/dell/practice/ibc-bdd/ibctest/chain/icon/BMC-0.1.0-optimized.jar",
		"/home/dell/practice/ibc-bdd/ibctest/chain/icon/keystore.json",
		"_net=btp://0x1.icon/")
	fmt.Println("Score Address:", scoreAddress)
	output, _ := iconChain.QueryContract(ctx, scoreAddress, "getOwners", "")
	fmt.Println(output)
	time.Sleep(1 * time.Second)
	hash, _ := iconChain.ExecuteContract(ctx, scoreAddress, "/home/dell/practice/ibc-bdd/ibctest/chain/icon/keystore.json", "addOwner", "_addr=hxc088a2e09809ba05b75e06ed247935020a2bc0c5")
	fmt.Println(hash)
	output, _ = iconChain.QueryContract(ctx, scoreAddress, "getOwners", "")
	fmt.Println(output)
	iconChain.WaitForBlocks(ctx, 3)

}
