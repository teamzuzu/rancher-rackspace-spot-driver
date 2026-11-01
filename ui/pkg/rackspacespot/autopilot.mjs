export function normalizeAutopilotPools(pools) {
  return pools.map(pool => ({
    ...pool,
    vcpus: Number(pool.vcpus),
    memoryPerVCPU: String(pool.memoryPerVCPU),
    budgetPerHour: String(pool.budgetPerHour),
    minVCPUsPerNode: Number(pool.minVCPUsPerNode || 0),
    maxVCPUsPerNode: Number(pool.maxVCPUsPerNode || 0),
  }));
}

export function validateAutopilotPools(pools) {
  for (const pool of pools) {
    if (!Number.isSafeInteger(pool.vcpus) || pool.vcpus <= 0) return 'Autopilot total vCPUs must be a positive integer.';
    if (!['any', '2', '4', '8'].includes(pool.memoryPerVCPU)) return 'Autopilot memory per vCPU must be any, 2, 4, or 8 GB.';
    if (!/^\d+(\.\d+)?$/.test(pool.budgetPerHour) || Number(pool.budgetPerHour) <= 0) return 'Autopilot hourly budget must be a positive decimal.';
    const min = pool.minVCPUsPerNode;
    const max = pool.maxVCPUsPerNode;
    if (!Number.isSafeInteger(min) || !Number.isSafeInteger(max) || min < 0 || max < 0 || (max > 0 && min > max)) return 'Autopilot per-node vCPU bounds must be non-negative integers, with min no greater than max.';
  }
  return '';
}
