# Expressions

About the expression syntax we use.

You can read about the [expression language definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md). We additionally add functions:

| Function | Description |
|---|---|
| `bytes` | Convert into a byte array |
| `int` | Convert to an int |
| `json` | Converts an object to a JSON byte array |
| `object` | Converts JSON as string or byte arrays to an object |
| `string` | Convert to a string |

# Sprig

Like Argo Workflows, [Sprig functions](http://masterminds.github.io/sprig/) are available under 'sprig'. 

To invoke a Sprig function, change the name into a function call. The last parameter is usually your variable.

Examples:

| Description| Example  |
|---|---|
| Trim whitespace | `sprig.trim(" my-msg  ")` | 
| Replace "a" with "b" | `sprig.replace("a", "b", "my-msg")` | 

Warning! Sprig functions usually do not return errors on invalid parameters.
