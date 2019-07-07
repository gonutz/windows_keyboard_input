package main

import (
	"fmt"
	"runtime"
	"unicode/utf16"

	"github.com/gonutz/w32"
	"github.com/gonutz/win"
)

func main() {
	runtime.LockOSThread()
	events := make(chan keyboardEvent)

	send := func(e keyboardEvent) {
		fmt.Printf("    %#v\n", e)
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
				if l&(1<<30) == 0 { // no key repeat
					t := w32.GetMessageTime()
					fmt.Println(t, "WM_KEYDOWN", w, l)
					var next w32.MSG
					if w32.PeekMessage(&next, 0, 0, 0, w32.PM_NOREMOVE) &&
						next.Time == uint32(t) &&
						next.Message == w32.WM_CHAR {
						w32.PeekMessage(&next, 0, 0, 0, w32.PM_REMOVE)
						r := utf16.Decode([]uint16{uint16(next.WParam)})[0]
						send(keyboardEvent{
							key:  int(w),
							text: string(r),
						})
					}
				}
				return 0
			case w32.WM_KEYUP:
				fmt.Println(w32.GetMessageTime(), "WM_KEYUP", w, l)
				return 0
			case w32.WM_CHAR:
				if l&(1<<30) == 0 { // no key repeat
					r := utf16.Decode([]uint16{uint16(w)})[0]
					fmt.Println(w32.GetMessageTime(), "WM_CHAR_"+string(r), w, l)
					send(keyboardEvent{
						text: string(r),
					})
				}
				return 0
			case w32.WM_DEADCHAR:
				fmt.Println(w32.GetMessageTime(), "WM_DEADCHAR", w, l)
				return 0
			case w32.WM_SYSKEYDOWN:
				fmt.Println(w32.GetMessageTime(), "WM_SYSKEYDOWN", w, l)
				return 0
			case w32.WM_SYSKEYUP:
				fmt.Println(w32.GetMessageTime(), "WM_SYSKEYUP", w, l)
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
				fmt.Println(w32.GetMessageTime(), "WM_APPCOMMAND", w, l)
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
	text  string
	key   int
	ctrl  bool
	shift bool
	alt   bool
}
