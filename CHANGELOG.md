# Changelog

## v0.0.37 (2021-06-06)

 * [267fd6b](https://github.com/argoproj/argo-workflows/commit/267fd6b03450b70629c3d7b43674fb988d863a34) feat: added latency metric
 * [c1c88da](https://github.com/argoproj/argo-workflows/commit/c1c88da7ff9a4448f08541cdab6ddabfe61b8d88) fix: in_flight -> inflight

### Contributors

 * Alex Collins

## v0.0.36 (2021-06-05)

 * [ca7edf1](https://github.com/argoproj/argo-workflows/commit/ca7edf16c34075efb77d0cc1e76c40b943ab75d5) feat: add in_flight metric
 * [009731a](https://github.com/argoproj/argo-workflows/commit/009731a0bab298f1d50a6a0f64278ddfba7ca2a5) refactor: made code more testable
 * [a3142eb](https://github.com/argoproj/argo-workflows/commit/a3142eb7b5755a064acf47066a8c99604bfa73d9) fix: add bearer token

### Contributors

 * Alex Collins

## v0.0.35 (2021-06-04)

 * [421c142](https://github.com/argoproj/argo-workflows/commit/421c14269c618eb1a1c601715e6ff361d45d9921) fix: total/error calcs
 * [06fc5f0](https://github.com/argoproj/argo-workflows/commit/06fc5f0592ce74982864f133af5dd405948451f0) fix: only return pending for replica=0

### Contributors

 * Alex Collins

## v0.0.34 (2021-06-04)

 * [c73ce5b](https://github.com/argoproj/argo-workflows/commit/c73ce5b33a6bfc42abf772d6f794c6af885499a0) fix: change from patch to update for the controller
 * [4a32fb9](https://github.com/argoproj/argo-workflows/commit/4a32fb9fcd72181bb6299d7d123166e29369a8e7) fix: kafka pending messages

### Contributors

 * Alex Collins
 * Derek Wang

## v0.0.33 (2021-06-04)

 * [60eab31](https://github.com/argoproj/argo-workflows/commit/60eab313fe856ee34b44d8f7cfcb4057d01c6344) fix: prevent violent scale-up and scale-down by only scaling by 1 each time

### Contributors

 * Alex Collins

## v0.0.32 (2021-06-04)

 * [ca363fe](https://github.com/argoproj/argo-workflows/commit/ca363fe3638fa6c329dc786dedb4aaf8d230a8f3) feat: add pending metric
 * [f3a4813](https://github.com/argoproj/argo-workflows/commit/f3a4813aaf23d630f4d8292792d6b147d53515f0) fix: fix metrics

### Contributors

 * Alex Collins

## v0.0.31 (2021-06-04)

 * [3dd39f1](https://github.com/argoproj/argo-workflows/commit/3dd39f16bf6d1104af3d2a3e4c23b98c22170639) fix(sidecar): updated to use fixed counters
 * [6bef420](https://github.com/argoproj/argo-workflows/commit/6bef420f18718ca977365ff2ed0c46f213036ec6) fix: removed scrape annotations
 * [3a7ab5c](https://github.com/argoproj/argo-workflows/commit/3a7ab5c238c351941fb6cbc7771da7479c80f607) ok

### Contributors

 * Alex Collins

## v0.0.30 (2021-06-03)

 * [c6763e5](https://github.com/argoproj/argo-workflows/commit/c6763e5b1f7f8112d0e7806b2344e78b15863810) feat(runner): Request Bearer token
 * [9d8676b](https://github.com/argoproj/argo-workflows/commit/9d8676b5243c70ee48a211f30d1564403fa84fd7) feat(runner): Emit Prometheus metrics (#57)
 * [869e492](https://github.com/argoproj/argo-workflows/commit/869e4929c38dd5c2e1c68589273b10b1f9cb6a92) ok
 * [af151b6](https://github.com/argoproj/argo-workflows/commit/af151b66c03495706d10dc9030fd7e305cc36f06) fix: correct STAN durable name

### Contributors

 * Alex Collins

## v0.0.29 (2021-06-02)

 * [6156b57](https://github.com/argoproj/argo-workflows/commit/6156b57f71d829df6eb8717d4a34de52a49d9511) fix: bug where we were not getting kafka secret

### Contributors

 * Alex Collins

## v0.0.28 (2021-05-27)

 * [57567e3](https://github.com/argoproj/argo-workflows/commit/57567e3cfae15ebcbd2b3801a2b3db3b3c4b68e9) fix!: change rate to resource.Quantity

### Contributors

 * Alex Collins

## v0.0.27 (2021-05-27)

 * [98eadec](https://github.com/argoproj/argo-workflows/commit/98eadec373d57799e83377a0e048793ddec53f0a) fix!: change rate to resource.Quantity

### Contributors

 * Alex Collins

## v0.0.26 (2021-05-26)

 * [53c4d04](https://github.com/argoproj/argo-workflows/commit/53c4d040a61e04488c717d06c3cec2c1729658e7) feat: mark all Kafka messages
 * [dcee8f5](https://github.com/argoproj/argo-workflows/commit/dcee8f59a2bdf3861c7475d840ec31f05bdc808e) config: remove secrets for stan/kafka

### Contributors

 * Alex Collins

## v0.0.25 (2021-05-25)

 * [980c741](https://github.com/argoproj/argo-workflows/commit/980c741576235f8eb04fcecd48ef1de3c59e69d4) fix: try and avoid partition changes

### Contributors

 * Alex Collins

## v0.0.24 (2021-05-24)


### Contributors


## v0.0.23 (2021-05-24)

 * [a561027](https://github.com/argoproj/argo-workflows/commit/a5610275123a0ce82f52ca870749cde9135c238d) fix: fixed bug with pipeline conditions and messages computation
 * [45f51fa](https://github.com/argoproj/argo-workflows/commit/45f51fa86e36cf3a1240a73e0774a9a0f02e8149) feat: implement an ordered shutdown sequence

### Contributors

 * Alex Collins

## v0.0.22 (2021-05-22)

 * [2ad2f52](https://github.com/argoproj/argo-workflows/commit/2ad2f52651cb8b4283a6cc9ef71c92c0d7c70c24) feat: allow messages to be return to request
 * [59669ca](https://github.com/argoproj/argo-workflows/commit/59669cabf52bd479cfae32e2bc7a74c49062a340) fix: correct http POST to use keep-alives

### Contributors

 * Alex Collins

## v0.0.21 (2021-05-21)

 * [22b0b8b](https://github.com/argoproj/argo-workflows/commit/22b0b8b436ea3bef7e0f2aa17fe223e31e51eb30) fix: fix bouncy scaling bug

### Contributors

 * Alex Collins

## v0.0.20 (2021-05-21)

 * [de8a19f](https://github.com/argoproj/argo-workflows/commit/de8a19fcb6960af8096ddb3bb944a06c883482b4) fix: enhanced shutdown

### Contributors

 * Alex Collins

## v0.0.19 (2021-05-20)

 * [f3ba148](https://github.com/argoproj/argo-workflows/commit/f3ba1481e3cfa45e675ea7e875939b0d21346c79) fix: correct metrics over restart
 * [3c0b7fd](https://github.com/argoproj/argo-workflows/commit/3c0b7fd34e1828104cc8b6c4ecaa71bb21f751d4) config: pump more data thorough Kafka topic
 * [f753b2d](https://github.com/argoproj/argo-workflows/commit/f753b2df22485ee71d771e91083235e7e42e18ae) feat: update examples
 * [5c6832b](https://github.com/argoproj/argo-workflows/commit/5c6832bbddb4b20d8601eb7e5145d86ff2bead6a) feat: report rate
 * [b28c3a3](https://github.com/argoproj/argo-workflows/commit/b28c3a3d2c9ec2a51391e701ba26c00894a9b963) fix: report back errors

### Contributors

 * Alex Collins

## v0.0.18 (2021-05-20)

 * [9d04dcd](https://github.com/argoproj/argo-workflows/commit/9d04dcde97f04eed1f54852eb7aaf8b6f3b79e49) refactor: change from `for {}` to `wait.JitterUntil`
 * [e17f961](https://github.com/argoproj/argo-workflows/commit/e17f96180cb4af617303862b2a2ea212d5ddfff3) feat: only check Kafka partition for pending
 * [d3e02ea](https://github.com/argoproj/argo-workflows/commit/d3e02ead1da9dd5bfb38d7d1a538e9c3961cef4f) feat: reduced update interval to every 30s
 * [72f0aeb](https://github.com/argoproj/argo-workflows/commit/72f0aeb724d14b331e258059be4ce5a0ff845f86) feat(runner): change name of queue
 * [e25a33c](https://github.com/argoproj/argo-workflows/commit/e25a33c98adcdbf56bd08172a3ea69285acc6527) feat: report Kubernetes API errors related to pod/servic creation/deletion
 * [36dae38](https://github.com/argoproj/argo-workflows/commit/36dae3845cc885ae60960ded1deea255d314f6fa) feat: only update pending for replica zero to prevent disagreement

### Contributors

 * Alex Collins

## v0.0.17 (2021-05-19)

 * [e4a4b1a](https://github.com/argoproj/argo-workflows/commit/e4a4b1a0a523c88190f833c4c06ec7fc0dd4fde2) feat: longer status messages

### Contributors

 * Alex Collins

## v0.0.16 (2021-05-19)


### Contributors


## v0.0.15 (2021-05-19)

 * [9bef0df](https://github.com/argoproj/argo-workflows/commit/9bef0df2ec1eb398f63c9ba4b99be76b8b08aee8) feat: expose pod failure reason

### Contributors

 * Alex Collins

## v0.0.14 (2021-05-18)

 * [2a23cb8](https://github.com/argoproj/argo-workflows/commit/2a23cb8abc4343272d4dd04f493d888694923114) fix: scale to 1 rather than 0 on start

### Contributors

 * Alex Collins

## v0.0.13 (2021-05-18)

 * [6204cd1](https://github.com/argoproj/argo-workflows/commit/6204cd1b085b354647495ec9866b1629fc474eb4) fix: surface errors
 * [c000e41](https://github.com/argoproj/argo-workflows/commit/c000e4152710e9a1a1d78516e396b3892ed55e7f) fix: make minReplicas required

### Contributors

 * Alex Collins

## v0.0.12 (2021-05-18)

 * [3089119](https://github.com/argoproj/argo-workflows/commit/3089119f0447727cd493b1116244d90ac6dc3a1e) fix: only terminate sidecars if the main container exit with code 0
 * [6dc28a2](https://github.com/argoproj/argo-workflows/commit/6dc28a23802b922e9c074dfe93911e9ad4763012) fix: failed to record sink status
 * [3db2fbc](https://github.com/argoproj/argo-workflows/commit/3db2fbc9645da9c9c2dd682322906f5541ca96d2) fix: add missing RBAC for manager

### Contributors

 * Alex Collins

## v0.0.11 (2021-05-18)

 * [b22e9fd](https://github.com/argoproj/argo-workflows/commit/b22e9fd986499a6f92bb417e246f2fe8eeaed98c) fix: correct changelog order

### Contributors

 * Alex Collins

## v0.0.10 (2021-05-18)

 * [dd8efe3](https://github.com/argoproj/argo-workflows/commit/dd8efe31cccc613fd15490c7ab9922fd8f1e3896) feat: and support for stateless sources and sinks with HTTP

### Contributors

 * Alex Collins

## v0.0.9 (2021-05-17)

 * [940632a](https://github.com/argoproj/argo-workflows/commit/940632a82a7e43ff8eed63cb6a69949dbec85372) feat: and 1st-class support for expand and flatten
 * [9e6ab3d](https://github.com/argoproj/argo-workflows/commit/9e6ab3dfbd0713f443700bc9997384143a74264b) feat: support layout of cron source messages
 * [290cde4](https://github.com/argoproj/argo-workflows/commit/290cde40f84371707e9bb46b55fe2ba09791788d) feat: support HPA
 * [d6920cf](https://github.com/argoproj/argo-workflows/commit/d6920cff0727ca96c80bb225ca5f9bc4f38f62ba) refactor: move replicas to top level

### Contributors

 * Alex Collins

## v0.0.8 (2021-05-14)


### Contributors


## v0.0.7 (2021-05-14)

 * [7e3a574](https://github.com/argoproj/argo-workflows/commit/7e3a5747cf7c60f1bbf2012ca58736dc3552abaf) config: increase stan-default storage to 16Gi
 * [9166870](https://github.com/argoproj/argo-workflows/commit/91668702230ef7cf5f3a74989cd7ae1b587f3dc0) config: increase stan-default storage to 16Gi

### Contributors

 * Alex Collins

## v0.0.6 (2021-05-14)

 * [8c2dd33](https://github.com/argoproj/argo-workflows/commit/8c2dd33e45125ced12a322df952fb7d40dcae066) fix(manager): re-instate killing terminated steps
 * [98917eb](https://github.com/argoproj/argo-workflows/commit/98917eb248100eb87d098d52d36fbd8707e4da9f) fix: report sink errors

### Contributors

 * Alex Collins

## v0.0.5 (2021-05-13)

 * [1fbfbd1](https://github.com/argoproj/argo-workflows/commit/1fbfbd1fe690ddf508ca0b6aa8bedda3ed7ad7ba) fix: correct `error` to `lastError`

### Contributors

 * Alex Collins

## v0.0.4 (2021-05-13)

 * [0d0c6d5](https://github.com/argoproj/argo-workflows/commit/0d0c6d56c2ff5bf7a6bab48fd2d1066885125ced) fix: remove version file
 * [99bfd15](https://github.com/argoproj/argo-workflows/commit/99bfd151acb810e9b9452f21a76de68d403fc081) refactor: move api/util to ../shared/
 * [e26c3a9](https://github.com/argoproj/argo-workflows/commit/e26c3a98e5b1bccebd8ca8ea6867ab9576e5927d) refactor: move containerkiller
 * [cc175fc](https://github.com/argoproj/argo-workflows/commit/cc175fc635100641922918ddf3edbb95e515968d) refactor: move controller files
 * [8d8c2ae](https://github.com/argoproj/argo-workflows/commit/8d8c2aee08422278e82b722443c02c167219f343) feat: add errors to step/status
 * [9172dc7](https://github.com/argoproj/argo-workflows/commit/9172dc718dab036bdf573f3b448d04dd966b9aba) fix: only calculate installed hash after changes have been applied

### Contributors

 * Alex Collins

## v0.0.3 (2021-05-12)

 * [f341a7d](https://github.com/argoproj/argo-workflows/commit/f341a7deec13f5962a2ddd6604a1e0c07c1cd2ac) config: change argo-server to secure
 * [fd75eb3](https://github.com/argoproj/argo-workflows/commit/fd75eb3763b10ca8de54316bf085b4012c74e99e) config: change argo-server to secure
 * [dcb24d5](https://github.com/argoproj/argo-workflows/commit/dcb24d537aa05debe2396ef73ce9e676ed09da92) config: change argo-server to secure

### Contributors

 * Alex Collins

## v0.0.2 (2021-05-11)


### Contributors


