package godog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/strangelove-ventures/interchaintest/v6/ibc"
	"github.com/strangelove-ventures/interchaintest/v6/testutil"
)

func (c *chain) clientShouldBeCreatedOnBothChains() error {
	height, _ := c.source.Height(c.ctx)
	fmt.Println("source height: ", height)
	err := c.r.CreateClients(c.ctx, rep.RelayerExecReporter(c.t), ibcPath, ibc.DefaultClientOpts())
	if err != nil {
		return fmt.Errorf("Create client failed: %v", err)
	}
	testutil.WaitForBlocks(c.ctx, 2, c.source, c.dest)
	return nil
}

func (c *chain) relayCreatesAPath() error {
	err := c.r.GeneratePath(c.ctx, rep.RelayerExecReporter(c.t), c.source.Config().ChainID, c.dest.Config().ChainID, ibcPath)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return fmt.Errorf("Path creation failed: %v", err)
	}
	return nil
}

func (c *chain) clientIDShouldBeReturned() error {
	sourceClientOutput := sourceClientOutputs[0]
	sourceClientID := sourceClientOutput.ClientID
	destClientOutput := destClientOutputs[0]
	destClientID := destClientOutput.ClientID
	fmt.Println(sourceClientID, destClientID)
	return nil
}

func (c *chain) weQueryClient() error {
	sourceClientOutputs, _ = c.r.GetClients(c.ctx, rep.RelayerExecReporter(c.t), c.source.Config().ChainID)
	destClientOutputs, _ = c.r.GetClients(c.ctx, rep.RelayerExecReporter(c.t), c.dest.Config().ChainID)
	return nil
}

func (c *chain) clientStatusShouldBeReturned() error {
	sourceClientOutput := sourceClientOutputs[0]
	sourceClientState := sourceClientOutput.ClientState.ChainID
	destClientOutput := destClientOutputs[0]
	destClientState := destClientOutput.ClientState.ChainID
	fmt.Println(sourceClientState, destClientState)
	return nil
}

func (c *chain) latestHeightFromClientStateShouldBeReturned() error {
	sourceClientOutput := sourceClientOutputs[0]
	sourceLatestHeight := sourceClientOutput.ClientState.ChainID
	destClientOutput := destClientOutputs[0]
	destLatestHeight := destClientOutput.ClientState.ChainID
	fmt.Println(sourceLatestHeight, destLatestHeight)

	// query consensus state
	// chainID := c.source.Config().ChainID
	// cmdOutput, _ = c.r.ExecCmd(c.ctx, rep.RelayerExecReporter(c.t), "node-state", chainID)
	// if cmdOutput.Err != nil {
	// 	return fmt.Errorf("Query consensus state failed: %v", cmdOutput.Err)
	// }
	// clConsesnsus, _ := parseConsensusState(string(cmdOutput.Stdout), string(cmdOutput.Stderr))
	// fmt.Println(clConsesnsus.Timestamp, clConsesnsus.Root, clConsesnsus.NextValidatorsHash)
	return nil
}

func (c *chain) clientShouldNotBeCreated() error {
	return nil
}

func (c *chain) weCreateRelayWithNotExistingPath() error {
	err := c.r.CreateClients(c.ctx, rep.RelayerExecReporter(c.t), "invalid Path name", ibc.DefaultClientOpts())
	if err == nil {
		return fmt.Errorf("Create Client doesn't return an error when an invalid Path name is provided")
	}
	return nil
}

func (c *chain) allNumberOfClientsShouldBeReturned() error {
	fmt.Println("Number of clients on source chain: ", len(sourceClientOutputs))
	fmt.Println("Number of clients on dest chain: ", len(destClientOutputs))
	return nil
}

func (c *chain) clientShouldBeUpdated() error {
	fmt.Println("After ClientUpdate latest height from client state is : ")
	fmt.Println(sourceClientOutputs, destClientOutputs)
	fmt.Println(c.source.Config().GasPrices)
	return nil
}

func (c *chain) weUpdateClient() error {
	fmt.Println("Before ClientUpdate latest height from client state is : ")
	fmt.Println(sourceClientOutputs, destClientOutputs)

	err := c.r.UpdateClients(c.ctx, rep.RelayerExecReporter(c.t), ibcPath)
	if err != nil {
		return fmt.Errorf("Error updating client: %v", err)
	}
	return nil
}

func (c *chain) relayAccountBalanceShouldBeReturned() error {
	res := string(cmdOutput.Stdout)
	for _, output := range strings.Split(res, "}") {
		op := strings.Replace(output, "{", "", -1)
		fmt.Println(op)
	}
	return nil
}

func (c *chain) weQueryUsingRelay() error {
	// chainID := c.source.Config().ChainID
	// cmdOutput, _ = c.r.ExecCmd(c.ctx, rep.RelayerExecReporter(c.t), "balance", chainID)
	// if cmdOutput.Err != nil {
	// 	return fmt.Errorf("Query consensus state failed: %v", cmdOutput.Err)
	// }
	return nil
}

func TestClient(t *testing.T) {
	cotx := context.Background()
	chains := &chain{
		t:   t,
		ctx: cotx,
	}
	suite := godog.TestSuite{
		Name: "TestClient",
		TestSuiteInitializer: func(sc *godog.TestSuiteContext) {
			sc.BeforeSuite(func() {
				chains.relaySetup()
			})
		},
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^client should be created on both chains$`, chains.clientShouldBeCreatedOnBothChains)
			ctx.Step(`^couple of IBC chains running$`, chains.coupleOfIBCChainsRunning)
			ctx.Step(`^relay creates a path$`, chains.relayCreatesAPath)
			ctx.Step(`^Client ID should be returned$`, chains.clientIDShouldBeReturned)
			ctx.Step(`^we query client$`, chains.weQueryClient)
			ctx.Step(`^Client status should be returned$`, chains.clientStatusShouldBeReturned)
			ctx.Step(`^latest height from client state should be returned$`, chains.latestHeightFromClientStateShouldBeReturned)
			ctx.Step(`^client should not be created$`, chains.clientShouldNotBeCreated)
			ctx.Step(`^we create relay with not existing path$`, chains.weCreateRelayWithNotExistingPath)
			ctx.Step(`^all number of clients should be returned$`, chains.allNumberOfClientsShouldBeReturned)
			ctx.Step(`^client should be updated$`, chains.clientShouldBeUpdated)
			ctx.Step(`^we update client$`, chains.weUpdateClient)
			ctx.Step(`^relay account balance should be returned$`, chains.relayAccountBalanceShouldBeReturned)
			ctx.Step(`^we query using relay$`, chains.weQueryUsingRelay)

		},
		Options: &godog.Options{Format: "cucumber", Paths: []string{"features/client.feature"}, TestingT: t, Tags: "smoke"},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
