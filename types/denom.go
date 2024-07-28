package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
)

const maxDecBitLen = 315

// denomUnits contains a mapping of denomination mapped to their respective unit
// multipliers (e.g. 1atom = 10^-6uatom).
var denomUnits = map[string]types.Dec{}

// baseDenom is the denom of smallest unit registered
var baseDenom = map[string]string{}

// RegisterDenom registers a denomination with a corresponding unit. If the
// denomination is already registered, an error will be returned.
func RegisterDenom(denom string, unit types.Dec, bDenom string, bUnit types.Dec) error {
	if err := types.ValidateDenom(denom); err != nil {
		return err
	}

	if _, ok := denomUnits[denom]; ok {
		return fmt.Errorf("denom %s already registered", denom)
	}

	denomUnits[denom] = unit
	denomUnits[bDenom] = bUnit
	baseDenom[denom] = bDenom
	baseDenom[bDenom] = bDenom
	return nil
}

// GetDenomUnit returns a unit for a given denomination if it exists. A boolean
// is returned if the denomination is registered.
func GetDenomUnit(denom string) (types.Dec, bool) {
	if err := types.ValidateDenom(denom); err != nil {
		return types.ZeroDec(), false
	}

	unit, ok := denomUnits[denom]
	if !ok {
		return types.ZeroDec(), false
	}

	return unit, true
}

// GetBaseDenom returns the denom of smallest unit registered
func GetBaseDenom(denom string) (string, error) {
	if baseDenom[denom] == "" {
		return "", fmt.Errorf("no denom is registered")
	}
	return baseDenom[denom], nil
}

// ConvertCoin attempts to convert a coin to a given denomination. If the given
// denomination is invalid or if neither denomination is registered, an error
// is returned.
func ConvertCoin(coin types.Coin, denom string) (types.Coin, error) {
	if err := types.ValidateDenom(denom); err != nil {
		return types.Coin{}, err
	}

	srcUnit, ok := GetDenomUnit(coin.Denom)
	if !ok {
		return types.Coin{}, fmt.Errorf("source denom not registered: %s", coin.Denom)
	}

	dstUnit, ok := GetDenomUnit(denom)
	if !ok {
		return types.Coin{}, fmt.Errorf("destination denom not registered: %s", denom)
	}

	if srcUnit.Equal(dstUnit) {
		return types.NewCoin(denom, coin.Amount), nil
	}

	return types.NewCoin(denom, types.NewDecFromInt(coin.Amount).Mul(srcUnit).Quo(dstUnit).TruncateInt()), nil
}

// ConvertDecCoin attempts to convert a decimal coin to a given denomination. If the given
// denomination is invalid or if neither denomination is registered, an error
// is returned.
func ConvertDecCoin(coin types.DecCoin, denom string) (types.DecCoin, error) {
	if err := types.ValidateDenom(denom); err != nil {
		return types.DecCoin{}, err
	}

	srcUnit, ok := GetDenomUnit(coin.Denom)
	if !ok {
		return types.DecCoin{}, fmt.Errorf("source denom not registered: %s", coin.Denom)
	}

	dstUnit, ok := GetDenomUnit(denom)
	if !ok {
		return types.DecCoin{}, fmt.Errorf("destination denom not registered: %s", denom)
	}

	if srcUnit.Equal(dstUnit) {
		return types.NewDecCoinFromDec(denom, coin.Amount), nil
	}

	return types.NewDecCoinFromDec(denom, coin.Amount.Mul(srcUnit).Quo(dstUnit)), nil
}

// NormalizeCoin try to convert a coin to the smallest unit registered,
// returns original one if failed.
func NormalizeCoin(coin types.Coin) types.Coin {
	base, err := GetBaseDenom(coin.Denom)
	if err != nil {
		return coin
	}
	newCoin, err := ConvertCoin(coin, base)
	if err != nil {
		return coin
	}
	return newCoin
}

// NormalizeDecCoin try to convert a decimal coin to the smallest unit registered,
// returns original one if failed.
func NormalizeDecCoin(coin types.DecCoin) types.DecCoin {
	base, err := GetBaseDenom(coin.Denom)
	if err != nil {
		return coin
	}
	newCoin, err := ConvertDecCoin(coin, base)
	if err != nil {
		return coin
	}
	return newCoin
}

// NormalizeCoins normalize and truncate a list of decimal coins
func NormalizeCoins(coins []types.DecCoin) types.Coins {
	if coins == nil {
		return nil
	}
	result := make([]types.Coin, 0, len(coins))

	for _, coin := range coins {
		newCoin, _ := NormalizeDecCoin(coin).TruncateDecimal()
		result = append(result, newCoin)
	}

	return result
}

// ParseCoinNormalized parses and normalize a cli input for one coin type, returning errors if invalid or on an empty string
// as well.
// Expected format: "{amount}{denomination}"
func ParseCoinNormalized(coinStr string) (coin types.Coin, err error) {
	decCoin, err := types.ParseDecCoin(coinStr)
	if err != nil {
		return types.Coin{}, err
	}

	coin, _ = NormalizeDecCoin(decCoin).TruncateDecimal()
	return coin, nil
}
func ParseCoinsNormalized(coinStr string) (types.Coins, error) {
	coins, err := types.ParseDecCoins(coinStr)
	if err != nil {
		return types.Coins{}, err
	}
	return NormalizeCoins(coins), nil
}
