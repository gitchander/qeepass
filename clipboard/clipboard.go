package clipboard

import (
	"sync"
	"time"

	atotto_cb "github.com/atotto/clipboard"
)

/*

	Clear clipboard (use xsel):

	sudo apt-get install xsel

	xsel -bc


	//---------------------------

	Clear (use xclip):

	echo -n | xclip -selection clipboard

*/

type clipboard struct {
	text string
}

var (
	cb   *clipboard
	once sync.Once
)

func Clipboard() *clipboard {
	once.Do(func() {
		if _, err := atotto_cb.ReadAll(); err != nil {
			panic(err)
		}
		cb = &clipboard{}
	})
	return cb
}

func (cb *clipboard) SetText(text string) error {

	cb.text = text
	err := atotto_cb.WriteAll(text)
	if err != nil {
		return err
	}

	quit := make(chan struct{})

	go func() {
		select {
		case <-quit:
			return

		case <-time.After(time.Second):

		}
	}()

	return nil
}

func (cb *clipboard) clearClipboard() {

}

func (cb *clipboard) Clear() error {
	text, err := atotto_cb.ReadAll()
	if err != nil {
		return err
	}
	if text == cb.text {
		atotto_cb.WriteAll("") // clear
	}
	cb.text = ""
	return nil
}
