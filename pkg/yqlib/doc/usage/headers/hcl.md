# HCL

Encode and decode to and from [HashiCorp Configuration Language (HCL)](https://github.com/hashicorp/hcl).

HCL is commonly used in HashiCorp tools like Terraform for configuration files. The yq HCL encoder and decoder support:
- Blocks and attributes
- String interpolation and expressions (preserved without quotes)
- Comments (leading, head, and line comments)
- Nested structures (maps and lists)
- Syntax colorisation when enabled

A nested map is always rendered one way: a plain map under a key becomes a block (`key { ... }`), never both a block and an attribute assignment for the same key. Scalar string values are always written as quoted, escaped HCL strings (embedded quotes, backslashes, newlines and other control characters are escaped) so they round-trip back to the same string, rather than being emitted as bare, unquoted tokens.

