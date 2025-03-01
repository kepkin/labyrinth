package image

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"

	lab "github.com/kepkin/labyrinth"
)

type PlayerMap struct {
	cmap *CellMap
	pmap *lab.PlayerMap
}

func NewPlayerMap(cmap *CellMap, playerMap *lab.PlayerMap) *PlayerMap {
	return &PlayerMap{cmap: cmap, pmap: playerMap}
}

func (pm *PlayerMap) Bounds() image.Rectangle {
	w, h := pm.pmap.Rect()

	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: w * pm.cmap.cellSize.X,
			Y: h * pm.cmap.cellSize.Y,
		},
	}
}

func (PlayerMap) ColorModel() color.Model {
	return color.RGBAModel
}

func (pm *PlayerMap) colorToCellMapPos(x, y int) (int, int) {
	xx := x + pm.pmap.LeftCorner.X*pm.cmap.cellSize.X
	yy := y + pm.pmap.LeftCorner.Y*pm.cmap.cellSize.Y
	return xx, yy
}

func (pm *PlayerMap) At(x, y int) color.Color {
	xx, yy := pm.colorToCellMapPos(x, y)

	p, _, _ := pm.cmap.colorToCellMapPos(xx, yy)
	if _, ok := pm.pmap.KnonwnCells[p]; ok {
		return pm.cmap.At(xx, yy)
	}

	return color.Black
}

type BlackImage struct {
	Width  int
	Height int
}

func (b BlackImage) Bounds() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: b.Width,
			Y: b.Height,
		},
	}
}

func (BlackImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (BlackImage) At(x, y int) color.Color {
	return color.Black
}

type CellMap struct {
	cmap     *lab.CellMap
	cellSize image.Point
	textures map[string]image.Image
}

func (cm *CellMap) Bounds() image.Rectangle {
	rows := cm.cmap.Rows()
	cols := cm.cmap.Cols()

	return image.Rectangle{
		Min: image.Point{},
		Max: image.Point{
			X: cols * cm.cellSize.X,
			Y: rows * cm.cellSize.Y,
		},
	}
}

func (cm *CellMap) ColorModel() color.Model {
	return cm.textures[lab.CellWall].ColorModel()
}

func (cm *CellMap) colorToCellMapPos(x, y int) (lab.Position, int, int) {
	p := lab.Position{X: x / cm.cellSize.X, Y: y / cm.cellSize.Y}

	xx := x - p.X*cm.cellSize.X
	yy := y - p.Y*cm.cellSize.Y
	return p, xx, yy
}

func (cm *CellMap) At(x, y int) color.Color {
	p, xx, yy := cm.colorToCellMapPos(x, y)

	cell := cm.cmap.Get(p)

	texture, ok := cm.textures[cell.Class]
	if !ok {
		texture = cm.textures[lab.CellWall]
	}
	return texture.At(xx, yy)
}

func NewCellMapImage(cmap *lab.CellMap) (*CellMap, error) {
	textureSize := 0

	res := &CellMap{
		cmap:     cmap,
		textures: map[string]image.Image{},
	}

	fileMap := map[string]string{
		lab.CellEarth:    "./image/Grass_01.jpg",
		lab.CellRiver:    "./image/water.jpg",
		lab.CellWall:     "./image/stone2.jpg",
		lab.CellWormHole: "./image/Grass_portal.jpg",
	}

	for key, path := range fileMap {
		var err error
		f, err := os.Open(path)

		if err != nil {
			return nil, err
		}

		res.textures[key], err = jpeg.Decode(f)
		if textureSize == 0 {
			textureSize = res.textures[key].Bounds().Dx()
		} else if textureSize != res.textures[key].Bounds().Dx() {
			panic("texture has different sizes")
		}

		if err != nil {
			return nil, err
		}

	}

	res.textures[lab.CellExit] = res.textures[lab.CellEarth]
	res.textures["unknown"] = &BlackImage{Width: textureSize, Height: textureSize}

	res.cellSize = image.Point{textureSize, textureSize}

	return res, nil
}
