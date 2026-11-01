#!/usr/bin/env python3
"""Refresh non-secret Rackspace option metadata. Requires SPOT_REFRESH_TOKEN."""
import argparse
import json
import os
from pathlib import Path
import re
import sys
import tempfile
import urllib.error
import urllib.parse
import urllib.request

BASE_URL = "https://spot.rackspace.com"
SCHEMA_URL = BASE_URL + "/docs/rackspace-spot-public-api"
OUTPUT = Path(__file__).resolve().parents[1] / "driver/catalog.json"
VERSION = re.compile(r"^\d+\.\d+\.\d+$")


def fetch_json(url, *, token=None, data=None):
    headers = {"Accept": "application/json"}
    if token:
        headers["Authorization"] = "Bearer " + token
    if data is not None:
        headers["Content-Type"] = "application/x-www-form-urlencoded"
    request = urllib.request.Request(url, data=data, headers=headers)
    with urllib.request.urlopen(request, timeout=60) as response:
        return json.load(response)


def cni_choices(document):
    # DeveloperHub exposes the OpenAPI document as Markdown; its multiline JSON
    # descriptions can contain literal newlines, hence strict=False.
    start = document.find("{")
    if start < 0:
        raise ValueError("Rackspace API documentation contains no OpenAPI schema")
    schema, _ = json.JSONDecoder(strict=False).raw_decode(document[start:])
    choices = schema["components"]["schemas"]["io.rxt.ngpc.v1.CloudSpace"]["properties"]["spec"]["properties"]["cni"]["enum"]
    if not choices or not all(isinstance(value, str) and value for value in choices):
        raise ValueError("Rackspace API schema contains no valid CNI choices")
    return sorted(set(choices))


def build_catalog(regions, classes, matrix, cnis):
    versions = json.loads(matrix["data"]["version-matrix"])
    if not versions or not all(VERSION.fullmatch(v) for v in versions):
        raise ValueError("Invalid or empty Kubernetes version matrix")
    for targets in versions.values():
        if not isinstance(targets, list) or not targets or not all(isinstance(v, str) and VERSION.fullmatch(v) for v in targets):
            raise ValueError("Invalid Kubernetes version matrix targets")
    versions = sorted(set(versions) | {v for targets in versions.values() for v in targets}, key=lambda v: tuple(map(int, v.split('.'))), reverse=True)
    region_options = sorted([
        {"name": r["metadata"]["name"], "description": r["spec"]["description"], "generation": r["spec"]["generation"]}
        for r in regions["items"]
    ], key=lambda r: r["name"])
    names = {r["name"] for r in region_options}
    if not names or len(names) != len(region_options):
        raise ValueError("Empty or duplicate region list")
    server_classes = []
    for resource in classes["items"]:
        spec = resource["spec"]
        if spec.get("availability") != "available":
            continue
        if spec["region"] not in names:
            raise ValueError("Server class refers to an unknown region")
        server_class = {
            "name": resource["metadata"]["name"],
            "region": spec["region"],
            "category": spec["category"],
            "cpu": str(spec["resources"]["cpu"]),
            "memory": spec["resources"]["memory"],
        }
        for source, target in [("disk", "disk"), ("gpu", "gpu")]:
            if source in spec["resources"]:
                server_class[target] = str(spec["resources"][source])
        if spec.get("minBidPricePerHour"):
            server_class["minBidPricePerHour"] = spec["minBidPricePerHour"]
        server_classes.append(server_class)
    server_classes.sort(key=lambda c: (c["region"], c["name"]))
    if not server_classes or len({c["name"] for c in server_classes}) != len(server_classes):
        raise ValueError("Empty or duplicate active server-class list")
    if {c["region"] for c in server_classes} != names:
        raise ValueError("Region list includes regions with no active server classes")
    if not cnis or not all(isinstance(cni, str) and cni for cni in cnis):
        raise ValueError("Empty or invalid CNI list")
    default_region = "us-east-iad-1" if "us-east-iad-1" in names else region_options[0]["name"]
    regional = [c for c in server_classes if c["region"] == default_region]
    if not regional:
        raise ValueError("Default region has no active server classes")
    default_class = next((c["name"] for c in regional if c["name"] == "gp.vs1.medium-iad"), regional[0]["name"])
    return {
        "schemaVersion": 1,
        "sources": {"versions": BASE_URL + "/api/v1/namespaces/default/configmaps/cloudspace-version-matrix", "regions": BASE_URL + "/apis/ngpc.rxt.io/v1/regions", "serverClasses": BASE_URL + "/apis/ngpc.rxt.io/v1/serverclasses", "cnis": SCHEMA_URL},
        "defaults": {"kubernetesVersion": versions[0], "region": default_region, "serverClass": default_class, "cni": "calico" if "calico" in cnis else sorted(cnis)[0]},
        "kubernetesVersions": versions,
        "cnis": sorted(set(cnis)),
        "regions": region_options,
        "serverClasses": server_classes,
    }


def refresh():
    refresh_token = os.environ.get("SPOT_REFRESH_TOKEN")
    if not refresh_token:
        raise ValueError("Set SPOT_REFRESH_TOKEN (GitHub Actions: add the repository secret SPOT_REFRESH_TOKEN)")
    body = urllib.parse.urlencode({"grant_type": "refresh_token", "client_id": "mwG3lUMV8KyeMqHe4fJ5Bb3nM1vBvRNa", "refresh_token": refresh_token}).encode()
    token = fetch_json("https://login.spot.rackspace.com/oauth/token", data=body)["id_token"]
    regions = fetch_json(BASE_URL + "/apis/ngpc.rxt.io/v1/regions", token=token)
    classes = fetch_json(BASE_URL + "/apis/ngpc.rxt.io/v1/serverclasses", token=token)
    matrix = fetch_json(BASE_URL + "/api/v1/namespaces/default/configmaps/cloudspace-version-matrix", token=token)
    request = urllib.request.Request(SCHEMA_URL, headers={"Accept": "text/markdown"})
    with urllib.request.urlopen(request, timeout=60) as response:
        cnis = cni_choices(response.read().decode())
    return build_catalog(regions, classes, matrix, cnis)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="Fail if the catalogue differs, without writing it")
    parser.add_argument("--output", type=Path, default=OUTPUT)
    args = parser.parse_args()
    try:
        catalog = refresh()
        content = json.dumps(catalog, indent=2) + "\n"
        if args.output.exists() and args.output.read_text() == content:
            print("Rackspace catalogue is unchanged")
            return 0
        if args.check:
            print("Rackspace catalogue needs updating")
            return 1
        # Replace only after all endpoints have succeeded and validated. Avoid
        # timestamps, prices and capacity counters that would create daily noise.
        with tempfile.NamedTemporaryFile(mode="w", dir=args.output.parent, delete=False) as file:
            temporary = Path(file.name)
            file.write(content)
        try:
            temporary.replace(args.output)
        finally:
            temporary.unlink(missing_ok=True)
        print(f"Updated catalogue: {len(catalog['kubernetesVersions'])} Kubernetes versions, {len(catalog['regions'])} regions, {len(catalog['serverClasses'])} server classes")
        return 0
    except (KeyError, TypeError, ValueError, OSError, urllib.error.URLError) as error:
        # Never include request headers, tokens or response bodies in logs.
        print(f"Catalogue refresh failed: {type(error).__name__}: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
