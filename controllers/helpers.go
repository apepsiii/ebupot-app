package controllers

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/big"
	"os"
	"time"

	"ebupot-app/config"

	xdraw "golang.org/x/image/draw"
)

const defaultLogoPath = "uploads/logo.png"

func logoPath() string {
	if config.Cfg != nil && config.Cfg.Upload.LogoPath != "" {
		return config.Cfg.Upload.LogoPath
	}
	return defaultLogoPath
}

const alphanumericSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomAlphanumeric(n int) string {
	b := make([]byte, n)
	max := big.NewInt(int64(len(alphanumericSet)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			b[i] = alphanumericSet[0]
			continue
		}
		b[i] = alphanumericSet[idx.Int64()]
	}
	return string(b)
}

// generateEbupotFileName menghasilkan nama file sesuai format:
// M_01-DOC002_Ebupot_2126_BPA1_25087FU42
//   M_01   -> M_ + UserID (2 digit, zero-padded)
//   DOC002 -> DOC + nomor dokumen user (3 digit, zero-padded)
//   2126   -> tanggal(1-2) + bulan(1-2) + tahun(2 digit)
//   BPA1   -> fixed
//   25087FU42 -> 9 karakter random alfanumerik
func generateEbupotFileName(userID uint, docNumber int, bulan, tahun int) string {
	uidPart := fmt.Sprintf("M_%02d", userID)
	docPart := fmt.Sprintf("DOC%03d", docNumber)
	tanggal := time.Now().Day()
	tahun2 := tahun % 100
	datePart := fmt.Sprintf("%d%d%02d", tanggal, bulan, tahun2)
	randPart := randomAlphanumeric(9)
	return fmt.Sprintf("%s-%s_Ebupot_%s_BPA1_%s", uidPart, docPart, datePart, randPart) + ".pdf"
}

// logoExists memeriksa apakah file logo sudah diunggah.
func logoExists() bool {
	if _, err := os.Stat(logoPath()); err == nil {
		return true
	}
	return false
}

// overlayLogoOnQR menempelkan logo di tengah QR code.
// Jika logo tidak ada atau gagal decode, mengembalikan QR asli.
func overlayLogoOnQR(qrPNG []byte) []byte {
	if !logoExists() {
		return qrPNG
	}

	qrImg, err := png.Decode(bytes.NewReader(qrPNG))
	if err != nil {
		return qrPNG
	}

	logoFile, err := os.Open(logoPath())
	if err != nil {
		return qrPNG
	}
	defer logoFile.Close()

	logoImg, _, err := image.Decode(logoFile)
	if err != nil {
		return qrPNG
	}

	bounds := qrImg.Bounds()
	qrSize := bounds.Dx()
	// Logo ~22% dari ukuran QR agar tidak merusak pembacaan
	logoSize := qrSize * 22 / 100
	if logoSize < 24 {
		logoSize = 24
	}

	canvas := image.NewRGBA(bounds)
	draw.Draw(canvas, bounds, qrImg, image.Point{}, draw.Src)

	// Posisi tengah
	cx := (qrSize - logoSize) / 2
	cy := (qrSize - logoSize) / 2
	pad := logoSize / 12
	if pad < 2 {
		pad = 2
	}

	// Background putih di belakang logo untuk kontras
	bgRect := image.Rect(cx-pad, cy-pad, cx+logoSize+pad, cy+logoSize+pad)
	draw.Draw(canvas, bgRect, image.NewUniform(color.White), image.Point{}, draw.Src)

	// Resize logo
	resized := image.NewRGBA(image.Rect(0, 0, logoSize, logoSize))
	xdraw.BiLinear.Scale(resized, resized.Bounds(), logoImg, logoImg.Bounds(), xdraw.Src, nil)
	draw.Draw(canvas, image.Rect(cx, cy, cx+logoSize, logoSize+cy), resized, image.Point{}, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return qrPNG
	}
	return buf.Bytes()
}
