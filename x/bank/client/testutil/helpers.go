package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	clitest "github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
)

func NewTxSend(f *clitest.Fixtures, from sdk.AccAddress, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	clientCtx := client.Context{}.WithChainID(f.ChainID)
	clientCtx = clientCtx.WithFromAddress(from)
	cmd := cli.NewSendTxCmd(clientCtx)
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	args := fmt.Sprintf("--keyring-backend=test %s %s %s %v", from, to, amount, f.Flags())
	args = clitest.AddFlags(args, flags)
	cmd.SetArgs([]string{args})
	out, err := ioutil.ReadAll(b)
	if err != nil {
		return false, "", err.Error()
	}
	err = cmd.Execute()
	if err != nil {
		return false, "", err.Error()
	}
	return true, string(out), ""
}

// TxSend is simcli tx send
func TxSend(f *clitest.Fixtures, from string, to sdk.AccAddress, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx send --keyring-backend=test %s %s %s %v", f.SimcliBinary, from,
		to, amount, f.Flags())
	return clitest.ExecuteWriteRetStdStreams(f.T, clitest.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryAccount is simcli query account
func QueryAccount(f *clitest.Fixtures, address sdk.AccAddress, flags ...string) authtypes.BaseAccount {
	cmd := fmt.Sprintf("%s query account %s %v", f.SimcliBinary, address, f.Flags())

	out, _ := tests.ExecuteT(f.T, clitest.AddFlags(cmd, flags), "")

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(f.T, err, "out %v, err %v", out, err)
	value := initRes["value"]

	var acc authtypes.BaseAccount
	err = f.Cdc.UnmarshalJSON(value, &acc)
	require.NoError(f.T, err, "value %v, err %v", string(value), err)

	return acc
}

// QueryBalances executes the bank query balances command for a given address and
// flag set.
func QueryBalances(f *clitest.Fixtures, address sdk.AccAddress, flags ...string) sdk.Coins {
	cmd := fmt.Sprintf("%s query bank balances %s %v", f.SimcliBinary, address, f.Flags())
	out, _ := tests.ExecuteT(f.T, clitest.AddFlags(cmd, flags), "")

	var balances sdk.Coins

	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &balances), "out %v\n", out)

	return balances
}

// QueryTotalSupply returns the total supply of coins
func QueryTotalSupply(f *clitest.Fixtures, flags ...string) (totalSupply sdk.Coins) {
	cmd := fmt.Sprintf("%s query bank total %s", f.SimcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	err := f.Cdc.UnmarshalJSON([]byte(res), &totalSupply)
	require.NoError(f.T, err)
	return totalSupply
}

// QueryTotalSupplyOf returns the total supply of a given coin denom
func QueryTotalSupplyOf(f *clitest.Fixtures, denom string, flags ...string) sdk.Int {
	cmd := fmt.Sprintf("%s query bank total %s %s", f.SimcliBinary, denom, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var supplyOf sdk.Int
	err := f.Cdc.UnmarshalJSON([]byte(res), &supplyOf)
	require.NoError(f.T, err)
	return supplyOf
}
