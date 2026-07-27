# dot Examples

Practical examples demonstrating dot usage.

## Available Examples

### Simple Package

Minimal example with a single configuration file, demonstrating package name
mapping and `dot-` prefix translation.

See: [basic/simple-package/](basic/simple-package/)

## Running an Example

```bash
cd examples/basic/simple-package
dot --dry-run manage vim   # preview
dot manage vim             # apply
```

## Contributing Examples

To add an example:

1. Create a directory under `basic/`
2. Add a README.md explaining the example
3. Include all necessary files
4. Document the expected output, verified against current behavior
5. Add the example to the list above
6. Submit a pull request

See [CONTRIBUTING.md](../CONTRIBUTING.md) for details.

## Navigation

**[Back to Main README](../README.md)**
