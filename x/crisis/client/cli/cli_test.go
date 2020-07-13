package cli

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"

	"github.com/cosmos/cosmos-sdk/x/crisis/client/testutil"

	"github.com/stretchr/testify/suite"
)

func TestIntegrationTestSuite(t *testing.T) {
	cfg := network.DefaultConfig()
	cfg.NumValidators = 1

	suite.Run(t, testutil.NewIntegrationTestSuite(cfg))
}
