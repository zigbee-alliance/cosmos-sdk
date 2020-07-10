# Context

## Supply is one blob of Coins

## SetBalance, etc. should not exist

## Vesting is handled with accounts

## Permissions are very basic, denom's are permissioned

# Decision

## Bank Keeper

```go
type ViewKeeper interface {
	ValidateBalance(ctx sdk.Context, addr sdk.AccAddress) error
	HasBalance(ctx sdk.Context, addr sdk.AccAddress, amt sdk.Coin) bool

	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins

	IterateAccountBalances(ctx sdk.Context, addr sdk.AccAddress, cb func(coin sdk.Coin) (stop bool))
	IterateAllBalances(ctx sdk.Context, cb func(address sdk.AccAddress, coin sdk.Coin) (stop bool))

    // these get moved here from SendKeeper because they are read-only
	GetParams(ctx sdk.Context) types.Params
	SendEnabledCoin(ctx sdk.Context, coin sdk.Coin) bool
	SendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
	BlockedAddr(addr sdk.AccAddress) bool
}

type SendKeeper interface {
	ViewKeeper

	MultiSendCoins(ctx sdk.Context, inputs []types.Input, outputs []types.Output) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type Keeper interface {
    SendKeeper

	SetParams(ctx sdk.Context, params types.Params)

    GetMinter(denomPrefixes []string) Minter
    GetBurner(denomPrefixes []string) Burner
    GetMinterBurner(denomPrefixes []string) MinterBurner
    GetStakingPool(denomPrefixes []string) Staker
}

type Minter interface {
    SendKeeper

	MintCoins(ctx sdk.Context, acct sdk.AccAddress, amt sdk.Coins) error
}

type Burner interface {
    SendKeeper

	BurnCoins(ctx sdk.Context, acct sdk.AccAddress, amt sdk.Coins) error
}

type MinterBurner interface {
	MintCoins(ctx sdk.Context, acct sdk.AccAddress, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, acct sdk.AccAddress, amt sdk.Coins) error
}

type StakingPool interface {
    Address() sdk.AccAddress
	DelegateCoins(ctx sdk.Context, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoins(ctx sdk.Context, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
}
```

Remove
* `SubtractCoins`
* `AddCoins`
* `SetBalance(s)`
* `GetSupply`
* `SetSupply`

```go
// simapp/app.go
atomMinter := bankKeeper.GetMinter("atom")
ibcMinter := bankKeeper.GetMinterBurner("ibc:")
creditMinter := bankKeeper.GetMinterBurner("credit:")

```