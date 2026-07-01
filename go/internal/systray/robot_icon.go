//go:build windows

package systray

import (
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// GDI & User32 API declarations for icon creation and font drawing
var (
	gdi32                   = windows.NewLazySystemDLL("gdi32.dll")
	pCreateBitmap           = gdi32.NewProc("CreateBitmap")
	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pCreateFontW            = gdi32.NewProc("CreateFontW")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pSetBkMode              = gdi32.NewProc("SetBkMode")

	pCreateIconIndirect = user32.NewProc("CreateIconIndirect")
	pGetDC              = user32.NewProc("GetDC")
	pReleaseDC          = user32.NewProc("ReleaseDC")
	pDrawTextW          = user32.NewProc("DrawTextW")
)

type ICONINFO struct {
	FIcon    int32
	XHotspot uint32
	YHotspot uint32
	HbmMask  windows.Handle
	HbmColor windows.Handle
}

const (
	ICON_SIZE = 32
)

var (
	robotNormalIcon windows.Handle
	robotAlertIcon  windows.Handle
	robotIconOnce   sync.Once
)

// initRobotIcons generates unicode emoji icons programmatically.
func initRobotIcons() {
	robotIconOnce.Do(func() {
		robotNormalIcon = drawEmojiIcon("🤖") // Normal: Robot emoji
		robotAlertIcon = drawEmojiIcon("🚨")  // Activity/Alert: Flashing siren emoji
		if robotNormalIcon == 0 {
			hIcon, _, _ := pLoadIcon.Call(0, uintptr(IDI_APPLICATION))
			robotNormalIcon = windows.Handle(hIcon)
		}
		if robotAlertIcon == 0 {
			robotAlertIcon = robotNormalIcon
		}
	})
}

// drawEmojiIcon creates a 32x32 icon containing the specified unicode emoji character.
func drawEmojiIcon(emoji string) windows.Handle {
	hDC, _, _ := pGetDC.Call(0)
	if hDC == 0 {
		return 0
	}
	defer pReleaseDC.Call(0, hDC)

	hMemDC, _, _ := pCreateCompatibleDC.Call(hDC)
	if hMemDC == 0 {
		return 0
	}
	defer pDeleteDC.Call(hMemDC)

	// Create color bitmap matching display compatibility
	hbmColor, _, _ := pCreateCompatibleBitmap.Call(hDC, ICON_SIZE, ICON_SIZE)
	if hbmColor == 0 {
		return 0
	}
	defer pDeleteObject.Call(hbmColor)

	// Select color bitmap into MemDC
	hOldObj, _, _ := pSelectObject.Call(hMemDC, hbmColor)
	defer pSelectObject.Call(hMemDC, hOldObj)

	// Create monochrome mask bitmap (all white/transparent initially)
	// For icons, mask = 1 means transparent, mask = 0 means opaque.
	// Since we want transparent background, we fill it with 1s.
	maskBits := make([]byte, (ICON_SIZE*ICON_SIZE)/8)
	for i := range maskBits {
		maskBits[i] = 0xFF
	}
	hbmMask, _, _ := pCreateBitmap.Call(ICON_SIZE, ICON_SIZE, 1, 1, uintptr(unsafe.Pointer(&maskBits[0])))
	if hbmMask == 0 {
		return 0
	}
	defer pDeleteObject.Call(hbmMask)

	// Select font supporting color emojis
	fontName, _ := syscall.UTF16PtrFromString("Segoe UI Emoji")
	hFont, _, _ := pCreateFontW.Call(
		26,                     // Height
		0,                      // Width
		0, 0,
		400,                    // FW_NORMAL
		0, 0, 0,
		1,                      // DEFAULT_CHARSET (Segoe UI Emoji relies on outline font)
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(fontName)),
	)
	if hFont != 0 {
		pSelectObject.Call(hMemDC, hFont)
		defer pDeleteObject.Call(hFont)
	}

	pSetBkMode.Call(hMemDC, 1) // TRANSPARENT

	rect := struct {
		Left, Top, Right, Bottom int32
	}{0, 0, ICON_SIZE, ICON_SIZE}

	utf16Str, _ := syscall.UTF16FromString(emoji)

	// DT_CENTER = 1, DT_VCENTER = 4, DT_SINGLELINE = 32
	pDrawTextW.Call(
		hMemDC,
		uintptr(unsafe.Pointer(&utf16Str[0])),
		uintptr(len(utf16Str)-1),
		uintptr(unsafe.Pointer(&rect)),
		1|4|32,
	)

	// Recreate the monochrome mask based on drawn pixels
	// (Any non-black pixel in hbmColor is part of the emoji, so make it opaque in mask)
	// We can let Windows handle transparency if we use a transparent bitmap, but for standard
	// system tray icons, we can also generate a mask.
	// Let's draw the emoji into a mask DC to mark drawn pixels as opaque.
	hMaskDC, _, _ := pCreateCompatibleDC.Call(0)
	if hMaskDC != 0 {
		hOldMask, _, _ := pSelectObject.Call(hMaskDC, hbmMask)
		if hFont != 0 {
			pSelectObject.Call(hMaskDC, hFont)
		}
		pSetBkMode.Call(hMaskDC, 1)
		
		// Fill mask bitmap with white (transparent = 1)
		// Draw emoji in black text (opaque = 0) to mask
		pSetTextColor := gdi32.NewProc("SetTextColor")
		pSetTextColor.Call(hMaskDC, 0x00000000)
		
		pDrawTextW.Call(
			hMaskDC,
			uintptr(unsafe.Pointer(&utf16Str[0])),
			uintptr(len(utf16Str)-1),
			uintptr(unsafe.Pointer(&rect)),
			1|4|32,
		)
		
		pSelectObject.Call(hMaskDC, hOldMask)
		pDeleteDC.Call(hMaskDC)
	}

	// Create the icon from the two bitmaps
	var ii ICONINFO
	ii.FIcon = 1
	ii.HbmMask = windows.Handle(hbmMask)
	ii.HbmColor = windows.Handle(hbmColor)

	hIcon, _, _ := pCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	return windows.Handle(hIcon)
}

func getNormalIcon() windows.Handle {
	initRobotIcons()
	return robotNormalIcon
}

func getAlertIcon() windows.Handle {
	initRobotIcons()
	return robotAlertIcon
}
