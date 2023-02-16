package godog

import (
	"context"
	"fmt"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/cucumber/godog"
	interchaintest "github.com/strangelove-ventures/interchaintest/v6"
	"github.com/strangelove-ventures/interchaintest/v6/ibc"
	"github.com/stretchr/testify/require"
)

func (c *chain) coupleOfIBCChainsRunning() error {
	if c.source == nil {
		return fmt.Errorf("source issue")
	} else if c.dest == nil {
		return fmt.Errorf("osmosis issue")
	} else {
		fmt.Println(c.source, c.dest)
		return nil
	}
}

func (c *chain) userWalletIsFunded() error {
	users := interchaintest.GetAndFundTestUsers(c.t, c.ctx, "default", fundAmount, c.source, c.dest)
	sourceUser = users[0]
	destUser = users[1]
	sourceUserBalInitial, _ = c.source.GetBalance(c.ctx, sourceUser.FormattedAddress(), c.source.Config().Denom)
	// Get Channel ID
	sourceChannelInfo, err := c.r.GetChannels(c.ctx, eRep, c.source.Config().ChainID)
	if err != nil {
		return fmt.Errorf("source channel issue: %v", err)
	}
	sourceChannelID = sourceChannelInfo[0].ChannelID

	osmoChannelInfo, err := c.r.GetChannels(c.ctx, eRep, c.dest.Config().ChainID)
	if err != nil {
		return fmt.Errorf("Destination channel issue: %v", err)
	}
	destChannelID = osmoChannelInfo[0].ChannelID
	return nil
}

func (c *chain) weSendIBCTransferAndRelayPackets() error {
	dstAddress := destUser.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   c.source.Config().Denom,
		Amount:  amountToSend,
	}
	tx, err := c.source.SendIBCTransfer(c.ctx, sourceChannelID, sourceUser.KeyName(), transfer, ibc.TransferOptions{})
	if err != nil {
		return fmt.Errorf("IBC transfer error: %v", err)
	}
	resp := tx.Validate()
	if resp != nil {
		return fmt.Errorf("IBC transfer is not well performed: %v", resp)
	}
	// relay packets and acknoledgments
	packErr := c.r.FlushPackets(c.ctx, eRep, ibcPath, destChannelID)
	if packErr != nil {
		return fmt.Errorf("flushpacket error: %v", packErr)
	}
	ackErr := c.r.FlushAcknowledgements(c.ctx, eRep, ibcPath, sourceChannelID)
	if packErr != nil {
		return fmt.Errorf("ackpacket error: %v", ackErr)
	}

	return nil
}

func (c *chain) fundsShouldBeTransferredAndAmountShouldBeDebittedFromAccount() error {
	// test source wallet has decreased funds
	expectedBal := sourceUserBalInitial - amountToSend
	sourceUserBalNew, err := c.source.GetBalance(c.ctx, sourceUser.FormattedAddress(), c.source.Config().Denom)
	require.NoError(c.t, err)
	require.Equal(c.t, expectedBal, sourceUserBalNew)

	// Trace IBC Denom
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", sourceChannelID, c.source.Config().Denom))
	dstIbcDenom := srcDenomTrace.IBCDenom()

	// Test destination wallet has increased funds
	destUserBalNew, err := c.dest.GetBalance(c.ctx, destUser.FormattedAddress(), dstIbcDenom)
	require.NoError(c.t, err)
	require.Equal(c.t, amountToSend, destUserBalNew)

	return nil
}

func transactionShouldFail() error {
	if transactionStatus == false {
		fmt.Println("transaction failed and Test case passed")
		return nil
	} else {
		return fmt.Errorf("trasaction is successfull so test case failed")
	}
}

func (c *chain) weSendAmountGreaterThanBalance() error {
	transactionStatus = true
	fmt.Println(sourceUserBalInitial)
	amount := sourceUserBalInitial + sourceUserBalInitial // failing this case by sending correct amount
	fmt.Println(amount)
	dstAddress := destUser.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   c.source.Config().Denom,
		Amount:  amount,
	}
	tx, _ := c.source.SendIBCTransfer(c.ctx, sourceChannelID, sourceUser.KeyName(), transfer, ibc.TransferOptions{})
	result := tx.Validate()
	c.r.FlushPackets(c.ctx, eRep, ibcPath, destChannelID)
	if result != nil {
		fmt.Println(result)
		transactionStatus = false
	}

	return nil
}

func (c *chain) weSendNegativeAmountForTranser() error {
	transactionStatus = true
	fmt.Println(sourceUserBalInitial)
	amount := int64(-1_000_000)
	fmt.Println(amount)
	dstAddress := destUser.FormattedAddress()
	transfer := ibc.WalletAmount{
		Address: dstAddress,
		Denom:   c.source.Config().Denom,
		Amount:  amount,
	}
	tx, _ := c.source.SendIBCTransfer(c.ctx, sourceChannelID, sourceUser.KeyName(), transfer, ibc.TransferOptions{})
	result := tx.Validate()
	c.r.FlushPackets(c.ctx, eRep, ibcPath, destChannelID)
	if result != nil {
		fmt.Println(result)
		transactionStatus = false
	}

	return nil
}

func TestFeatures(t *testing.T) {
	// InitializeScenario(t)
	cotx := context.Background()
	chains := &chain{
		t:   t,
		ctx: cotx,
	}
	suite := godog.TestSuite{
		Name: "TestFeatures",
		TestSuiteInitializer: func(sc *godog.TestSuiteContext) {
			sc.BeforeSuite(func() {
				chains.createChain()
			})
		},
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^couple of IBC chains running$`, chains.coupleOfIBCChainsRunning)
			ctx.Step(`^user wallet is funded$`, chains.userWalletIsFunded)
			ctx.Step(`^we send IBC transfer and relay packets$`, chains.weSendIBCTransferAndRelayPackets)
			ctx.Step(`^funds should be transferred and amount should be debitted from account$`, chains.fundsShouldBeTransferredAndAmountShouldBeDebittedFromAccount)
			ctx.Step(`^transaction should fail$`, transactionShouldFail)
			ctx.Step(`^we send amount greater than balance$`, chains.weSendAmountGreaterThanBalance)
			ctx.Step(`^we send negative amount for transer$`, chains.weSendNegativeAmountForTranser)
		},
		Options: &godog.Options{Format: "pretty", Paths: []string{"features/e2e.feature"}, TestingT: t, Tags: "smoke"},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}
