# ローカル開発環境のセットアップ

## Goの環境を準備する

[goenv](https://github.com/syndbg/goenv)がオススメです。

## MySQLを起動する

DockerでMySQLコンテナを起動します。

```bash
$ make run-db-local
```

正しくアクセスできることが確認できたらOKです。

```bash
$ mysql -u app -pPassword -h 127.0.0.1 -P 3306 -e "SELECT 1"
mysql: [Warning] Using a password on the command line interface can be insecure.
+---+
| 1 |
+---+
| 1 |
+---+
```

## マイグレーション

1. skeemaをインストール

```bash
$ GO111MODULE=off go get -u github.com/skeema/skeema
```

2. マイグレート

```bash
$ skeema push local -p
```

## サーバを起動する

TBD
