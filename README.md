# anilist-synchronizer

[AniList](https://anilist.co) アカウント間でリストを同期します。

同期元のアカウント (A) から同期先のアカウント (B) への同期は次のルールで行われます。

- A と B の共通エントリーが対象です。
  - A にはあるが、B にはないエントリーは新規作成されません。
  - また B にはあるが、A にはないエントリーは削除されません。
- 「視聴ステータス」「スコア」「話数」のみを同期します。

## 環境変数

- `ANILIST_CLIENT_ID`, `ANILIST_CLIENT_SECRET`
  - AniList の OAuth クライアントです。https://anilist.co/settings/developer で発行してください。
- `TOKEN_DIRECTORY`
  - トークン情報を格納するディレクトリを指定します。未指定の場合はカレントディレクトリに格納します。
- `INTERVAL_MINUTES`
  - 指定した分ごとに同期を行います。未指定の場合は一度同期して終了します。

## Build

```console
$ make build
```

## Run

```console
$ make run
```
