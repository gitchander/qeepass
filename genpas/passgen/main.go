package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

	"github.com/gitchander/qeepass/genpas"
)

func main() {
	// Инициализируем GTK.
	gtk.Init(nil)

	// Создаём билдер
	b, err := gtk.BuilderNew()
	checkError(err)

	// Загружаем в билдер окно из файла Glade
	err = b.AddFromFile("assets/glade/test.glade")
	checkError(err)

	// Получаем объект главного окна по ID
	obj, err := b.GetObject("window_main")
	checkError(err)

	// Преобразуем из объекта именно окно типа gtk.Window
	// и соединяем с сигналом "destroy" чтобы можно было закрыть
	// приложение при закрытии окна
	w := obj.(*gtk.Window)
	w.Connect("destroy", func() {
		gtk.MainQuit()
	})

	w.SetTitle("passgen")

	gw, err := NewGenWidgets(b)
	checkError(err)

	gw.buttonGenerate.Connect("clicked", func() {
		gw.generate()
	})
	gw.buttonClear.Connect("clicked", func() {
		gw.clear()
	})
	gw.buttonCopyToClipboard.Connect("clicked", func() {
		gw.copyToClipboard()
	})

	// Отображаем все виджеты в окне
	w.ShowAll()

	// Выполняем главный цикл GTK (для отрисовки). Он остановится когда
	// выполнится gtk.MainQuit()
	gtk.Main()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type genWidgets struct {
	entryNumberOfPasswords *gtk.Entry
	entryPasswordLength    *gtk.Entry

	checkbuttonUpperLetters   *gtk.CheckButton
	checkbuttonLowerLetters   *gtk.CheckButton
	checkbuttonDigits         *gtk.CheckButton
	checkbuttonSpecialSymbols *gtk.CheckButton

	checkbuttonExcludeSimilar *gtk.CheckButton
	checkbuttonHasEveryGroup  *gtk.CheckButton

	buttonGenerate        *gtk.Button
	buttonClear           *gtk.Button
	buttonCopyToClipboard *gtk.Button

	textViewPasswords   *gtk.TextView
	textBufferPasswords *gtk.TextBuffer

	clipboard *gtk.Clipboard
}

func NewGenWidgets(b *gtk.Builder) (*genWidgets, error) {

	clipboard, err := gtk.ClipboardGet(gdk.SELECTION_CLIPBOARD)
	if err != nil {
		return nil, err
	}

	textViewPasswords := getTextViewMust(b, "textviewPasswords")
	textBufferPasswords, err := textViewPasswords.GetBuffer()
	if err != nil {
		return nil, err
	}

	return &genWidgets{
		entryNumberOfPasswords:    getEntryMust(b, "entryNumberOfPasswords"),
		entryPasswordLength:       getEntryMust(b, "entryPasswordLength"),
		checkbuttonUpperLetters:   getCheckButtonMust(b, "checkbuttonUpperLetters"),
		checkbuttonLowerLetters:   getCheckButtonMust(b, "checkbuttonLowerLetters"),
		checkbuttonDigits:         getCheckButtonMust(b, "checkbuttonDigits"),
		checkbuttonSpecialSymbols: getCheckButtonMust(b, "checkbuttonSpecialSymbols"),
		checkbuttonExcludeSimilar: getCheckButtonMust(b, "checkbuttonExcludeSimilar"),
		checkbuttonHasEveryGroup:  getCheckButtonMust(b, "checkbuttonHasEveryGroup"),
		buttonGenerate:            getButtonMust(b, "buttonGenerate"),
		buttonClear:               getButtonMust(b, "buttonClear"),
		buttonCopyToClipboard:     getButtonMust(b, "buttonCopyToClipboard"),

		textViewPasswords:   textViewPasswords,
		textBufferPasswords: textBufferPasswords,

		clipboard: clipboard,
	}, nil
}

func (gw *genWidgets) clear() {
	gw.textBufferPasswords.SetText("")
}

func (gw *genWidgets) copyToClipboard() {
	var (
		startIter = gw.textBufferPasswords.GetStartIter()
		endIter   = gw.textBufferPasswords.GetEndIter()
	)
	text, err := gw.textBufferPasswords.GetText(startIter, endIter, true)
	checkError(err)

	gw.clipboard.SetText(text)
}

func (gw *genWidgets) generate() {
	text, err := gw.entryNumberOfPasswords.GetText()
	checkError(err)
	numberOfPasswords, _ := strconv.Atoi(text)
	numberOfPasswords = cropInt(numberOfPasswords, 1, 100)
	gw.entryNumberOfPasswords.SetText(strconv.Itoa(numberOfPasswords))
	//fmt.Println("NumberOfPasswords:", numberOfPasswords)

	text, err = gw.entryPasswordLength.GetText()
	checkError(err)
	passwordLength, _ := strconv.Atoi(text)
	passwordLength = cropInt(passwordLength, 8, 64)
	gw.entryPasswordLength.SetText(strconv.Itoa(passwordLength))
	//fmt.Println("PasswordLength:", passwordLength)

	p := genpas.Params{
		Upper:   gw.checkbuttonUpperLetters.GetActive(),
		Lower:   gw.checkbuttonLowerLetters.GetActive(),
		Digits:  gw.checkbuttonDigits.GetActive(),
		Special: gw.checkbuttonSpecialSymbols.GetActive(),

		ExcludeSimilar: gw.checkbuttonExcludeSimilar.GetActive(),
		HasEveryGroup:  gw.checkbuttonHasEveryGroup.GetActive(),
	}

	if (!p.Upper) && (!p.Lower) && (!p.Digits) && (!p.Special) {
		gw.checkbuttonUpperLetters.SetActive(true)
		gw.checkbuttonLowerLetters.SetActive(true)
		gw.checkbuttonDigits.SetActive(true)
		gw.checkbuttonSpecialSymbols.SetActive(true)

		p.Upper = true
		p.Lower = true
		p.Digits = true
		p.Special = true
	}

	g, err := genpas.NewGenerator(p, genpas.NewRandom())
	checkError(err)

	passwords := make([]string, numberOfPasswords)
	for i := range passwords {
		passwords[i] = g.Generate(passwordLength)
	}

	//fmt.Println("UpperLetters:", gw.checkbuttonUpperLetters.GetActive())
	//fmt.Println("LowerLetters:", gw.checkbuttonLowerLetters.GetActive())

	endIter := gw.textBufferPasswords.GetEndIter()
	for _, password := range passwords {
		gw.textBufferPasswords.Insert(endIter, password+"\n")
	}
}

func getEntry(b *gtk.Builder, name string) (*gtk.Entry, error) {
	obj, err := b.GetObject(name)
	if err != nil {
		return nil, err
	}
	p, ok := obj.(*gtk.Entry)
	if !ok {
		return nil, errConvert(name, "Entry")
	}
	return p, nil
}

func getEntryMust(b *gtk.Builder, name string) *gtk.Entry {
	p, err := getEntry(b, name)
	checkError(err)
	return p
}

func getCheckButton(b *gtk.Builder, name string) (*gtk.CheckButton, error) {
	obj, err := b.GetObject(name)
	if err != nil {
		return nil, err
	}
	p, ok := obj.(*gtk.CheckButton)
	if !ok {
		return nil, errConvert(name, "CheckButton")
	}
	return p, nil
}

func getCheckButtonMust(b *gtk.Builder, name string) *gtk.CheckButton {
	p, err := getCheckButton(b, name)
	checkError(err)
	return p
}

func getButton(b *gtk.Builder, name string) (*gtk.Button, error) {
	obj, err := b.GetObject(name)
	if err != nil {
		return nil, err
	}
	p, ok := obj.(*gtk.Button)
	if !ok {
		return nil, errConvert(name, "Button")
	}
	return p, nil
}

func getButtonMust(b *gtk.Builder, name string) *gtk.Button {
	p, err := getButton(b, name)
	checkError(err)
	return p
}

func getTextView(b *gtk.Builder, name string) (*gtk.TextView, error) {
	obj, err := b.GetObject(name)
	if err != nil {
		return nil, err
	}
	p, ok := obj.(*gtk.TextView)
	if !ok {
		return nil, errConvert(name, "TextView")
	}
	return p, nil
}

func getTextViewMust(b *gtk.Builder, name string) *gtk.TextView {
	p, err := getTextView(b, name)
	checkError(err)
	return p
}

func errConvert(name, tp string) error {
	return fmt.Errorf("invalid convert tag %s to %s", name, tp)
}

func cropInt(x int, min, max int) int {
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return x
}
