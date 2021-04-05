# Expressions

About the expression syntax we use.

You can read about the [expression language definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md). We additionally add functions:

* The function `string` is provided to convert the message to a string.
* The function `bytes` is provided to convert the message back into a byte array.
* The function `object` converts JSON as string or byte arrays to an object.
* The function `json` converts an object to a JSON byte array.

