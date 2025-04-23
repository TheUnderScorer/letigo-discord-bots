# [2.18.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.17.0...v2.18.0) (2025-04-23)


### Features

* improve logging ([2b69d40](https://github.com/TheUnderScorer/letigo-discord-bots/commit/2b69d4061bb2cf64a2bfe6d3348a7425a6a65de5))

# [2.17.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.16.0...v2.17.0) (2025-04-23)


### Features

* add chat scanner ([426eeea](https://github.com/TheUnderScorer/letigo-discord-bots/commit/426eeeac70f6ba0fd41c60d46e4671c65b9f2263))
* add vision support ([a4e6c4f](https://github.com/TheUnderScorer/letigo-discord-bots/commit/a4e6c4f29d041a106eefb6ffeceed0dcc7aa34c7))
* remove worthness check in new message handler ([cc7f1c6](https://github.com/TheUnderScorer/letigo-discord-bots/commit/cc7f1c6219dd696d0397649cb163374ad3c97a66))
* rework memory system to use batching and intervals ([4c03fa9](https://github.com/TheUnderScorer/letigo-discord-bots/commit/4c03fa9bf522c20f29065e1ad6d9c23053546c83))

# [2.16.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.15.0...v2.16.0) (2025-04-18)


### Features

* support DMs for chat ([7039bdd](https://github.com/TheUnderScorer/letigo-discord-bots/commit/7039bddf92fabdb0b3d23741133661b32464113e))

# [2.15.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.14.0...v2.15.0) (2025-04-18)


### Bug Fixes

* **trivia:** handle cases where player nomination channel is not listening ([0c31bff](https://github.com/TheUnderScorer/letigo-discord-bots/commit/0c31bffc0b0dc25fdd32a6b4163f12792f2aece7))


### Features

* handle forgetting ([b544985](https://github.com/TheUnderScorer/letigo-discord-bots/commit/b54498581ed7852b385f27efb0ba4cfceae6fee1))
* use embed for reporting errors ([5a85008](https://github.com/TheUnderScorer/letigo-discord-bots/commit/5a850086b34e6ab5ac2898deb93b077f972972b5))
* use embeds in memory updated handler ([17dda4e](https://github.com/TheUnderScorer/letigo-discord-bots/commit/17dda4e3390dbe14e3a8bfc3c39dc2e5cba21375))

# [2.14.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.13.2...v2.14.0) (2025-04-15)


### Features

* add chat functionality to the bot ([254a65d](https://github.com/TheUnderScorer/letigo-discord-bots/commit/254a65de3bad361bcc8a103f298ae1da456db331))
* add memory functionality to the chat ([b59ed76](https://github.com/TheUnderScorer/letigo-discord-bots/commit/b59ed76792c695e32aebdae0ed12ab7aeb617dd2))

## [2.13.2](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.13.1...v2.13.2) (2025-03-05)


### Bug Fixes

* use correct channel for greeting ([3f41279](https://github.com/TheUnderScorer/letigo-discord-bots/commit/3f412796517e0032adbc87e60fbdc62145f547b4))

## [2.13.1](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.13.0...v2.13.1) (2024-12-27)


### Bug Fixes

* use fallback format if opus is not found ([80cc9ea](https://github.com/TheUnderScorer/letigo-discord-bots/commit/80cc9ea75b446124e45fc2fa9826b5bd23813a7e))

# [2.13.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.12.0...v2.13.0) (2024-11-24)


### Features

* add logging to scheduler ([15ce228](https://github.com/TheUnderScorer/letigo-discord-bots/commit/15ce228fdb2ca54d77c98caaa04a7c39c5b3d6e7))

# [2.12.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.11.0...v2.12.0) (2024-11-17)


### Features

* add logging to file ([3f664e2](https://github.com/TheUnderScorer/letigo-discord-bots/commit/3f664e27ceb510fb2076553f25a103fc9c2eeee7))

# [2.11.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.10.0...v2.11.0) (2024-11-14)


### Bug Fixes

* remove logs for processed requests ([d953da6](https://github.com/TheUnderScorer/letigo-discord-bots/commit/d953da6f3307e6165d57f0eb11ff612c811b2854))


### Features

* add linux build to release ([d75ef58](https://github.com/TheUnderScorer/letigo-discord-bots/commit/d75ef5837db82368099fd6f565ecd9088048f568))

# [2.10.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.9.0...v2.10.0) (2024-11-14)


### Bug Fixes

* don't panic when command fails to create ([232cecf](https://github.com/TheUnderScorer/letigo-discord-bots/commit/232cecfadf3ffe2bc1982d10664a8dda980cb91b))


### Features

* add endpoint for registering interactions ([f2f75bd](https://github.com/TheUnderScorer/letigo-discord-bots/commit/f2f75bd2fef1d97d10f74623450f76dc4fa5f3ab))


### Performance Improvements

* init interactions in separate go routine ([c0ecd75](https://github.com/TheUnderScorer/letigo-discord-bots/commit/c0ecd75db60f7cfc61d6e6febb8ff18a035c7260))

# [2.9.0](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.8.2...v2.9.0) (2024-11-08)


### Features

* Add "1 dzban z dziesieciu" in beta ([62f7243](https://github.com/TheUnderScorer/letigo-discord-bots/commit/62f724332a822c9fd9378b06391e99374971974b))
* use GO lang ([a665cbf](https://github.com/TheUnderScorer/letigo-discord-bots/commit/a665cbf0f7b2dbc6a4d26032a1892f2285db004c))

## [2.8.2](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.8.1...v2.8.2) (2024-10-19)


### Bug Fixes

* use custom agent in ytdl ([0c79dfc](https://github.com/TheUnderScorer/letigo-discord-bots/commit/0c79dfcd0b52ec36e18f5c318b99a12909522e84))

## [2.8.1](https://github.com/TheUnderScorer/letigo-discord-bots/compare/v2.8.0...v2.8.1) (2024-10-19)


### Bug Fixes

* fix player ([cf84b6e](https://github.com/TheUnderScorer/letigo-discord-bots/commit/cf84b6ee6c614371128faab46cbd222ebda4686c))

# [2.8.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.7.1...v2.8.0) (2024-01-07)


### Features

* add more messages ([63adf73](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/63adf73646b269393e3e9a147303274b5d7a3460))

## [2.7.1](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.7.0...v2.7.1) (2023-03-17)


### Bug Fixes

* correctly dispose voice connection after disconnect ([153ff11](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/153ff11e004263bce60132a9efedc12c29088186))
* use channel id rather than guild id for managing players ([9ccb868](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/9ccb868110fb72ece262f96508f4c2e8cfdc6e20))

# [2.7.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.6.0...v2.7.0) (2023-03-06)


### Features

* improve ai context ([2662a7b](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/2662a7b2e18a5cd901eece09ba769d9d4a195c15))
* increase model temperature ([e70e821](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/e70e821f35b1dfc2738f0a81841754d1ee9620b5))
* set max_tokens ([d2c2d79](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/d2c2d7985b76f1a39159c71f8658caba5a8ded4e))

# [2.6.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.5.0...v2.6.0) (2023-03-06)


### Features

* use new chat gpt model ([688f8ab](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/688f8ab531de251751dac74acec6f541d6560332))

# [2.5.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.4.0...v2.5.0) (2023-02-28)


### Features

* add beta OpenAI integration ([bd73536](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/bd73536f24986f383d330ea04bd6af8d36cfd113))

# [2.4.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.3.0...v2.4.0) (2023-02-25)


### Features

* add twin tails reactions ([60885a7](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/60885a76b1ad346c8e7237b118e757f37ae7bd8f))
* add twin tails reminder ([8a1ee59](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/8a1ee59875c93ea5f574bc94bb095d959f3580f6))
* support multiple daily reminder messages ([883745d](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/883745d4df3071c1054b7e7de3cf31b78fe10fb8))

# [2.3.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.2.0...v2.3.0) (2022-09-04)


### Features

* support skipped daily report days ([c5d083c](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/c5d083c5af74a797ac6682e0beee0c4e10a1c631))

# [2.2.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.1.0...v2.2.0) (2022-09-03)


### Features

* reply to daily reports ([493f54b](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/493f54b9cc6b17ab4cfd3f53715b4fa1d36642cd))

# [2.1.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v2.0.0...v2.1.0) (2022-07-17)


### Features

* support new daily report format ([a921990](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/a921990174e35493b88df0c2101c9be6effd994a))

# [2.0.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v1.3.0...v2.0.0) (2022-07-10)


### Features

* introduce music player available via "!kolego player <command>" ([40e9dc2](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/40e9dc2329ee0814437822d5f44f1c1e718bba71))
* Remove lambdas ([de40c1a](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/de40c1a9c1a36be543447a64da65775bf1cc85d0))


### Performance Improvements

* **player:** find best quality for playing ([d77ef1c](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/d77ef1c9c3c8faa6e6297a7e3363715c6fd727ec))


### BREAKING CHANGES

* API Endpoints are no longer available

# [1.3.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v1.2.0...v1.3.0) (2022-05-02)


### Features

* Support insulting other users ([e49f14d](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/e49f14da4952c07f42ff33cc4b0f7a002dbfdbab))

# [1.2.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v1.1.0...v1.2.0) (2022-05-02)


### Features

* add /cotam command ([e1f982b](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/e1f982bc7458eb06a31673d11247640c631b169c))

# [1.1.0](https://github.com/TheUnderScorer/wojciech-discord-bot/compare/v1.0.0...v1.1.0) (2022-04-22)


### Features

* Add daily greeting ([4e6b38d](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/4e6b38d0e8f17bbbdda99b4aba757ee5d7f41e9d))

# 1.0.0 (2022-04-19)


### Features

* Add initial interactions handler ([d82a67b](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/d82a67be617319bb31209eb3490089b8c1a26c60))
* Handle questions ([8e350fb](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/8e350fbe5dbb995a6c759c197dc5608927f1e3bc))
* improve daily report checking logic ([596cfb8](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/596cfb8d9fe8483e1032fe143dc86e50110365e4))
* Initial commit ([d30e249](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/d30e2492ee6186e3d4623379015dbd6c3b2076ac))
* Register initial command ([5b36955](https://github.com/TheUnderScorer/wojciech-discord-bot/commit/5b3695534c9de40c45ee84b748a43be1bfc201c8))
