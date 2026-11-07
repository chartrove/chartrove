# Reported units in lab trends

Lab history groups observations by code, which does not establish equal units.
The numeric chart plots the supplied numbers without conversion. Synthetic
same-code quantities with `mg/dL` and `mmol/L` previously shared that axis.

The visualization selector now requires identical supplied Quantity `system`,
`code` and `unit` metadata across the series before choosing a shared chart.
It compares literal metadata, without normalization or inferred equivalence.
Differing metadata, or a quantity mixed with a non-quantity, uses the existing
table with each original value and unit. Identical quantities, a single quantity
and the existing uniformly unitless integer series retain their chart behavior.

This is a conservative display guard, not a unit conversion or clinical
comparability engine. It does not validate reference-bound units or change the
legacy string-measurement parser. Equal metadata alone does not certify clinical
comparability. Raw records, query/source authorization, ordering, persistence,
provenance and source identity remain unchanged.

Seven synthetic real-component browser regressions cover four mismatches and
three chart controls. The original test-only source produced four semantic
failures before the selector correction; that failure log is retained. No real
records, provider operations, recovery, deployment or release are involved.
