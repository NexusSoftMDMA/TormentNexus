//go:build windows

package systray

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Additional GDI API declarations for icon creation
var (
	gdi32            = windows.NewLazySystemDLL("gdi32.dll")
	pCreateBitmap    = gdi32.NewProc("CreateBitmap")
	pCreateCompatibleDC = gdi32.NewProc("CreateCompatibleDC")
	pSelectObject    = gdi32.NewProc("SelectObject")
	pDeleteDC        = gdi32.NewProc("DeleteDC")
	pDeleteObject    = gdi32.NewProc("DeleteObject")
	pSetPixel        = gdi32.NewProc("SetPixel")
	pCreateIconIndirect = user32.NewProc("CreateIconIndirect")
	pGetDC           = user32.NewProc("GetDC")
	pReleaseDC       = user32.NewProc("ReleaseDC")
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

// initRobotIcons generates robot face icons programmatically.
func initRobotIcons() {
	robotIconOnce.Do(func() {
		robotNormalIcon = drawRobotFace(0x00B4FF) // bright blue
		robotAlertIcon = drawRobotFace(0xFF4444)  // alert red
		if robotNormalIcon == 0 {
			hIcon, _, _ := pLoadIcon.Call(0, uintptr(IDI_APPLICATION))
			robotNormalIcon = windows.Handle(hIcon)
		}
		if robotAlertIcon == 0 {
			robotAlertIcon = robotNormalIcon
		}
	})
}

// drawRobotFace creates a 32x32 icon of a simple robot face in the given color.
// Returns 0 on failure.
func drawRobotFace(color uint32) windows.Handle {
	// Create a monochrome mask bitmap (1 = transparent, 0 = opaque)
	maskBits := make([]byte, (ICON_SIZE*ICON_SIZE)/8)
	// Create a color bitmap (32bpp BGRA)
	colorBits := make([]uint32, ICON_SIZE*ICON_SIZE)

	// Draw the robot face pixel by pixel
	// Pattern: rounded rectangle head with eyes and mouth
	for y := 0; y < ICON_SIZE; y++ {
		for x := 0; x < ICON_SIZE; x++ {
			px := isRobotPixel(x, y)
			bitIdx := y*ICON_SIZE + x
			byteIdx := bitIdx / 8
			bitOffset := uint(bitIdx % 8)

			if px {
				// Opaque pixel
				// Mask bit = 0 (opaque)
				maskBits[byteIdx] &^= (1 << bitOffset)
				// Color pixel in the given color (BGRA format)
				r := byte((color >> 16) & 0xFF)
				g := byte((color >> 8) & 0xFF)
				b := byte(color & 0xFF)
				colorBits[bitIdx] = uint32(b) | uint32(g)<<8 | uint32(r)<<16 | 0xFF000000
			} else {
				// Transparent pixel
				// Mask bit = 1 (transparent)
				maskBits[byteIdx] |= (1 << bitOffset)
				colorBits[bitIdx] = 0
			}
		}
	}

	// Create the color bitmap (using a DIB section approach)
	// Since we need raw pixel access, we use CreateBitmap with a byte array
	// Then SetBitmapBits to set the pixels
	
	// Create monochrome mask bitmap
	// Each row must be DWORD-aligned
	maskStride := ((ICON_SIZE + 31) / 32) * 4
	maskPadded := make([]byte, maskStride*ICON_SIZE)
	for y := 0; y < ICON_SIZE; y++ {
		srcStart := y * (ICON_SIZE / 8)
		dstStart := y * maskStride
		copy(maskPadded[dstStart:], maskBits[srcStart:srcStart+ICON_SIZE/8])
	}

	hbmMask, _, _ := pCreateBitmap.Call(
		ICON_SIZE, ICON_SIZE, 1, 1,
		uintptr(unsafe.Pointer(&maskPadded[0])),
	)
	if hbmMask == 0 {
		return 0
	}

	// For the color bitmap, we create a 32bpp bitmap and set its bits
	// Color bitmap: 32 bits per pixel
	colorStride := ICON_SIZE * 4
	colorRow := make([]byte, colorStride)
	var allColorData []byte

	for y := 0; y < ICON_SIZE; y++ {
		for x := 0; x < ICON_SIZE; x++ {
			c := colorBits[y*ICON_SIZE+x]
			colorRow[x*4+0] = byte(c & 0xFF)         // B
			colorRow[x*4+1] = byte((c >> 8) & 0xFF)  // G
			colorRow[x*4+2] = byte((c >> 16) & 0xFF) // R
			colorRow[x*4+3] = byte((c >> 24) & 0xFF) // A
		}
		allColorData = append(allColorData, colorRow...)
	}

	hbmColor, _, _ := pCreateBitmap.Call(
		ICON_SIZE, ICON_SIZE, 1, 32,
		uintptr(unsafe.Pointer(&allColorData[0])),
	)
	if hbmColor == 0 {
		pDeleteObject.Call(hbmMask)
		return 0
	}

	// Create the icon from the two bitmaps
	var ii ICONINFO
	ii.FIcon = 1 // true = icon
	ii.HbmMask = windows.Handle(hbmMask)
	ii.HbmColor = windows.Handle(hbmColor)

	hIcon, _, _ := pCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	if hIcon == 0 {
		pDeleteObject.Call(hbmMask)
		pDeleteObject.Call(hbmColor)
		return 0
	}

	return windows.Handle(hIcon)
}

// isRobotPixel returns true if pixel (x,y) should be part of the robot face.
// The robot face is a 32x32 pixel-art design.
func isRobotPixel(x, y int) bool {
	// Head: rounded rectangle with 2px border, slightly inset
	// Head bounds
	hLeft, hRight := 4, 27
	hTop, hBottom := 1, 30

	// Outside head area
	if x < hLeft || x > hRight || y < hTop || y > hBottom {
		return false
	}

	// Head outline (2px border) — always draw
	if x < hLeft+2 || x > hRight-2 || y < hTop+2 || y > hBottom-2 {
		return true
	}

	// Inner face area — draw eyes and mouth
	// Left eye (x:9-13, y:8-13)
	if x >= 9 && x <= 13 && y >= 8 && y <= 13 {
		// Eye socket (hollow circle)
		if x == 9 || x == 13 || y == 8 || y == 13 {
			return true
		}
		// Pupil: 2x2 center
		if x >= 10 && x <= 12 && y >= 9 && y <= 12 {
			return false
		}
		return false
	}

	// Right eye (x:18-22, y:8-13)
	if x >= 18 && x <= 22 && y >= 8 && y <= 13 {
		if x == 18 || x == 22 || y == 8 || y == 13 {
			return true
		}
		if x >= 19 && x <= 21 && y >= 9 && y <= 12 {
			return false
		}
		return false
	}

	// Mouth area (y:20-25, x:10-21)
	if y >= 20 && y <= 25 && x >= 10 && x <= 21 {
		// Mouth outline
		if y == 20 || y == 25 || x == 10 || x == 21 {
			return true
		}
		// Antenna on top of head
		if y >= 20 && y <= 25 && x >= 14 && x <= 17 {
			return false // hollow center
		}
		return false
	}

	// Antenna (on top of head, center)
	if x >= 14 && x <= 17 && y >= 0 && y <= 3 {
		return true
	}
	// Antenna tip
	if x >= 15 && x <= 16 && y >= 0 && y <= 0 {
		return false
	}

	// Ear nubs on sides
	if (x == 3 || x == 28) && y >= 12 && y <= 18 {
		return true
	}

	return false
}

func getNormalIcon() windows.Handle {
	initRobotIcons()
	return robotNormalIcon
}

func getAlertIcon() windows.Handle {
	initRobotIcons()
	return robotAlertIcon
}
