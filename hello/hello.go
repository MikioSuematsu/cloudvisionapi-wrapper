package hello

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"bytes"
	"strings"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/cloud/storage"
)


var API_KEY = "API_KEYを貼り付けてください"
var BUCKET_NAME = "作成したバケット名を記載してください"

var VISION_API_PATH = "https://vision.googleapis.com/v1/images:annotate?key=" + API_KEY
var MAX_RESULTS = "50"

func init() {
	http.HandleFunc("/", handler)
}

func handler(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	//画像URL
	url := r.FormValue("url")
	if url == "" {
		fmt.Fprint(w, "画像URL(url）を指定してください。.")
		return
	}

	/*
	* Cloud Vision API で指定可能な画像認識の種類
	*
	* FACE_DETECTION    顔認識
	* LANDMARK_DETECTION    ランドマークの認識 ［デフォルト］
	* LOGO_DETECTION    製品ロゴの認識
	* LABEL_DETECTION    画像コンテンツの認識 (ラベリング)
	* TEXT_DETECTION    画像内テキストの認識 (OCR)
	* SAFE_SEARCH_DETECTION    セーフサーチの判定
	* IMAGE_PROPERTIES  色解析
	 */
	featureType := r.FormValue("type")
	if featureType == "" {
		featureType = "LABEL_DETECTION"
	}

	/*
	 * キャッシュの有無
	 *
	 * 0 キャッシュせずに画像の取得 & CloudVisionAPIに問い合わせをします CloudStorage 上には何も残しません （デフォルト）
	 * 1 Cloud Storage のキャッシュを利用
	 */
	cache := r.FormValue("cache")
	if cache == "" {
		cache = "0"
	}

	/**
	 * jsonp 対応のための callback名
	 */
	callback := r.FormValue("callback")


	//Cloud Storage上に置くファイルのmimeTypeを判断
	var mimeType string
	var extension string
	if strings.HasSuffix(url, ".png") {
		mimeType = "image/png"
		extension = ".png"
	} else if strings.HasSuffix(url, ".jpeg") {
		mimeType = "image/jpg"
		extension = ".jpeg"
	} else if strings.HasSuffix(url, ".gif") {
		mimeType = "image/gif"
		extension = ".gif"
	} else {
		mimeType = "image/jpg"
		extension = ".jpeg"
	}
	fileName := strings.Replace(url, "/", "_", -1) + "_" + featureType + extension

	//キャッシュの利用
	if (cache == "1") {
		log.Debugf(c, "キャッシュデータを検索します。")
		client, err := storage.NewClient(c)
		//キャッシュ有効時は、キャッシュ済みのレスポンスファイルを取得し、処理終了
		r, err := client.Bucket(BUCKET_NAME).Object(fileName + "_.json").NewReader(c)
		if err == nil {
			buf := new(bytes.Buffer)
			if _, err := buf.ReadFrom(r); err == nil {
				log.Debugf(c, "キャッシュデータを取得します。")
				responsData := buf.Bytes()
				//レスポンスを返す
				responsDataStr := string(responsData)
				log.Debugf(c, "キャッシュデータのレスポンスを返します。"+ responsDataStr)
				fmt.Fprint(w, callback + "(" + responsDataStr + ");")
				defer r.Close()
				return
			}
		}
		log.Debugf(c, "キャッシュデータが無かったため Cloud Vision API の利用します。")
	}


	//画像ファイルをダウンロード
	log.Debugf(c, "画像のダウンロードを開始します。")
	downlocadClient := urlfetch.Client(c)
	response, err := downlocadClient.Get(url)
	if err != nil {
		fmt.Fprint(w, "画像のダウンロードに失敗しました。" + err.Error())
		return
	}
	downloadBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprint(w, "画像のファイルの読み込みに失敗しました。" + err.Error())
		return
	}
	log.Debugf(c, "画像のダウンロードが完了しました。")


	//Cloud Storage にダウンロード画像を格納
	log.Debugf(c, "Cloud Storage に画像を格納します。")
	log.Debugf(c, "画像格納先：" + BUCKET_NAME + "/" + fileName)
	imageClient, err := storage.NewClient(c)
	wc := imageClient.Bucket(BUCKET_NAME).Object(fileName).NewWriter(c)
	wc.ContentType = mimeType
	if _, err := wc.Write(downloadBody); err != nil {
		fmt.Fprint(w, "画像の格納に失敗しました。" + err.Error())
		return
	}
	if err := wc.Close(); err != nil {
		fmt.Fprint(w, "画像の格納に失敗しました。" + err.Error())
		return
	}
	log.Debugf(c, "Cloud Storage に画像を格納完了しました。")


	//Cloud Vision API を使う
	log.Debugf(c, "Cloud Vision APIの利用を開始します。")
	cloudVisionClient := urlfetch.Client(c)
	bodyTipe := "application/json"
	requestJson := "{ \"requests\": [ { \"features\": [ { \"maxResults\": " + MAX_RESULTS + ", \"type\": \"" + featureType + "\" } ], \"image\": { \"source\": { \"gcsImageUri\": \"gs://" + BUCKET_NAME + "/" + fileName + "\" } } } ] }"
	log.Debugf(c, "Cloud Vision APIのリクエストパラメータ :" + requestJson)
	bs := []byte(requestJson)
	buf := bytes.NewBuffer(bs)
	responseCloudVision, err := cloudVisionClient.Post(VISION_API_PATH, bodyTipe, buf)
	cloudvisionData, err := ioutil.ReadAll(responseCloudVision.Body)
	if err != nil {
		fmt.Fprint(w, "Cloud Vision API のリクエストに失敗しました。" + err.Error())
		return
	}
	responseStr := string(cloudvisionData)
	log.Debugf(c, "Cloud Vision APIの利用を完了しました。 レスポンスデータ：" + responseStr)

	//キャッシュデータとしてレスポンスを格納する。キャッシュ無効時には、ダウンロードした画像を削除。
	if (cache == "1") {
		log.Debugf(c, "レスポンスデータをキャッシュします。")
		cacheClient, err := storage.NewClient(c)
		if err != nil {
			fmt.Fprint(w, "レスポンスデータのキャッシュに失敗しました。" + err.Error())
			return
		}
		////レスポンスデータをCloud Storage に格納
		wc = cacheClient.Bucket(BUCKET_NAME).Object(fileName + "_.json").NewWriter(c)
		wc.ContentType = "application/json"
		if _, err := wc.Write(cloudvisionData); err != nil {
			fmt.Fprint(w, "レスポンスデータのキャッシュに失敗しました。" + err.Error())
			return
		}
		if err := wc.Close(); err != nil {
			fmt.Fprint(w, "レスポンスデータのキャッシュに失敗しました。" + err.Error())
			return
		}
		log.Debugf(c, "レスポンスデータのキャッシュが完了しました。")
	} else {
		//CloudVison APIで利用したファイルを削除
		imageClient, err := storage.NewClient(c)
		err = imageClient.Bucket(BUCKET_NAME).Object(fileName).Delete(c)
		if err == nil {
			log.Debugf(c, "利用した画像ファイルを削除しました。BUCKET_NAME : " + BUCKET_NAME + " : " + fileName)
		} else {
			log.Debugf(c, "利用した画像ファイルを削除できませんでした。" + err.Error())
		}
	}

	//Cloud Vision API の レスポンスを返す callback関数を指定されていた場合はラップする（JSONP形式で返却）
	if callback != "" {
		fmt.Fprint(w, callback + "(" + responseStr + ");")
	} else {
		fmt.Fprint(w, responseStr)
	}
	return
}
