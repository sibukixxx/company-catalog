"""Offline conformance check against a pinned public Insight JSON schema.
Install jsonschema first. No engine, DB, credentials, or internal imports.
"""
import json
import pathlib
import subprocess
import jsonschema

root = pathlib.Path(__file__).resolve().parents[1]
schema = json.loads((root / "docs/contracts/insight-public-engine.v1.schema.json").read_text())
request = json.loads(subprocess.check_output([
    "go", "run", "./examples/insight_handoff",
    "docs/examples/company-source-package.v1.json"], cwd=root))
request_schema = {"$schema": schema["$schema"], "$defs": schema["$defs"],
                  "$ref": "#/$defs/AddEvidenceRequest"}
validator = jsonschema.Draft202012Validator(request_schema)
validator.validate(request)
bad = dict(request, contractVersion="2")
assert not validator.is_valid(bad), "must detect unsupported public contract version"
package = json.loads((root / "docs/examples/company-source-package.v1.json").read_text())
assert json.loads(request["documents"][0]["content"]) == package
assert "analyticalArtifacts" not in request
print("Public Engine v1 schema: PASS; complete source package retained; version mismatch rejected")
