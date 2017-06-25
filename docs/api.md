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

## ユーザー追加 [/api/createuser{?username,password,authlevel}]
### POST

* 新しいユーザーを追加する
* すでに登録されているユーザーの場合はエラーとなる
* この操作は管理者権限があるユーザーでのみ可能

+ Parameters
    + username: name (string, required) - ユーザー名（半角英数字）
    + password: abcdefg (string, required) - パスワード（半角英数字）
    + authlevel: 1 (number, required) - 権限レベル（1=ユーザー、100=管理者）

+ Response 200 (application/json)
    + Attributes
        + status: 0 (number, required) - 処理結果（0=正常）

## ユーザー削除 [/api/deleteuser{?username}]
### POST

* 指定したユーザーを削除する
* この操作は管理者権限があるユーザーのみ可能

+ Parameters
    + username: name (string, required) - ユーザー名（半角英数字）

+ Response 200 (application/json)
    + Attributes
        + status: 0 (number, required) - 処理結果（0=正常）

## ユーザー一覧 [/api/userlist]
### GET

* 現在登録されているユーザーをすべて取得する

+ Response 200 (application/json)
    + Attributes
        + count: 取得ユーザー数
        + users (array) - ユーザー情報リスト
            + (object)
                + name: name.zip (string) - ユーザー名
                + authlevel: 1 (number)  - 権限レベル

# Group ファイル取得API

## ファイル・フォルダ一覧取得 [/api/filelist{?hash,offset,limit}]
### POST

* 指定したフォルダ以下のファイルまたはフォルダを一覧で取得する
* ".." ファイル名で一つ上の階層のフォルダも取得する（ルート時は返さない）

+ Parameters
    + hash: zzzzzzzzz (string, required) - ファイルを取得するフォルダのハッシュ値（空文字時はルートを取得）
    + offset: 0 (number, required) - 取得開始位置
    + limit: 10 (number, required) - 取得最大数
    
+ Response 200 (application/json)
    + Attributes
        + name: 指定したハッシュのファイル・フォルダ名
        + allcount: 所属ファイルの最大数(親フォルダは含まない)
        + count: 取得ファイル数
        + files (array) - ファイル情報リスト
            + (object)
                + hash: xxxxx (string) - ファイル・フォルダのハッシュ値
                + name: name.zip (string) - ファイル・フォルダ名
                + size: 4000000 (number)  - ファイルサイズ（フォルダ時は0）
                + page: 194 (number)  - ページ数（フォルダ時は0）
                + isdir: false (boolean)  - フォルダかどうか（フォルダ時はtrue ファイル時はfalse）
                + modtime: 2017-01-01T02:44:33 (datetime)  - 最終更新日
                + readtime: 2017-05-06T23:44:33 (datetime)  - 最終閲覧日時
                + index: 45 (number)  - 既読位置（フォルダ時は0）
                + reaction: 1 (number)  - リアクションタイプ（フォルダ時は0）

## ファイル・フォルダ一覧取得 [/api/parentlist{?hash}]
### POST

* 指定したフォルダより上のフォルダを一覧で取得する
* 自分自身のフォルダ情報を含み下位階層から順番に登録する

+ Response 200 (application/json)
    + Attributes
        + count: 取得フォルダ数
        + folders (array) - フォルダ情報リスト
            + (object)
                + hash: xxxxx (string) - ファイル・フォルダのハッシュ値
                + name: name.zip (string) - ファイル・フォルダ名

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

## サムネイル画像取得 [/api/thumbnail/{hash}{?base64}]
### GET

* サムネイル画像を取得する

+ Parameters
    + hash: xxxxxxxxxxx (string, required) - ファイルハッシュ（フォルダも可能）
    + base64: false (boolean, required) - base64文字列で返却するかどうか

+ Response 200 (image/jpeg) 
    * base64 == false の時は画像データとして返す

+ Response 200 (text/plain)
    * base64 == true の時はBASE64文字列として返す

## ページ画像取得 [/api/page/{hash}{?index,maxheight,maxwidth,base64}]
### POST

* リアクション登録されたファイルを一覧で取得する

+ Parameters
    + hash: xxxxxxxxxxx (string, required) - ファイルハッシュ
    + index: 1 (number, required) - ページ番号（1～）
    + maxheight: 1280 (number, required) - 最大高さ
    + maxwidth: 720 (number, required) - 最大幅
    + base64: false (boolean, required) - base64文字列で返却するかどうか

+ Response 200 (image/jpeg) 
    * base64 == false の時は画像データとして返す

+ Response 200 (text/plain)
    * base64 == true の時はBASE64文字列として返す

+ Response 403 (text/plain)
    * まだファイルの展開が行われていないとき返却する。事件経過後にアクセスを行うことで取得できる


## ページ状態保存 [/api/savebook{?hash,index,reaction}]
### POST

+ Parameters
    + hash: xxxxxxxxxxx (string, required) - ファイルハッシュ
    + index: 1 (number, required) - 既読位置
    + reaction: 1 (number, required) - リアクションタイプ

+ Response 200 (application/json)
    + Attributes
        + status: 0 (number, required) - 保存結果

