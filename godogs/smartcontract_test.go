package godog

import (
	"context"
	"fmt"
	"testing"
	"time"

	"interchaintest/chain/icon"

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
	cfg := LoadConfig(".")
	if cfg.Environment == "local" {
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
	} else {
		it := icon.IconTestnet{}
		it.LoadConfig()
		s, _ := it.GetLastBlock()
		fmt.Println(s)
		t, _ := it.GetBlockByHeight(19585147)
		fmt.Println(t)

		// test deploy contract
		// scoreAddress, _ := it.DeployContract("/home/dell/practice/ibc-bdd/ibctest/chain/icon/BMC-0.1.0-optimized.jar",
		// 	"/home/dell/practice/ibc-bdd/ibctest/chain/icon/keystore.json", "_net=btp://0x1.icon/")
		// fmt.Println("Score Address: ", scoreAddress)

		// Query contract - cxe3c22462c5ec53d2de928b6923700c3ce9473db0
		result, _ := it.QueryContract("cxe3c22462c5ec53d2de928b6923700c3ce9473db0", "getOwners", "")
		fmt.Println(result)
		hash, _ := it.ExecuteContract("cxe3c22462c5ec53d2de928b6923700c3ce9473db0", "/home/dell/practice/ibc-bdd/ibctest/chain/icon/keystore.json", "addOwner", "_addr=hxc088a2e09809ba05b75e06ed247935020a2bc0c5")
		time.Sleep(3 * time.Second)
		tResult, _ := it.GetTransactionResult(hash)
		fmt.Println(tResult)
		result, _ = it.QueryContract("cxe3c22462c5ec53d2de928b6923700c3ce9473db0", "getOwners", "")
		fmt.Println(result)
		hash, _ = it.ExecuteContract("cxe3c22462c5ec53d2de928b6923700c3ce9473db0", "/home/dell/practice/ibc-bdd/ibctest/chain/icon/keystore.json", "removeOwner", "_addr=hxc088a2e09809ba05b75e06ed247935020a2bc0c5")
		time.Sleep(3 * time.Second)
		tResult, _ = it.GetTransactionResult(hash)
		fmt.Println(tResult)
		result, _ = it.QueryContract("cxe3c22462c5ec53d2de928b6923700c3ce9473db0", "getOwners", "")
		fmt.Println(result)
		return nil
	}
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
			ctx.Step(`^Contract should be deployed on chain$`, chains.contractShouldBeDeployedOnOsmosis)
			ctx.Step(`^Chain running$`, chains.osmosisChainRunning)
			ctx.Step(`^we Deploy SmartContract on chain$`, chains.weDeploySmartContractOnOsmosis)
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"features/smartcontract.feature"}, TestingT: t},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
