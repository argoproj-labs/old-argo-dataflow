# Contributing

Install Golang v1.16 and have a Kubernetes cluster ready.

Start:

```
make start
```

To access the user interface, you must use Argo Workflows.

```
cd ../../argoproj/argo-workflows
git checkout dev-dataflow
make ./dist/argo DEV_BRANCH=true
./dist/argo server --secure=false --namespaced --auth-mode=server 
killall node
yarn --cwd ui start
```
