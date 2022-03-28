package main

import (
	"embed"
	"fmt"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"time"
)

//go:embed assets/*
var EmbeddedAssets embed.FS

var (
	normal font.Face
	bold   font.Face
)

const (
	GameWidth   = 1250
	GameHeight  = 750
	PlayerSpeed = 8
)

type Sprite struct {
	pict *ebiten.Image
	xloc int
	yloc int
	dX   int
	dY   int
}

type enemySprite struct {
	pict   *ebiten.Image
	xloc   int
	yloc   int
	putout bool
}

type Game struct {
	player           Sprite
	enemy            []enemySprite
	score            int
	drawOps          ebiten.DrawImageOptions
	gameName         string
	extinguishedFire bool
}

func init() {
	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 68
	normal, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    20,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
	bold, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    40,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func (g *Game) Update() error {
	processPlayerInput(g)
	for i := 0; i < len(g.enemy); i++ {
		if g.extinguishedFire == false {
			if len(g.enemy) == 1 {
				if gotFire(g.player, g.enemy[0]) {
					g.score += 1
					g.extinguishedFire = true
					//os.Exit(0)
				}
			}
		}
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colornames.Black)
	text.Draw(screen, fmt.Sprintf("Score: %d", g.score), normal, 20, 40, color.White)
	if g.extinguishedFire {
		text.Draw(screen, "Game Over", bold, GameWidth/2-100, GameHeight/2, color.White)

	}
	g.drawOps.GeoM.Reset()
	g.drawOps.GeoM.Translate(float64(g.player.xloc), float64(g.player.yloc))
	screen.DrawImage(g.player.pict, &g.drawOps)
	for i := 0; i < len(g.enemy); i++ {
		if !g.extinguishedFire {
			g.drawOps.GeoM.Reset()
			g.drawOps.GeoM.Translate(float64(g.enemy[i].xloc), float64(g.enemy[i].yloc))
			screen.DrawImage(g.enemy[i].pict, &g.drawOps)
			if hasCollided(g.player, g.enemy[i]) {
				g.enemy[i].putout = true
				g.enemy = remove(g.enemy, i)
				g.score += 1
			}
		}
	}
}

func (g Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return GameWidth, GameHeight
}

func main() {
	fireFighting := Game{gameName: "Fire Fighting"}
	ebiten.SetWindowSize(GameWidth, GameHeight)
	ebiten.SetWindowTitle(fireFighting.gameName)
	ebiten.SetFullscreen(false)
	fireFighting.enemy = fireFighting.enemiesSlice(fireFighting.enemy)
	loadImages(&fireFighting)
	fireFighting.player.yloc = GameHeight / 2
	if err := ebiten.RunGame(&fireFighting); err != nil {
		log.Fatal("Oh no! something terrible happened and the game crashed", err)
	}
}

func loadPNGImageFromEmbedded(name string) *ebiten.Image {
	pictNames, err := EmbeddedAssets.ReadDir("assets")
	if err != nil {
		log.Fatal("failed to read embedded dir ", pictNames, " ", err)
	}
	embeddedFile, err := EmbeddedAssets.Open("assets/" + name)
	if err != nil {
		log.Fatal("failed to load embedded image ", embeddedFile, err)
	}
	rawImage, err := png.Decode(embeddedFile)
	if err != nil {
		log.Fatal("failed to load embedded image ", name, err)
	}
	gameImage := ebiten.NewImageFromImage(rawImage)
	return gameImage
}

func processPlayerInput(theGame *Game) {
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		theGame.player.dY = -PlayerSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		theGame.player.dX = PlayerSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		theGame.player.dX = -PlayerSpeed
	} else if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		theGame.player.dY = PlayerSpeed
	} else if inpututil.IsKeyJustReleased(ebiten.KeyRight) || inpututil.IsKeyJustReleased(ebiten.KeyLeft) {
		theGame.player.dX = 0
	} else if inpututil.IsKeyJustReleased(ebiten.KeyUp) || inpututil.IsKeyJustReleased(ebiten.KeyDown) {
		theGame.player.dY = 0
	}
	theGame.player.yloc += theGame.player.dY
	theGame.player.xloc += theGame.player.dX
	if theGame.player.yloc <= 0 {
		theGame.player.dY = 0
		theGame.player.yloc = 0
	} else if theGame.player.yloc+theGame.player.pict.Bounds().Size().Y > GameHeight {
		theGame.player.dY = 0
		theGame.player.yloc = GameHeight - theGame.player.pict.Bounds().Size().Y
	}
	if theGame.player.xloc <= 0 {
		theGame.player.dX = 0
		theGame.player.xloc = 0
	} else if theGame.player.xloc+theGame.player.pict.Bounds().Size().X > GameWidth {
		theGame.player.dX = 0
		theGame.player.xloc = GameWidth - theGame.player.pict.Bounds().Size().X
	}
}

func loadImages(g *Game) {
	firetruck := loadPNGImageFromEmbedded("firetruck.png")
	g.player.pict = firetruck
	fire := loadPNGImageFromEmbedded("fire.png")
	for i := 0; i < len(g.enemy); i++ {
		g.enemy[i].pict = fire
	}
}

func (g *Game) enemiesSlice(slice []enemySprite) []enemySprite {
	rand.Seed(int64(time.Now().Second()))
	newSlice := make([]enemySprite, 10)
	for i := 0; i < len(newSlice); i++ {
		newSlice[i].xloc = rand.Intn(GameWidth - 50)
		newSlice[i].yloc = rand.Intn(GameHeight - 50)
	}
	return newSlice
}

func hasCollided(player Sprite, enemy enemySprite) bool {
	canWidth, canHeight := enemy.pict.Size()
	playerWidth, playerHeight := player.pict.Size()
	if player.xloc < enemy.xloc+canWidth &&
		player.xloc+playerWidth > enemy.xloc &&
		player.yloc < enemy.yloc+canHeight &&
		player.yloc+playerHeight > enemy.yloc {
		enemy.putout = true
		return true
	}
	return false
}

func gotFire(player Sprite, enemy enemySprite) bool {
	return hasCollided(player, enemy)
}

func remove(slice []enemySprite, s int) []enemySprite {
	return append(slice[:s], slice[s+1:]...)
}
