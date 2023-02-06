package godog

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/strangelove-ventures/ibctest/v6/ibc"
	"github.com/strangelove-ventures/ibctest/v6/testutil"
)

func (c *chain) channelIsCreated() error {
	err := c.r.CreateChannel(c.ctx, rep.RelayerExecReporter(c.t), ibcPath, ibc.DefaultChannelOpts())
	if err != nil {
		return fmt.Errorf("Create connection failed: %v", err)
	}
	testutil.WaitForBlocks(c.ctx, 2, c.source, c.dest)
	return nil
}

func (c *chain) channelShouldBeEstablished() error {
	channelOutput, _ := c.r.GetChannels(c.ctx, rep.RelayerExecReporter(c.t), c.source.Config().ChainID)
	fmt.Println("Channel ID :", channelOutput)
	return nil
}

func TestChannel(t *testing.T) {
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
			ctx.Step(`^channel is created$`, chains.channelIsCreated)
			ctx.Step(`^channel should be established$`, chains.channelShouldBeEstablished)
			ctx.Step(`^client should be created on both chains$`, chains.clientShouldBeCreatedOnBothChains)
			ctx.Step(`^couple of IBC chains running$`, chains.coupleOfIBCChainsRunning)
			ctx.Step(`^relay creates a connection$`, chains.relayCreatesAConnection)
			ctx.Step(`^relay creates a path$`, chains.relayCreatesAPath)
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"features/channel.feature"}, TestingT: t},
	}
	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
