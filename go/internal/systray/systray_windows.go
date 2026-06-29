//go:build windows

package systray

import (
	"fmt"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/tormentnexushq/tormentnexus-go/internal/eventbus"
)

// Win32 API declarations
var (
	user32           = windows.NewLazySystemDLL("user32.dll")
	shell32          = windows.NewLazySystemDLL("shell32.dll")
	kernel32         = windows.NewLazySystemDLL("kernel32.dll")

	pRegisterClassEx   = user32.NewProc("RegisterClassExW")
	pCreateWindowEx    = user32.NewProc("CreateWindowExW")
	pDefWindowProc     = user32.NewProc("DefWindowProcW")
	pGetMessage        = user32.NewProc("GetMessageW")
	pTranslateMessage  = user32.NewProc("TranslateMessage")
	pDispatchMessage   = user32.NewProc("DispatchMessageW")
	pPostQuitMessage   = user32.NewProc("PostQuitMessage")
	pShowWindow        = user32.NewProc("ShowWindow")
	pUpdateWindow      = user32.NewProc("UpdateWindow")
	pSetWindowText     = user32.NewProc("SetWindowTextW")
	pSendMessage       = user32.NewProc("SendMessageW")
	pLoadIcon          = user32.NewProc("LoadIconW")
	pDestroyWindow     = user32.NewProc("DestroyWindow")

	pShellNotifyIcon   = shell32.NewProc("Shell_NotifyIconW")
)

const (
	NIM_ADD        = 0
	NIM_MODIFY     = 1
	NIM_DELETE     = 2
	NIF_MESSAGE    = 1
	NIF_ICON       = 2
	NIF_TIP        = 4

	WM_CREATE        = 0x0001
	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_COMMAND       = 0x0111
	WM_USER          = 0x0400
	WM_TRAY          = WM_USER + 100
	WM_LBUTTONUP     = 0x0202
	WM_RBUTTONUP     = 0x0208
	WM_LBUTTONDBLCLK = 0x0203

	IDI_APPLICATION = 32512
	IDI_INFORMATION = 32516
	IDI_WARNING     = 32515

	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_VSCROLL          = 0x00200000
	ES_MULTILINE        = 0x0004
	ES_AUTOVSCROLL      = 0x0040
	ES_READONLY         = 0x0800
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     windows.Handle
	HIcon         windows.Handle
	HCursor       windows.Handle
	HbrBackground windows.Handle
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       windows.Handle
}

type NOTIFYICONDATAW struct {
	CbSize            uint32
	HWnd              windows.HWND
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             windows.Handle
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          windows.GUID
	HBalloonIcon      windows.Handle
}

// Global state
var (
	hHiddenWnd       windows.HWND
	hLogWnd          windows.HWND
	hEditWnd         windows.HWND
	hNormalIcon      windows.Handle
	hActivityIcon    windows.Handle
	activityTimeout  *time.Timer
	activityMutex    sync.Mutex
	logMutex         sync.Mutex
	logLines         []string
	maxLogLines      = 200
)

func init() {
	// Preload programmatically generated robot icons
	hNormalIcon = getNormalIcon()
	hActivityIcon = getAlertIcon()
}

// Start launches the system tray helper and message loops
func Start(eb *eventbus.EventBus) {
	go runMessageLoop()

	// Subscribe to eventbus to display logs in real-time
	eb.OnGlobal(func(ev eventbus.SystemEvent) {
		addLogLine(fmt.Sprintf("[%s] %s (Source: %s)", time.Now().Format("15:04:05"), ev.Type, ev.Source))
		// Treat any eventbus activity as I/O flow
		NotifyActivity("in")
	})
}

// NotifyActivity changes the system tray icon to show activity
func NotifyActivity(dir string) {
	activityMutex.Lock()
	defer activityMutex.Unlock()

	if hHiddenWnd == 0 {
		return
	}

	// Change icon to warning (flashing) and update tip
	updateTrayIcon(hActivityIcon, fmt.Sprintf("TormentNexus - Active I/O (%s)", dir))

	if activityTimeout != nil {
		activityTimeout.Stop()
	}

	activityTimeout = time.AfterFunc(300*time.Millisecond, func() {
		activityMutex.Lock()
		defer activityMutex.Unlock()
		updateTrayIcon(hNormalIcon, "TormentNexus (Running)")
	})
}

func addLogLine(line string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	logLines = append(logLines, line)
	if len(logLines) > maxLogLines {
		logLines = logLines[len(logLines)-maxLogLines:]
	}

	if hEditWnd != 0 {
		// Update edit window text
		fullText := ""
		for _, l := range logLines {
			fullText += l + "\r\n"
		}
		ptr, _ := syscall.UTF16PtrFromString(fullText)
		pSetWindowText.Call(uintptr(hEditWnd), uintptr(unsafe.Pointer(ptr)))
		// Scroll to end
		const EM_SETSEL = 0x00B1
		const EM_SCROLLCARET = 0x00B7
		pSendMessage.Call(uintptr(hEditWnd), EM_SETSEL, uintptr(len(fullText)), uintptr(len(fullText)))
		pSendMessage.Call(uintptr(hEditWnd), EM_SCROLLCARET, 0, 0)
	}
}

func updateTrayIcon(hIcon windows.Handle, tooltip string) {
	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hHiddenWnd
	nid.UID = 1
	nid.UFlags = NIF_ICON | NIF_TIP
	nid.HIcon = hIcon

	copy(nid.SzTip[:], syscall.StringToUTF16(tooltip))

	pShellNotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&nid)))
}

func runMessageLoop() {
	hInstance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)

	// 1. Register hidden message window class
	className, _ := syscall.UTF16PtrFromString("TormentNexusTrayClass")
	var wc WNDCLASSEXW
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(hiddenWndProc)
	wc.HInstance = windows.Handle(hInstance)
	wc.LpszClassName = className
	wc.HIcon = hNormalIcon

	pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wc)))

	// 2. Create hidden window
	hHiddenWndRaw, _, _ := pCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(className)),
		0, 0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	hHiddenWnd = windows.HWND(hHiddenWndRaw)

	// 3. Register notification icon
	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hHiddenWnd
	nid.UID = 1
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAY
	nid.HIcon = hNormalIcon
	copy(nid.SzTip[:], syscall.StringToUTF16("TormentNexus (Running)"))

	pShellNotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))

	// 4. Register log window class
	logClassName, _ := syscall.UTF16PtrFromString("TormentNexusLogClass")
	var wcLog WNDCLASSEXW
	wcLog.CbSize = uint32(unsafe.Sizeof(wcLog))
	wcLog.LpfnWndProc = syscall.NewCallback(logWndProc)
	wcLog.HInstance = windows.Handle(hInstance)
	wcLog.LpszClassName = logClassName
	wcLog.HIcon = hNormalIcon
	wcLog.HbrBackground = windows.Handle(6) // COLOR_WINDOW

	pRegisterClassEx.Call(uintptr(unsafe.Pointer(&wcLog)))

	// 5. Message Loop
	var msg struct {
		HWnd    windows.HWND
		Message uint32
		WParam  uintptr
		LParam  uintptr
		Time    uint32
		Pt      struct{ X, Y int32 }
	}

	for {
		ret, _, _ := pGetMessage.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		pDispatchMessage.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Cleanup icon on exit
	pShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&nid)))
}

func hiddenWndProc(hWnd windows.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAY:
		switch lParam {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			showLogWindow()
		}
		return 0
	case WM_DESTROY:
		pPostQuitMessage.Call(0)
		return 0
	}
	ret, _, _ := pDefWindowProc.Call(uintptr(hWnd), uintptr(msg), wParam, lParam)
	return ret
}

func showLogWindow() {
	if hLogWnd != 0 {
		// Just bring to front/restore
		const SW_RESTORE = 9
		pShowWindow.Call(uintptr(hLogWnd), SW_RESTORE)
		return
	}

	hInstance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)
	logClassName, _ := syscall.UTF16PtrFromString("TormentNexusLogClass")
	title, _ := syscall.UTF16PtrFromString("TormentNexus Internal Event Logs")

	hLogRaw, _, _ := pCreateWindowEx.Call(
		0,
		uintptr(unsafe.Pointer(logClassName)),
		uintptr(unsafe.Pointer(title)),
		WS_OVERLAPPEDWINDOW,
		100, 100, 600, 400,
		0, 0, hInstance, 0,
	)
	hLogWnd = windows.HWND(hLogRaw)

	pShowWindow.Call(uintptr(hLogWnd), 1) // SW_SHOWNORMAL
	pUpdateWindow.Call(uintptr(hLogWnd))
}

func logWndProc(hWnd windows.HWND, msg uint32, wParam uintptr, lParam uintptr) uintptr {
	switch msg {
	case WM_CREATE:
		// Create the EDIT control
		editClass, _ := syscall.UTF16PtrFromString("EDIT")
		hInstance, _, _ := kernel32.NewProc("GetModuleHandleW").Call(0)

		hEditRaw, _, _ := pCreateWindowEx.Call(
			0,
			uintptr(unsafe.Pointer(editClass)),
			0,
			WS_CHILD|WS_VISIBLE|WS_VSCROLL|ES_MULTILINE|ES_AUTOVSCROLL|ES_READONLY,
			0, 0, 580, 360,
			uintptr(hWnd),
			1001,
			hInstance,
			0,
		)
		hEditWnd = windows.HWND(hEditRaw)

		// Populate initial logs
		logMutex.Lock()
		fullText := ""
		for _, l := range logLines {
			fullText += l + "\r\n"
		}
		logMutex.Unlock()

		ptr, _ := syscall.UTF16PtrFromString(fullText)
		pSetWindowText.Call(uintptr(hEditWnd), uintptr(unsafe.Pointer(ptr)))
		return 0

	case WM_SIZE:
		// Resize edit control to match client area
		width := int32(lParam & 0xFFFF)
		height := int32((lParam >> 16) & 0xFFFF)
		if hEditWnd != 0 {
			const SWP_NOZORDER = 0x0004
			const SWP_NOMOVE = 0x0002
			pMoveWindow := user32.NewProc("MoveWindow")
			pMoveWindow.Call(uintptr(hEditWnd), 0, 0, uintptr(width), uintptr(height), 1)
		}
		return 0

	case WM_DESTROY:
		hLogWnd = 0
		hEditWnd = 0
		return 0
	}
	ret, _, _ := pDefWindowProc.Call(uintptr(hWnd), uintptr(msg), wParam, lParam)
	return ret
}
