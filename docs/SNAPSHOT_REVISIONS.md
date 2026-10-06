# Snapshot revision and observed delta design (#8)

Source Package v1 already separates content_hash, dataset_definition_revision,
snapshot_revision and observation/export times. These are inputs to a future
comparison, not claims about why companies changed.

- SnapshotRevision: definition + normalized content; reordering/reacquisition time
  alone does not change it. A revision of record values changes content_hash.
- DatasetDefinitionRevision: query, source/API version, record schema and
  normalization version. Changed scope/version blocks an observed-population delta.
- ObservedDelta: deterministic record additions/removals/corrections only for two
  validated COMPLETE packages with identical definition revisions. It reports
  observed differences, never business founding, bankruptcy or causal conclusions.

Acceptance for the next small implementation:

1. Compare identical content at different observation times: unchanged observation,
   no additions/removals. Reordered rows produce the same result.
2. Compare changed records under the same definition: deterministic differences,
   preserving both package IDs, observation times and content hashes.
3. Changed query/provider/schema/normalization: DEFINITION_CHANGED; do not expose
   an observed population count change. Upstream silent definition changes remain
   UNKNOWN unless authoritative source revision information is available.
4. LIMITED, failed/malformed acquisition or unsupported schema: NOT_COMPARABLE;
   do not turn missing records into removals or an empty successful population.
5. Zero records remains observed zero, optional unavailable fields remain unknown.
6. Define handling of repeated corporate numbers (NTA correction/history records)
   explicitly before implementing per-company pairing; do not overwrite one event.
7. Publish two file-only fixtures + a delta fixture. Validate downstream Public
   Engine handoff separately; keep raw package/legacy CSV/JSONL compatibility.

No delta algorithm is shipped in this change. It needs event pairing and downstream
fixture verification; the boundary above prevents revision categories being mixed.
