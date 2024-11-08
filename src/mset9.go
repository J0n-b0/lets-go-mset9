package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func mset9() { // Main MSET9 Page
	_, err := os.OpenFile(filepath.Join(id1, "extdata", "002F003A.txt"), os.O_RDONLY, 0644)
	if os.IsExist(err) {
		injectionStatus = "Yes"
	}
	mset9Header := widget.NewLabel(fmt.Sprintf(`
	System Type: %s
	Inital Setup Started?: %s
	Trigger File Injected?: %s
	
	Setup MSET9 - Start MSET9 process.
	Inject Trigger File - Create exploit trigger. Do NOT enable this before specified.
	Remove MSET9 - Remove MSET9 from SD. Don't forget this step. 
	`, sysType, mset9Started, injectionStatus))

	mset9Screen.RemoveAll()
	mset9Screen.Add(mset9Header)
	mset9Screen.Add(widget.NewButton("Setup MSET9", setupMset9))
	mset9Screen.Add(widget.NewButton("Inject Trigger File", inject))
	mset9Screen.Add(widget.NewButton("Remove MSET9", func() {
		removeMset9(true, true)
		dialog.ShowInformation("Success!", "Enjoy your console :)", window)
	}))

	window.SetContent(mset9Screen)
}

func setupMset9() {

	// Insert Embedded Files
	fileCount := 0
	shas := [4]string{
		"10e68c74cdd84141de64cf7f47d1c3a5c2aec17d37b6e8b5d4ab1cda622454b6",
		"22af6174c8b055cf3a9c5d7d35bcc26a6188f65fa51ba176d8a7dda23861dc28",
		"9c95ef995e34f7fce2dab7bf3d6ff19d952837d032586a11256f5ae0b2a3fee6",
		"d380eff72c437e1de1bbc0398a8d55fedb36a39927c1e2fb6305b8962d48a872"} // SafeB9S.bin, b9, boot.3dsx, boot.firm

	fs.WalkDir(files, "exploit-files", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return err
		}
		if err != nil {
			return fucked("Failed to read embedded files!\nPlease reinstall this program and try again.", window)
		}

		data, err := files.ReadFile(path)
		if err != nil {
			return fucked(fmt.Sprintf("Failed to read %s!\nPlease reinstall this program and try again.", d.Name()), window)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != shas[fileCount] {
			return fucked(fmt.Sprintf("Hash mismatch on %s located on SD! Try again. If this keeps happening, format your SD Card.", d.Name()), window)
		}

		err = os.WriteFile(filepath.Join(sdRoot, d.Name()), data, 0644)
		if err != nil {
			return fucked(fmt.Sprintf("Failed to create %s! Try again.\n If this keeps happening, format your SD Card.", d.Name()), window)
		}

		data, err = os.ReadFile(filepath.Join(sdRoot, d.Name()))
		if fmt.Sprintf("%x", sha256.Sum256(data)) != shas[fileCount] {
			return fucked(fmt.Sprintf("Hash mismatch of %s located on SD! Try again. If this keeps happening, format your SD Card.", d.Name()), window)
		}
		fileCount++
		return err
	})

	// Add Hax ID1
	os.Rename(id1, id1+"_user-id1")
	os.Mkdir(haxId1, 0755)
	os.Mkdir(filepath.Join(haxId1, "dbs"), 0755)
	os.WriteFile(filepath.Join(haxId1, "dbs", "title.db"), []byte(""), 0644)
	os.WriteFile(filepath.Join(haxId1, "dbs", "import.db"), []byte(""), 0644)

	mset9Started = "Yes"
	dialog.ShowInformation("Success!", "You may now continue the guide.", window)
}

func inject() { // Includes main sanity checks
	var uhoh []string
	sanityBroken := false

	// Check if MSET9 setup is removed for some reason
	_, haxId1Err := os.ReadDir(haxId1)
	_, userId1Err := os.ReadDir(id1 + "_user-id1")

	if os.IsNotExist(haxId1Err) || os.IsNotExist(userId1Err) {
		fucked("MSET9 install removed or broken. Rerun \"MSET9 Setup\" and reread the guide", window)
		removeMset9(!os.IsNotExist(haxId1Err), !os.IsNotExist(userId1Err))
		mset9Started = "No"
		mset9()
		return
	}

	// Database Checks
	dbCheck, err := os.ReadFile(filepath.Join(haxId1, "dbs", "title.db"))
	if err != nil {
		uhoh = append(uhoh, "- Title database inaccessible\n")
		sanityBroken = true
	} else {
		titleSize := len(dbCheck)
		if titleSize != 0x31E400 {
			uhoh = append(uhoh, "- Title database not initialized or deformed\n")
			sanityBroken = true
		}
	}
	dbCheck, err = os.ReadFile(filepath.Join(haxId1, "dbs", "import.db"))
	if err != nil {
		uhoh = append(uhoh, "- Import database inaccessible\n")
		sanityBroken = true
	} else {
		importSize := len(dbCheck)
		if importSize != 0x31E400 {
			uhoh = append(uhoh, "- Import database not initialized or deformed\n")
			sanityBroken = true
		}
	}

	// Extdata Checks
	_, err = os.ReadDir(filepath.Join(haxId1, "exdata"))
	if os.IsNotExist(err) {
		uhoh = append(uhoh, "- Extdata folder missing\n")
		sanityBroken = true
	} else {
		extdCheck, err := os.ReadDir(filepath.Join(haxId1, "exdata", "00000000"))
		if os.IsNotExist(err) {
			uhoh = append(uhoh, "- Home Menu extdata not found\n")
			uhoh = append(uhoh, "- Mii Maker extdata not found\n")
			sanityBroken = true
		} else {
			for _, data := range extdCheck {
				found := false
				// Home Menu
				for _, extdata := range homeMenuExtdata {
					if data.Name() == string(extdata) {
						found = true
					}
				}
				if !found {
					uhoh = append(uhoh, "- Home Menu extdata not found\n")
					sanityBroken = true
				}
				// Mii Maker
				for _, extdata := range miiMakerExtdata {
					if data.Name() == string(extdata) {
						found = true
					}
				}
				if !found {
					uhoh = append(uhoh, "- Mii Maker extdata not found\n")
					sanityBroken = true
				}
			}
		}
	}
	if sanityBroken {
		var final string
		for _, data := range uhoh {
			final += fmt.Sprintf("\n%s", data)
		}
		final = fmt.Sprintf(
			`Requirements to inject not met!

The following sanity checks failed:
		%s`, final)

		fucked(final, window)
	} else {
		os.WriteFile(filepath.Join(haxId1, "extdata", "002F003A.txt"), []byte("get haxxed says I, the swordbearer j0n_b0!\n\n\n\nPS Bonk Gabbi"), 0644)
		injectionStatus = "Yes"
	}
	mset9()
}

func removeMset9(remHax bool, editUserId1 bool) {
	os.Remove(sdRoot + "b9")
	os.Remove(sdRoot + "SafeB9S.bin")
	if remHax {
		os.RemoveAll(haxId1)
	}
	if editUserId1 {
		os.Rename(id1+"_user-id1", id1)
	}
}
