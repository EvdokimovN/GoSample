package writer

/* Copyright (C) Nikita Evdokimov - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited
 * Proprietary and confidential
 * Written by Nikita Evokimov <nevdokimovm@gmail.com>, 2017
 */
import (
	"image/draw"
	"io/ioutil"

	"image"

	"fmt"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/math/fixed"
)

const spacing = 1.5

type TextDrawer interface {
	SetFontFile(string) error
	SetFont([]byte) error
	DrawString(draw.Image, int, int, int, int, []string) error
	DrawStringCenter(dst draw.Image, paddingLeft, paddingRight int, text []string) error
}

type Texter struct {
	context  *freetype.Context
	font     *truetype.Font
	fontSize float64
}

//SetFontFile reads font from file and sets it as Texter font
func (t *Texter) SetFontFile(fontfile string) error {
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		return err
	}
	return t.SetFont(fontBytes)
}

//SetFont used to draw text on image
func (t *Texter) SetFont(fontBytes []byte) error {
	f, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return err
	}
	t.context.SetFont(f)
	t.font = f
	return nil
}

func (t *Texter) SetFontSize(fontSize float64) {
	t.context.SetFontSize(fontSize)
	t.fontSize = fontSize
}

func (t Texter) DrawStringCenter(dst draw.Image, paddingLeft, paddingRight int, text []string) error {
	l := len(text)
	rq := dst.Bounds()
	height := t.context.PointToFixed(float64(rq.Dy()))
	var textHeight fixed.Int26_6
	for l > 0 {
		textHeight += t.linearStep(spacing)
		l--
	}
	spaceLeft := height - textHeight
	if spaceLeft < 0 {
		return fmt.Errorf("no space left to center")
	}
	pad := spaceLeft.Mul(t.context.PointToFixed(0.5))
	return t.DrawString(dst, paddingLeft, paddingRight, pad.Floor(), pad.Floor(), text)
}

func (t Texter) DrawString(dst draw.Image, paddingLeft, paddingRight, paddingTop, paddingBottom int, text []string) error {
	rq := dst.Bounds()
	width := t.context.PointToFixed(float64(rq.Dx()))
	height := t.context.PointToFixed(float64(rq.Dy()))
	// TODO: (evdokimovn) Refactor
	t.context.SetClip(rq)
	t.context.SetDst(dst)
	t.context.SetSrc(image.Black)

	//pT := t.context.PointToFixed(float64(paddingTop))
	pB := t.context.PointToFixed(float64(paddingBottom))
	pL := t.context.PointToFixed(float64(paddingLeft))
	pR := t.context.PointToFixed(float64(paddingRight))
	textSpace := width - (pL + pR)

	// Draw the text.
	pt := freetype.Pt(paddingLeft, paddingTop+int(t.context.PointToFixed(t.fontSize)>>6))
	for _, s := range text {
		advance := t.calculateAdvance(s)
		// Since left center text
		newPositionX := pL + advance
		// For now simply skip such long strings
		if (width - newPositionX) < pR {
			overFlow := advance - textSpace
			meanCharWidth := advance.Ceil() / len(s)
			n := overFlow.Floor() / meanCharWidth

			s = fmt.Sprintf("%s...", s[:len(s)-n])
		}
		_, err := t.context.DrawString(s, pt)
		if err != nil {
			return err
		}
		pt.Y += t.linearStep(spacing)
		if pt.Y > height-pB {
			break
		}
	}
	return nil
}

func (t Texter) linearStep(spacing float64) fixed.Int26_6 {
	return t.context.PointToFixed(t.fontSize * spacing)
}

func NewTextDrawer(dpi, size float64) TextDrawer {
	c := freetype.NewContext()
	c.SetDPI(dpi)
	fg := image.White
	c.SetSrc(fg)
	c.SetFontSize(size)
	return &Texter{context: c, fontSize: size}
}

func (t Texter) calculateAdvance(line string) fixed.Int26_6 {
	var l fixed.Int26_6
	// It looks like scale is simply size
	scale := t.context.PointToFixed(t.fontSize)
	for _, c := range line {
		h := t.font.HMetric(scale, t.font.Index(c))
		l += h.AdvanceWidth
	}
	return l
}
