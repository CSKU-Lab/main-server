# [0.14.0](https://github.com/CSKU-Lab/main-server/compare/v0.13.0...v0.14.0) (2026-05-24)


### Features

* add og resync internal endpoint ([b6efdae](https://github.com/CSKU-Lab/main-server/commit/b6efdaeae19219cd97d995e5c10bcb8dbb42be0f))

# [0.13.0](https://github.com/CSKU-Lab/main-server/compare/v0.12.2...v0.13.0) (2026-05-24)


### Features

* add GET /api/v1/lsp/token endpoint ([9e8aaa3](https://github.com/CSKU-Lab/main-server/commit/9e8aaa305f1f92f61222d6e372366c805f52307a))

## [0.12.2](https://github.com/CSKU-Lab/main-server/compare/v0.12.1...v0.12.2) (2026-05-24)


### Bug Fixes

* reconnect postgres pubsub and bump queue to v0.3.2 ([165122e](https://github.com/CSKU-Lab/main-server/commit/165122e2a684052e4da66ab1df747e63aec9dc03))
* use dedicated og service ([b9753ae](https://github.com/CSKU-Lab/main-server/commit/b9753aede43554d3148c7ce7f2a5822f8fbee448))

## [0.12.1](https://github.com/CSKU-Lab/main-server/compare/v0.12.0...v0.12.1) (2026-05-23)


### Bug Fixes

* change typo from Limit -> Limits ([114176d](https://github.com/CSKU-Lab/main-server/commit/114176d4fb427bafca2a536d0f46099030f269d7))

# [0.12.0](https://github.com/CSKU-Lab/main-server/compare/v0.11.0...v0.12.0) (2026-05-23)


### Features

* implement rate limit middleware ([78bf470](https://github.com/CSKU-Lab/main-server/commit/78bf4705243e2c0376868d5b03b961c1eb48f736))

# [0.11.0](https://github.com/CSKU-Lab/main-server/compare/v0.10.0...v0.11.0) (2026-05-23)


### Bug Fixes

* empty submission gradebook return null ([992a9d1](https://github.com/CSKU-Lab/main-server/commit/992a9d17b3a73d9d72a585a492f736a5715b833e))


### Features

* add section lab affected entities ([71280b6](https://github.com/CSKU-Lab/main-server/commit/71280b6fa0cce3ed977b219608b17b646dbd2966))
* **core:** add search route for command palette ([e5a5da8](https://github.com/CSKU-Lab/main-server/commit/e5a5da8ed305237002e25dff987d6da11957530d))

# [0.10.0](https://github.com/CSKU-Lab/main-server/compare/v0.9.1...v0.10.0) (2026-05-23)


### Bug Fixes

* enforce user validation on create ([1d32147](https://github.com/CSKU-Lab/main-server/commit/1d32147e0854a04cb7649a0c58ac8d2e44a41499))
* response error back when add non student user to the section ([b7f8b7c](https://github.com/CSKU-Lab/main-server/commit/b7f8b7c3a940d4008d374509540ccffce0e65e82))


### Features

* add section lab search ([42c7377](https://github.com/CSKU-Lab/main-server/commit/42c7377ee0be6a7b12c87a53a817a25da61dab02))
* improve section lab status to able to force status ([fa77188](https://github.com/CSKU-Lab/main-server/commit/fa771882501992e6feeee1bffc0d8b03d41f67ea))

## [0.9.1](https://github.com/CSKU-Lab/main-server/compare/v0.9.0...v0.9.1) (2026-05-23)


### Bug Fixes

* import user ([6ca1695](https://github.com/CSKU-Lab/main-server/commit/6ca1695e98ed4d28011f645952e928a2615089cc))

# [0.9.0](https://github.com/CSKU-Lab/main-server/compare/v0.8.0...v0.9.0) (2026-05-22)


### Features

* **search:** implement fuzzy search for cms route ([6a2240e](https://github.com/CSKU-Lab/main-server/commit/6a2240ecc9d4bb46fb2f62f6ddfefcfe6e84b9fe))

# [0.8.0](https://github.com/CSKU-Lab/main-server/compare/v0.7.0...v0.8.0) (2026-05-22)


### Features

* **material:** add a new document material type ([508c78d](https://github.com/CSKU-Lab/main-server/commit/508c78dfd205dcf106b7f64d6709e8ee1f196196))

# [0.7.0](https://github.com/CSKU-Lab/main-server/compare/v0.6.1...v0.7.0) (2026-05-21)


### Features

* **course:** update course lab materials logic follow frontend ([da2ec28](https://github.com/CSKU-Lab/main-server/commit/da2ec28df56fee0fa87b1c5d8df79babdf908a85))

## [0.6.1](https://github.com/CSKU-Lab/main-server/compare/v0.6.0...v0.6.1) (2026-05-20)


### Bug Fixes

* reconnect RabbitMQ on channel/connection closed error in submission worker ([541f9c3](https://github.com/CSKU-Lab/main-server/commit/541f9c32caf23eb40d82976b1cd1796685483977))

# [0.6.0](https://github.com/CSKU-Lab/main-server/compare/v0.5.0...v0.6.0) (2026-05-19)


### Features

* update material domain and fix atlas schema ([7194e25](https://github.com/CSKU-Lab/main-server/commit/7194e256d881b60881d26bd72843efb0ac6383e4))
