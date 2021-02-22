package keeper

import (
	"fmt"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/cns/x/cns/types"
)

type (
	Keeper struct {
		cdc      codec.Marshaler
		storeKey sdk.StoreKey
		memKey   sdk.StoreKey

		bankKeeper  types.BankKeeper
		distrKeeper types.DistributionKeeper
	}
)

func NewKeeper(cdc codec.Marshaler, storeKey, memKey sdk.StoreKey) *Keeper {
	return &Keeper{
		cdc:      cdc,
		storeKey: storeKey,
		memKey:   memKey,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) Register(ctx sdk.Context, cInfo types.ChainInfo) error {
	if err := cInfo.ValidateBasic(); err != nil {
		return err
	}

	err := k.SetChainInfo(ctx, cInfo)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) Update(ctx sdk.Context, cInfo types.ChainInfo) error {
	storeInfo, err := k.GetChainInfo(ctx, cInfo.ChainName, cInfo.Owner)
	if err != nil {
		return err
	}

	err = cInfo.ValidateBasic()
	if err != nil {
		return err
	}

	storeInfo.Update(cInfo)

	return k.SetChainInfo(ctx, storeInfo)
}

func (k Keeper) GetChainInfo(ctx sdk.Context, cName, owner string) (types.ChainInfo, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetStoreKey(cName, owner))
	if bz == nil {
		return types.ChainInfo{}, fmt.Errorf("chain info not found for name %s", cName)
	}
	var cInfo types.ChainInfo
	err := k.cdc.UnmarshalBinaryBare(bz, &cInfo)
	if err != nil {
		return types.ChainInfo{}, err
	}

	return cInfo, err
}

func (k Keeper) SetChainInfo(ctx sdk.Context, info types.ChainInfo) error {
	store := ctx.KVStore(k.storeKey)

	bytes, err := k.cdc.MarshalBinaryBare(&info)
	if err != nil {
		return err
	}

	store.Set(types.GetStoreKey(info.ChainName, info.Owner), bytes)
	return nil
}

func (k Keeper) GetChainInfoFromName(ctx sdk.Context, name string) types.ChainInfo {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, []byte(name))
	defer iterator.Close()

	var info types.ChainInfo
	for ; iterator.Valid(); iterator.Next() {
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &info)
	}

	return info
}

func (k Keeper) IterateAllInfos(ctx sdk.Context, cb func(info types.ChainInfo) bool) {
	store := ctx.KVStore(k.storeKey)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var info types.ChainInfo
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &info)

		if cb(info) {
			break
		}
	}
}
