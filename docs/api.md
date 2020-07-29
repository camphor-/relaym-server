# API仕様

## エンドポイント

ローカル開発環境でのエンドポイントは `relaym.local:8080/api/v3` です。

## データ形式

`application/json` を利用します。

## エラーレスポンス

エラーが発生した時は以下のオブジェクトを含んだエラーレスポンスが返却されます。

```json
{
    "code": 404,
    "message": "user not found"
}
```

### 共通エラーコード

全てのAPIは以下のエラーを返す可能性があります。

| code | message | 補足 |
| -------- | -------- | -------- |
| 401 | Unauthorized | ログインしていない |
| 500 | Internal Server Error | 不明な内部エラー |


## CSRF対策

CSRF対策としてプリフライトリクエストを発生させるために、カスタムヘッダが必要です。

```
X-CSRF-Token: relaym
```

をHTTPヘッダに付与してAPIリクエストを行ってください。

## POST /sessions

### 概要
新しいセッションを作成します。

### 認証
事前に`GET /login`で認証を済ませ、Cookieをつけた状態でリクエストを送る必要があります。

### リクエスト

```json
{
  "name" : "CAMPHOR- HOUSE",
  "allow_to_control_by_others": true
}
```

### レスポンス
  
```json
{
  "id": "xxxxxxxxxxxxxxxxxxxxxxx",
  "name": "CAMPHOR- HOUSE",
  "allow_to_control_by_others": true,
  "creator": {
    "id": "p1ass",
    "display_name": "p1ass"
  },
  "playback": {
    "state": {
      "type": "STOP",
    },
    "device": null
  },
  "queue": {
    "head": 0,
    "tracks": []
  }
}
```

| code  |   補足    |
| ----- | -------- | 
| 201   |          |

### エラー 

| code | message | 補足 |
| ---- | -------- | -------- |
| 400 | empty name | セッション名がリクエストに含まれていない | 



## GET /sessions/:id

### 概要

指定されたidのセッションを返します。

### パスパラメータ

| key | 説明 |
| --- | ------- |
| :id | 参加するsessionのID |

### レスポンス
  
```json5
{
  "id": "xxxxxxxxxxxxxxxxxxxxxxx",
  "name": "CAMPHOR- HOUSE",
  "creator": {
    "id": "p1ass",
    "display_name": "p1ass"
  },
  "playback": {
    "state": {
      "type": "PLAY",
      "length": 12345, // trackの全長 (ms)
      "progress": 10000, // 再生位置 (ms)
      "remaining": 2345, // 再生残り時間 (ms)
    },
    "device": {
      "id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "is_restricted": false,
      "name": "my-device",
    },
  },
  "queue": {
    "head": 1, // 0-indexedなプレイヤーにセットされている曲の番号
    "tracks": [
      { // 0番目: 再生済み
        "uri" : "spotify:track:7zHq5ayXLxpJ89392EYm1L",
        "id": "7zHq5ayXLxpJ89392EYm1L",
        "name" : "Pixel Galaxy",
        "duration_ms": 254165,
        "artists": [{"name": "Snail's House"}],
        "external_url": "https://open.spotify.com/track/7zHq5ayXLxpJ89392EYm1L"
        "album": {
          "name": "Pixel Galaxy",
          "images" : [
            {
              "url" : "https://i.scdn.co/image/ab67616d0000b273ee9b82c65c9a4195f653f063",
              "height" : 640,
              "width" : 640
            }, 
            {
              "url" : "https://i.scdn.co/image/ab67616d00001e02ee9b82c65c9a4195f653f063",
              "height" : 300,
              "width" : 300
            }, {
              "url" : "https://i.scdn.co/image/ab67616d00004851ee9b82c65c9a4195f653f063",
              "height" : 64,
              "width" : 64
            } 
          ],
        },
      },
      { // 1番目: プレイヤーにセット
        "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1",
        // 以下省略
      },
      { // 2番目: 未再生
        "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1",
        // 以下省略
      },
    ]
  }
}
```

| code  |   補足    |
| ----- | -------- | 
| 200   |          |

### エラー 
    
| code | message | 補足 |
| ---- | -------- | -------- |
| 404 | session not found | 指定されたidのセッションが存在しない |



## PUT /sessions/:id/devices

### 概要

指定されたidのセッションの再生に使うデバイスを指定します。

### 認証
事前に`GET /login`で認証を済ませ、Cookieをつけた状態でリクエストを送る必要があります。

### リクエスト

```json
{
  "device_id": "xxxxxxxxxx"
}
```

### レスポンス
空

| code  |   補足    |
| ----- | -------- | 
| 204   |          |

### エラー 

| code | message | 補足 |
| ---- | -------- | -------- |
| 400 | empty device id | デバイスIDがリクエストに含まれていない |
| 403 | user is not session's creator | セッションの作成者ではない |
| 404 | session not found | 指定されたidのセッションが存在しない |


## PUT /sessions/:id/state

### 概要

与えられたセッションのstateを操作します。

### リクエスト

```json5
{
  "state": "PLAY" // 再生の状態: PLAY, PAUSE, STOP, ARCHIVED
}
```

### レスポンス

空

| code  |   補足    |
| ----- | -------- | 
| 202   |          |

非同期的にレスポンスを返すので、実際に状態が反映されたかWebSocketのメッセージか別のAPIリクエストを通して取得する必要があります。

### エラー 

| code | message | 補足 |
| ---- | -------- | -------- |
| 400 | invalid state     | 不正なstate |
| 400 | queue track not found | キューが存在しないので操作を開始できない |
| 400 | requested state is not allowed | 許可されていないstateへの変更(許可されているstateの変更は[PRD](prd.md)を参照) |
| 400 | session is not allowed to control by others | 作成者以外によるstateの操作が許可されていない | 
| 400 | next queue track not found | 再生が終了してStopになったが次のキューが無いので再生を開始できない |   
| 403 | active device not found | アクティブなデバイスが存在しないので操作ができない |
| 404 | session not found | 指定されたidのセッションが存在しない |


**注意(解決方法を調査中)**

[Spotify APIの不思議な挙動](sotify_api_problem.md)

## POST /sessions/:id/queue

### 概要

指定したセッションに曲を追加します。

### リクエスト

```json
{
  "uri": "spotify:track:xxxxxxxxx", 
}
```

### レスポンス

空

| code  |   補足    |
| ----- | -------- | 
| 204   |          |

### エラー 

| code | message | 補足 |
| ---- | -------- | -------- |
| 400 | invalid track id | 指定されたIDが不正 |
| 404 | session not found | 指定されたidのセッションが存在しない |


## GET /users/me

### 概要

ログイン中のユーザ情報を取得します。

### 認証
事前に`GET /login`で認証を済ませ、Cookieをつけた状態でリクエストを送る必要があります。

### リクエスト
空

### レスポンス
  
```json
{
  "id": "p1ass",
  "uri": "spotify:user:p1ass",
  "display_name": "p1ass",
  "is_premium": true
}
```

| code  |   補足    |
| ----- | -------- | 
| 200   |          |



## GET /sessions/:id/devices

### 概要

セッションの作成者のデバイス一覧を取得します。

Spotifyのアプリが起動していないと一覧に現れません。


### リクエスト
空

### レスポンス
  
```json
{
  "devices": [
    {
      "id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "is_restricted": false,
      "name": "my-device",
    },
    {
      "id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
      "is_restricted": false,
      "name": "my-device",
    }
    ]
}
```

| code  |   補足    |
| ----- | -------- | 
| 200   |          |



## GET /sessions/:id/search

### 概要

Spotifyで曲の検索を行います。

### クエリパラメータ

| key | 説明 |
| --- | ------- |
| q | 検索キーワード |


### レスポンス
   
```json5
{
  "tracks": [
    { 
      "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1L",
      // 以下省略 /sessions/:idのtracksを参考
    },
    { 
      "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1L",
      // 以下省略 /sessions/:idのtracksを参考
    },
    {
      "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1L",
      // 以下省略 /sessions/:idのtracksを参考
    },
  ]

}
```

| code  |   補足    |
| ----- | -------- | 
| 200   |          |

### エラー 

| code | message | 補足 |
| ---- | -------- | -------- |
| 400 | query is empty | 検索キーワードが空 |



## GET /sessions/:id/ws

### 概要
指定したセッションに関連するイベントを配信するWebSocketエンドポイントです。

### パスパラメータ

| key | 説明 |
| --- | ------- |
| :id | 参加するsessionのID |

### レスポンス

| code  |   補足    |
| ----- | -------- | 
| 101   |          |
  

### イベント

#### ADDTRACK
セッションに曲が追加された際に発されるイベントです。
  
```json
{
  "type": "ADDTRACK"
}
```

#### NEXTTRACK
セッションの曲の再生が (正常に) 次の曲に移った際に発されるイベント。キューの現在再生している曲の位置が含まれますです。
  
```json
{
  "type": "NEXTTRACK",
  "head": "1"
}
```
  
#### PLAY
セッションの再生が開始された際に発されるイベントです。

```json
{
  "type": "PLAY" 
}
```

#### PAUSE
セッションが一時停止された際に発されるイベントです。
```json
{
  "type": "PAUSE"
}
```

### STOP
全ての曲の再生が終了した際に発されるイベントです。
```json
{
  "type": "STOP"
}
```

#### INTERRUPT
Spotifyの本体アプリ側で操作されて、Relaym側との同期が取れなくなったタイミングで発されるイベントです。

セッションはSTOP状態になり、再度state APIでPLAYにする必要があります。

```json
{
"type": "INTERRUPT"
}
```

#### ARCHIVED
セッションがARCHIVEされた際に発されるイベントです。
```json
{
"type": "ARCHIVED"
}
```

#### UNARCHIVED
セッションのARCHIVEが解除された際に発されるイベントです。
```json
{
"type": "UNARCHIVED"
}
```

### エラー 
    
| code | message | 補足 |
| ---- | -------- | -------- |
| 404 | session not found | 指定されたidのセッションが存在しない |

## GET /login

### 概要
SpotifyのOAuthログインをスタートします。内部で処理が終わったらSpotifyの認証画面にリダイレクトされます。

JavaScriptで非同期にリクエストするのではなく、aタグで同期的にアクセスしてください。

### クエリパラメータ

| key | 説明 |
| --- | ------- |
| redirect_url | Spotifyの認証が終わった後リダイレクトされるクライアントのURLを指定します |

#### レスポンス
| code | 補足 |
| - | - |
|302 | Spotifyの認証画面にリダイレクトします|


## GET /callback

### 概要

Spotifyの認証が終わった際にリダイレクトされてくるエンドポイントです。

クライアントが直接叩くことはありません。

### レスポンス
| code | 補足 |
| - | - |
|302 | GET /login で受け取ったredirect_url に認証用のクッキーをつけてリダイレクトします |

## POST /batch/archive

### 概要

以下の条件に当てはまるsessionのstateをARCHIVEDに変更します
夜間にcronで叩かれることを想定しています

- `expired_at`に現在の時刻より前に設定されている

### レスポンス
| code | 補足 |
| - | - |
| 200 | |
| 500 | 何らかのエラーが発生 |
