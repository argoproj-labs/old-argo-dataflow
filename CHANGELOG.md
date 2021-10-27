# Changelog

## v0.4.0 (2021-10-25)

 * [5ec67f6](https://github.com/argoproj-labs/argo-dataflow/commit/5ec67f66010481a91e8743b9ef841c51de813fe2) feat: Make Kafka group.id configurable. Fixes #89 (#464)
 * [5572212](https://github.com/argoproj-labs/argo-dataflow/commit/557221254f9cd8c74032dd3302955b339685ece4) fix: demote Kafka error messages to info. Fixes #466 (#469)

### Contributors

 * Alex Collins

## v0.3.0 (2021-10-21)

 * [9014b30](https://github.com/argoproj-labs/argo-dataflow/commit/9014b30b889daf684c145ab22a542459d5845d31) feat: Enable DLQ (#444)
 * [42034f8](https://github.com/argoproj-labs/argo-dataflow/commit/42034f8c63425299c203cea7865f5fa22472b1f3) fix: CRD metadata.annotations size (#457)

### Contributors

 * Saravanan Balasubramanian

## v0.2.0 (2021-10-15)

 * [f81b33d](https://github.com/argoproj-labs/argo-dataflow/commit/f81b33dfc44683525d9c31ca0f94cbd07fd45a08) feat: do not use finalizer to stop metrics cache loop (#450)

### Contributors

 * Derek Wang

## v0.1.0 (2021-10-08)


### Contributors


## v0.0.128 (2021-10-08)

 * [b5114d6](https://github.com/argoproj-labs/argo-dataflow/commit/b5114d61f2e5632d72590d69d7d601ce5e38d114) fix: close jetstream connection (#441)
 * [049ec21](https://github.com/argoproj-labs/argo-dataflow/commit/049ec21fd159f586babd0c7595a648212cbd5924) feat: fix ID for cron and HTTP sources
 * [3c45487](https://github.com/argoproj-labs/argo-dataflow/commit/3c4548723760d7ab34d835e16d1dccf96057a533) feat: tune CPU resources
 * [d5259ea](https://github.com/argoproj-labs/argo-dataflow/commit/d5259ea3017e683140313af65b510ca285d9f760) fix: log `giveUp=false` at `level=info`

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.127 (2021-10-07)

 * [31e721d](https://github.com/argoproj-labs/argo-dataflow/commit/31e721d392b0220bccf5a3e42598e006e8672211) feat: add `enable.idempotence=true` for Kafka sink

### Contributors

 * Alex Collins

## v0.0.126 (2021-10-06)

 * [0b2455f](https://github.com/argoproj-labs/argo-dataflow/commit/0b2455f7bcea2c3e08b5960d6280946e2e53b169) fix: try to fix memory leak in Kafka sink

### Contributors

 * Alex Collins

## v0.0.125 (2021-10-06)

 * [6423c9e](https://github.com/argoproj-labs/argo-dataflow/commit/6423c9e339897b929bcb6b2d23239847c2bd73b3) fix: sensible linger Kafka defaults (#432)
 * [11251f5](https://github.com/argoproj-labs/argo-dataflow/commit/11251f5806ed8bfbd9e6b5017259c19b006118ca) feat: jetstream source and sink (#400)
 * [6b36b07](https://github.com/argoproj-labs/argo-dataflow/commit/6b36b07f65c06b6d53c1fd044792d6cf6cd769fe) feat: log Kafka structured. Fixes #431
 * [695e117](https://github.com/argoproj-labs/argo-dataflow/commit/695e117b8945fdd18c93621ebf401b2eedfb13b9) feat: enable Kafka debug with ARGO_DATAFLOW_DEBUG=kafka.generic
 * [7d601a4](https://github.com/argoproj-labs/argo-dataflow/commit/7d601a45e1a78f1b7d0e855457563422a0649618) feat!: remove unused Redis/monitor code (#429)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.124 (2021-10-05)

 * [29b4deb](https://github.com/argoproj-labs/argo-dataflow/commit/29b4debf4bc38592b97dbb5e8128bee1bfdbd1ed) feat: export Prom metrics for logs. Fixes #418
 * [64f8f10](https://github.com/argoproj-labs/argo-dataflow/commit/64f8f105b3da231639cbc147fd76b31b7e52b14c) feat: log metrics labels on pod term
 * [f318ebf](https://github.com/argoproj-labs/argo-dataflow/commit/f318ebf4ddc4ede41fa22b0660427430defef5ca) feat: send error logs to stderr
 * [fa64809](https://github.com/argoproj-labs/argo-dataflow/commit/fa648092a97097f1dc91c21fbdc53d1475770e15) fix: correctly log Kafka async error
 * [fa7373b](https://github.com/argoproj-labs/argo-dataflow/commit/fa7373b329f5dc17addb71c78da5694b466d483b) feat: expose Kafka config (#414)

### Contributors

 * Alex Collins

## v0.0.123 (2021-10-05)

 * [5bbe50a](https://github.com/argoproj-labs/argo-dataflow/commit/5bbe50a4c2e42c8bf1ae0d6093db6b95cdeac8ea) feat: cammit Kafka offset async (#411)
 * [e056b27](https://github.com/argoproj-labs/argo-dataflow/commit/e056b272003ac74fd94e8a17c82dcaa438b52cec) fix: do not re-balance channels (#410)

### Contributors

 * Alex Collins

## v0.0.122 (2021-10-04)

 * [4d2ad98](https://github.com/argoproj-labs/argo-dataflow/commit/4d2ad98f1bbbebb7ed8502aea8c30f1b0dfd545c) fix: make Kafka consumers channels size 1

### Contributors

 * Alex Collins

## v0.0.121 (2021-10-04)

 * [a46256f](https://github.com/argoproj-labs/argo-dataflow/commit/a46256f7eb4734168d06788244edd8adbb1092a3) fix: remove lock, and fix loop. Fixes #404 (#407)

### Contributors

 * Alex Collins

## v0.0.120 (2021-10-04)

 * [0b05359](https://github.com/argoproj-labs/argo-dataflow/commit/0b05359a0325cfe8b5745aec21b92da7e752cd5a) fix: lock assign/revoke partition code to prevent multiple consumers starting for same partition. Fixes #404 (#405)
 * [d515a7e](https://github.com/argoproj-labs/argo-dataflow/commit/d515a7e43749e5a2eea5525860d90777b8afca49) feat: Kafka thoughput (#402)

### Contributors

 * Alex Collins

## v0.0.119 (2021-10-04)

 * [7376f50](https://github.com/argoproj-labs/argo-dataflow/commit/7376f5063b1f3d37103536d304cefc3ef1afa12e) fix: fix Kafka failure mode (#395)
 * [9c05fcc](https://github.com/argoproj-labs/argo-dataflow/commit/9c05fccc8ce824f344c6b7dad1a37c4084c0c8bc) fix: fix Kafka source (#394)
 * [5a1d295](https://github.com/argoproj-labs/argo-dataflow/commit/5a1d29564d2d4645072d2b95a732969c7f4245f9) feat: tune Kafka producer and consumer (4x TPS)

### Contributors

 * Alex Collins

## v0.0.118 (2021-10-01)

 * [647d0a6](https://github.com/argoproj-labs/argo-dataflow/commit/647d0a6cba13cfa21063661c33be12f18c50c8f5) fix: change Kafka to use stats for pending
 * [dc03e6c](https://github.com/argoproj-labs/argo-dataflow/commit/dc03e6ce3aef0e395346afd2bb64932179abd7e8) fix: change Kafka to use stats for pending
 * [63ea3fa](https://github.com/argoproj-labs/argo-dataflow/commit/63ea3fa7800f3f05070d4306d7900e575020ecfa) feat: change update interval from 1m to 15s
 * [f1e7b43](https://github.com/argoproj-labs/argo-dataflow/commit/f1e7b43a28925ef241fe3041155fd13dcb8f8204) fix: race in monitor
 * [cb439be](https://github.com/argoproj-labs/argo-dataflow/commit/cb439be606b7adf681ed4e91b2bd21168489260a) fix: logging metrics

### Contributors

 * Alex Collins

## v0.0.117 (2021-10-01)

 * [2720622](https://github.com/argoproj-labs/argo-dataflow/commit/27206226fae64ec28e34e826ff0842eb732a98aa) feat: fix bugs in monitor (#391)
 * [b82ef2f](https://github.com/argoproj-labs/argo-dataflow/commit/b82ef2ff0736aed62149c53cbaf72fb17fa4d63e) feat: migrate to confluent-kafka-go from Sarama (#387)
 * [a070247](https://github.com/argoproj-labs/argo-dataflow/commit/a07024722aa6bb5c844b34c7dc4715ee86173b0e) feat: log metrics on runner stop
 * [2ac7baf](https://github.com/argoproj-labs/argo-dataflow/commit/2ac7baf94bf9dfac0202c0948cc599637e9a7a12) fix: enhance missing/duplicate monitor
 * [67f5024](https://github.com/argoproj-labs/argo-dataflow/commit/67f502408d0496917423e81474cc1159d58b4235) fix: more start-up logging detail (#386)
 * [f738044](https://github.com/argoproj-labs/argo-dataflow/commit/f7380441a79db00d5969f5f81c7fa7c7db8a9c38) fix: correct report of `missing`  when multiple messages go missing
 * [e55b950](https://github.com/argoproj-labs/argo-dataflow/commit/e55b950f05bfc869f5ecf0c247af5a898634620b) fix: updated retry/error messages with clearer detail

### Contributors

 * Alex Collins

## v0.0.116 (2021-09-28)

 * [ebf38e6](https://github.com/argoproj-labs/argo-dataflow/commit/ebf38e61fc0597516ddbda1132156b2149d7e69d) fix: correct mis-reporting of `missing`

### Contributors

 * Alex Collins

## v0.0.115 (2021-09-23)

 * [1b91c71](https://github.com/argoproj-labs/argo-dataflow/commit/1b91c71e798e676addc29b23eb7c68c925be33b1) feat: ignored duplicate Kafka messages (#373)

### Contributors

 * Alex Collins

## v0.0.114 (2021-09-22)

 * [9e7332a](https://github.com/argoproj-labs/argo-dataflow/commit/9e7332a55c6d016299b1bbbb0b88497cf96bcf99) feat: Auto-detect missing/duplicate messages (#369)
 * [504a63d](https://github.com/argoproj-labs/argo-dataflow/commit/504a63dcbcd72d4218c339400565c9256249275c) fix: delete Unix domain socket

### Contributors

 * Alex Collins

## v0.0.113 (2021-09-21)

 * [57a1634](https://github.com/argoproj-labs/argo-dataflow/commit/57a1634b42b8712f25c0e6a66d9037222cdcc06a) fix: not upating status even replicas changed (#366)
 * [50742ca](https://github.com/argoproj-labs/argo-dataflow/commit/50742ca82409c84e75b05d0fdc89fd25f68b24ae) feat: update Python DSL to support multiple sources and sinks (#365)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.112 (2021-09-20)

 * [40eac37](https://github.com/argoproj-labs/argo-dataflow/commit/40eac37cbcbc3a9e92c34b45b8146714a9ecd11c) fix: Kafka commit session end of consume claim

### Contributors

 * Alex Collins

## v0.0.111 (2021-09-20)

 * [a618b38](https://github.com/argoproj-labs/argo-dataflow/commit/a618b38ab6d628cc87c80f365ce3f1cd7a717aab) fix: Partial revert of 76e1d95 (#360)
 * [2176e72](https://github.com/argoproj-labs/argo-dataflow/commit/2176e72210a724f2f52c90324c4c0f2c553b2606) fix: more reasonable desired replicas calculation (#359)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.110 (2021-09-20)

 * [6cd0639](https://github.com/argoproj-labs/argo-dataflow/commit/6cd06390f2bf0c82b123aa7e982a0d31f0958641) fix: requeue interval change (#357)
 * [124d6cb](https://github.com/argoproj-labs/argo-dataflow/commit/124d6cb13f722994ad3c0466c3ddac2d11224446) fix: Auto create HTTP auth secret. Fixes #319 (#353)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.109 (2021-09-20)

 * [32f3e63](https://github.com/argoproj-labs/argo-dataflow/commit/32f3e631a501c5906aefe93305b6fdf264940aab) fix: Correct typo that meant ARGO_DATAFLOW_UNIX_DOMAIN_SOCKET was ignored.

### Contributors

 * Alex Collins

## v0.0.108 (2021-09-19)

 * [c09d10a](https://github.com/argoproj-labs/argo-dataflow/commit/c09d10a65545e2d639f3eb79907ab6735b8c88b0) fix: correct Unix Domain Socket env var name
 * [ad85dff](https://github.com/argoproj-labs/argo-dataflow/commit/ad85dff00f69d8025c82c9bc3b1a5ca74bf19b51) fix: update NATS dep to fix Snyk failure (#349)

### Contributors

 * Alex Collins

## v0.0.107 (2021-09-17)

 * [657062a](https://github.com/argoproj-labs/argo-dataflow/commit/657062ae24e5b578c8c1846becd7523756050b35) fix: short-term fix for pending metric not being updated (#346)

### Contributors

 * Alex Collins

## v0.0.106 (2021-09-17)

 * [06bdf8d](https://github.com/argoproj-labs/argo-dataflow/commit/06bdf8d5ba00e8ed02ff060ea0f70372f8c8f29c) fix: run only one Kafka offset committer per step (#343)
 * [0733e32](https://github.com/argoproj-labs/argo-dataflow/commit/0733e3224d31a6f294a8f88381383e5dbfea9f1a) feat: update `dedupe` to use `NewCounter` rather than `NewCounterFunc` (#342)
 * [0764aa1](https://github.com/argoproj-labs/argo-dataflow/commit/0764aa134a60365f0cb4c13d3b16736631c8e117) feat: scale based on metrics (#337)
 * [489ded6](https://github.com/argoproj-labs/argo-dataflow/commit/489ded6fcca010368594c44f192d9963b2079d1b) feat: Add dedupe to Python DSL (#341)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.105 (2021-09-17)

 * [cc7706d](https://github.com/argoproj-labs/argo-dataflow/commit/cc7706d99ca28933f51bd8ba44b9efad12f4e6f3) fix: fix logging an signal handling (#339)
 * [ae8ad3f](https://github.com/argoproj-labs/argo-dataflow/commit/ae8ad3ff68ce33ed0077034aad403c77832c42a8) feat: make meta just plain string
 * [d2bf012](https://github.com/argoproj-labs/argo-dataflow/commit/d2bf01250d9638077d320a6644b1d0c43eecf120) feat: Run Kafka offset commit loop every 1s, 2x faster, more robust.
 * [13d5b50](https://github.com/argoproj-labs/argo-dataflow/commit/13d5b500280edbc5e8712cce76a013ef062f6010) fix: golang1-16 runtimes
 * [84bf124](https://github.com/argoproj-labs/argo-dataflow/commit/84bf1240fdf9096531785606ab9255caa3386207) feat: add ARGO_DATAFLOW_UNIX_DOMAIN_SOCKET config (#335)

### Contributors

 * Alex Collins

## v0.0.104 (2021-09-16)

 * [f50e8a7](https://github.com/argoproj-labs/argo-dataflow/commit/f50e8a730b63f529e9cb833195ce0bf9109c0bba) feat: add ARGO_DATAFLOW_UNIX_DOMAIN_SOCKET config
 * [4cb372e](https://github.com/argoproj-labs/argo-dataflow/commit/4cb372e893ffcffd022c3e3d06f2879ff7efb5ab) feat!: Remove `step..metrics` (#334)
 * [884ae0d](https://github.com/argoproj-labs/argo-dataflow/commit/884ae0d4e6fd098deae5d4b35da09d436714547d) Revert "fix: Remove Golang SDK dependencies"

### Contributors

 * Alex Collins

## v0.0.103 (2021-09-16)

 * [0fe072e](https://github.com/argoproj-labs/argo-dataflow/commit/0fe072e335ffe3a5d805b3f9b33cbd301c5fda9e) fix: Remove Golang SDK dependencies
 * [76e1d95](https://github.com/argoproj-labs/argo-dataflow/commit/76e1d95354f60938514679fb625e665ed4c7da80) fix!: Remove Kafka auto-commit. Fixes #327

### Contributors

 * Alex Collins

## v0.0.102 (2021-09-16)

 * [69e608a](https://github.com/argoproj-labs/argo-dataflow/commit/69e608a53a4f326eda0c00063d20950b441879b0) feat!: Remove `SunkMessages` condition. Add `sinks_total` metric. (#328)
 * [4168dd0](https://github.com/argoproj-labs/argo-dataflow/commit/4168dd0eb796ffffcc70095779bf1928aa94dec6) feat!: Remove `metrics.rate` field.
 * [2192cd5](https://github.com/argoproj-labs/argo-dataflow/commit/2192cd55233377fdeb66cef9cf23d62c0e497b13) feat!: Remove `metrics.rate` field.
 * [b2dd247](https://github.com/argoproj-labs/argo-dataflow/commit/b2dd247170cb2997028e9b32b0b244f1c160a809) feat!: Remove `metrics.rate` field.
 * [ae75900](https://github.com/argoproj-labs/argo-dataflow/commit/ae75900d38370c3284195458b7e11dacc8e00718) feat!: Remove `metrics.rate` field.
 * [91a9d39](https://github.com/argoproj-labs/argo-dataflow/commit/91a9d3964831f1f1d596a7d5827a0ac7ed18c937) feat!: Remove `metrics.rate` field.
 * [7c80fea](https://github.com/argoproj-labs/argo-dataflow/commit/7c80feafef64dddb1e5f798c1dddbb0586f0bf69) feat: switch `/var/run/...` to be `tmpfs`
 * [3665bdb](https://github.com/argoproj-labs/argo-dataflow/commit/3665bdb3917cbd3fe8864890c9dc20ae94f7cae2) feat: use unix domain socket
 * [7db2d38](https://github.com/argoproj-labs/argo-dataflow/commit/7db2d38f1ea7042332c42c94306c2d6c2ae2a733) Revert "feat: use unix domain socket"
 * [44b321e](https://github.com/argoproj-labs/argo-dataflow/commit/44b321eba08a17858e6eacc34e55ce64beff523a) feat: use unix domain socket
 * [9574072](https://github.com/argoproj-labs/argo-dataflow/commit/9574072c95241cdca0f2d00ea9b7b9c755a34f92) fix: use 32 connections
 * [c0c3aac](https://github.com/argoproj-labs/argo-dataflow/commit/c0c3aac9416bb0a595210e8a02f091ed32b7e3ae) fix: close http source request body
 * [259cf36](https://github.com/argoproj-labs/argo-dataflow/commit/259cf362d75c225c1c33a83b73b11dfab7e7bc6e) fix: close HTTP request body
 * [3a6e5f0](https://github.com/argoproj-labs/argo-dataflow/commit/3a6e5f010025b400f5d84f161d7c942d931c328a) feat: Meta-data. Fixes #161  (#314)
 * [0ac1ec2](https://github.com/argoproj-labs/argo-dataflow/commit/0ac1ec24eacdf60fc82d6a6c8ae6f432a2d7fb3b) feat: Add OpenTracing. Fixes #295 (#322)
 * [8a16a02](https://github.com/argoproj-labs/argo-dataflow/commit/8a16a0245c4b6639cd700b8eea25824700699e3a) feat: expose replicas metric with a new approach (#315)
 * [dfefd4e](https://github.com/argoproj-labs/argo-dataflow/commit/dfefd4e0fe6156e10e8ade4e8c887eb10c222978) refactor: make volumes consistent (#317)
 * [1a73899](https://github.com/argoproj-labs/argo-dataflow/commit/1a73899e86458ad283425ef8d25519012b1d23d7) Revert "fix: uses signal.NotifyContext"
 * [5c7ed77](https://github.com/argoproj-labs/argo-dataflow/commit/5c7ed772a2f597f15af0b0e5c5634d7f3a3a8ce0) fix: uses signal.NotifyContext
 * [8e992cc](https://github.com/argoproj-labs/argo-dataflow/commit/8e992cc5f6b35f3671b2edc99cb6bd96d714abc4) chore!: change clusterName to cluster. Fixes #313
 * [935cf15](https://github.com/argoproj-labs/argo-dataflow/commit/935cf15cbdb3968bdd24ad4cd26d13c31bbed420) fix: quick-start add create argo-server step and set default ARGO_DATAFLOW_CLUSTER_NAME env (#233)
 * [fe7c683](https://github.com/argoproj-labs/argo-dataflow/commit/fe7c683f0a58ed3419ed7e502e27fa3869c3c594) feat: export prometheus metrics from memory (#311)

### Contributors

 * Alex Collins
 * Derek Wang
 * meijin

## v0.0.101 (2021-09-08)

 * [9a0d7cf](https://github.com/argoproj-labs/argo-dataflow/commit/9a0d7cf8cc7a787a45df554d8730e024d27d462b) fix: fix-up context handling (#305)
 * [fbae663](https://github.com/argoproj-labs/argo-dataflow/commit/fbae6639d91a29f7da4f12fd066d665fc0955287) fix: report error correctly
 * [c3cb97f](https://github.com/argoproj-labs/argo-dataflow/commit/c3cb97fd53194335bea4b8133f1a8ff2e64cdd83) feat: support insecureSkipVerify for HTTP sink (#304)
 * [19bc8c8](https://github.com/argoproj-labs/argo-dataflow/commit/19bc8c8281f4f495b8e51097b383c769b1d96965) feat: Updating Python SDK to support Generator Steps & Asyncio handlers. (#299)

### Contributors

 * Alex Collins
 * Dominik Deren

## v0.0.100 (2021-09-07)

 * [ab6d38a](https://github.com/argoproj-labs/argo-dataflow/commit/ab6d38a41806fdc85908d0d3bb63480199bf5c10) fix: log failed to process message

### Contributors

 * Alex Collins

## v0.0.99 (2021-09-02)

 * [45f7847](https://github.com/argoproj-labs/argo-dataflow/commit/45f784725a15eb432bf5fe2d6ea558e93ee6b312) feat: Expand Kafka sink auto-commit config (#293)
 * [6b406c2](https://github.com/argoproj-labs/argo-dataflow/commit/6b406c2a4c62f655d9ae4ae3dce480559266bd29) feat: Added options  MaxMessageSize and Resource for step container for processing large messages (#282)
 * [01dc7a6](https://github.com/argoproj-labs/argo-dataflow/commit/01dc7a66733ad9ad2e0559ad8c9d7ccac4b7ad4b) fix: Argo-Server should still use old image repository (#290)
 * [f4d0ca5](https://github.com/argoproj-labs/argo-dataflow/commit/f4d0ca5d9f640c1f60f936ccd3a4938734716275) fix: fix image name in config

### Contributors

 * Alex Collins
 * Dominik Deren
 * Saravanan Balasubramanian

## v0.0.98 (2021-08-30)

 * [50efb44](https://github.com/argoproj-labs/argo-dataflow/commit/50efb447c85952cb3d62408ff38d564304e99dea) feat: support for AWS session token for S3 (#281)

### Contributors

 * Vigith Maurice

## v0.0.97 (2021-08-30)

 * [93b7f76](https://github.com/argoproj-labs/argo-dataflow/commit/93b7f762c122c466ce5b1c6b63dc05822ded9d4a) fix: listen on HTTPS
 * [0d5175c](https://github.com/argoproj-labs/argo-dataflow/commit/0d5175c37ec1c68fc336af1a1af4f58ce7b1f9c3) chore!: change `map` and `filter` from string to object
 * [b4376d0](https://github.com/argoproj-labs/argo-dataflow/commit/b4376d0afd8a321b733476816c247068e3e5a4ad) fix: port TLS cert code from workflows
 * [64e66d7](https://github.com/argoproj-labs/argo-dataflow/commit/64e66d7777ab310a4dfa3bcee39841bdd757dbe9) fix: port TLS cert code from workflows
 * [3e69afd](https://github.com/argoproj-labs/argo-dataflow/commit/3e69afd5e7ed49e5e08b037dfe3d594be7d5cb75) feat: `make pre-commit -B`

### Contributors

 * Alex Collins

## v0.0.96 (2021-08-23)

 * [be2b63d](https://github.com/argoproj-labs/argo-dataflow/commit/be2b63d63172f0864874945f18347680028fcbbf) feat: Use pod priority class to prioritize the lead replica. Fixes #269 (#275)
 * [8bdfbf2](https://github.com/argoproj-labs/argo-dataflow/commit/8bdfbf2b274fde2b37c7b1d1bef12b9181703358) add kafka async metrics (#276)
 * [dcccd6f](https://github.com/argoproj-labs/argo-dataflow/commit/dcccd6fff2eed86f4807cf5fb47cf8a123134af3) use default k3d cluster

### Contributors

 * Alex Collins
 * Vigith Maurice

## v0.0.95 (2021-08-19)

 * [7dc5e75](https://github.com/argoproj-labs/argo-dataflow/commit/7dc5e753431cbea900f51b4fe62a787cf2dcd372) chore `make pre-commit -B`
 * [744a563](https://github.com/argoproj-labs/argo-dataflow/commit/744a5636fe90c9f242116dd8968dd022e19156ac) fix: fix scaling logic
 * [1869fb1](https://github.com/argoproj-labs/argo-dataflow/commit/1869fb11cab71f5287a762cdefec71324cdc370d) fix: correct error in currentReplicas
 * [53269f7](https://github.com/argoproj-labs/argo-dataflow/commit/53269f710708e49f8bda1b71aade40711f9bef0e) feat: Enable TotalBytes metrics (#258)
 * [53a319d](https://github.com/argoproj-labs/argo-dataflow/commit/53a319d21c3e35fedda5db9b5e9d4d1f6add0f38) feat: Unique Kafka Consumer GroupID (#251)

### Contributors

 * Alex Collins
 * Saravanan Balasubramanian

## v0.0.94 (2021-08-18)


### Contributors


## v0.0.93 (2021-08-18)

 * [d7f9eb4](https://github.com/argoproj-labs/argo-dataflow/commit/d7f9eb44d3b6b2eed085b8b80054295f83cc14ca) fix: correct copy back logic

### Contributors

 * Alex Collins

## v0.0.92 (2021-08-18)

 * [db6f875](https://github.com/argoproj-labs/argo-dataflow/commit/db6f875ed422f5313f352c3597e9916624fde647) fix: correct copy back logic

### Contributors

 * Alex Collins

## v0.0.91 (2021-08-18)

 * [a3c961a](https://github.com/argoproj-labs/argo-dataflow/commit/a3c961a8b676cf456b4118769879e1ed305692c1) fix!: give scaling variable long, easy to understand, names
 * [7bdff21](https://github.com/argoproj-labs/argo-dataflow/commit/7bdff21e62c9f67c5e5d1eab1d29ae44559f06be) feat: add volume sink. Fixes #265
 * [2afbc66](https://github.com/argoproj-labs/argo-dataflow/commit/2afbc66377c534da31a1b58af5dad8f2bdf9219b) feat: adds volume source. Fixes #262
 * [061efa5](https://github.com/argoproj-labs/argo-dataflow/commit/061efa52117ab5e7395437a49968f82febaa1b9a) fix: change to `metav1.Duration` pointer types so defaults work
 * [104460f](https://github.com/argoproj-labs/argo-dataflow/commit/104460f14ca52c6d14255619b702b065203779e1) feat: add `limit` for desired replicas to allow more flexibility is sc… (#261)
 * [dc46d4d](https://github.com/argoproj-labs/argo-dataflow/commit/dc46d4daac7ac1e1be8a0e0fab7801a8bdd209cb) feat: surface S3 pending
 * [67d14aa](https://github.com/argoproj-labs/argo-dataflow/commit/67d14aa9bc04c616d6e549ec3dcf1b331d17397b) feat: surface S3 concurrency configuration
 * [0786c0e](https://github.com/argoproj-labs/argo-dataflow/commit/0786c0e0f839eba4ea8a2e6784a53bdbf43fcaf0) fix: fix peek and scaling delay to Python DSL
 * [531f4b8](https://github.com/argoproj-labs/argo-dataflow/commit/531f4b8c6bb061bf79122aa09f9953f12f9cc094) fix: add peek and scaling delay to Python DSL
 * [af42349](https://github.com/argoproj-labs/argo-dataflow/commit/af423498462889b08aa63114243ad180412b3a67) fix: correct logging of peek/scaling delays

### Contributors

 * Alex Collins

## v0.0.90 (2021-08-16)

 * [dab337c](https://github.com/argoproj-labs/argo-dataflow/commit/dab337c8caaebf535177759edc3277d4837592c8) feat!: Expression based scaling. Fixes #249 (#250)
 * [1e3ad21](https://github.com/argoproj-labs/argo-dataflow/commit/1e3ad216c783622e0306f9d5e5a5ad1104fb8a49) fix: correct deletionDelay time (must be pointer to marshall correctly)
 * [acc648f](https://github.com/argoproj-labs/argo-dataflow/commit/acc648fa5bb24ca98b69bc7e42a6584e4ebae82d) fix: log deletion delay
 * [3be7f7f](https://github.com/argoproj-labs/argo-dataflow/commit/3be7f7f8ed7d4d1f88f2e440fc66768311db44a0) fix: Correct logging of resource names
 * [a81dde5](https://github.com/argoproj-labs/argo-dataflow/commit/a81dde52e5dcbd9ce9c2687bbff6eef80692ab95) feat: Move deletionDelay to pipeline spec
 * [2e807b0](https://github.com/argoproj-labs/argo-dataflow/commit/2e807b01904f389454d0a49aca800fb635437892) config: use argocli:latest

### Contributors

 * Alex Collins

## v0.0.89 (2021-08-12)

 * [3ba7708](https://github.com/argoproj-labs/argo-dataflow/commit/3ba77087e5fd48a38e7ed6cd9eef44b5a854cf25) fix: Prevent scaling from deleting pods
 * [5523aa3](https://github.com/argoproj-labs/argo-dataflow/commit/5523aa33d93d657d7f4cd15f2a1cf0ff98d75730) fix: Required debug enabled for pprof. Fixes #241

### Contributors

 * Alex Collins

## v0.0.88 (2021-08-10)

 * [85f4d1f](https://github.com/argoproj-labs/argo-dataflow/commit/85f4d1fa8537e2e0d7a783b4f89a73a766a6830e) feat: db source (#215)
 * [5e07456](https://github.com/argoproj-labs/argo-dataflow/commit/5e074569add0536a1d968f13d1025bdef3b7a72e) feat: add log sink message truncation (#225)
 * [c47c023](https://github.com/argoproj-labs/argo-dataflow/commit/c47c023a7fbafc190e30eee9e938a8e78becbbf7) feat: Adding ability to specify imagePullSecrets for pipelines (#214)
 * [f5b7751](https://github.com/argoproj-labs/argo-dataflow/commit/f5b7751ef29da843310d5f2b23837d7988377fa6) fix: update sidecar locking to try and avoid sawing metric values. Fixes #224

### Contributors

 * Alex Collins
 * Derek Wang
 * Dominik Deren

## v0.0.87 (2021-08-09)

 * [95f8cb2](https://github.com/argoproj-labs/argo-dataflow/commit/95f8cb25bf831ab1c99e4b73cc2dedff13da1a76) feat: Expose async and sidecarResources in Python DSL

### Contributors

 * Alex Collins

## v0.0.86 (2021-08-09)

 * [d3a7a72](https://github.com/argoproj-labs/argo-dataflow/commit/d3a7a7222261174c1e777d6507b184bae598b0e7) feat: Enable config of sidecar resources
 * [00c3178](https://github.com/argoproj-labs/argo-dataflow/commit/00c3178e7dba8868d2cf64238c5d0b0aac497898) feat: Add support for Kafka async producer. Fixes #216 (#220)

### Contributors

 * Alex Collins

## v0.0.85 (2021-08-07)

 * [bbec883](https://github.com/argoproj-labs/argo-dataflow/commit/bbec8838892668cdb52008a224be3ccc3e35830e) feat: Updating Python runtime to use Python SDK (#212)
 * [d7a0540](https://github.com/argoproj-labs/argo-dataflow/commit/d7a0540a642dbd1c6d6145983a91815c11398db3) feat: NodeJS Runtime & working examples for NodeJS & Python runtimes & sdks. (#211)

### Contributors

 * Dominik Deren

## v0.0.84 (2021-08-06)

 * [ab5ccde](https://github.com/argoproj-labs/argo-dataflow/commit/ab5ccded626f6f9a6fd90203c8346eb3a69ba47b) fix: set Kafka consumer max limit to 16x default (16m)

### Contributors

 * Alex Collins

## v0.0.83 (2021-08-06)


### Contributors


## v0.0.82 (2021-08-06)

 * [c31e5e6](https://github.com/argoproj-labs/argo-dataflow/commit/c31e5e648e2f6942d20be50f2bb18f25aa611fae) fix: use one Kafka client for consumer/producer
 * [1b18462](https://github.com/argoproj-labs/argo-dataflow/commit/1b18462a4a58f6ddbd92371315dd962a484d6f6c) fix: use one Kafka client for consumer/producer

### Contributors

 * Alex Collins

## v0.0.81 (2021-08-05)

 * [616753f](https://github.com/argoproj-labs/argo-dataflow/commit/616753f06b1cc70e3ab438bb4d9e4c1298d764e6) fix: nats disconnect log (#206)
 * [c52ff23](https://github.com/argoproj-labs/argo-dataflow/commit/c52ff230bea111eb145957b0cabe5fde512539a9) fix: Cache and re-use Kafka clients. Fixes #199 (#200)
 * [a5ba4fe](https://github.com/argoproj-labs/argo-dataflow/commit/a5ba4fe83cd83fe89931de375afcecd750426f76) fix: Change Kafka max-processing time from 100ms to 10s
 * [efa15a6](https://github.com/argoproj-labs/argo-dataflow/commit/efa15a63429caf6262b3ea8f9bb596d8532a0ff9) feat: change hash from b64 to hex (#203)
 * [0066afb](https://github.com/argoproj-labs/argo-dataflow/commit/0066afbedbaaaa2000557bdd6db437fc3d4e0805) fix!: Mandate configured cluster name (#198)
 * [4a645de](https://github.com/argoproj-labs/argo-dataflow/commit/4a645de51495eba2aa1442351f5bd0290a7dc91b) Revert "fix: Removed noisy Kafka logging"
 * [a4d7977](https://github.com/argoproj-labs/argo-dataflow/commit/a4d7977b83d399258f986349ca93f6752f1dbe33) fix!: Update Kafka group name/id to include namespace.
 * [07cb751](https://github.com/argoproj-labs/argo-dataflow/commit/07cb7517d385bf7e82f280de5b9f658f38490b3a) feat: Set `allowPrivilegeEscalation: false`
 * [7a25bb8](https://github.com/argoproj-labs/argo-dataflow/commit/7a25bb826afa8bebb4d6bc37dc560ca8fe240bcd) feat: Generate self-signed certificates (#196)
 * [9aac4e8](https://github.com/argoproj-labs/argo-dataflow/commit/9aac4e84755edb916ff75527d06d148cc2e3c42b) fix: Fix for #190, broken nodejs sdk test. (#194)

### Contributors

 * Alex Collins
 * Derek Wang
 * Dominik Deren

## v0.0.80 (2021-08-03)

 * [2cfee32](https://github.com/argoproj-labs/argo-dataflow/commit/2cfee3216d0668897ea502cb9562cd9e5e6291cf) feat!: Change HTTP endpoints to have TLS v1.2. Fixes #178 (#193)
 * [68084f2](https://github.com/argoproj-labs/argo-dataflow/commit/68084f28f32c67810dc4ddf49f82f5a70b4d6496) fix!: Add authentication to HTTP source. Fixes #152 (#187)
 * [656af83](https://github.com/argoproj-labs/argo-dataflow/commit/656af835773fc0f2dd5172610f0ea87a76019270) feat: NodeJS SDK for handling the container contract (#188)
 * [2ba29ce](https://github.com/argoproj-labs/argo-dataflow/commit/2ba29cef0ca874bb54a2802a49ec2590e2343698) fix: Removed noisy Kafka logging
 * [f6b8915](https://github.com/argoproj-labs/argo-dataflow/commit/f6b8915a817c271e5d0c8b08278a61fb3f475f0f) fix!: Move bearer token from environment variable to file.

### Contributors

 * Alex Collins
 * Dominik Deren

## v0.0.79 (2021-08-02)

 * [2f3cdbc](https://github.com/argoproj-labs/argo-dataflow/commit/2f3cdbc32da2df910a21ea94337fdda4aae1a0c9) feat!: updated default backoff to cap at 1 day
 * [7b00de2](https://github.com/argoproj-labs/argo-dataflow/commit/7b00de2535de42c55ca2acdbb646cbc38779ae2f) feat: Drop-all capabilites by default. Fixes #142 (#181)
 * [78e1b32](https://github.com/argoproj-labs/argo-dataflow/commit/78e1b32fe5276fd04bb20d76718320eb9825bba1) feat: log `backoffSteps`, so we can know if an error will retry

### Contributors

 * Alex Collins

## v0.0.78 (2021-07-30)

 * [2e80ead](https://github.com/argoproj-labs/argo-dataflow/commit/2e80ead67787089cea5ee7d72465f4c502d39836) fix: allow Kafka TLS without certs

### Contributors

 * Alex Collins

## v0.0.77 (2021-07-30)

 * [a3ee6a7](https://github.com/argoproj-labs/argo-dataflow/commit/a3ee6a7426da45ac696e4d7de77a24c61f700b7b) fix: add ssh configmap

### Contributors

 * Alex Collins

## v0.0.76 (2021-07-30)

 * [507f1e7](https://github.com/argoproj-labs/argo-dataflow/commit/507f1e7831ded71888cb1058f793f348ddbd812e) feat: db sink (#165)
 * [f1ab11c](https://github.com/argoproj-labs/argo-dataflow/commit/f1ab11ce7801d046930c51fe4cff41e57c254858) fix: Adding missing examples. (#166)
 * [7e73134](https://github.com/argoproj-labs/argo-dataflow/commit/7e73134ef9bae9dac62ceef84ac88f91946e6246) fix: fix python runtime
 * [76054c2](https://github.com/argoproj-labs/argo-dataflow/commit/76054c261e3742c74388dd3e17d120c5216af0e2) feat: Adding Python SDK for git step with Python3-9 runtime. (#163)
 * [c42f9be](https://github.com/argoproj-labs/argo-dataflow/commit/c42f9bee49544dca454961251f45d7012502f4cc) fix: fix S3 test
 * [5acf7dd](https://github.com/argoproj-labs/argo-dataflow/commit/5acf7dd9014742858980771614e1e62d802f1e28) Update README.md (#157)
 * [82c5f4c](https://github.com/argoproj-labs/argo-dataflow/commit/82c5f4c00c05b214584596df6f4c3bb849dfa00a) feat: updated S3 sink to use shared path rather than messages. Fixes #159
 * [aa70b35](https://github.com/argoproj-labs/argo-dataflow/commit/aa70b359eed48a5f602fe25b847596e81e95b8b2) fix: Fixing incorrectly passed auth arguments in git basic auth. (#156)
 * [ed0cc74](https://github.com/argoproj-labs/argo-dataflow/commit/ed0cc74f2d54794c7559770dcfc7e75899da33da) feat: Updating git step action to use credentials based authentication. (#151)

### Contributors

 * Alex Collins
 * Derek Wang
 * Dominik Deren

## v0.0.75 (2021-07-27)

 * [4ef8477](https://github.com/argoproj-labs/argo-dataflow/commit/4ef847768c5af4ce4406f7f440c113a5fb1da64f) fix: improve/fix s3 support by using workqueue to avoid repeated work
 * [4981e06](https://github.com/argoproj-labs/argo-dataflow/commit/4981e06cc13ca7f675b1680843b3b4406e11f533) fix: add known_hosts for git cloning private repos
 * [e9b9ae6](https://github.com/argoproj-labs/argo-dataflow/commit/e9b9ae65287bdfd52be9d81ab0db1746ad212e36) feat: add S3 sink

### Contributors

 * Alex Collins

## v0.0.74 (2021-07-26)

 * [eaa0e22](https://github.com/argoproj-labs/argo-dataflow/commit/eaa0e2277765d9048a28d6d2f5db6a0f381d029f) feat: upgrade to Kubernetes v0.20.5
 * [a9f59e6](https://github.com/argoproj-labs/argo-dataflow/commit/a9f59e68fed3ed9a8c7c5aaaa3a57ec82b1bf1cd) fix: Fix S3 poll period type
 * [3e9334c](https://github.com/argoproj-labs/argo-dataflow/commit/3e9334c37dac92eb9254d2a2d855a4e1b333b46e) feat: Add S3 source
 * [9e17d95](https://github.com/argoproj-labs/argo-dataflow/commit/9e17d951a63828052afc2b715c80dbceb359fc69) fix: do not Kafka commit error message
 * [bb3fcaa](https://github.com/argoproj-labs/argo-dataflow/commit/bb3fcaa6a5b9a642e9b4234bc2afe6b3580d9f88) feat: git ssh private key

### Contributors

 * Alex Collins

## v0.0.73 (2021-07-23)

 * [e166364](https://github.com/argoproj-labs/argo-dataflow/commit/e1663641925815d836990c26777061c115aba923) fix: make codegen
 * [3456b18](https://github.com/argoproj-labs/argo-dataflow/commit/3456b1834813c1249299491579b81f63a7ae2b20) feat: kafka authentication (#141)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.72 (2021-07-21)

 * [40211bf](https://github.com/argoproj-labs/argo-dataflow/commit/40211bfd5f59f84bc7522f3d1d9c2d71f2748625) chore!: remove message and time from status to reduce complexity
 * [18815b5](https://github.com/argoproj-labs/argo-dataflow/commit/18815b51f53b755d44c111d841683a1472c9c74e) feat: add command to GitStep in Python DSL
 * [003d94e](https://github.com/argoproj-labs/argo-dataflow/commit/003d94ead7e949597424259c748cce306abccf18) feat: add terminator to Python DSL
 * [8293394](https://github.com/argoproj-labs/argo-dataflow/commit/82933943b322931e602b791af9eb2c038122c3d9) feat: add terminator to Python DSL
 * [e5face5](https://github.com/argoproj-labs/argo-dataflow/commit/e5face59462aa88a5e736fbfce6a6a1e28bf4971) feat: label service
 * [e4215b3](https://github.com/argoproj-labs/argo-dataflow/commit/e4215b3812d4ea310e402ed41311ba7a27b54fe6) Revert "Revert "test: add expand_step_test.go. Fixes #131""
 * [d5cb429](https://github.com/argoproj-labs/argo-dataflow/commit/d5cb429a1c1fe57b69335a68ab18abe785f101ac) Revert "test: add expand_step_test.go. Fixes #131"
 * [9bc52a6](https://github.com/argoproj-labs/argo-dataflow/commit/9bc52a638d0653e366c34eafdd132e6e3a423ebe) fix: use dumb-init to correctly handle signals with entrypoint.sh scripts (#134)

### Contributors

 * Alex Collins

## v0.0.71 (2021-07-20)

 * [7dce579](https://github.com/argoproj-labs/argo-dataflow/commit/7dce5795f66aa9f26156b846d63c35b5e5677b0d) fix: do not loop logging error
 * [bc8f72c](https://github.com/argoproj-labs/argo-dataflow/commit/bc8f72c095b059750335dd31b9f5d75fc53136f8) chore!: Change `go1-16` to `golang1-16`
 * [937e874](https://github.com/argoproj-labs/argo-dataflow/commit/937e874402037cbc677ac010ddc83e30152bc087) feat: add Golang SDKs draft
 * [391db24](https://github.com/argoproj-labs/argo-dataflow/commit/391db243aef2e41095b3247509831c58a21ab094) feat!: rename "handler" to "code". Fixes #99
 * [29741e5](https://github.com/argoproj-labs/argo-dataflow/commit/29741e550a205fd0f2692c58d461ab94da774036) tests: add container step test. Fixes #123
 * [f06c4c8](https://github.com/argoproj-labs/argo-dataflow/commit/f06c4c89818bb6a506b9e11091edb40e9fc87391) tests: add map test. Fixes #126

### Contributors

 * Alex Collins

## v0.0.70 (2021-07-20)

 * [d71b06d](https://github.com/argoproj-labs/argo-dataflow/commit/d71b06dc965fcbb3bcc6131119871becabd2f1ca) fix: report pending Kafka messages
 * [c4aa21b](https://github.com/argoproj-labs/argo-dataflow/commit/c4aa21bcb95ff627ff66ea39d4b506b22610db73) Sarama (#121)

### Contributors

 * Alex Collins

## v0.0.69 (2021-07-19)

 * [2c3e65f](https://github.com/argoproj-labs/argo-dataflow/commit/2c3e65ff4715e4129ec7b8af6bc5529f8b6b5214) fix: change to use logrus rather than zap

### Contributors

 * Alex Collins

## v0.0.68 (2021-07-16)

 * [6f3cc61](https://github.com/argoproj-labs/argo-dataflow/commit/6f3cc61fda2938432317606ea168d505225b29d7) feat: support configurable Kafka starting offset
 * [4cdf4bb](https://github.com/argoproj-labs/argo-dataflow/commit/4cdf4bb117833fee39872b3740053a91e32ab038) fix: Kafka should start at LastOffset by default
 * [af46d66](https://github.com/argoproj-labs/argo-dataflow/commit/af46d66cc6a09b4d1286b06aad714c8c731aa63e) fix: correct source hook to be pre-stop hook
 * [5544da5](https://github.com/argoproj-labs/argo-dataflow/commit/5544da5a08e64e92f9b68bf02666de506f3175fc) refactor: refactor sources
 * [4398de4](https://github.com/argoproj-labs/argo-dataflow/commit/4398de44ae42d51607b74f9fa432ba938a50b98d) feat: expose maxInflight for stan config (#117)
 * [51c21ad](https://github.com/argoproj-labs/argo-dataflow/commit/51c21add95e2176c33016066101ebb57adbd3426) feat: Partially migrate from sarama to kafka-go (#116)
 * [39cd13a](https://github.com/argoproj-labs/argo-dataflow/commit/39cd13afff9d5aa460c6411cca6035a97cbf566c) feat: make Kafka commitN configurable. Fixes #114

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.67 (2021-07-15)

 * [573b31c](https://github.com/argoproj-labs/argo-dataflow/commit/573b31c50db92c85aad9fbf640a79bbfb67410cd) feat: stan reliable auto reconnection (#112)

### Contributors

 * Derek Wang

## v0.0.66 (2021-07-15)

 * [8f81586](https://github.com/argoproj-labs/argo-dataflow/commit/8f81586a710505decd2855b02629a60d7d4ac276) feat: add version metrics. Closes #109
 * [1adec35](https://github.com/argoproj-labs/argo-dataflow/commit/1adec358710bf0af3c84f5de5be7263b0b3bc6ca) fix: add retries labels
 * [5465d3c](https://github.com/argoproj-labs/argo-dataflow/commit/5465d3c2a8a67f906198fcd65c54ef38516ebfef) fix: wait for Kafka

### Contributors

 * Alex Collins

## v0.0.65 (2021-07-13)

 * [e8fa437](https://github.com/argoproj-labs/argo-dataflow/commit/e8fa437ccd0f82ed6a2b50a194edb1b738444a22) fix: use shared counter for Kafka commit
 * [3ad7cf1](https://github.com/argoproj-labs/argo-dataflow/commit/3ad7cf1ddcf70752ebdd9684bb8ccb10aed0d8f1) fix: stop Kafka dropping messages on disruption
 * [9a86950](https://github.com/argoproj-labs/argo-dataflow/commit/9a8695069c1d68923e6340d75d6a420c1805a339) fix: return 503 if HTTP source not ready due to pre-stop
 * [2155591](https://github.com/argoproj-labs/argo-dataflow/commit/2155591cbbe026a403e690a55b51cfece68e2adf) feat: change default deletion delay to 720h (~30d)
 * [8fb4c17](https://github.com/argoproj-labs/argo-dataflow/commit/8fb4c17eb838023e43e1e647205c185745ac434b) fix!: harmonize `retries`
 * [0aa7aa4](https://github.com/argoproj-labs/argo-dataflow/commit/0aa7aa40281d0f1844c3a253659c428b6df7e823) fix: add timeout to HTTP sink
 * [f5e89b9](https://github.com/argoproj-labs/argo-dataflow/commit/f5e89b99ad46240b0f4b175f95476c2f9d4edec0) fix: correct hash

### Contributors

 * Alex Collins

## v0.0.64 (2021-07-12)

 * [11b2033](https://github.com/argoproj-labs/argo-dataflow/commit/11b2033b81a4ac2ecdd9e449cd3c11bba652e653) feat: update resources to only apply to built-ins

### Contributors

 * Alex Collins

## v0.0.63 (2021-07-12)

 * [d389e44](https://github.com/argoproj-labs/argo-dataflow/commit/d389e447caa132a668596e5f0c159f9368b572a4) feat: change pod hash when image changes

### Contributors

 * Alex Collins

## v0.0.62 (2021-07-12)


### Contributors


## v0.0.61 (2021-07-12)

 * [f306c0c](https://github.com/argoproj-labs/argo-dataflow/commit/f306c0c2bada95e71a44a6a3045812508256e77a) fix: allow lastUpdated to be empty

### Contributors

 * Alex Collins

## v0.0.60 (2021-07-12)

 * [6ac6810](https://github.com/argoproj-labs/argo-dataflow/commit/6ac68101dc04c6aefb6cf645331040e0668f25e9) feat(controller): use version from build
 * [858d717](https://github.com/argoproj-labs/argo-dataflow/commit/858d71754e542564728772d96863549294696a7d) feat: support HTTP headers for HTTP sink
 * [ad31740](https://github.com/argoproj-labs/argo-dataflow/commit/ad3174034d0c02217c0689acd423aa6e2faaa314) fix: correctly step
 * [4e1d2ff](https://github.com/argoproj-labs/argo-dataflow/commit/4e1d2ff4c19bdda65d635f001ca8faa757c835fb) fix: delay start-up until Kafka is ready
 * [810aef7](https://github.com/argoproj-labs/argo-dataflow/commit/810aef721f745f7a90a6aee395a294ef3ff65d75) feat: automatic GC of pipelines 30m after completion
 * [2135e1d](https://github.com/argoproj-labs/argo-dataflow/commit/2135e1d6f52b1d2d5d9caf45a0e21bc0ee2bb5e0) feat: add lastUpdated field to pipeline
 * [e9b2218](https://github.com/argoproj-labs/argo-dataflow/commit/e9b2218ad3c20264ecafc78908ad9d979342f2e2) fix: fix HTTP shutdown
 * [8ac27da](https://github.com/argoproj-labs/argo-dataflow/commit/8ac27daed0feebddc6f0af5799972eb537c9e222) fix: fix HTTP shutdown
 * [d13284f](https://github.com/argoproj-labs/argo-dataflow/commit/d13284fd5a484beea98ca84c225ba23e7560dee9) config: update config
 * [7da3df6](https://github.com/argoproj-labs/argo-dataflow/commit/7da3df67b0f626fd357823397a5009ddf9953d31) feat: Update retry (#91)
 * [9db646a](https://github.com/argoproj-labs/argo-dataflow/commit/9db646a2ef38d498761cbb6cb91a7969bb6a64b9) feat: Enable the Message retry metrics (#75)
 * [0299cd8](https://github.com/argoproj-labs/argo-dataflow/commit/0299cd80aaa6bafa126ad8c4efc6f98acf212f41) feat: stan supports token auth (#87)

### Contributors

 * Alex Collins
 * Derek Wang
 * Saravanan Balasubramanian

## v0.0.59 (2021-06-29)

 * [778f24a](https://github.com/argoproj-labs/argo-dataflow/commit/778f24a5ab34ad6f8b3b422010e0c149e0be1c0d) fix: prevent Kafka pending loop dieing on disconnection
 * [f289249](https://github.com/argoproj-labs/argo-dataflow/commit/f2892494ca727f0541b44378796d18beb9bde134) refactor: container killer refactor
 * [7e793bf](https://github.com/argoproj-labs/argo-dataflow/commit/7e793bf7afdc4a2a4ee2ab647a995e7905274fec) fix: improved dedupe

### Contributors

 * Alex Collins

## v0.0.58 (2021-06-28)

 * [70acfa7](https://github.com/argoproj-labs/argo-dataflow/commit/70acfa7895545e6c93116b3ae6a918d0549be73e) fix: improved dedupe
 * [69e43d6](https://github.com/argoproj-labs/argo-dataflow/commit/69e43d6fbf0858fd042ebafb84c53cfaf4b7e377) test http fmea

### Contributors

 * Alex Collins

## v0.0.57 (2021-06-24)

 * [972ebf0](https://github.com/argoproj-labs/argo-dataflow/commit/972ebf03c232bb37ac72bcd99d39b072381eeca2) fix: increased default CPU resources requests to 250m
 * [786bc05](https://github.com/argoproj-labs/argo-dataflow/commit/786bc05ceaaa7059202ea20881f7bd00953c1f5e) fix: move patching step status earlier in shutdown sequence
 * [2f9898b](https://github.com/argoproj-labs/argo-dataflow/commit/2f9898b69c0e580868863153b8877b11cf3d59c8) config: removed invalid resource
 * [3645436](https://github.com/argoproj-labs/argo-dataflow/commit/3645436785ab591b82b594a980c9ebf44b5378c7) fix: correct metrics
 * [8a169ca](https://github.com/argoproj-labs/argo-dataflow/commit/8a169ca4d81e75b5b165e2eab9bb0faeb14e01e4) feat: fix test
 * [752086b](https://github.com/argoproj-labs/argo-dataflow/commit/752086b3fdd517598ad21434d5c8f4edbf941ff0) feat: re-order logging
 * [4b6a9d6](https://github.com/argoproj-labs/argo-dataflow/commit/4b6a9d6981174d5b24d86a2ef52b5fd931d37b93) feat: pass GODEBUG to runner
 * [7a08e5e](https://github.com/argoproj-labs/argo-dataflow/commit/7a08e5ea9127573b21693e1ad029d012fa91a6bd) feat: dedupe
 * [6904aa7](https://github.com/argoproj-labs/argo-dataflow/commit/6904aa7b996849dd0e870f3176201282fd182a9a) feat: add `sha1` func

### Contributors

 * Alex Collins

## v0.0.56 (2021-06-22)

 * [ed68cb0](https://github.com/argoproj-labs/argo-dataflow/commit/ed68cb02c8ca114c9e39708ec7586b056ce76948) config: ARGO_DATAFLOW_UPDATE_INTERVAL=10s for dev
 * [c564148](https://github.com/argoproj-labs/argo-dataflow/commit/c564148e90a2fb7cdcc0038f0469e2956b450ec5) fix(controller): correct service name

### Contributors

 * Alex Collins

## v0.0.55 (2021-06-21)

 * [e5fa079](https://github.com/argoproj-labs/argo-dataflow/commit/e5fa0791858fda3bdc43ccfd8f7348975e9dc905) feat: pprof sidecar
 * [3c0d169](https://github.com/argoproj-labs/argo-dataflow/commit/3c0d169b2fd111e313d9a76b62405f1795665edd) fix: Kafka pending wrong when re-using pipeline
 * [f37aea4](https://github.com/argoproj-labs/argo-dataflow/commit/f37aea4f1cd3283194d0e7a2d7bc59612085188e) fix: revert to auto-commit
 * [24d91a1](https://github.com/argoproj-labs/argo-dataflow/commit/24d91a120ae641909201bfcc816a5d9944e446fb) fix: negative Kafka pending
 * [32a73a7](https://github.com/argoproj-labs/argo-dataflow/commit/32a73a7ae8c6e0993ed2a9f96f1bc4e24607caac) fix: expose stan monitoring port (#78)

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.54 (2021-06-17)

 * [0d68892](https://github.com/argoproj-labs/argo-dataflow/commit/0d688927826574de93aaa8fb9351b108408291ee) feat: add HTTP source service name
 * [e2728c6](https://github.com/argoproj-labs/argo-dataflow/commit/e2728c66fd766f2ebf2f0b24df693c9592cc25df) feat(sidecar): disable Kafka auto-commit
 * [2b90f11](https://github.com/argoproj-labs/argo-dataflow/commit/2b90f11a918d8d63d3156c89cd9c2f3ab312878b) fix(sidecar): fix Kafka pending tiny over-estimate
 * [f296a83](https://github.com/argoproj-labs/argo-dataflow/commit/f296a83301bfff3b54956b67782bc9508c15a4ba) config: added natsMonitoringUrl
 * [04fe2f8](https://github.com/argoproj-labs/argo-dataflow/commit/04fe2f8955162f80905b90b824606830a390316e) feat(sidecar): added sidecar readiness probe
 * [b395ffe](https://github.com/argoproj-labs/argo-dataflow/commit/b395ffef3aec02d74e8f093ffbf5deed98978a6c) feat(sidecar): set kafka max-processing time to 30s
 * [0fc71c2](https://github.com/argoproj-labs/argo-dataflow/commit/0fc71c2a1ff8320d19e6ce20c9334da2397fc00b) chore(sidecar): Refactor sidecar (#70)
 * [994e80e](https://github.com/argoproj-labs/argo-dataflow/commit/994e80e588e46288a4a29e8ad3f7ceafa54d379d) ok
 * [efb20b1](https://github.com/argoproj-labs/argo-dataflow/commit/efb20b11cc378ed4b0ffe01268609601950e5016) feat: smaller binaries without DWARF/symbols
 * [2332aeb](https://github.com/argoproj-labs/argo-dataflow/commit/2332aebce3522e4f43f40b775b5aa748b267088a) feat: smaller binaries without DWARF/symbols

### Contributors

 * Alex Collins

## v0.0.53 (2021-06-16)

 * [454c0d7](https://github.com/argoproj-labs/argo-dataflow/commit/454c0d7e23767f868f00d56e4c4499c4a1db6668) fix: set MaxInFlight and AckWait for stan
 * [6b7be67](https://github.com/argoproj-labs/argo-dataflow/commit/6b7be6738c49daf757f09f96e18cdb1c9710913c) fix: fix logger

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.52 (2021-06-15)

 * [12a98cd](https://github.com/argoproj-labs/argo-dataflow/commit/12a98cd3e549e3fafc727341b7093ee555614911) fix: stan manual ack

### Contributors

 * Derek Wang

## v0.0.51 (2021-06-15)

 * [88f77f7](https://github.com/argoproj-labs/argo-dataflow/commit/88f77f7e308d2c1a3ab8277b9ed8011d3de7f4ec) feat: change Errors to RecentErrors
 * [a4db306](https://github.com/argoproj-labs/argo-dataflow/commit/a4db30654f2c198e8582b2a9d424156081dc1842) fix: bug in STAN ack

### Contributors

 * Alex Collins

## v0.0.50 (2021-06-15)

 * [6f7bb50](https://github.com/argoproj-labs/argo-dataflow/commit/6f7bb505b636d5de21921040d1e9f6ca20a66d0a) config: change argo-server to HTTP
 * [4770ed6](https://github.com/argoproj-labs/argo-dataflow/commit/4770ed65d9e0fedd5215157acf6ada268bfd02d4) config: change argo-server to HTTP
 * [8d4ee45](https://github.com/argoproj-labs/argo-dataflow/commit/8d4ee45871d8308a8a84efff7c5e593dbeeceec0) config: change argo-server to HTTP
 * [7679590](https://github.com/argoproj-labs/argo-dataflow/commit/76795903ebe52a8ea9ca957e35cba5669d3b19e5) config: change argo-server to HTTP
 * [c63a715](https://github.com/argoproj-labs/argo-dataflow/commit/c63a715d14c3216089d6faae972f2a69279ba37f) refactor: rename variables
 * [316abc9](https://github.com/argoproj-labs/argo-dataflow/commit/316abc97b680bc4753474bd6276fb67290094e3d) Update README.md (#68)

### Contributors

 * Alex Collins
 * wanghong230

## v0.0.49 (2021-06-10)

 * [23fc919](https://github.com/argoproj-labs/argo-dataflow/commit/23fc9195e87a74e798b5045a758a06b16a26d975) fix: Add missing kubernetes requirement in setup.py (#64)
 * [dbc403d](https://github.com/argoproj-labs/argo-dataflow/commit/dbc403d4d23a1a047392861e26d1aedcda04f061) fix: fix Python
 * [a2fdaaf](https://github.com/argoproj-labs/argo-dataflow/commit/a2fdaafbf848434f90f6f5fdd338716ef96d832f) feat: publish Python library
 * [a411be8](https://github.com/argoproj-labs/argo-dataflow/commit/a411be82d23310376b416b23ad9a353473ee300a) feat: publish Python library
 * [72b3484](https://github.com/argoproj-labs/argo-dataflow/commit/72b34845eae532e0c7663dd8eb7f867b2ccb9e71) feat: publish Python library
 * [361c47c](https://github.com/argoproj-labs/argo-dataflow/commit/361c47c3f39a8e6ac67005f0020af43559db2717) feat: add a method to run pipelines in DSL

### Contributors

 * Alex Collins
 * Yuan Tang

## v0.0.48 (2021-06-10)

 * [e61c307](https://github.com/argoproj-labs/argo-dataflow/commit/e61c307d9bed9ae33892c57cde4da5a06c26ffbb) feat: add context to Java runtime
 * [5becca6](https://github.com/argoproj-labs/argo-dataflow/commit/5becca695ef81e26e116d578f6385262ed881638) feat: add context to Git example
 * [6b922eb](https://github.com/argoproj-labs/argo-dataflow/commit/6b922ebd6cb73b352bd3c022167fadee4b3e43bd) feat: add context to Python runtime

### Contributors

 * Alex Collins

## v0.0.47 (2021-06-10)

 * [34b75ae](https://github.com/argoproj-labs/argo-dataflow/commit/34b75ae05de30b1e80fa1eb0312d398c59095c77) fix: change default retryPolicy=Always
 * [a9e83fa](https://github.com/argoproj-labs/argo-dataflow/commit/a9e83fa0b033ae9a7c1355d14a4b30acbb946bea) feat: add Golang context

### Contributors

 * Alex Collins

## v0.0.46 (2021-06-09)

 * [268b1ec](https://github.com/argoproj-labs/argo-dataflow/commit/268b1ecd89e9cd86be2384e44eeda441d4d190b8) fix: removed parallel
 * [4cd5e8b](https://github.com/argoproj-labs/argo-dataflow/commit/4cd5e8b248142a8baec7170935de4658fca7ea62) feat: switch to manual ack
 * [e09e8e9](https://github.com/argoproj-labs/argo-dataflow/commit/e09e8e9084900013b1d13fe78832d07135c53af6) fix: clean up parallel for stan

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.45 (2021-06-09)

 * [b8e88fe](https://github.com/argoproj-labs/argo-dataflow/commit/b8e88fe3ba1aa1c62cb965560ccbcfddbeb6e61b) fix: nats pending messages type assersion error

### Contributors

 * Derek Wang

## v0.0.44 (2021-06-09)

 * [12f0e5e](https://github.com/argoproj-labs/argo-dataflow/commit/12f0e5e9f4cf2dbd792f94a5656bddf7aba6052b) fix: correct total counter
 * [bf7b993](https://github.com/argoproj-labs/argo-dataflow/commit/bf7b99324199fc26de068ded51a01f4aba312657) fix: nats pending messages

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.43 (2021-06-08)

 * [a03615d](https://github.com/argoproj-labs/argo-dataflow/commit/a03615d7ccd5d7ed68479cdebff76b0ce5e05a17) ok

### Contributors

 * Alex Collins

## v0.0.42 (2021-06-08)

 * [d2f219f](https://github.com/argoproj-labs/argo-dataflow/commit/d2f219f806277636ab5b4c41dd26ea87c887ec3a) ok
 * [abbf7d0](https://github.com/argoproj-labs/argo-dataflow/commit/abbf7d09b6ef6ca36d7538acfc663907233b08b6) fix: various issues
 * [c6a9e6e](https://github.com/argoproj-labs/argo-dataflow/commit/c6a9e6e7bbee5a60d43ddc0d3b69713a90e8de72) feat: add retryPolicy
 * [8170a88](https://github.com/argoproj-labs/argo-dataflow/commit/8170a8847fe43f9a79eaaf73cd6d39a886294cb2) feat: more work on Python DLS
 * [e306739](https://github.com/argoproj-labs/argo-dataflow/commit/e3067398a57919791a531fd4ed90e27d69028b5e) feat: more work on Python DLS

### Contributors

 * Alex Collins

## v0.0.41 (2021-06-07)

 * [cd2082e](https://github.com/argoproj-labs/argo-dataflow/commit/cd2082e893ac130a4367894b8f1a924fad070f84) feat: simplify so only replica 0 report metrics

### Contributors

 * Alex Collins

## v0.0.40 (2021-06-07)

 * [da6a1e0](https://github.com/argoproj-labs/argo-dataflow/commit/da6a1e0e396a9d225ec1b8b25213597d69b995eb) fix: fix permissions

### Contributors

 * Alex Collins

## v0.0.39 (2021-06-07)

 * [e1cc03c](https://github.com/argoproj-labs/argo-dataflow/commit/e1cc03c070744057d2f4cce0bfb32aa19861a696) config: expose 9090
 * [5f9c09c](https://github.com/argoproj-labs/argo-dataflow/commit/5f9c09c62a72cb779b5595fae24dbaf4df6be400) config: expose 9090

### Contributors

 * Alex Collins

## v0.0.38 (2021-06-07)

 * [fcfb45f](https://github.com/argoproj-labs/argo-dataflow/commit/fcfb45f880d7b414136f16be32dcd5bcdaf303e3) fix: bugs in grouping
 * [bf55468](https://github.com/argoproj-labs/argo-dataflow/commit/bf55468731dfb63244b66062ab30c93c9edfb992) refactor: de-couple scaling
 * [2695c7a](https://github.com/argoproj-labs/argo-dataflow/commit/2695c7a8f5493bdf5e895e86b030e45100a3009d) feat: more Python DSL
 * [f6a6d0d](https://github.com/argoproj-labs/argo-dataflow/commit/f6a6d0d381611bb0f03eaf1e76a3b09fc007b93e) feat: more Python DSL
 * [85908bb](https://github.com/argoproj-labs/argo-dataflow/commit/85908bb2263de3b8353fb6acda4b6e3574b5c149) feat: add Python DSL

### Contributors

 * Alex Collins

## v0.0.37 (2021-06-06)

 * [267fd6b](https://github.com/argoproj-labs/argo-dataflow/commit/267fd6b03450b70629c3d7b43674fb988d863a34) feat: added latency metric
 * [c1c88da](https://github.com/argoproj-labs/argo-dataflow/commit/c1c88da7ff9a4448f08541cdab6ddabfe61b8d88) fix: in_flight -> inflight

### Contributors

 * Alex Collins

## v0.0.36 (2021-06-05)

 * [ca7edf1](https://github.com/argoproj-labs/argo-dataflow/commit/ca7edf16c34075efb77d0cc1e76c40b943ab75d5) feat: add in_flight metric
 * [009731a](https://github.com/argoproj-labs/argo-dataflow/commit/009731a0bab298f1d50a6a0f64278ddfba7ca2a5) refactor: made code more testable
 * [a3142eb](https://github.com/argoproj-labs/argo-dataflow/commit/a3142eb7b5755a064acf47066a8c99604bfa73d9) fix: add bearer token

### Contributors

 * Alex Collins

## v0.0.35 (2021-06-04)

 * [421c142](https://github.com/argoproj-labs/argo-dataflow/commit/421c14269c618eb1a1c601715e6ff361d45d9921) fix: total/error calcs
 * [06fc5f0](https://github.com/argoproj-labs/argo-dataflow/commit/06fc5f0592ce74982864f133af5dd405948451f0) fix: only return pending for replica=0

### Contributors

 * Alex Collins

## v0.0.34 (2021-06-04)

 * [c73ce5b](https://github.com/argoproj-labs/argo-dataflow/commit/c73ce5b33a6bfc42abf772d6f794c6af885499a0) fix: change from patch to update for the controller
 * [4a32fb9](https://github.com/argoproj-labs/argo-dataflow/commit/4a32fb9fcd72181bb6299d7d123166e29369a8e7) fix: kafka pending messages

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.33 (2021-06-04)

 * [60eab31](https://github.com/argoproj-labs/argo-dataflow/commit/60eab313fe856ee34b44d8f7cfcb4057d01c6344) fix: prevent violent scale-up and scale-down by only scaling by 1 each time

### Contributors

 * Alex Collins

## v0.0.32 (2021-06-04)

 * [ca363fe](https://github.com/argoproj-labs/argo-dataflow/commit/ca363fe3638fa6c329dc786dedb4aaf8d230a8f3) feat: add pending metric
 * [f3a4813](https://github.com/argoproj-labs/argo-dataflow/commit/f3a4813aaf23d630f4d8292792d6b147d53515f0) fix: fix metrics

### Contributors

 * Alex Collins

## v0.0.31 (2021-06-04)

 * [3dd39f1](https://github.com/argoproj-labs/argo-dataflow/commit/3dd39f16bf6d1104af3d2a3e4c23b98c22170639) fix(sidecar): updated to use fixed counters
 * [6bef420](https://github.com/argoproj-labs/argo-dataflow/commit/6bef420f18718ca977365ff2ed0c46f213036ec6) fix: removed scrape annotations
 * [3a7ab5c](https://github.com/argoproj-labs/argo-dataflow/commit/3a7ab5c238c351941fb6cbc7771da7479c80f607) ok

### Contributors

 * Alex Collins

## v0.0.30 (2021-06-03)

 * [c6763e5](https://github.com/argoproj-labs/argo-dataflow/commit/c6763e5b1f7f8112d0e7806b2344e78b15863810) feat(runner): Request Bearer token
 * [9d8676b](https://github.com/argoproj-labs/argo-dataflow/commit/9d8676b5243c70ee48a211f30d1564403fa84fd7) feat(runner): Emit Prometheus metrics (#57)
 * [869e492](https://github.com/argoproj-labs/argo-dataflow/commit/869e4929c38dd5c2e1c68589273b10b1f9cb6a92) ok
 * [af151b6](https://github.com/argoproj-labs/argo-dataflow/commit/af151b66c03495706d10dc9030fd7e305cc36f06) fix: correct STAN durable name

### Contributors

 * Alex Collins

## v0.0.29 (2021-06-02)

 * [6156b57](https://github.com/argoproj-labs/argo-dataflow/commit/6156b57f71d829df6eb8717d4a34de52a49d9511) fix: bug where we were not getting kafka secret

### Contributors

 * Alex Collins

## v0.0.28 (2021-05-27)

 * [57567e3](https://github.com/argoproj-labs/argo-dataflow/commit/57567e3cfae15ebcbd2b3801a2b3db3b3c4b68e9) fix!: change rate to resource.Quantity

### Contributors

 * Alex Collins

## v0.0.27 (2021-05-27)

 * [98eadec](https://github.com/argoproj-labs/argo-dataflow/commit/98eadec373d57799e83377a0e048793ddec53f0a) fix!: change rate to resource.Quantity

### Contributors

 * Alex Collins

## v0.0.26 (2021-05-26)

 * [53c4d04](https://github.com/argoproj-labs/argo-dataflow/commit/53c4d040a61e04488c717d06c3cec2c1729658e7) feat: mark all Kafka messages
 * [dcee8f5](https://github.com/argoproj-labs/argo-dataflow/commit/dcee8f59a2bdf3861c7475d840ec31f05bdc808e) config: remove secrets for stan/kafka

### Contributors

 * Alex Collins

## v0.0.25 (2021-05-25)

 * [980c741](https://github.com/argoproj-labs/argo-dataflow/commit/980c741576235f8eb04fcecd48ef1de3c59e69d4) fix: try and avoid partition changes

### Contributors

 * Alex Collins

## v0.0.24 (2021-05-24)


### Contributors


## v0.0.23 (2021-05-24)

 * [a561027](https://github.com/argoproj-labs/argo-dataflow/commit/a5610275123a0ce82f52ca870749cde9135c238d) fix: fixed bug with pipeline conditions and messages computation
 * [45f51fa](https://github.com/argoproj-labs/argo-dataflow/commit/45f51fa86e36cf3a1240a73e0774a9a0f02e8149) feat: implement an ordered shutdown sequence

### Contributors

 * Alex Collins

## v0.0.22 (2021-05-22)

 * [2ad2f52](https://github.com/argoproj-labs/argo-dataflow/commit/2ad2f52651cb8b4283a6cc9ef71c92c0d7c70c24) feat: allow messages to be return to request
 * [59669ca](https://github.com/argoproj-labs/argo-dataflow/commit/59669cabf52bd479cfae32e2bc7a74c49062a340) fix: correct http POST to use keep-alives

### Contributors

 * Alex Collins

## v0.0.21 (2021-05-21)

 * [22b0b8b](https://github.com/argoproj-labs/argo-dataflow/commit/22b0b8b436ea3bef7e0f2aa17fe223e31e51eb30) fix: fix bouncy scaling bug

### Contributors

 * Alex Collins

## v0.0.20 (2021-05-21)

 * [de8a19f](https://github.com/argoproj-labs/argo-dataflow/commit/de8a19fcb6960af8096ddb3bb944a06c883482b4) fix: enhanced shutdown

### Contributors

 * Alex Collins

## v0.0.19 (2021-05-20)

 * [f3ba148](https://github.com/argoproj-labs/argo-dataflow/commit/f3ba1481e3cfa45e675ea7e875939b0d21346c79) fix: correct metrics over restart
 * [3c0b7fd](https://github.com/argoproj-labs/argo-dataflow/commit/3c0b7fd34e1828104cc8b6c4ecaa71bb21f751d4) config: pump more data thorough Kafka topic
 * [f753b2d](https://github.com/argoproj-labs/argo-dataflow/commit/f753b2df22485ee71d771e91083235e7e42e18ae) feat: update examples
 * [5c6832b](https://github.com/argoproj-labs/argo-dataflow/commit/5c6832bbddb4b20d8601eb7e5145d86ff2bead6a) feat: report rate
 * [b28c3a3](https://github.com/argoproj-labs/argo-dataflow/commit/b28c3a3d2c9ec2a51391e701ba26c00894a9b963) fix: report back errors

### Contributors

 * Alex Collins

## v0.0.18 (2021-05-20)

 * [9d04dcd](https://github.com/argoproj-labs/argo-dataflow/commit/9d04dcde97f04eed1f54852eb7aaf8b6f3b79e49) refactor: change from `for {}` to `wait.JitterUntil`
 * [e17f961](https://github.com/argoproj-labs/argo-dataflow/commit/e17f96180cb4af617303862b2a2ea212d5ddfff3) feat: only check Kafka partition for pending
 * [d3e02ea](https://github.com/argoproj-labs/argo-dataflow/commit/d3e02ead1da9dd5bfb38d7d1a538e9c3961cef4f) feat: reduced update interval to every 30s
 * [72f0aeb](https://github.com/argoproj-labs/argo-dataflow/commit/72f0aeb724d14b331e258059be4ce5a0ff845f86) feat(runner): change name of queue
 * [e25a33c](https://github.com/argoproj-labs/argo-dataflow/commit/e25a33c98adcdbf56bd08172a3ea69285acc6527) feat: report Kubernetes API errors related to pod/servic creation/deletion
 * [36dae38](https://github.com/argoproj-labs/argo-dataflow/commit/36dae3845cc885ae60960ded1deea255d314f6fa) feat: only update pending for replica zero to prevent disagreement

### Contributors

 * Alex Collins

## v0.0.17 (2021-05-19)

 * [e4a4b1a](https://github.com/argoproj-labs/argo-dataflow/commit/e4a4b1a0a523c88190f833c4c06ec7fc0dd4fde2) feat: longer status messages

### Contributors

 * Alex Collins

## v0.0.16 (2021-05-19)


### Contributors


## v0.0.15 (2021-05-19)

 * [9bef0df](https://github.com/argoproj-labs/argo-dataflow/commit/9bef0df2ec1eb398f63c9ba4b99be76b8b08aee8) feat: expose pod failure reason

### Contributors

 * Alex Collins

## v0.0.14 (2021-05-18)

 * [2a23cb8](https://github.com/argoproj-labs/argo-dataflow/commit/2a23cb8abc4343272d4dd04f493d888694923114) fix: scale to 1 rather than 0 on start

### Contributors

 * Alex Collins

## v0.0.13 (2021-05-18)

 * [6204cd1](https://github.com/argoproj-labs/argo-dataflow/commit/6204cd1b085b354647495ec9866b1629fc474eb4) fix: surface errors
 * [c000e41](https://github.com/argoproj-labs/argo-dataflow/commit/c000e4152710e9a1a1d78516e396b3892ed55e7f) fix: make minReplicas required

### Contributors

 * Alex Collins

## v0.0.12 (2021-05-18)

 * [3089119](https://github.com/argoproj-labs/argo-dataflow/commit/3089119f0447727cd493b1116244d90ac6dc3a1e) fix: only terminate sidecars if the main container exit with code 0
 * [6dc28a2](https://github.com/argoproj-labs/argo-dataflow/commit/6dc28a23802b922e9c074dfe93911e9ad4763012) fix: failed to record sink status
 * [3db2fbc](https://github.com/argoproj-labs/argo-dataflow/commit/3db2fbc9645da9c9c2dd682322906f5541ca96d2) fix: add missing RBAC for manager

### Contributors

 * Alex Collins

## v0.0.11 (2021-05-18)

 * [b22e9fd](https://github.com/argoproj-labs/argo-dataflow/commit/b22e9fd986499a6f92bb417e246f2fe8eeaed98c) fix: correct changelog order

### Contributors

 * Alex Collins

## v0.0.10 (2021-05-18)

 * [dd8efe3](https://github.com/argoproj-labs/argo-dataflow/commit/dd8efe31cccc613fd15490c7ab9922fd8f1e3896) feat: and support for stateless sources and sinks with HTTP

### Contributors

 * Alex Collins

## v0.0.9 (2021-05-17)

 * [940632a](https://github.com/argoproj-labs/argo-dataflow/commit/940632a82a7e43ff8eed63cb6a69949dbec85372) feat: and 1st-class support for expand and flatten
 * [9e6ab3d](https://github.com/argoproj-labs/argo-dataflow/commit/9e6ab3dfbd0713f443700bc9997384143a74264b) feat: support layout of cron source messages
 * [290cde4](https://github.com/argoproj-labs/argo-dataflow/commit/290cde40f84371707e9bb46b55fe2ba09791788d) feat: support HPA
 * [d6920cf](https://github.com/argoproj-labs/argo-dataflow/commit/d6920cff0727ca96c80bb225ca5f9bc4f38f62ba) refactor: move replicas to top level

### Contributors

 * Alex Collins

## v0.0.8 (2021-05-14)


### Contributors


## v0.0.7 (2021-05-14)

 * [7e3a574](https://github.com/argoproj-labs/argo-dataflow/commit/7e3a5747cf7c60f1bbf2012ca58736dc3552abaf) config: increase stan-default storage to 16Gi
 * [9166870](https://github.com/argoproj-labs/argo-dataflow/commit/91668702230ef7cf5f3a74989cd7ae1b587f3dc0) config: increase stan-default storage to 16Gi

### Contributors

 * Alex Collins

## v0.0.6 (2021-05-14)

 * [8c2dd33](https://github.com/argoproj-labs/argo-dataflow/commit/8c2dd33e45125ced12a322df952fb7d40dcae066) fix(manager): re-instate killing terminated steps
 * [98917eb](https://github.com/argoproj-labs/argo-dataflow/commit/98917eb248100eb87d098d52d36fbd8707e4da9f) fix: report sink errors

### Contributors

 * Alex Collins

## v0.0.5 (2021-05-13)

 * [1fbfbd1](https://github.com/argoproj-labs/argo-dataflow/commit/1fbfbd1fe690ddf508ca0b6aa8bedda3ed7ad7ba) fix: correct `error` to `lastError`

### Contributors

 * Alex Collins

## v0.0.4 (2021-05-13)

 * [0d0c6d5](https://github.com/argoproj-labs/argo-dataflow/commit/0d0c6d56c2ff5bf7a6bab48fd2d1066885125ced) fix: remove version file
 * [99bfd15](https://github.com/argoproj-labs/argo-dataflow/commit/99bfd151acb810e9b9452f21a76de68d403fc081) refactor: move api/util to ../shared/
 * [e26c3a9](https://github.com/argoproj-labs/argo-dataflow/commit/e26c3a98e5b1bccebd8ca8ea6867ab9576e5927d) refactor: move containerkiller
 * [cc175fc](https://github.com/argoproj-labs/argo-dataflow/commit/cc175fc635100641922918ddf3edbb95e515968d) refactor: move controller files
 * [8d8c2ae](https://github.com/argoproj-labs/argo-dataflow/commit/8d8c2aee08422278e82b722443c02c167219f343) feat: add errors to step/status
 * [9172dc7](https://github.com/argoproj-labs/argo-dataflow/commit/9172dc718dab036bdf573f3b448d04dd966b9aba) fix: only calculate installed hash after changes have been applied

### Contributors

 * Alex Collins

## v0.0.3 (2021-05-12)

 * [f341a7d](https://github.com/argoproj-labs/argo-dataflow/commit/f341a7deec13f5962a2ddd6604a1e0c07c1cd2ac) config: change argo-server to secure
 * [fd75eb3](https://github.com/argoproj-labs/argo-dataflow/commit/fd75eb3763b10ca8de54316bf085b4012c74e99e) config: change argo-server to secure
 * [dcb24d5](https://github.com/argoproj-labs/argo-dataflow/commit/dcb24d537aa05debe2396ef73ce9e676ed09da92) config: change argo-server to secure

### Contributors

 * Alex Collins

## v0.0.2 (2021-05-11)


### Contributors


