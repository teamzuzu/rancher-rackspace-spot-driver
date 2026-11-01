export const MINIMUM_BID = 0.01;

export function meetsMinimumBid(value) {
  const decimal = String(value ?? '').trim();
  return /^\d+(\.\d+)?$/.test(decimal) && Number.isFinite(Number(decimal)) && Number(decimal) >= MINIMUM_BID;
}

export function validateSpotBids(config, additionalPools) {
  if (!config.disablePrimarySpotPool && !meetsMinimumBid(config.spotBidPrice)) {
    return 'Primary spot pool bid must be at least $0.01 per hour.';
  }
  for (const [index, pool] of additionalPools.entries()) {
    if (!meetsMinimumBid(pool.bidPrice)) return `Spot pool ${index + 2} bid must be at least $0.01 per hour.`;
  }
  return '';
}
