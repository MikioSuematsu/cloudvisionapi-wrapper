#  Cloud Vision Wrapper API
 これは、[Cloud Vision API ](https://cloud.google.com/vision/docs/)　をお手軽に Google App Engine上で試せるソースコードになります。

## API仕様


### API のパラメータ

- url 画像のURL
- cache: レスポンスや画像をキャッシュするかどうか。１の場合はキャッシュする。０の場合はキャッシュしない。
- callback: JSONPを利用する場合にコールバック関数名を指定する
- type: `Cloud Vision API  の Future Type。以下のパラメータが指定可能
 - FACE_DETECTION    顔認識
 - LANDMARK_DETECTION    ランドマークの認識
 - LOGO_DETECTION    製品ロゴの認識
 - LABEL_DETECTION    画像コンテンツの認識 (ラベリング)
 - TEXT_DETECTION    画像内テキストの認識 (OCR)
 - SAFE_SEARCH_DETECTION    セーフサーチの判定
 - IMAGE_PROPERTIES  色解析

### APIサンプルリクエスト
`http://xxxxxxxxxxx.appspot.com/?url=http://xxxxxxxxxx.com/image.jpg`


# API DEMO Page
[DEMOページはこちら](./demo/demo.html)


# Cloud Vision Wrapper API の環境構築について

## 環境準備
 [App Engineの管理画面](https://console.cloud.google.com/appengine)からGo言語のサンプル・アプリケーションをダウンロードするか、本ソースをダウンロードしてください。

## Cloud Storageの準備
このアプリケーションは、 Cloud Storage を利用するため、Cloud Storage上でバケットを作成してください。
また、作成したバケットについては、 `AllUsers` ユーザーに対して `読み取り` の権限を付与してください。


## ソースの修正
サンプル・アプリケーションをダウンロードした場合は、本ソースの hallo.go をコピー＆ペーストして以下の部分を修正してください。本ソースをそのままダウンロードした場合は、以下の部分と `app.yaml` の `application` 項目の変更をお願いします。


```
var API_KEY = "API_KEYを貼り付けてください";
var BUCKET_NAME = "作成したバケット名を記載してください"
```


#デプロイ方法については、以下のコマンドでデプロイしてください。

```
goapp deploy -application [ProjectName] app.yaml
```