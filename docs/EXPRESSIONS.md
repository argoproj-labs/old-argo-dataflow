# Expressions

About the expression syntax we use.

You can read about the [expression language definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md). We additionally add functions:

* The function `int` is provided to convert to an int.
* The function `string` is provided to convert to a string.
* The function `bytes` is provided to convert into a byte array.
* The function `object` converts JSON as string or byte arrays to an object.
* The function `json` converts an object to a JSON byte array.

# Sprig

Like Argo Workflows, [Sprig functions](http://masterminds.github.io/sprig/) are available under 'sprig'. 

To invoke a Sprig function, change the name into a function call. The last parameter is usually your variable.

Examples:

| Description| Expression  |
|---|---|
| Trim whitespace | `sprig.trim(" my-msg  ")` | 
| Replace "a" with "b" | `sprig.replace("a", "b", "my-msg")` | 

Warning! Sprig functions usually do not return errors on invalid parameters.
