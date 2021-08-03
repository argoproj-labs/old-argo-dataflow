const ProcessHandler = require('argo-dataflow-sdk')

const handler = require('./handler')

async function main() {
  ProcessHandler.start(handler)
}

main()