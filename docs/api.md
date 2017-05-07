FORMAT: 1A

# squidgirl-go API

# Group ユーザー処理API

## ユーザーログイン [/login{?username,password}]
### POST

* 指定したページURLから画像のURLを抽出して返却する
* サイズ指定を行うとそのサイズ範囲内に合致した画像のみを返却する

+ Parameters
    + username: user_name (string, required) - ユーザー名
    + password: hogehoge (string, required) - パスワード

+ Response 200 (application/json)
    + Attributes
        + token: xxxxxxxxxxxxxx (string, required) - ログイントークン

## ユーザー作成 [/createuser]
### POST

* 内部で定義したユーザーを追加する

+ Parameters

+ Response 200 (application/json)
    + Attributes
        + status: OK (string, required) - 登録結果


# Group ファイル取得API

## ファイル・フォルダ一覧取得 [/api/filelist{?folder_hash}]
### POST

* 指定したフォルダ以下のファイルまたはフォルダを一覧で取得する

+ Parameters
    + hash: zzzzzzzzz (string, required) - ファイルを取得するフォルダのハッシュ値（空文字時はルートを取得）

+ Response 200 (application/json)
    + Attributes
        + files (array) - ファイル情報リスト
            + (object)
                + hash: xxxxx (string) - ファイル・フォルダのハッシュ値
                + name: name.zip (string) - ファイル・フォルダ名
                + size: 4000000 (number)  - ファイルサイズ（フォルダ時は0）
                + page: 194 (number)  - ページ数（フォルダ時は0）
                + isdir: false (boolean)  - フォルダかどうか（フォルダ時はtrue ファイル時はfalse）
                + modtime: 2017-01-01T02:44:33 (datetime)  - 最終更新日
                + readtime: 2017-05-06T23:44:33 (datetime)  - 最終閲覧日時
                + readpos: 45 (number)  - 既読位置（フォルダ時は0）
                + reaction: 1 (number)  - リアクションタイプ（フォルダ時は0）

## 履歴一覧取得 [/api/historylist{?token}]

* 履歴として残っているファイルを一覧で取得する

+ Parameters

+ Response 200 (application/json)
    ファイル・フォルダ一覧取得と同じ

## リアクション一覧取得 [/api/reactionlist{?token}]

* リアクション登録されたファイルを一覧で取得する

+ Parameters
    + reaction: 1 (number, required) - リアクションタイプ

+ Response 200 (application/json)
    ファイル・フォルダ一覧取得と同じ


# Group ページ関連API

## ページ画像取得 [/api/getbookimage{?token,path,index,maxheight,maxwidth}]
### POST

* リアクション登録されたファイルを一覧で取得する

+ Parameters
    + token: xxxxxxxxxxx (string, required) - ログインで取得したトークン
    + path: 1 (number, required) - ファイルパス
    + index: 1 (number, required) - ページ番号
    + maxheight: 1 (number, required) - 最大高さ
    + maxwidth: 1 (number, required) - 最大幅

+ Response 200 (image/jpeg)

## ページ状態保存 [/api/setbook{?token,path,index,reaction}]
### POST

+ Parameters
    + token: xxxxxxxxxxx (string, required) - ログインで取得したトークン
    + path: 1 (number, required) - ファイルパス
    + index: 1 (number, required) - 既読位置
    + reaction: 1 (number, required) - リアクションタイプ

+ Response 200 (application/json)
    + Attributes
        + status: 0 (number, required) - 保存結果

