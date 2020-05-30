# ローカル開発環境のセットアップ

## Goの環境を準備する

[goenv](https://github.com/syndbg/goenv) がオススメです。

## Dockerを準備する

Docker及びDocker Composeを準備してください。

https://www.docker.com/get-started

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

## 環境変数の準備

Spotify Web APIのトークンが必要です。[こちら](https://developer.spotify.com/dashboard) から作成してください。

作成後、 `env.secret` に取得したトークンを設定してください。

```bash
$ cp env.secret.example env.secret
$ ${EDITOR} env.secret
```

## `/etc/hosts` を修正する

`localhost` ではクッキーを使えないので、別名を割り当てる必要があります。

 ```bash
$ sudo vim /etc/hosts

127.0.0.1 relaym.local # これを追加
::1 relaym.local # これを追加
 ```


## サーバを起動する

```bash
$ make serve
```

## モックの生成

repositoryやspotifyインタフェースのモックを生成して、テスタビリティの向上を図ります。

1. [mockgen](https://github.com/golang/mock) をインストール

```bash
$ GO111MODULE=on go get github.com/golang/mock/mockgen
```

2. インターフェースに `go generate` の記述をする。

```go
//go:generate mockgen -source=$GOFILE -destination=../mock_$GOPACKAGE/$GOFILE
package repository
// ...
```

3. go generateする 

```bash
$ make generate
```

### 使い方

[/web/handler/user_test.go](../web/handler/user_test.go)を参照。

## テストの実行

### ユニットテスト

```bash
$ make test
```

### インテグレーションテスト
SpotifyのAPIを実体に叩くことが出来ます。CIで実行することは想定しておらず、ローカルでデバッグするための機能として存在しています。

1. `env.secret` の `SPOTIFY_REFRESH_TOKEN_FOR_TEST` にリフレッシュトークンをセットする。(/loginを叩いてDBから取得するのが簡単です。)
2. `make integration-test`

Spotifyのクライアントを起動しているかどうかや再生しているかどうかで、テストが通るかどうかは変わるので、テストが落ちるのは仕様です。

### E2Eテスト
ローカルで起動したAPIサーバに対して、簡単なシナリオを元にしたAPIリクエストを送信することができます。

1. `make serve`
1. Chromeで `http://relaym.local:8080/api/v3/login` にアクセスしてログイン処理を実行
1. http://relaym.local:8080/* を開いた状態でChrome Dev Consoleに [e2e.js](../testdata/e2e.js)を貼り付ける。
