package cmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/stretchr/testify/suite"

	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	crisistestutil "github.com/cosmos/cosmos-sdk/x/crisis/client/testutil"
)

func TestIntegrationTestSuites(t *testing.T) {
	t.Parallel()

	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	suite.Run(t, banktestutil.NewIntegrationTestSuite(cfg))
	suite.Run(t, crisistestutil.NewIntegrationTestSuite(cfg))
}
