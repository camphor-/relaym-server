# データベース

データベースにはMySQL 8.0を使用しています。

## Goの構造体へのマッピングライブラリ

マッピングライブラリは[gorp](https://github.com/go-gorp/gorp)を使用しています。選定理由は@p1assの[ブログ](https://blog.p1ass.com/posts/go-database-sql-wrapper/)をご覧ください。

クエリのSQL文を作る際はGoのファイル上で書くより、GUIのツールで書くと便利です。
Golandを使用しているなら[データベースツール](https://pleiades.io/help/go/relational-databases.html)が使いやすいです。

## スキーマの管理

[skeema](https://github.com/skeema/skeema)というCLIツールを使って管理しています。

スキーマファイルは[./mysql/schemas](../mysql/schemas)にあります。

### 特徴
- 差分のSQLを積んていくのではなく、現在のスキーマの状態をSQLファイルとして管理します。
- 差分の取り込み、反映はコマンド経由で行います。
- パット見で現在のスキーマが分かるところがメリットです。


### 差分をローカルのDBに反映する

upstreamでスキーマが更新された場合は、ローカルのDBにも反映させる必要があります。

`SUPER`か`SYSTEM_VARIABLES_ADMIN`、`SESSION_VARIABLES_ADMIN`の権限が必要です。

```bash
$ skeema diff local -p
$ skeema push local -p
```

破壊的な変更がある場合は失敗します。   
`--allow-unsafe` オプションを付けることで適用することができますが、リリースしてからは注意して行うようにしてください。

### スキーマを変更するPRを作る

好きな方法(mysqlコマンド、GUIクライアント)などでスキーマを変更してください。その後以下のコマンドで差分を取り込んでください。

```bash
$ skeema pull local -p
```
