# アーキテクチャ

## ユーザアクセス

ユーザからのアクセスをAPIサーバが受け取り、曲に関連する情報は[Spotify Web API](https://developer.spotify.com/documentation/web-api/) を使って取得しています。

```
ユーザの端末
↓ HTTP 1.1
APIサーバ
↓
Spotify Web API or MySQL
```

## Spotify Web API

- APIドキュメントは[こちら](https://developer.spotify.com/documentation/web-api/reference-beta/) から確認できます。
- 再生関連のAPIなどはプレミアム会員でないと `403 FORBIDDEN` になります。
    - 非プレミアム会員の場合、テストを書くときはモックを活用してください。  

### OAuth

[Authorization Code](https://developer.spotify.com/documentation/general/guides/authorization-guide/#authorization-code-flow) を使ったApp Authorizationを利用しています。詳しくは公式ドキュメントをお読みください。

https://developer.spotify.com/documentation/general/guides/authorization-guide/

## 本番環境
TBD