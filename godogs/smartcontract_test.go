package godog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cucumber/godog"
	interchaintest "github.com/strangelove-ventures/interchaintest/v6"
	"github.com/strangelove-ventures/interchaintest/v6/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v6/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func (c *chain) contractShouldBeDeployedOnOsmosis() error {
	return nil
}

func (c *chain) osmosisChainRunning() error {
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(c.t), chains)
	if cf == nil {
		return fmt.Errorf("chain factory failed")
	}
	chains, _ := cf.Chains(c.t.Name())
	c.dest = chains[1]
	client, network := interchaintest.DockerSetup(c.t)
	ic := interchaintest.NewInterchain().
		AddChain(c.dest)
	// Log location
	f, err := interchaintest.CreateLogFile(fmt.Sprintf("%d.json", time.Now().Unix()))
	require.NoError(c.t, err)
	// Reporter/logs
	rep := testreporter.NewReporter(f)
	eRep := rep.RelayerExecReporter(c.t)

	// Build interchain
	require.NoError(c.t, ic.Build(c.ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          c.t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),

		SkipPathCreation: false},
	),
	)
	return nil
}

func (c *chain) weDeploySmartContractOnOsmosis() error {
	users := interchaintest.GetAndFundTestUsers(c.t, c.ctx, "default", fundAmount, c.dest)
	destUser = users[0]
	balance, _ := c.dest.GetBalance(c.ctx, destUser.FormattedAddress(), c.dest.Config().Denom)
	fmt.Println(balance, c.dest.Config().Denom)
	osmosis := c.dest.(*cosmos.CosmosChain)
	keyName := destUser.KeyName()
	codeId, err := osmosis.StoreContract(c.ctx, keyName, "/home/dell/practice/ibc-bdd/ibctest/godogs/cw_tpl_osmosis.wasm")
	if err != nil {
		return fmt.Errorf("error storing: %v", err)
	}
	fmt.Println(codeId)
	return nil
}

func TestSmartContract(t *testing.T) {
	cotx := context.Background()
	chains := &chain{
		t:   t,
		ctx: cotx,
	}
	suite := godog.TestSuite{
		Name: "TestSmartContract",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^Contract should be deployed on Osmosis$`, chains.contractShouldBeDeployedOnOsmosis)
			ctx.Step(`^Osmosis Chain running$`, chains.osmosisChainRunning)
			ctx.Step(`^we Deploy SmartContract on Osmosis$`, chains.weDeploySmartContractOnOsmosis)
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"features/smartcontract.feature"}, TestingT: t},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
