# Changelog

## v0.0.12 (2021-05-18)

 * [3089119](https://github.com/argoproj/argo-workflows/commit/3089119f0447727cd493b1116244d90ac6dc3a1e) fix: only terminate sidecars if the main container exit with code 0
 * [6dc28a2](https://github.com/argoproj/argo-workflows/commit/6dc28a23802b922e9c074dfe93911e9ad4763012) fix: failed to record sink status
 * [3db2fbc](https://github.com/argoproj/argo-workflows/commit/3db2fbc9645da9c9c2dd682322906f5541ca96d2) fix: add missing RBAC for manager
 * [97c9d4b](https://github.com/argoproj/argo-workflows/commit/97c9d4b402a29ab978c536c0b2e5231c183cdd20) docs: updated CHANGELOG.md

### Contributors

 * Alex Collins

## v0.0.11 (2021-05-18)

 * [b22e9fd](https://github.com/argoproj/argo-workflows/commit/b22e9fd986499a6f92bb417e246f2fe8eeaed98c) fix: correct changelog order
 * [b487a16](https://github.com/argoproj/argo-workflows/commit/b487a1633b005c1084e79282665e1a7340c15580) chore: make HTTP sinks unique

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
 * [bb32570](https://github.com/argoproj/argo-workflows/commit/bb32570b14a9756713eee63bf3f5a46f4055e901) docs: add CHANGELOG.md
 * [d6920cf](https://github.com/argoproj/argo-workflows/commit/d6920cff0727ca96c80bb225ca5f9bc4f38f62ba) refactor: move replicas to top level

### Contributors

 * Alex Collins

## v0.0.8 (2021-05-14)

 * [1e57f5c](https://github.com/argoproj/argo-workflows/commit/1e57f5c59a9e7cf0aec3526a58b5add4c8cd5018) chore: break dependency on util
 * [8ca5215](https://github.com/argoproj/argo-workflows/commit/8ca52152226a0e231aeb3bde875750ae022970ce) chore: make lint

### Contributors

 * Alex Collins

## v0.0.7 (2021-05-14)

 * [59a5ac6](https://github.com/argoproj/argo-workflows/commit/59a5ac621c24162f1838764cb4985d7d94af63da) chore: remove embed so we can work with Golang v1.15.7
 * [7e3a574](https://github.com/argoproj/argo-workflows/commit/7e3a5747cf7c60f1bbf2012ca58736dc3552abaf) config: increase stan-default storage to 16Gi
 * [9166870](https://github.com/argoproj/argo-workflows/commit/91668702230ef7cf5f3a74989cd7ae1b587f3dc0) config: increase stan-default storage to 16Gi
 * [16e7687](https://github.com/argoproj/argo-workflows/commit/16e76875ff75ac460a35d6ad4ac52e5524b2580e) chore: remove unused variables

### Contributors

 * Alex Collins

## v0.0.6 (2021-05-14)

 * [2a44fcb](https://github.com/argoproj/argo-workflows/commit/2a44fcb302d5ccbffb56549ae0d39c6291e979e3) chore: Remove installer code
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
 * [940e449](https://github.com/argoproj/argo-workflows/commit/940e449cff9858ee8902c81ba7b639ab3d3c5008) chore: lint
 * [99bfd15](https://github.com/argoproj/argo-workflows/commit/99bfd151acb810e9b9452f21a76de68d403fc081) refactor: move api/util to ../shared/
 * [e26c3a9](https://github.com/argoproj/argo-workflows/commit/e26c3a98e5b1bccebd8ca8ea6867ab9576e5927d) refactor: move containerkiller
 * [cc175fc](https://github.com/argoproj/argo-workflows/commit/cc175fc635100641922918ddf3edbb95e515968d) refactor: move controller files
 * [8d8c2ae](https://github.com/argoproj/argo-workflows/commit/8d8c2aee08422278e82b722443c02c167219f343) feat: add errors to step/status
 * [9172dc7](https://github.com/argoproj/argo-workflows/commit/9172dc718dab036bdf573f3b448d04dd966b9aba) fix: only calculate installed hash after changes have been applied
 * [0c055d2](https://github.com/argoproj/argo-workflows/commit/0c055d2d9878e8905788bd85c30470ab5611c77b) test: add `imageName` test

### Contributors

 * Alex Collins

## v0.0.3 (2021-05-12)

 * [1bf0b1f](https://github.com/argoproj/argo-workflows/commit/1bf0b1fa8cc45c7d1995b609530babe6a1e61d0a) ci: fetch-depth=0
 * [f341a7d](https://github.com/argoproj/argo-workflows/commit/f341a7deec13f5962a2ddd6604a1e0c07c1cd2ac) config: change argo-server to secure
 * [fd75eb3](https://github.com/argoproj/argo-workflows/commit/fd75eb3763b10ca8de54316bf085b4012c74e99e) config: change argo-server to secure
 * [dcb24d5](https://github.com/argoproj/argo-workflows/commit/dcb24d537aa05debe2396ef73ce9e676ed09da92) config: change argo-server to secure
 * [126dfc9](https://github.com/argoproj/argo-workflows/commit/126dfc9731efbe5e7585e43af26e77bbf35df512) ci: harmonize release with workflows

### Contributors

 * Alex Collins

## v0.0.2 (2021-05-11)

 * [8eaf82d](https://github.com/argoproj/argo-workflows/commit/8eaf82d979563058389950314919799bc902086d) ci: only attach default.yaml and quick-start.yaml
 * [0935fe1](https://github.com/argoproj/argo-workflows/commit/0935fe15b4835c17027387ae997b778c37308553) ci: only push images on main or `v`
 * [7251471](https://github.com/argoproj/argo-workflows/commit/7251471cee6f69501eb80a456fb8e316fa26cdb9) ci: only publish v releases
 * [3dd3708](https://github.com/argoproj/argo-workflows/commit/3dd3708a89734575859cbbd23d36baeb08f12ae2) test: add exclusion
 * [418ec34](https://github.com/argoproj/argo-workflows/commit/418ec34f3a1f11d893c7b917258c5af53b2ba49e) docs: Update CI badge in README.md

### Contributors

 * Alex Collins

