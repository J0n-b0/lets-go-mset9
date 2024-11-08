package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed exploit-files/*
var files embed.FS

//go:embed assets/newvold.png
var conRef embed.FS

var a fyne.App
var window, dirSel fyne.Window
var startScreen, consoleSelectScreen, mset9Screen *fyne.Container

var sdRoot string           // SD Root
var sysType string          // New or Old
var id0, id1, haxId1 string // ID0, ID1, and hax ID1 full path
var mset9Started, injectionStatus string = "No", "No"

var homeMenuExtdata = []int{0x8F, 0x98, 0x82, 0xA1, 0xA9, 0xB1} // U, E, J, C, K, T
var miiMakerExtdata = []int{0x217, 0x227, 0x207, 0x267, 0x277, 0x287}

var haxs = [2]string{ // Old and New
	"쀁ʏ＜į餑䠋䚅敩ꄇ∁䬄䞘䙨䙙ꫀᰗ䙃䰂䞠䞸ꁱࠅ캙ࠄsdmc退ࠊb9",
	"쀁ʏ＜į餑䠋䚅敩ꄇ∁䬄䞘䙨䙙ꫀᰗ䙃䰂䞠䞸ꁱࠅ칝ࠄsdmc退ࠊb9",
}

func main() {
	a = app.NewWithID("com.j0n_b0.lets-go-mset9")
	window = a.NewWindow("MSET9")
	window.Resize(fyne.NewSize(400, 400))
	window.CenterOnScreen()

	mset9Screen = container.New(layout.NewVBoxLayout())

	setupStart()
	window.SetContent(startScreen)
	window.ShowAndRun()
}

func setupStart() {
	// Start Screen
	startHeader := widget.NewLabel("MSET9")
	startHeader.Alignment = fyne.TextAlignCenter
	startReading := widget.NewLabel("This application will allow to run MSET9, an exploit for the 3DS. When you're ready, select your SD when asked.")

	dirSel = a.NewWindow("Select your SD")
	dirSel.Resize(fyne.NewSize(700, 500))
	startScreen = container.New(layout.NewVBoxLayout(), startHeader, startReading, widget.NewButton("CHOOSE SD", func() {
		dirSel.Show()
		dialog.NewFolderOpen(func(dir fyne.ListableURI, err error) {
			dirSel.Hide()
			if dir == nil {
				fucked("SD Selection Cancelled", window)
				return
			} else {
				sdRoot = dir.Path()

				// Check for Write Protection (to the best of my abilites, golang is flawed here :P)
				file, _ := os.Create(filepath.Join(sdRoot, "bonk"))
				file.Close()
				file, err := os.OpenFile(filepath.Join(sdRoot, "bonk"), os.O_RDONLY, 0644)
				if os.IsNotExist(err) {
					fucked("Write check failed.\nVerify your SD is not locked and try again. If this keeps happening, format your SD.", window)
					window.SetContent(startScreen)
					return
				}
				file.Close()
				os.Remove(filepath.Join(sdRoot, "bonk"))

				ninDir := filepath.Join(sdRoot, "Nintendo 3DS")
				// Check for Nintendo 3DS folder
				entries, err := os.ReadDir(ninDir)
				if os.IsNotExist(err) {
					fucked("Provided path missing Nintendo 3DS!\nVerify you have inserted the SD into the console once and that you selected the SD root.\nIf this keeps happening, format your SD.", window)
					return
				}
				// Check Amount of ID0s
				id0Count := 0
				for _, entry := range entries {
					if entry.IsDir() && len(entry.Name()) == 32 {
						id0Count++
						id0 = filepath.Join(ninDir, entry.Name())
					}
				}
				if id0Count != 1 {
					fucked(fmt.Sprintf("%d ID0s found, expected 1.\nPlease remove them and try again.", id0Count), window)
					return
				}
				// Check Amount of ID1s
				entries, _ = os.ReadDir(id0)
				id1Count := 0
				haxId1Present := false
				userId1Edited := false
				for _, entry := range entries {

					length := len(entry.Name())
					if length >= 9 {
						check := entry.Name()[length-9:]
						if entry.IsDir() && check == "_user-id1" {
							userId1Edited = true
							id1 = filepath.Join(id0, entry.Name()[:length-9])
						}
					}
					if entry.IsDir() && entry.Name() == string(haxs[0]) {
						haxId1Present = true
						haxId1 = filepath.Join(id0, entry.Name())
						sysType = "Old 3DS"
					}
					if entry.IsDir() && entry.Name() == string(haxs[1]) {
						haxId1Present = true
						haxId1 = filepath.Join(id0, entry.Name())
						sysType = "New 3DS"
					}

					if entry.IsDir() && len(entry.Name()) == 32 {
						id1Count++
						id1 = filepath.Join(id0, entry.Name())
					}
				}
				dirSel.Close()

				if haxId1Present || userId1Edited {
					if haxId1Present && userId1Edited { // On close and reopen
						dialog.ShowConfirm("MSET9 Already Found", "Would You like us to remove it and start over?", func(confirm bool) {
							if confirm {
								removeMset9(haxId1Present, userId1Edited)
								consoleSelectScreen = setupConsoleSelect()
								window.SetContent(consoleSelectScreen)
								window.CenterOnScreen()
							} else {
								mset9Started = "Yes"
								mset9()
							}
						}, window)
						return
					} else {
						removeMset9(haxId1Present, userId1Edited)
						consoleSelectScreen = setupConsoleSelect()
						window.SetContent(consoleSelectScreen)
						window.CenterOnScreen()
					}
				} else if id1Count != 1 {
					fucked(fmt.Sprintf("%d ID1s found, expected 1.\nPlease remove them and try again.", id1Count), window)
					return
				} else {
					consoleSelectScreen = setupConsoleSelect()
					window.SetContent(consoleSelectScreen)
					window.CenterOnScreen()
				}
			}
		}, dirSel).Show()
	}))
}

func setupConsoleSelect() *fyne.Container {
	imgData, _ := conRef.ReadFile("assets/newvold.png")
	imgRes := fyne.NewStaticResource("newvold.png", imgData)
	img := canvas.NewImageFromResource(imgRes)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(400, 400))

	selection := container.New(layout.NewGridLayout(2),
		widget.NewButton("Old 3DS", func() {
			sysType = "Old 3DS"
			haxId1 = filepath.Join(id0, haxs[0])
			mset9()
		}),
		widget.NewButton("New 3DS", func() {
			sysType = "New 3DS"
			haxId1 = filepath.Join(id0, haxs[1])
			mset9()
		}))

	return container.New(layout.NewVBoxLayout(), widget.NewLabel("What type of 3DS do you have? (See image for reference)"), img, selection)

}

func fucked(mes string, window fyne.Window) error {
	err := fmt.Errorf("%s", mes)
	dialog.ShowError(err, window)
	return err
}
