package godog

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/strangelove-ventures/ibctest/v6/testutil"
)

func (c *chain) relayCreatesAConnection() error {
	err := c.r.CreateConnections(c.ctx, rep.RelayerExecReporter(c.t), ibcPath)
	if err != nil {
		return fmt.Errorf("Create connection failed: %v", err)
	}
	testutil.WaitForBlocks(c.ctx, 2, c.source, c.dest)
	return nil
}

func (c *chain) connectionShouldBeEstablished() error {
	connectionOutputs, _ := c.r.GetConnections(c.ctx, rep.RelayerExecReporter(c.t), c.source.Config().ChainID)
	connectionOutput := connectionOutputs[0]
	fmt.Println("Connection ID: ", connectionOutput.ID)
	return nil
}

func TestConnection(t *testing.T) {
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
			ctx.Step(`^connection should be established$`, chains.connectionShouldBeEstablished)
			ctx.Step(`^couple of IBC chains running$`, chains.coupleOfIBCChainsRunning)
			ctx.Step(`^relay creates a connection$`, chains.relayCreatesAConnection)
			ctx.Step(`^relay creates a path$`, chains.relayCreatesAPath)

		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"features/connection.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
