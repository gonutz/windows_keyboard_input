package main

import (
	"fmt"
	"runtime"
	"strconv"
	//"unicode"
	"unicode/utf16"

	"github.com/gonutz/w32"
	"github.com/gonutz/win"
)

func main() {
	runtime.LockOSThread()
	events := make(chan keyboardEvent)

	send := func(e keyboardEvent) {
		//if e.down && e.text != "" {
		//	for _, r := range e.text {
		//		if unicode.IsPrint(r) {
		//			fmt.Print(string(r))
		//		}
		//	}
		//}
		fmt.Printf("    %v\n", e)
		if !e.down {
			fmt.Println()
		}
		//events <- e
	}
	_ = send

	escCount := 0
	opt := win.DefaultOptions()
	opt.X = 800
	opt.Y = 100
	window, err := win.NewWindow(
		opt,
		func(window w32.HWND, msg uint32, w, l uintptr) uintptr {
			extended := l&(1<<24) != 0
			if w32.WM_KEYFIRST <= msg && msg < w32.WM_KEYLAST {
				if l&(1<<24) != 0 {
					//fmt.Println("extended")
				}
			}

			switch msg {
			case w32.WM_RBUTTONDOWN:
				win.CloseWindow(window)
				return 0

			case w32.WM_INPUT:
				fmt.Println(w32.GetMessageTime(), "WM_INPUT", w, l)
				return 0
			case w32.WM_KEYDOWN:
				if w == 27 {
					escCount++
					if escCount >= 2 {
						win.CloseWindow(window)
					}
				}
				//fmt.Println(w32.GetMessageTime(), "WM_KEYDOWN", w, l)

				if l&(1<<30) == 0 { // no key repeat
					t := w32.GetMessageTime()
					var next w32.MSG
					var text string
					if w32.PeekMessage(&next, 0, 0, 0, w32.PM_NOREMOVE) &&
						next.Time == uint32(t) &&
						next.Message == w32.WM_CHAR {
						w32.PeekMessage(&next, 0, 0, 0, w32.PM_REMOVE)
						text = string(utf16.Decode([]uint16{uint16(next.WParam)}))
					}

					key := Key(w)
					if key == w32.VK_SHIFT {
						if extended || l&0xFF0000 == 0x360000 {
							key = w32.VK_RSHIFT
						} else {
							key = w32.VK_LSHIFT
						}
					}
					if key == w32.VK_MENU {
						if extended {
							key = w32.VK_RMENU
						} else {
							key = w32.VK_LMENU
						}
					}
					if key == w32.VK_CONTROL {
						if extended {
							key = w32.VK_RCONTROL
						} else {
							key = w32.VK_LCONTROL
						}
					}
					altDown := w32.GetKeyState(w32.VK_MENU)&0x8000 != 0
					ctrlDown := w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0
					shiftDown := w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0
					if key == w32.VK_CANCEL && ctrlDown {
						if extended {
							key = w32.VK_PAUSE
						} else {
							key = w32.VK_SCROLL
						}
					} else if key == w32.VK_PAUSE && ctrlDown && !altDown && !shiftDown {
						key = w32.VK_NUMLOCK
					}
					send(keyboardEvent{
						key:   key,
						text:  text,
						down:  true,
						alt:   altDown,
						ctrl:  ctrlDown,
						shift: shiftDown,
					})
				}
				return 0
			case w32.WM_KEYUP:
				//fmt.Println(w32.GetMessageTime(), "WM_KEYUP", w, l)
				key := Key(w)
				if key == w32.VK_SHIFT {
					if extended || l&0xFF0000 == 0x360000 {
						key = w32.VK_RSHIFT
					} else {
						key = w32.VK_LSHIFT
					}
				}
				if key == w32.VK_MENU {
					if extended {
						key = w32.VK_RMENU
					} else {
						key = w32.VK_LMENU
					}
				}
				if key == w32.VK_CONTROL {
					if extended {
						key = w32.VK_RCONTROL
					} else {
						key = w32.VK_LCONTROL
					}
				}
				altDown := w32.GetKeyState(w32.VK_MENU)&0x8000 != 0
				ctrlDown := w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0
				shiftDown := w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0
				if key == w32.VK_CANCEL && ctrlDown {
					if extended {
						key = w32.VK_PAUSE
					} else {
						key = w32.VK_SCROLL
					}
				} else if key == w32.VK_PAUSE && ctrlDown && !altDown && !shiftDown {
					key = w32.VK_NUMLOCK
				}
				if key == w32.VK_SNAPSHOT {
					// The print key does not send a key down message, only a
					// key up message. We just send the down ourselves right
					// before the up.
					send(keyboardEvent{
						down:  true,
						key:   key,
						alt:   altDown,
						ctrl:  ctrlDown,
						shift: shiftDown,
					})
				}
				send(keyboardEvent{
					down:  false,
					key:   key,
					alt:   w32.GetKeyState(w32.VK_MENU)&0x8000 != 0,
					ctrl:  w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0,
					shift: w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0,
				})
				return 0
			case w32.WM_CHAR:
				if l&(1<<30) == 0 { // no key repeat
					r := utf16.Decode([]uint16{uint16(w)})[0]
					fmt.Println(w32.GetMessageTime(), "WM_CHAR_"+string(r), w, l)
					send(keyboardEvent{
						text:  string(r),
						down:  true,
						alt:   w32.GetKeyState(w32.VK_MENU)&0x8000 != 0,
						ctrl:  w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0,
						shift: w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0,
					})
				}
				return 0
			case w32.WM_DEADCHAR:
				fmt.Println(w32.GetMessageTime(), "WM_DEADCHAR", w, l)
				return 0
			case w32.WM_SYSKEYDOWN:
				//fmt.Println(w32.GetMessageTime(), "WM_SYSKEYDOWN", w, l)
				if l&(1<<30) == 0 { // no key repeat
					t := w32.GetMessageTime()
					var next w32.MSG
					var text string
					if w32.PeekMessage(&next, 0, 0, 0, w32.PM_NOREMOVE) &&
						next.Time == uint32(t) &&
						next.Message == w32.WM_SYSCHAR {
						w32.PeekMessage(&next, 0, 0, 0, w32.PM_REMOVE)
						text = string(utf16.Decode([]uint16{uint16(next.WParam)}))
					}

					send(keyboardEvent{
						key:   Key(w),
						text:  text,
						down:  true,
						alt:   w32.GetKeyState(w32.VK_MENU)&0x8000 != 0,
						ctrl:  w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0,
						shift: w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0,
					})
				}
				return 0
			case w32.WM_SYSKEYUP:
				//fmt.Println(w32.GetMessageTime(), "WM_SYSKEYUP", w, l)
				send(keyboardEvent{
					down:  false,
					key:   Key(w),
					alt:   w32.GetKeyState(w32.VK_MENU)&0x8000 != 0,
					ctrl:  w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0,
					shift: w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0,
				})
				return 0
			case w32.WM_SYSCHAR:
				fmt.Println(w32.GetMessageTime(), "WM_SYSCHAR", w, l)
				return 0
			case w32.WM_SYSDEADCHAR:
				fmt.Println(w32.GetMessageTime(), "WM_SYSDEADCHAR", w, l)
				return 0
			case w32.WM_COMMAND:
				fmt.Println(w32.GetMessageTime(), "WM_COMMAND", w, l)
				return 0
			case w32.WM_SYSCOMMAND:
				fmt.Println(w32.GetMessageTime(), "WM_SYSCOMMAND", w, l)
				return 0
			case w32.WM_MENUCHAR:
				fmt.Println(w32.GetMessageTime(), "WM_MENUCHAR", w, l)
				return 0
			case w32.WM_HOTKEY:
				fmt.Println(w32.GetMessageTime(), "WM_HOTKEY", w, l)
				return 0
			case w32.WM_APPCOMMAND:
				//fmt.Println(w32.GetMessageTime(), "WM_APPCOMMAND", w, l)
				return 0

			case w32.WM_DESTROY:
				w32.PostQuitMessage(0)
				return 0
			default:
				return w32.DefWindowProc(window, msg, w, l)
			}
		},
	)
	check(err)

	go func() {
		var title string
		for e := range events {
			title += fmt.Sprintf(" %v", e)
			const maxLen = 180
			for len(title) > maxLen {
				title = title[1:]
			}
			w32.SetWindowText(window, title)
		}
	}()

	win.RunMainLoop()
	close(events)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type keyboardEvent struct {
	down  bool
	key   Key
	text  string
	ctrl  bool
	shift bool
	alt   bool
}

func (e keyboardEvent) String() string {
	var s string
	if e.ctrl {
		s += "Ctrl"
	}
	if e.shift {
		if s != "" {
			s += "+"
		}
		s += "Shift"
	}
	if e.alt {
		if s != "" {
			s += "+"
		}
		s += "Alt"
	}
	if s != "" {
		s += "+"
	}
	s += e.key.String()
	if e.text != "" {
		s += fmt.Sprintf(" (%q)", e.text)
	}
	if e.down {
		s = "v " + s
	} else {
		s = "^ " + s
	}
	return s
}

type Key int

func (k Key) String() string {
	if 'A' <= k && k <= 'Z' {
		return string(k)
	}
	switch k {
	case w32.VK_LBUTTON:
		return "LBUTTON"
	case w32.VK_RBUTTON:
		return "RBUTTON"
	case w32.VK_CANCEL:
		return "CANCEL"
	case w32.VK_MBUTTON:
		return "MBUTTON"
	case w32.VK_XBUTTON1:
		return "XBUTTON1"
	case w32.VK_XBUTTON2:
		return "XBUTTON2"
	case w32.VK_BACK:
		return "BACK"
	case w32.VK_TAB:
		return "TAB"
	case w32.VK_CLEAR:
		return "CLEAR"
	case w32.VK_RETURN:
		return "RETURN"
	case w32.VK_SHIFT:
		return "SHIFT"
	case w32.VK_CONTROL:
		return "CONTROL"
	case w32.VK_MENU:
		return "MENU"
	case w32.VK_PAUSE:
		return "PAUSE"
	case w32.VK_CAPITAL:
		return "CAPITAL"
	case w32.VK_KANA:
		return "KANA"
	case w32.VK_JUNJA:
		return "JUNJA"
	case w32.VK_FINAL:
		return "FINAL"
	case w32.VK_HANJA:
		return "HANJA"
	case w32.VK_ESCAPE:
		return "ESCAPE"
	case w32.VK_CONVERT:
		return "CONVERT"
	case w32.VK_NONCONVERT:
		return "NONCONVERT"
	case w32.VK_ACCEPT:
		return "ACCEPT"
	case w32.VK_MODECHANGE:
		return "MODECHANGE"
	case w32.VK_SPACE:
		return "SPACE"
	case w32.VK_PRIOR:
		return "PRIOR"
	case w32.VK_NEXT:
		return "NEXT"
	case w32.VK_END:
		return "END"
	case w32.VK_HOME:
		return "HOME"
	case w32.VK_LEFT:
		return "LEFT"
	case w32.VK_UP:
		return "UP"
	case w32.VK_RIGHT:
		return "RIGHT"
	case w32.VK_DOWN:
		return "DOWN"
	case w32.VK_SELECT:
		return "SELECT"
	case w32.VK_PRINT:
		return "PRINT"
	case w32.VK_EXECUTE:
		return "EXECUTE"
	case w32.VK_SNAPSHOT:
		return "SNAPSHOT"
	case w32.VK_INSERT:
		return "INSERT"
	case w32.VK_DELETE:
		return "DELETE"
	case w32.VK_HELP:
		return "HELP"
	case w32.VK_LWIN:
		return "LWIN"
	case w32.VK_RWIN:
		return "RWIN"
	case w32.VK_APPS:
		return "APPS"
	case w32.VK_SLEEP:
		return "SLEEP"
	case w32.VK_NUMPAD0:
		return "NUMPAD0"
	case w32.VK_NUMPAD1:
		return "NUMPAD1"
	case w32.VK_NUMPAD2:
		return "NUMPAD2"
	case w32.VK_NUMPAD3:
		return "NUMPAD3"
	case w32.VK_NUMPAD4:
		return "NUMPAD4"
	case w32.VK_NUMPAD5:
		return "NUMPAD5"
	case w32.VK_NUMPAD6:
		return "NUMPAD6"
	case w32.VK_NUMPAD7:
		return "NUMPAD7"
	case w32.VK_NUMPAD8:
		return "NUMPAD8"
	case w32.VK_NUMPAD9:
		return "NUMPAD9"
	case w32.VK_MULTIPLY:
		return "MULTIPLY"
	case w32.VK_ADD:
		return "ADD"
	case w32.VK_SEPARATOR:
		return "SEPARATOR"
	case w32.VK_SUBTRACT:
		return "SUBTRACT"
	case w32.VK_DECIMAL:
		return "DECIMAL"
	case w32.VK_DIVIDE:
		return "DIVIDE"
	case w32.VK_F1:
		return "F1"
	case w32.VK_F2:
		return "F2"
	case w32.VK_F3:
		return "F3"
	case w32.VK_F4:
		return "F4"
	case w32.VK_F5:
		return "F5"
	case w32.VK_F6:
		return "F6"
	case w32.VK_F7:
		return "F7"
	case w32.VK_F8:
		return "F8"
	case w32.VK_F9:
		return "F9"
	case w32.VK_F10:
		return "F10"
	case w32.VK_F11:
		return "F11"
	case w32.VK_F12:
		return "F12"
	case w32.VK_F13:
		return "F13"
	case w32.VK_F14:
		return "F14"
	case w32.VK_F15:
		return "F15"
	case w32.VK_F16:
		return "F16"
	case w32.VK_F17:
		return "F17"
	case w32.VK_F18:
		return "F18"
	case w32.VK_F19:
		return "F19"
	case w32.VK_F20:
		return "F20"
	case w32.VK_F21:
		return "F21"
	case w32.VK_F22:
		return "F22"
	case w32.VK_F23:
		return "F23"
	case w32.VK_F24:
		return "F24"
	case w32.VK_NUMLOCK:
		return "NUMLOCK"
	case w32.VK_SCROLL:
		return "SCROLL"
	case w32.VK_OEM_NEC_EQUAL:
		return "OEM_NEC_EQUAL"
	case w32.VK_OEM_FJ_MASSHOU:
		return "OEM_FJ_MASSHOU"
	case w32.VK_OEM_FJ_TOUROKU:
		return "OEM_FJ_TOUROKU"
	case w32.VK_OEM_FJ_LOYA:
		return "OEM_FJ_LOYA"
	case w32.VK_OEM_FJ_ROYA:
		return "OEM_FJ_ROYA"
	case w32.VK_LSHIFT:
		return "LSHIFT"
	case w32.VK_RSHIFT:
		return "RSHIFT"
	case w32.VK_LCONTROL:
		return "LCONTROL"
	case w32.VK_RCONTROL:
		return "RCONTROL"
	case w32.VK_LMENU:
		return "LMENU"
	case w32.VK_RMENU:
		return "RMENU"
	case w32.VK_BROWSER_BACK:
		return "BROWSER_BACK"
	case w32.VK_BROWSER_FORWARD:
		return "BROWSER_FORWARD"
	case w32.VK_BROWSER_REFRESH:
		return "BROWSER_REFRESH"
	case w32.VK_BROWSER_STOP:
		return "BROWSER_STOP"
	case w32.VK_BROWSER_SEARCH:
		return "BROWSER_SEARCH"
	case w32.VK_BROWSER_FAVORITES:
		return "BROWSER_FAVORITES"
	case w32.VK_BROWSER_HOME:
		return "BROWSER_HOME"
	case w32.VK_VOLUME_MUTE:
		return "VOLUME_MUTE"
	case w32.VK_VOLUME_DOWN:
		return "VOLUME_DOWN"
	case w32.VK_VOLUME_UP:
		return "VOLUME_UP"
	case w32.VK_MEDIA_NEXT_TRACK:
		return "MEDIA_NEXT_TRACK"
	case w32.VK_MEDIA_PREV_TRACK:
		return "MEDIA_PREV_TRACK"
	case w32.VK_MEDIA_STOP:
		return "MEDIA_STOP"
	case w32.VK_MEDIA_PLAY_PAUSE:
		return "MEDIA_PLAY_PAUSE"
	case w32.VK_LAUNCH_MAIL:
		return "LAUNCH_MAIL"
	case w32.VK_LAUNCH_MEDIA_SELECT:
		return "LAUNCH_MEDIA_SELECT"
	case w32.VK_LAUNCH_APP1:
		return "LAUNCH_APP1"
	case w32.VK_LAUNCH_APP2:
		return "LAUNCH_APP2"
	case w32.VK_OEM_1:
		return "OEM_1"
	case w32.VK_OEM_PLUS:
		return "OEM_PLUS"
	case w32.VK_OEM_COMMA:
		return "OEM_COMMA"
	case w32.VK_OEM_MINUS:
		return "OEM_MINUS"
	case w32.VK_OEM_PERIOD:
		return "OEM_PERIOD"
	case w32.VK_OEM_2:
		return "OEM_2"
	case w32.VK_OEM_3:
		return "OEM_3"
	case w32.VK_OEM_4:
		return "OEM_4"
	case w32.VK_OEM_5:
		return "OEM_5"
	case w32.VK_OEM_6:
		return "OEM_6"
	case w32.VK_OEM_7:
		return "OEM_7"
	case w32.VK_OEM_8:
		return "OEM_8"
	case w32.VK_OEM_AX:
		return "OEM_AX"
	case w32.VK_OEM_102:
		return "OEM_102"
	case w32.VK_ICO_HELP:
		return "ICO_HELP"
	case w32.VK_ICO_00:
		return "ICO_00"
	case w32.VK_PROCESSKEY:
		return "PROCESSKEY"
	case w32.VK_ICO_CLEAR:
		return "ICO_CLEAR"
	case w32.VK_OEM_RESET:
		return "OEM_RESET"
	case w32.VK_OEM_JUMP:
		return "OEM_JUMP"
	case w32.VK_OEM_PA1:
		return "OEM_PA1"
	case w32.VK_OEM_PA2:
		return "OEM_PA2"
	case w32.VK_OEM_PA3:
		return "OEM_PA3"
	case w32.VK_OEM_WSCTRL:
		return "OEM_WSCTRL"
	case w32.VK_OEM_CUSEL:
		return "OEM_CUSEL"
	case w32.VK_OEM_ATTN:
		return "OEM_ATTN"
	case w32.VK_OEM_FINISH:
		return "OEM_FINISH"
	case w32.VK_OEM_COPY:
		return "OEM_COPY"
	case w32.VK_OEM_AUTO:
		return "OEM_AUTO"
	case w32.VK_OEM_ENLW:
		return "OEM_ENLW"
	case w32.VK_OEM_BACKTAB:
		return "OEM_BACKTAB"
	case w32.VK_ATTN:
		return "ATTN"
	case w32.VK_CRSEL:
		return "CRSEL"
	case w32.VK_EXSEL:
		return "EXSEL"
	case w32.VK_EREOF:
		return "EREOF"
	case w32.VK_PLAY:
		return "PLAY"
	case w32.VK_ZOOM:
		return "ZOOM"
	case w32.VK_NONAME:
		return "NONAME"
	case w32.VK_PA1:
		return "PA1"
	case w32.VK_OEM_CLEAR:
		return "OEM_CLEAR"
	default:
		return "Key(" + strconv.Itoa(int(k)) + ")"
	}
}
