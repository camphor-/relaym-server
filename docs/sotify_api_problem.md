# Spotify Web APIの不思議な挙動

## 曲の再生スタート時の `active device not found` に関して

### 関連issue

https://github.com/spotify/web-api/issues/1325

### 1つ目

#### 動作

1. Spotify アプリを閉じる (kill)
2. 再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く
3. `active device not found` になる
4. `GET /users/me/devices` を叩く
5. 空
6. Spotify アプリを開く
7. `GET /users/me/devices` を叩く
8. アプリを開いたデバイスが存在する
9. 再度、再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く
10. `active device not found` になる
11. なにかしら曲を再生する
12. 曲を止める
13. 再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く
14. 正しく再生される

#### 要因

アプリをkillして開くと、active deviceにはなるが、再生準備は出来ていない扱いになるっぽい。

(なら8のタイミングでactive deviceにしないでくれ〜)

#### 対応

再生API `PUT /sessions/:id/state` で叩いているSpotify API `PUT https://api.spotify.com/v1/me/player/play` で明示的に `device_id` をクエリパラメータに含めることで、9のタイミングで表示されるようになる。

操作の流れとしては、

1. Relaymの再生ボタンを押す
2. `active device not found` エラーになる
3. Spotifyのアプリを開いてもらう
4. Relaymに戻ってくる
5. デバイスを選択する
6. 再度Relaymの再生ボタンを押す
7. 再生が始まる

が良さそう

### 2つ目

#### 動作

1. Spotify アプリを開く
2. なにかしら曲を再生する
3. 曲を止める
4. 再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く
5. 正しく再生される
6. Spotify アプリで再生を止める
7. アプリをバックグラウンドに移行する
8. 1分ほど待つ
9. `GET /users/me/devices` を叩く
10. 先程まで再生していたデバイスが存在する
11. 再度、再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く
12. Spotify APIが500 で `Server Error` が返ってくる
13. `GET /users/me/devices` を叩く
14. 先程まで再生していたデバイスが消える
15. Spotify アプリをフォアグラウンドにする
16. ここで自動で再生が開始される (11のAPIの処理が行われる)

#### 動作パターン2

10までは一緒

11'. Spotify アプリをフォアグラウンドにする  
12'. 再度、再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く  
13'. `active device not found` になる  
14'. なにかしら曲を再生する  
15'. 曲を止める  
16'. 再生API `PUT /sessions/:id/state` を `state: PLAY` で叩く  
17'. 正しく再生される  

#### 要因

アプリを一定時間バックグラウンドにするとデバイスがinactiveになるが、デバイス取得APIはキャッシュされていてそのままactiveの状態を返してしまっているっぽい。

#### 厄介なところ

- `Server Error` として500を返してくる
    - 素直に`active device not found` を返してくれと思う。

#### 対応

サーバ側だけで対応するのは難しい。

- `Server Error` はサーバサイドで `active device not found` エラーと同様のものとして扱う
- `PUT /sessions/:id/state` で `active device not found` になったときは、ユーザにSpotify アプリを開いてもらうダイアログを出す。アプリを開けば自動で再生が始めるので、あまりユーザの負担にならない。


