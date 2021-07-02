# Garbage Collection

Pipelines will be, by default, deleted 30m after they complete. To prevent this, [add a finalizer](https://kubernetes.io/blog/2021/05/14/using-finalizers-to-control-deletion/). 