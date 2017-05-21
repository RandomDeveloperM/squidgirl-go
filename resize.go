package main

import (
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" //image.Decode でPNGファイルを読み込むのに必要
	"io"
	"os"

	imageresize "github.com/nfnt/resize"
)

//Resize はリサイズ処理を行うデータを保持する
type Resize struct {
	maxHeight   uint
	maxWidth    uint
	height      uint
	width       uint
	jpegQuality int
}

//NewResize はリサイズ処理を行う構造体を初期値をセットして返す
func NewResize(height uint, width uint, jpegQuality int) *Resize {
	resize := new(Resize)
	resize.maxHeight = height
	resize.maxWidth = width

	h, w := resize.getResizeParam(resize.maxHeight, resize.maxWidth)
	resize.height = h
	resize.width = w
	resize.jpegQuality = jpegQuality
	return resize
}

//ResizeFile は指定した画像ファイルから設定値に従って縮小して保存する
func (resize *Resize) ResizeFile(reader io.Reader, writePath string) error {
	//画像読み込み
	image, _, err := image.Decode(reader)
	if err != nil {
		fmt.Printf("画像読み込みエラー err:%s\n", err)
		return err
	}
	resizeImage := imageresize.Resize(resize.width, resize.height, image, imageresize.Lanczos3)

	//書き込み用ファイル作成
	outFile, err := os.Create(writePath)
	if err != nil {
		fmt.Printf("ファイル作成エラー err:%s\n", err)
		return err
	}
	defer outFile.Close()

	//JPEGとして保存
	opts := &jpeg.Options{Quality: resize.jpegQuality}
	err = jpeg.Encode(outFile, resizeImage, opts)
	if err != nil {
		fmt.Printf("JPEG変換エラー err:%s\n", err)
	}

	return nil
}

//getResizeParam は指定した縦横から縦横比を維持した縮小サイズに設定するための値に変換して返す
func (resize *Resize) getResizeParam(height uint, width uint) (uint, uint) {
	if width != 0 && height != 0 {
		if width > height {
			width = 0 //横幅は自動
		} else if height > width {
			height = 0 //高さは自動
		}
	}

	return height, width
}
