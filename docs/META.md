# Meta-data

Messages come with meta-data that can be accessed via the messages context:

| Name | Description |
|---|---|
| `source` | A URN for the source the message came from |
| `id` | A unique identifier for the messages within the source |

`source+id` is intended to be globally unique.