export function withCurrentOptions(options, current = []) {
  const result = [...options];
  for (const value of [current].flat().filter(Boolean)) {
    if (!result.some(option => option.value === value)) result.push({ label: `${value} (current)`, value });
  }
  return result;
}

export function regionOptions(catalog, current) {
  return withCurrentOptions(catalog.regions.map(region => ({
    label: `${region.name} — ${region.description}`,
    value: region.name,
  })), current);
}

export function versionOptions(catalog, current) {
  return withCurrentOptions(catalog.kubernetesVersions.map(version => ({ label: version, value: version })), current);
}

export function cniOptions(catalog, current) {
  return withCurrentOptions(catalog.cnis.map(cni => ({ label: cni, value: cni })), current);
}

export function serverClassOptions(catalog, region, current = []) {
  return withCurrentOptions(catalog.serverClasses.filter(serverClass => serverClass.region === region).map(serverClass => ({
    label: `${serverClass.name} — ${serverClass.category}, ${serverClass.cpu} CPU, ${serverClass.memory}${serverClass.disk ? `, ${serverClass.disk} disk` : ''}`,
    value: serverClass.name,
  })), current);
}

export function defaultServerClass(catalog, region) {
  const classes = catalog.serverClasses.filter(serverClass => serverClass.region === region);
  return classes.find(serverClass => serverClass.name === catalog.defaults.serverClass)?.name
    || classes.find(serverClass => /^gp\.vs\d+\.medium-/.test(serverClass.name))?.name
    || classes[0]?.name
    || '';
}
