import copy
import json
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

import update_catalog as updater


class CatalogTests(unittest.TestCase):
    def setUp(self):
        self.regions = {"items": [{"metadata": {"name": "test-region"}, "spec": {"description": "Test", "generation": "v2"}}]}
        self.classes = {"items": [{"metadata": {"name": "active"}, "spec": {"availability": "available", "region": "test-region", "category": "General", "resources": {"cpu": "4", "memory": "16GB"}}, "status": {"capacity": 10}}]}
        self.matrix = {"data": {"version-matrix": json.dumps({"1.9.9": ["1.10.1"], "1.10.1": ["1.10.1"]})}}

    def build(self):
        return updater.build_catalog(self.regions, self.classes, self.matrix, ["cilium", "calico"])

    def test_numeric_order_and_defaults(self):
        catalog = self.build()
        self.assertEqual(catalog["kubernetesVersions"], ["1.10.1", "1.9.9"])
        self.assertEqual(catalog["defaults"]["kubernetesVersion"], "1.10.1")
        self.assertEqual(catalog["defaults"]["region"], "test-region")

    def test_excludes_deprecated_classes_and_volatile_metadata(self):
        retired = copy.deepcopy(self.classes["items"][0])
        retired["metadata"]["name"] = "retired"
        retired["spec"]["availability"] = "deprecated"
        self.classes["items"].append(retired)
        catalog = self.build()
        self.assertEqual([c["name"] for c in catalog["serverClasses"]], ["active"])
        self.classes["items"][0]["status"]["capacity"] = 999
        self.assertEqual(catalog, self.build())

    def test_rejects_empty_invalid_or_inconsistent_responses(self):
        for field in ["regions", "classes"]:
            with self.subTest(field=field):
                with patch.object(self, field, {"items": []}):
                    with self.assertRaises(ValueError):
                        self.build()
        self.classes["items"][0]["spec"]["region"] = "unknown"
        with self.assertRaises(ValueError):
            self.build()
        self.matrix["data"]["version-matrix"] = '{"invalid": []}'
        with self.assertRaises(ValueError):
            self.build()

    def test_extracts_cnis_from_published_openapi_schema(self):
        schema = {"components": {"schemas": {"io.rxt.ngpc.v1.CloudSpace": {"properties": {"spec": {"properties": {"cni": {"enum": ["calico", "cilium", "byocni"]}}}}}}}}
        self.assertEqual(updater.cni_choices("# Public API\n" + json.dumps(schema)), ["byocni", "calico", "cilium"])

    def test_failure_never_overwrites_existing_catalogue(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "catalog.json"
            output.write_text("original\n")
            with patch("sys.argv", ["update_catalog.py", "--output", str(output)]), patch.object(updater, "refresh", side_effect=ValueError("invalid upstream data")):
                self.assertEqual(updater.main(), 1)
            self.assertEqual(output.read_text(), "original\n")

    def test_check_and_idempotent_writes(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory) / "catalog.json"
            with patch.object(updater, "refresh", return_value=self.build()):
                with patch("sys.argv", ["update_catalog.py", "--output", str(output)]):
                    self.assertEqual(updater.main(), 0)
                    before = output.stat().st_mtime_ns
                    self.assertEqual(updater.main(), 0)
                    self.assertEqual(output.stat().st_mtime_ns, before)
                with patch("sys.argv", ["update_catalog.py", "--check", "--output", str(output)]):
                    self.assertEqual(updater.main(), 0)
                    output.write_text("{}\n")
                    self.assertEqual(updater.main(), 1)
                    self.assertEqual(output.read_text(), "{}\n")


if __name__ == "__main__":
    unittest.main()
