package main

import (
	"bytes"
	"fmt"
	"log"
	"strconv"

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
	err = b.AddFromFile("test.glade")
	checkError(err)

	// Получаем объект главного окна по ID
	obj, err := b.GetObject("window_main")
	checkError(err)

	// Преобразуем из объекта именно окно типа gtk.Window
	// и соединяем с сигналом "destroy" чтобы можно было закрыть
	// приложение при закрытии окна
	win := obj.(*gtk.Window)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	gw, err := NewGenWidgets(b)
	checkError(err)

	gw.buttonGenerate.Connect("clicked", func() {
		gw.generate()
	})

	// Отображаем все виджеты в окне
	win.ShowAll()

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

	buttonGenerate    *gtk.Button
	textviewPasswords *gtk.TextView
}

func NewGenWidgets(b *gtk.Builder) (*genWidgets, error) {
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
		textviewPasswords:         getTextViewMust(b, "textviewPasswords"),
	}, nil
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

	var buf bytes.Buffer
	for i := 0; i < numberOfPasswords; i++ {
		password := g.Generate(passwordLength)
		buf.WriteString(password)
		buf.WriteByte('\n')
	}

	//fmt.Println("UpperLetters:", gw.checkbuttonUpperLetters.GetActive())
	//fmt.Println("LowerLetters:", gw.checkbuttonLowerLetters.GetActive())

	textBuffer, err := gw.textviewPasswords.GetBuffer()
	checkError(err)
	textBuffer.SetText(buf.String())
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
