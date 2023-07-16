# jp.nvim

jp.nvim is an interface to JMESPath, an expression language for manipulating JSON.

## Requirements

- Neovim 0.9.1 or higher
- Go 1.20 or higher

## Installation

Using packer.nvim,

```lua
use {
  'kit494way/jp.nvim',
  run = 'go build',
}
```

## Usage

Open a JSON file and execute `:JP query`.

For example, open a JSON file with the following contents,

```json
{
  "foo": 1,
  "bar": [2, 3]
}
```

and execute the command `:JP foo`.
The following result will be displayed in a scratch buffer.

```json
1
```

For more information on JMESPath, please see [here](https://jmespath.site/main/).
