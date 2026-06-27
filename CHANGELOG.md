## [0.28.2](https://github.com/CSKU-Lab/main-server/compare/v0.28.1...v0.28.2) (2026-06-27)


### Bug Fixes

* **document-material:** derive status and per-embed scores from embedded code submissions ([4aefdb3](https://github.com/CSKU-Lab/main-server/commit/4aefdb3d68d121b422c54658c22b2853cf7a37d5))

## [0.28.1](https://github.com/CSKU-Lab/main-server/compare/v0.28.0...v0.28.1) (2026-06-27)


### Bug Fixes

* **storage:** remove HTTP proxy handler, MinIO now served directly via ingress ([c964d59](https://github.com/CSKU-Lab/main-server/commit/c964d59677c9239900d56b7b42ccf844cce83f73))

# [0.28.0](https://github.com/CSKU-Lab/main-server/compare/v0.27.4...v0.28.0) (2026-06-27)


### Features

* **submission:** compute document material scores from embedded code submissions ([6d6537c](https://github.com/CSKU-Lab/main-server/commit/6d6537c274e814f555faaeed8537bdc9e279c3e2))

## [0.27.4](https://github.com/CSKU-Lab/main-server/compare/v0.27.3...v0.27.4) (2026-06-27)


### Bug Fixes

* **storage:** add public /storage proxy route ([fb01d46](https://github.com/CSKU-Lab/main-server/commit/fb01d4622736a377edb44063509f757f7fb229cb))

## [0.27.3](https://github.com/CSKU-Lab/main-server/compare/v0.27.2...v0.27.3) (2026-06-27)


### Bug Fixes

* **ip:** read real client IP from X-Forwarded-For behind Traefik ([b187567](https://github.com/CSKU-Lab/main-server/commit/b187567ad56d3d6b6ee5c39054bfbc8684a7065f))

## [0.27.2](https://github.com/CSKU-Lab/main-server/compare/v0.27.1...v0.27.2) (2026-06-27)


### Bug Fixes

* **submission:** close subscribe-after-publish race in listen SSE ([9c8d212](https://github.com/CSKU-Lab/main-server/commit/9c8d2125438757b9c46ff45177863f8c0928e062))

## [0.27.1](https://github.com/CSKU-Lab/main-server/compare/v0.27.0...v0.27.1) (2026-06-27)


### Bug Fixes

* **submission:** stop panic in typing CMS submissions view ([e9a6d64](https://github.com/CSKU-Lab/main-server/commit/e9a6d6441f8f36f79c056b17944283bd7df19c4b))

# [0.27.0](https://github.com/CSKU-Lab/main-server/compare/v0.26.1...v0.27.0) (2026-06-27)


### Features

* **material:** per-field testcase visibility (hide input/output) (CS-233) ([614147e](https://github.com/CSKU-Lab/main-server/commit/614147e2bd257ac2ed11bdbfe1e729bc2ea8d388))

## [0.26.1](https://github.com/CSKU-Lab/main-server/compare/v0.26.0...v0.26.1) (2026-06-27)


### Bug Fixes

* **material:** hide grader-only segment data from students (CS-232) ([33a52cb](https://github.com/CSKU-Lab/main-server/commit/33a52cb8bc10b4267bf9d5822cec752bdac48a6b))


### Performance Improvements

* **analytics:** index created_at on submissions and auth_logs ([9995ebb](https://github.com/CSKU-Lab/main-server/commit/9995ebb85305212809480d1d45bcf15a4ed5b2b1))

# [0.26.0](https://github.com/CSKU-Lab/main-server/compare/v0.25.0...v0.26.0) (2026-06-27)


### Features

* **analytics:** admin-only analytics overview endpoint ([39f2696](https://github.com/CSKU-Lab/main-server/commit/39f2696a546629c21e32f3cb6c2a2ad01ff56a7c))

# [0.25.0](https://github.com/CSKU-Lab/main-server/compare/v0.24.0...v0.25.0) (2026-06-26)


### Features

* **material:** add clone endpoint for same-course duplication ([0119ace](https://github.com/CSKU-Lab/main-server/commit/0119ace9a9a12e16387ddac057fc5819fbccd0c6))

# [0.24.0](https://github.com/CSKU-Lab/main-server/compare/v0.23.2...v0.24.0) (2026-06-25)


### Features

* **submissions:** hide hidden segments from student submission view ([cf3bba9](https://github.com/CSKU-Lab/main-server/commit/cf3bba9d01e285f82983d846e3400b1c957ecbd9))

## [0.23.2](https://github.com/CSKU-Lab/main-server/compare/v0.23.1...v0.23.2) (2026-06-25)


### Bug Fixes

* **lab-material:** renumber positions atomically on reorder ([f3ab6eb](https://github.com/CSKU-Lab/main-server/commit/f3ab6eb6bc05483b579b3f4ef56af6a06efe96fc))

## [0.23.1](https://github.com/CSKU-Lab/main-server/compare/v0.23.0...v0.23.1) (2026-06-25)


### Bug Fixes

* **regrade:** use detached context for regrade goroutines ([cc4f60a](https://github.com/CSKU-Lab/main-server/commit/cc4f60a7dad440e4a66de900c9ca14d4e799c5c4))

# [0.23.0](https://github.com/CSKU-Lab/main-server/compare/v0.22.4...v0.23.0) (2026-06-24)


### Features

* show best typing submissions in CMS material submissions view ([8059cf4](https://github.com/CSKU-Lab/main-server/commit/8059cf4bb96987bac2270fa84337bcd6cb000a45))

## [0.22.4](https://github.com/CSKU-Lab/main-server/compare/v0.22.3...v0.22.4) (2026-06-24)


### Bug Fixes

* evaluate typing submission status on pass criteria for practice and exam modes ([be7b0ee](https://github.com/CSKU-Lab/main-server/commit/be7b0eed5382222cc639ce86f6b2c8ab56e2f9cb))

# [0.22.0](https://github.com/CSKU-Lab/main-server/compare/v0.21.0...v0.22.0) (2026-06-23)


### Bug Fixes

* **CS-189:** expose auto_score in submission responses for typing materials ([50d2854](https://github.com/CSKU-Lab/main-server/commit/50d2854a1f665c80351f6e8f92c5d57ca3be22ca))
* **CS-214:** register legacy "type" alias and cleanup orphan materials ([92f4d9f](https://github.com/CSKU-Lab/main-server/commit/92f4d9f2427dd07585f5a724141a9a9a1748a42b))
* **CS-215:** register DocumentSubmission handler in submission registry ([e4ad202](https://github.com/CSKU-Lab/main-server/commit/e4ad202189fa9b7ae5781aafafbdedd8173bb299))
* **CS-216:** double core API rate limit to reduce false 429s on material switch ([622e75f](https://github.com/CSKU-Lab/main-server/commit/622e75f6141b05981dbc7d746f6b85a028aee14b))
* merge CS-214 and CS-215 registry changes ([fd9ae96](https://github.com/CSKU-Lab/main-server/commit/fd9ae96158935bbaa2f399d4b5216e430ce44cc2))


### Features

* **CS-189:** replace typing thresholds with practice/exam mode and evaluate formula ([bf41072](https://github.com/CSKU-Lab/main-server/commit/bf41072cc298c357dc7f4f738dcb707b3134a0c5))
* **CS-191:** add section-scoped typing submissions export endpoint ([35d8c9f](https://github.com/CSKU-Lab/main-server/commit/35d8c9fddc024dd620288f8a0f849aa7f4d71f94))
* **CS-210:** score propagation for document materials with embedded code problems ([306dc63](https://github.com/CSKU-Lab/main-server/commit/306dc63989cb446058266624857e108a435d333b))
* **CS-211:** add segment support to code material and submission assembly ([37bb5c3](https://github.com/CSKU-Lab/main-server/commit/37bb5c3270bee369e5b73c5bda52eab44e6e6285))
* **CS-211:** regenerate genproto with Segment message in File ([ae47032](https://github.com/CSKU-Lab/main-server/commit/ae4703281818ff6f6cdcad58a18ff8dee0356b4a))
* expose section name and semester in my courses response ([53b0e40](https://github.com/CSKU-Lab/main-server/commit/53b0e40f17824b4b9cc711fbd302a6ebfe66e2cc))

# [0.21.0](https://github.com/CSKU-Lab/main-server/compare/v0.20.1...v0.21.0) (2026-06-21)


### Features

* add multi-provider auth support per user ([dfa3f8a](https://github.com/CSKU-Lab/main-server/commit/dfa3f8a62d6f3c17078d89a3b95f6412c1877d85))

## [0.20.1](https://github.com/CSKU-Lab/main-server/compare/v0.20.0...v0.20.1) (2026-06-21)


### Bug Fixes

* prioritize admin role over instructor in permission checks ([a785649](https://github.com/CSKU-Lab/main-server/commit/a785649ac1cd787c8b4a09fd09b670b21560529f))

# [0.20.0](https://github.com/CSKU-Lab/main-server/compare/v0.19.0...v0.20.0) (2026-06-21)


### Bug Fixes

* allow course creators to create sections ([b244443](https://github.com/CSKU-Lab/main-server/commit/b2444438956c39274c2f15c01ac58f9989281998))
* **CS-206:** add backend tag search and fix material type filter validation ([a1d1a2d](https://github.com/CSKU-Lab/main-server/commit/a1d1a2d033dfbc17535e37e4af6c116c5323ac43))
* **CS-208:** fix section lab mutation permission and status transition ([7c9bbe5](https://github.com/CSKU-Lab/main-server/commit/7c9bbe5317fe1691b62d46b92fd9ff94dda63596))
* support array fields in filter builder for roles filter ([80ad4e5](https://github.com/CSKU-Lab/main-server/commit/80ad4e528995463449b4f5036b752be765209f49))
* use section banner instead of course banner for private sections in my-courses API ([3e9c674](https://github.com/CSKU-Lab/main-server/commit/3e9c67486fd819017d071d4700906fc50e268d3b))


### Features

* **CS-208:** enforce instructor role restrictions on CMS API ([2317da6](https://github.com/CSKU-Lab/main-server/commit/2317da6cb204f6e3b7dfaf523dcb10240809e6d8))
* strip score and auto_score from core submission routes ([e757811](https://github.com/CSKU-Lab/main-server/commit/e7578117d2c36807088bca27fd3fb80d65c30720))

# [0.19.0](https://github.com/CSKU-Lab/main-server/compare/v0.18.0...v0.19.0) (2026-06-20)


### Features

* **CS-197:** add system settings with default compare script fallback for grading ([96ff0b4](https://github.com/CSKU-Lab/main-server/commit/96ff0b4079bfbf9fd62d9e9b4fe02ee3fc135aa5))
* **CS-197:** set default compare script on code material creation ([2da4c68](https://github.com/CSKU-Lab/main-server/commit/2da4c6811cb0a36e2f9dd3adcf037d95e0aeeb78))

# [0.18.0](https://github.com/CSKU-Lab/main-server/compare/v0.17.0...v0.18.0) (2026-06-20)


### Features

* **CS-199:** add instructor delete submission endpoint ([21d072c](https://github.com/CSKU-Lab/main-server/commit/21d072ca8231c9650d7135809cbef30d4d187ba5))

# [0.17.0](https://github.com/CSKU-Lab/main-server/compare/v0.16.1...v0.17.0) (2026-06-20)


### Bug Fixes

* allow mistyped chars in typing submission and fix adjWPM ([53b0f56](https://github.com/CSKU-Lab/main-server/commit/53b0f56fcdda3ee7e266ea4f9016eb04f1a6eeec))


### Features

* **cms:** add Regrade All endpoint for code materials ([6b00d4f](https://github.com/CSKU-Lab/main-server/commit/6b00d4fcd1f7761bbda0ec653985c2702a9eadc0))

## [0.16.1](https://github.com/CSKU-Lab/main-server/compare/v0.16.0...v0.16.1) (2026-06-20)


### Bug Fixes

* use correct material ID for student status lookup and sort by position ([05c8656](https://github.com/CSKU-Lab/main-server/commit/05c865615573cbb53fc9c00de6a452a6ea06269b))

# [0.16.0](https://github.com/CSKU-Lab/main-server/compare/v0.15.0...v0.16.0) (2026-06-20)


### Features

* support upsert on user import route ([8311dfa](https://github.com/CSKU-Lab/main-server/commit/8311dfa8ed10f4df281e6e2c4c12fbc128608b57))

# [0.15.0](https://github.com/CSKU-Lab/main-server/compare/v0.14.3...v0.15.0) (2026-06-20)


### Features

* extend JWT access token expiry to 6 hours ([6645295](https://github.com/CSKU-Lab/main-server/commit/6645295efe00168412b248a6fddbc7a85bfba941))

## [0.14.3](https://github.com/CSKU-Lab/main-server/compare/v0.14.2...v0.14.3) (2026-06-19)


### Bug Fixes

* allow enrolled students to create submissions ([b50eefe](https://github.com/CSKU-Lab/main-server/commit/b50eefe69beef7e2bd8a75fc7975f9def7154c27))

## [0.14.2](https://github.com/CSKU-Lab/main-server/compare/v0.14.1...v0.14.2) (2026-06-03)


### Bug Fixes

* typing submission error rates ([222a4ae](https://github.com/CSKU-Lab/main-server/commit/222a4ae45ad149d8b264aa24a485ea83d718f9d3))

## [0.14.1](https://github.com/CSKU-Lab/main-server/compare/v0.14.0...v0.14.1) (2026-05-24)


### Bug Fixes

* build failed ([87e41d0](https://github.com/CSKU-Lab/main-server/commit/87e41d099c535b06295826c4b365a27b3031bc31))

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
