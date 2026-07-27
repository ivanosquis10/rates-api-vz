# Changelog

## [1.5.1](https://github.com/ivanosquis10/rates-api-vz/compare/v1.5.0...v1.5.1) (2026-07-27)


### Bug Fixes

* **scraper:** disable TLS verification for BCV scraper ([67361ed](https://github.com/ivanosquis10/rates-api-vz/commit/67361edfaf326aa6965dcc7ecc8ee6e77320c1c3))

## [1.5.0](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.4.0...v1.5.0) (2026-07-27)


### Features

* **docker:** add Dockerfile and docker-compose configuration for production ([248d56c](https://github.com/ivanosquis10/api-rates-venezuela/commit/248d56cb6b2df18405488269e28765619c9d92e8))

## [1.4.0](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.3.0...v1.4.0) (2026-07-16)


### Features

* run initial scrape at startup before first cron tick ([df01172](https://github.com/ivanosquis10/api-rates-venezuela/commit/df01172fb25fb178c7fa39f6ef316f9ac679d441))
* scraping inicial al iniciar la aplicacion ([edfb8b8](https://github.com/ivanosquis10/api-rates-venezuela/commit/edfb8b81d8d920848291651c3decbacd9011b1f2))


### Bug Fixes

* update RunNow tests for new NewScheduler signature ([e7741b9](https://github.com/ivanosquis10/api-rates-venezuela/commit/e7741b9b77df7ce41063fe9c09fc64305ba3673f))

## [1.3.0](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.2.1...v1.3.0) (2026-07-16)


### Features

* simplify rates api to bcv official reference rates only ([896587f](https://github.com/ivanosquis10/api-rates-venezuela/commit/896587f5e8c2b4fb2642888fe025677af79e9061))
* simplify rates api to bcv official reference rates only ([5622758](https://github.com/ivanosquis10/api-rates-venezuela/commit/5622758d1bc3c9657613f8dee0b1015682451710))


### Bug Fixes

* map sql.ErrNoRows to domain.ErrNotFound and return empty slice on list endpoints ([9c9c16c](https://github.com/ivanosquis10/api-rates-venezuela/commit/9c9c16cc3ac8bc514a5c85d3ac6204c452ae9897))

## [1.2.1](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.2.0...v1.2.1) (2026-07-12)


### Bug Fixes

* **store:** return all rates when currency empty and handle duplicates ([e9adfc6](https://github.com/ivanosquis10/api-rates-venezuela/commit/e9adfc6f39ffd65fbff263ecb9c56b8a3db6c4f2))
* **store:** return all rates when currency empty and handle duplicates (closes [#27](https://github.com/ivanosquis10/api-rates-venezuela/issues/27)) ([9d0e95e](https://github.com/ivanosquis10/api-rates-venezuela/commit/9d0e95e2993aba635257b393eb3ff494f74f984f))

## [1.2.0](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.1.0...v1.2.0) (2026-07-11)


### Features

* **api:** add CORS middleware for cross-origin requests ([cf5948c](https://github.com/ivanosquis10/api-rates-venezuela/commit/cf5948ce0017d1de1343b752da2f66b681c2dcc5))
* **api:** add CORS middleware for cross-origin requests ([786766a](https://github.com/ivanosquis10/api-rates-venezuela/commit/786766a29700d2ca7432648017f191fa8e26886b))

## [1.1.0](https://github.com/ivanosquis10/api-rates-venezuela/compare/v1.0.0...v1.1.0) (2026-07-11)


### Features

* **api:** add health endpoint and /api/v1 prefix ([5cd4bc7](https://github.com/ivanosquis10/api-rates-venezuela/commit/5cd4bc70fd1e9ab45c14aa183a1930aa7722dfa2))
* **api:** add health endpoint and /api/v1 prefix ([9636ce2](https://github.com/ivanosquis10/api-rates-venezuela/commit/9636ce202302987079965acd40a7b96575d0e255))
