# Garbage Collection

The controller will, by default, try to delete any pipelines 30m after they complete. But, by default, the controller does not have permission to do this.

You need to add the permission `delete pipelines` to enable this feature. 

To prevent this for a single pipeline, [add a finalizer](https://kubernetes.io/blog/2021/05/14/using-finalizers-to-control-deletion/). 