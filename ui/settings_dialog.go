package ui

import (
	"GO-Clock/config"
	"GO-Clock/utils"
	"embed"
	"image/color"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"os"

	"github.com/lusingander/colorpicker"
)

func ShowSettingsDialog(a fyne.App, cfg *config.Config, saveConfig func(), onConfigChanged func(), loadResource, fileResource, playResource, stopResource, refreshResource fyne.Resource, alarmsFS embed.FS) {
	win := a.NewWindow("Settings")
	initialLabel := "No Alarm Selected"
	if cfg.Alarm.SoundFile != "" {
		// Strip path and add suffix
		filename := filepath.Base(cfg.Alarm.SoundFile)
		suffix := " (user)"
		_, err := alarmsFS.ReadFile("alarms/" + filename)
		if err == nil {
			suffix = " (embedded)"
		}
		initialLabel = filename + suffix
	}
	selectedAudioLabel := widget.NewLabel("Selected: " + initialLabel)
	selectedAudioLabel.Wrapping = fyne.TextTruncate

	// Sounds list
	var soundFiles []string
	refreshSoundFiles := func() {
		soundFiles = nil
		entries, err := alarmsFS.ReadDir("alarms")
		if err != nil {
			ShowFatalError(a, "Failed to read alarms directory: "+err.Error())
			return
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				soundFiles = append(soundFiles, entry.Name())
			}
		}

		soundFiles = append(soundFiles, "___________ USER DATA ___________")

		alarmsDir, err := config.GetAlarmsDir()
		if err == nil {
			userEntries, err := os.ReadDir(alarmsDir)
			if err == nil {
				for _, entry := range userEntries {
					if !entry.IsDir() {
						soundFiles = append(soundFiles, entry.Name())
					}
				}
			}
		}
	}
	refreshSoundFiles()

	// List to select sounds
	soundsList := widget.NewList(
		func() int {
			return len(soundFiles)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*widget.Label).SetText(soundFiles[id])
		},
	)
	soundsList.OnSelected = func(id widget.ListItemID) {
		selected := soundFiles[id]
		if selected == "___________ USER DATA ___________" {
			soundsList.Unselect(id)
			return
		}

		// Load into memory
		var data []byte
		var err error
		isEmbedded := false

		// Try loading from embedded FS
		data, err = alarmsFS.ReadFile("alarms/" + selected)
		if err == nil {
			isEmbedded = true
		} else {
			// If not found in embedded, try loading from user dir
			alarmsDir, errDir := config.GetAlarmsDir()
			if errDir == nil {
				data, err = os.ReadFile(filepath.Join(alarmsDir, selected))
			}
		}

		if err == nil {
			cfg.Alarm.SoundFile = selected
			cfg.Alarm.IsUser = !isEmbedded
			LoadSound(selected, data)

			suffix := " (user)"
			if isEmbedded {
				suffix = " (embedded)"
			}
			selectedAudioLabel.SetText("Selected: " + selected + suffix)
			saveConfig()
		}
	}

	// Pre-load current sound if exists
	if cfg.Alarm.SoundFile != "" {
		for i, name := range soundFiles {
			if name == cfg.Alarm.SoundFile {
				soundsList.Select(i)
				break
			}
		}

		// Load into memory
		var data []byte
		var err error

		// Try loading from embedded FS
		data, err = alarmsFS.ReadFile("alarms/" + cfg.Alarm.SoundFile)
		if err != nil {
			// If not found in embedded, try loading from user dir
			alarmsDir, errDir := config.GetAlarmsDir()
			if errDir == nil {
				data, err = os.ReadFile(filepath.Join(alarmsDir, cfg.Alarm.SoundFile))
			}
		}

		if err == nil {
			LoadSound(cfg.Alarm.SoundFile, data)
		}
	}

	var playBtn *widget.Button
	playBtn = widget.NewButtonWithIcon("", playResource, func() {
		if playBtn.Icon == playResource {
			PlaySound(cfg.Alarm.SoundFile, true)
			playBtn.SetIcon(stopResource)
		} else {
			StopSound()
			playBtn.SetIcon(playResource)
		}
	})

	loadBtn := widget.NewButtonWithIcon("", loadResource, func() {
		// File dialog to add a new sound
		d := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				// Copy to Alarms dir
				newPath, errCopy := config.CopyAlarmSound(reader.URI().Path())
				if errCopy != nil {
					ShowFatalError(a, "Failed to copy alarm sound: "+errCopy.Error())
					return
				}
				cfg.Alarm.SoundFile = filepath.Base(newPath)
				cfg.Alarm.IsUser = true
				saveConfig()
				// Update label
				suffix := " (user)"
				selectedAudioLabel.SetText("Selected: " + cfg.Alarm.SoundFile + suffix)
				refreshSoundFiles()
				soundsList.Refresh()
			}
		}, win)
		d.Resize(fyne.NewSize(800, 600))
		alarmsDir, errDir := config.GetAlarmsDir()
		if errDir != nil {
			ShowFatalError(a, "Failed to get alarms directory: "+errDir.Error())
			return
		}
		uri := storage.NewFileURI(alarmsDir)
		lister, errLister := storage.ListerForURI(uri)
		if errLister != nil {
			ShowFatalError(a, "Failed to create lister for URI: "+errLister.Error())
			return
		}
		d.SetLocation(lister)
		d.Show()
	})

	refreshBtn := widget.NewButtonWithIcon("", refreshResource, func() {
		refreshSoundFiles()
		soundsList.Refresh()
	})

	// Container for buttons
	buttonContainer := container.New(layout.NewGridLayout(3), refreshBtn, loadBtn, playBtn)

	// Helper to add border
	borderWrapper := func(content fyne.CanvasObject) fyne.CanvasObject {
		rect := canvas.NewRectangle(color.Transparent)
		rect.StrokeColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
		rect.StrokeWidth = 1
		return container.NewStack(rect, content)
	}

	leftPanel := container.NewBorder(borderWrapper(selectedAudioLabel), buttonContainer, nil, nil, borderWrapper(soundsList))

	// Color panel
	var colorEditorContainer fyne.CanvasObject
	picker := colorpicker.New(200, colorpicker.StyleHueCircle)
	pickerContainer := container.NewCenter(container.New(layout.NewGridWrapLayout(fyne.NewSize(210, 210)), picker))

	currentBox := canvas.NewRectangle(color.Transparent)
	currentBox.SetMinSize(fyne.NewSize(50, 20))
	currentBox.StrokeColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	currentBox.StrokeWidth = 1
	newBox := canvas.NewRectangle(color.Transparent)
	newBox.SetMinSize(fyne.NewSize(50, 20))
	newBox.StrokeColor = color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	newBox.StrokeWidth = 1
	currentLabel := widget.NewLabel("Current Color")
	newLabel := widget.NewLabel("New Color")

	// Keep track of which color is being edited
	var currentColorTarget *string
	var selectedColor color.Color
	var originalColor color.Color
	picker.SetOnChanged(func(c color.Color) {
		selectedColor = c
		newBox.FillColor = c
		newBox.Refresh()
	})

	defaultBtn := widget.NewButton("Default", func() {
		if currentColorTarget != nil {
			var defaultHex string
			if currentColorTarget == &cfg.Color.Background {
				defaultHex = config.DefaultColorConfig.Background
			} else if currentColorTarget == &cfg.Color.Digits {
				defaultHex = config.DefaultColorConfig.Digits
			} else if currentColorTarget == &cfg.Color.Sidebar {
				defaultHex = config.DefaultColorConfig.Sidebar
			}
			*currentColorTarget = defaultHex
			col, err := utils.ParseHexColor(defaultHex)
			if err != nil {
				ShowFatalError(a, "Failed to parse default color: "+err.Error())
				return
			}
			picker.SetColor(col)
			currentBox.FillColor = col
			newBox.FillColor = col
			currentBox.Refresh()
			newBox.Refresh()
			saveConfig()
			onConfigChanged()
			colorEditorContainer.Hide()
		}
	})

	revertBtn := widget.NewButton("Revert", func() {
		if currentColorTarget != nil {
			picker.SetColor(originalColor)
			currentBox.FillColor = originalColor
			newBox.FillColor = originalColor
			currentBox.Refresh()
			newBox.Refresh()
			hex := utils.ToHexColor(color.RGBAModel.Convert(originalColor).(color.RGBA))
			*currentColorTarget = hex
			saveConfig()
			onConfigChanged()
			colorEditorContainer.Hide()
		}
	})

	applyBtn := widget.NewButton("Apply", func() {
		if currentColorTarget != nil {
			hex := utils.ToHexColor(color.RGBAModel.Convert(selectedColor).(color.RGBA))
			*currentColorTarget = hex
			currentBox.FillColor = selectedColor
			currentBox.Refresh()
			saveConfig()
			onConfigChanged()
			colorEditorContainer.Hide()
		}
	})

	buttonBox := container.NewHBox(defaultBtn, revertBtn, applyBtn)
	colorEditorContainer = container.NewVBox(
		currentLabel,
		currentBox,
		newLabel,
		newBox,
		pickerContainer,
		buttonBox,
	)
	colorEditorContainer.Hide()

	colorPanel := container.NewScroll(container.NewVBox(
		widget.NewLabel("Colors"),
		widget.NewButton("Background Color", func() {
			currentColorTarget = &cfg.Color.Background
			col, err := utils.ParseHexColor(cfg.Color.Background)
			if err != nil {
				ShowFatalError(a, "Failed to parse background color: "+err.Error())
				return
			}
			originalColor = col
			picker.SetColor(col)
			currentBox.FillColor = col
			newBox.FillColor = col
			currentBox.Refresh()
			newBox.Refresh()
			currentLabel.SetText("Current Background Color")
			newLabel.SetText("New Background Color")
			colorEditorContainer.Show()
		}),
		widget.NewButton("Digits Color", func() {
			currentColorTarget = &cfg.Color.Digits
			col, err := utils.ParseHexColor(cfg.Color.Digits)
			if err != nil {
				ShowFatalError(a, "Failed to parse digit color: "+err.Error())
				return
			}
			originalColor = col
			picker.SetColor(col)
			currentBox.FillColor = col
			newBox.FillColor = col
			currentBox.Refresh()
			newBox.Refresh()
			currentLabel.SetText("Current Digits Color")
			newLabel.SetText("New Digits Color")
			colorEditorContainer.Show()
		}),
		widget.NewButton("Sidebar Color", func() {
			currentColorTarget = &cfg.Color.Sidebar
			col, err := utils.ParseHexColor(cfg.Color.Sidebar)
			if err != nil {
				ShowFatalError(a, "Failed to parse sidebar color: "+err.Error())
				return
			}
			originalColor = col
			picker.SetColor(col)
			currentBox.FillColor = col
			newBox.FillColor = col
			currentBox.Refresh()
			newBox.Refresh()
			currentLabel.SetText("Current Sidebar Color")
			newLabel.SetText("New Sidebar Color")
			colorEditorContainer.Show()
		}),
		colorEditorContainer,
	))

	content := container.NewGridWithColumns(2, leftPanel, colorPanel)

	win.SetContent(content)
	win.Resize(fyne.NewSize(800, 600))
	win.Show()
}
