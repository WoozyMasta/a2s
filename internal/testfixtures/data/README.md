# Protocol fixture corpus

The `.hex` files contain deterministic protocol bytes encoded as hexadecimal.
Whitespace is ignored by the fixture loader.

These fixtures are synthetic offline samples
derived from the existing parser tests and benchmark payloads.
They are not claimed to be packet captures from live servers
and must not be used as evidence for undocumented wire behavior.

The raw bytes are kept outside protocol packages
so decoder tests and future encoder round-trip tests can use the same inputs.
