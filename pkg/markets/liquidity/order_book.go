package liquidity

import (
	"github.com/stateshape/augur-analyzer/pkg/proto/markets"
)

type outcomeOrderBook struct {
	Bids []*markets.LiquidityAtPrice
	Asks []*markets.LiquidityAtPrice
}

func NewOutcomeOrderBook(bids []*markets.LiquidityAtPrice, asks []*markets.LiquidityAtPrice) OutcomeOrderBook {
	bidsCopy := make([]*markets.LiquidityAtPrice, len(bids))
	asksCopy := make([]*markets.LiquidityAtPrice, len(asks))
	for i, bid := range bids {
		bidsCopy[i] = &markets.LiquidityAtPrice{
			Price:  bid.Price,
			Amount: bid.Amount,
		}
	}
	for i, ask := range asks {
		asksCopy[i] = &markets.LiquidityAtPrice{
			Price:  ask.Price,
			Amount: ask.Amount,
		}
	}

	return &outcomeOrderBook{
		Bids: bidsCopy,
		Asks: asksCopy,
	}
}

func (oob *outcomeOrderBook) DeepClone() OutcomeOrderBook {
	bidsCopy := make([]*markets.LiquidityAtPrice, len(oob.Bids))
	asksCopy := make([]*markets.LiquidityAtPrice, len(oob.Asks))
	for i, bid := range oob.Bids {
		bidsCopy[i] = &markets.LiquidityAtPrice{
			Price:  bid.Price,
			Amount: bid.Amount,
		}
	}
	for i, ask := range oob.Asks {
		asksCopy[i] = &markets.LiquidityAtPrice{
			Price:  ask.Price,
			Amount: ask.Amount,
		}
	}
	return &outcomeOrderBook{
		Bids: bidsCopy,
		Asks: asksCopy,
	}
}

func (oob *outcomeOrderBook) CloseLongFillOnly(shares float64, market MarketData, dryRun bool) float64 {
	bids, proceeds := oob.TakeBest(oob.Bids, shares, market, dryRun, false)
	if !dryRun {
		oob.Bids = bids
	}
	return proceeds
}

func (oob *outcomeOrderBook) CloseShortFillOnly(shares float64, market MarketData, dryRun bool) float64 {
	asks, proceeds := oob.TakeBest(oob.Asks, shares, market, dryRun, true)
	if !dryRun {
		oob.Asks = asks
	}
	return proceeds
}

// NormalizeComplementPrice solves two problems: 1. when taking an Ask (closing a long), the price in order book isn't the actual proceeds for taker; the proceeds are MaxPrice-price; 2. when taking a Bid for a scalar, the price in order book isn't the actual proceeds for taker; it's price-MinPrice.
func (oob *outcomeOrderBook) NormalizeComplementPrice(price float32, market MarketData, closingShort bool) float64 {
	// Complement
	if closingShort {
		return market.MaxPrice - float64(price)
	}
	return float64(price) - market.MinPrice
}

func (oob *outcomeOrderBook) TakeBest(liquidity []*markets.LiquidityAtPrice, shares float64, market MarketData, dryRun bool, closingShort bool) ([]*markets.LiquidityAtPrice, float64) {
	proceeds := 0.0

	for shares > 0 {
		if len(liquidity) < 1 {
			return liquidity, proceeds
		}

		if float64(liquidity[0].Amount) > shares {
			price := oob.NormalizeComplementPrice(liquidity[0].Price, market, closingShort)
			proceeds += shares * price
			if !dryRun {
				liquidity[0].Amount -= float32(shares)
			}
			shares -= shares
			return liquidity, proceeds
		}

		price := oob.NormalizeComplementPrice(liquidity[0].Price, market, closingShort)
		proceeds += float64(liquidity[0].Amount) * price
		shares -= float64(liquidity[0].Amount)
		if !dryRun {
			liquidity[0].Amount = 0
		}
		liquidity = liquidity[1:]
	}
	return liquidity, proceeds
}
