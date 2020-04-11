# API仕様

## エンドポイント

ローカル開発環境でのエンドポイントは `localhost.local:8080/api/v3` です。

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

## 認証

全てのAPIで認証が必要です。

事前に`GET /login`で認証を済ませ、Cookieをつけた状態でリクエストを送る必要があります。


## POST /sessions

### 概要
新しいセッションを作成します。

### リクエスト

```json
{
  "name" : "CAMPHOR- HOUSE"
}
```

### レスポンス
  
```json
{
  "session": {
    "id": "xxxxxxxxxxxxxxxxxxxxxxx",
    "name": "CAMPHOR- HOUSE",
    "creator": {
      "id": "p1ass",
      "display_name": "p1ass"
    },
    "playback": {
      "state": "STOP",
      "device": null
    },
    "queue": {
      "head": 0,
      "tracks": []
    }
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
| 403 | already entered other session    | 既に他のセッションに参加している |



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
    "state": "PLAY",
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
        "external_urls": {
          "spotify": "https://open.spotify.com/track/7zHq5ayXLxpJ89392EYm1L"
        },
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
| 400 | invalid device id | 指定されたデバイスIDはオフライン or 不正 |
| 403 | user is not session's creator | セッションの作成者ではない |
| 404 | session not found | 指定されたidのセッションが存在しない |


## PUT /sessions/:id/playback

### 概要

参加しているセッションの再生状態を操作します。

### リクエスト

```json5
{
  "state": "PLAY" // 再生の状態: PLAY または PAUSE または STOP
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
| 400 | invalid device id | 指定されたデバイスIDはオフライン or 不正 |
| 404 | session not found | 指定されたidのセッションが存在しない |


**注意(解決方法を調査中)**

404で `"invalid device id"` が返ってきた時は、聞いていた端末でアプリを開き、再生ボタンを押して一時停止ボタンを押すという操作をしないと、一生再生できない。

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



## GET /users/me/devices

### 概要

ログイン中のユーザのデバイス一覧を取得します。

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



## GET /search

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
      // 以下省略; playbackにあるtrackと同様．
    },
    { 
      "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1L",
      // 以下省略; playbackにあるtrackと同様．
    },
    {
      "uri": "spotify:track:7zHq5ayXLxpJ89392EYm1L",
      // 以下省略; playbackにあるtrackと同様．
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



## GET /ws/:id

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

#### RESUME
sessionの再生が再開された際に発されるイベントです。
```json
{
  "type": "RESUME"
}
```

#### INTERRUPT
Spotifyの本体アプリ側で操作されて、Relaym側との同期が取れなくなったタイミングで発されるイベントです。

セッションは一時停止状態になり、RESUMEを送ることで再開されます。

```json
{
"type": "INTERRUPT",
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
