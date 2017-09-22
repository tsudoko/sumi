package ocr

import (
	"gopkg.in/GeertJohan/go.leptonica.v1"
	//"gopkg.in/GeertJohan/go.tesseract.v1"
	"github.com/tsudoko/go.tesseract"
)

func Detect(t *tesseract.Tess, path string) ([][]rune, error) {
	p, err := leptonica.NewPixFromFile(path)
	if err != nil {
		return [][]rune{}, err
	}
	defer p.Close()

	w, h, _, err := p.GetDimensions()
	if err != nil {
		return [][]rune{}, err
	}

	if w > h {
		t.SetPageSegMode(tesseract.PSM_SINGLE_BLOCK)
	} else {
		t.SetPageSegMode(tesseract.PSM_SINGLE_BLOCK_VERT_TEXT)
	}

	t.SetImagePix(p)
	t.Recognize()

	ri, err := t.Iterator()
	if err != nil {
		return [][]rune{}, err
	}
	level := tesseract.RIL_SYMBOL
	var matches [][]rune

	for {
		if _, err = ri.Text(level); err != nil {
			break
		}

		var cur []rune

		ci := ri.ChoiceIterator()
		for {
			c, err := ci.Text()
			if err != nil {
				break
			}

			cur = append(cur, []rune(c)[0])

			if !ci.Next() {
				break
			}
		}

		matches = append(matches, cur)
		if !ri.Next(level) {
			break
		}
	}

	t.Clear()
	return matches, nil
}
