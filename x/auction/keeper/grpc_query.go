package keeper

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	proto "github.com/gogo/protobuf/proto"

	"github.com/kava-labs/kava/x/auction/types"
)

var _ types.QueryServer = Keeper{}

// Params implements the gRPC service handler for querying x/auction parameters.
func (k Keeper) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.GetParams(sdkCtx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// Auction implements the Query/Auction gRPC method
func (k Keeper) Auction(c context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	auctionID, err := strconv.ParseUint(req.AuctionId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid auction ID %s", req.AuctionId))
	}

	auction, found := k.GetAuction(ctx, int64(auctionID))
	if !found {
		return &types.QueryAuctionResponse{}, nil
	}

	msg, ok := auction.(proto.Message)
	if !ok {
		return nil, status.Errorf(codes.Internal, "can't protomarshal %T", msg)
	}

	auctionAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAuctionResponse{
		Auction: auctionAny,
	}, nil
}

// Auctions implements the Query/Auctions gRPC method
func (k Keeper) Auctions(c context.Context, req *types.QueryAuctionsRequest) (*types.QueryAuctionsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	// k.GetAllEvidence(ctx)

	var auctions []*codectypes.Any
	auctionStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.AuctionKeyPrefix)

	pageRes, err := query.Paginate(auctionStore, req.Pagination, func(key []byte, value []byte) error {
		result, err := k.UnmarshalAuction(value)
		if err != nil {
			return err
		}

		msg, ok := result.(proto.Message)
		if !ok {
			return status.Errorf(codes.Internal, "can't protomarshal %T", msg)
		}

		auctionAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return err
		}
		auctions = append(auctions, auctionAny)
		return nil
	})
	if err != nil {
		return &types.QueryAuctionsResponse{}, err
	}

	return &types.QueryAuctionsResponse{
		Auction:    auctions,
		Pagination: pageRes,
	}, nil
}
