package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
)

type ZindexSorter []layout

func (a ZindexSorter) Len() int           { return len(a) }
func (a ZindexSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ZindexSorter) Less(i, j int) bool { return a[i].Zindex > a[j].Zindex }

type BikePartList struct {
	Image       interface{} `json:"Image"`
	Layout      []layout    `json:"Layout"`
	ContentType string      `json:"ContentType"`
	BaseUrl     string      `json:"BaseUrl"`
}
type ImagesKTM struct {
	Image       interface{}    `json:"Image"`
	Layout      []BikePartList `json:"Layout"`
	ContentType string         `json:"ContentType"`
	BaseURL     string         `json:"BaseUrl"`
}

type layout struct {
	Id struct {
		Value int `json:"Value"`
	} `json:"Id"`
	VehicleProductCode struct {
		Value string `json:"Value"`
	} `json:"VehicleProductCode"`
	IsStandardPart bool `json:"IsStandardPart"`
	PartCode       struct {
		Value string `json:"Value"`
	} `json:"PartCode"`
	FileName string `json:"FileName"`
	Zindex   int    `json:"Zindex"`
	Color    struct {
		IsSome bool `json:"IsSome"`
		IsNone bool `json:"IsNone"`
	} `json:"Color"`
	Perspective int `json:"Perspective"`
	LayerSide   int `json:"LayerSide"`
}

func main() {
	jsonFile, err := os.Open("ktm.json") //открываем json файл, получаем указатель на него (не данные!)
	if err != nil {                      // если не удалось найти файл или открыть его, то
		fmt.Println(err) // выводим в консоль ошибку, программа останавливается
	} else { // если всё гладко, то программа выполняется дальше
		fmt.Println("Был успешно открыт ktm.json") //выводим на экран текст
		defer jsonFile.Close()                     // отложенное закрытие файла, когда наша работа с ним будет окончена
		byteValue, _ := ioutil.ReadAll(jsonFile)   // читаем наш jsonFile как массив 0 и 1
		var Part BikePartList                      // объявляем переменную, Part - имя, BikePartList - тип
		err := json.Unmarshal(byteValue, &Part)    // преобразуем массив в структуру и заполняем переменную
		if err != nil {                            //если что-то пошло не так выводим текст ниже + программа стоп
			fmt.Println("Не удалось преобразовать массив байт в структуру ImagesKTM и заполнить переменную parts") // выводим текст
			fmt.Println(err)                                                                                       //если ошибка программа стоп
			return                                                                                                 // возвращаем значение
		}
		fmt.Println("Удалось преобразовать массив байт в структуру ImagesKTM и заполнить переменную parts") //если ок

		index := Part.Layout
		sort.Sort(ZindexSorter(index))
		log.Println("by Zindex:", index)

		partsCount := len(Part.Layout)
		fmt.Println("Количество шагов", partsCount) //выводим текст и кол-во шагов
		//for i := 0; i < partsCount; i++ {           // i переменная, которую мы используем как индекс
		for i, v := range index {
			fmt.Println("Номер шага - ", i, v)                                                                                                                                                //выводим текст
			fmt.Println("Название картинки: " + Part.Layout[i].FileName)                                                                                                                      // Выводим в консоль название картинки из активного слоя (из Layout)
			url := "https://configurator.ktm.com/rendering/api/getblobimage?companyCode=KTM&bikeSetupId=" + Part.Layout[i].VehicleProductCode.Value + "&imagePath=" + Part.Layout[i].FileName // Склеиваем статические и динамические части ссылки
			fmt.Println("URL картинки: " + url)                                                                                                                                               // выводим в консоль сформированный адрес картинки
			fileName := "./Фото/" + Part.Layout[i].FileName                                                                                                                                   //объявляем переменную
			_, err := os.Stat(fileName)                                                                                                                                                       // err - не удалось получить инфориацию о файле, поэтому считаем, что файла не существует. _ переменная которую мы не будем использовать (содержит ненужную информацию, которую потом удаляем)
			if err != nil {                                                                                                                                                                   //если err т.е файла не нашлось, то
				fmt.Printf("Файла не существует, создаём файл\n") // выводим текст
				out, err := os.Create(fileName)                   // и пытаемся создать пустой файл
				if err != nil {                                   // если создать файл не получилось
					log.Fatalf("Не удалось создать пустой файл: "+fileName, err) // выводим текст
				} else { //если всё ок, тогда идём дальше
					defer out.Close()        // отложенное закрытие файла, когда наша работа с ним будет окончена
					resp, _ := http.Get(url) //********** response - ответ сервера.
					// todo: нужно обработать коды ответа и ошибки(они скорее всего будут только если с сервером не удалось соединиться, а в случае 404 или подобных ошибок мы все еще получим полноценный resp, который не будет картинкой и будет в будущем ломать склейку картинок)
					defer resp.Body.Close()        // отложенное закрытие тела ответа от сервера, когда наша работа с ним будет окончена
					_, _ = io.Copy(out, resp.Body) // todo: использовать длину записанных данных в файл и обработать ошибку
				} //закрываем сценарий если файл удалось создать
			} // закрываем сценарий что файл создан
			_, err = os.Stat(fileName) // err - не удалось получить инфориацию о файле, поэтому считаем, что файла не существует. _ переменная которую мы не будем использовать
			//if err == nil && fileName != "images/Base_000_0000.png" { // **имитировали ошибку для проверки
			if err == nil { //если удалось получить файл
				fmt.Printf("Удалось получить файл, склеиваем\n")
				imageForGluing, err := os.Open(fileName) // получаем указатель на файл, если он существует и доступен для чтения. переменная файлнейм уже содержит путь к картинке: "images/" + parts.Layout[i].FileName
				if err != nil {                          //если не удалось открыть файл
					log.Fatalf("failed to open: %s", err) //фатал, программа стоп
				}
				first, err := png.Decode(imageForGluing) //декодирование изображения
				if err != nil {                          //если ошибка (конверт с ошибкой не пустой)
					log.Fatalf("failed to decode: %s", err) //если не получилось, фатал,программа стоп
				}
				defer imageForGluing.Close()                         //отложенное закрытие
				var resultImage *image.RGBA                          //приминение цвета
				b := first.Bounds()                                  //******
				resultImage = image.NewRGBA(b)                       //******
				draw.Draw(resultImage, b, first, image.ZP, draw.Src) //******
				if i != 0 {                                          //если это не первый слой, тогда
					lastResult, err := os.Open("result.png") //открываем изображение результ
					if err != nil {                          //если не получилось, то
						log.Fatalf("failed to open: %s", err) //выводим ошибку, стоп программа
					}
					second, err := png.Decode(lastResult) //декодируем изображение
					if err != nil {                       //если не получилось, то
						log.Fatalf("failed to decode: %s", err) //выводим ошибку, стоп программа
					}
					defer lastResult.Close()                                             //отложенное закрытие
					draw.Draw(resultImage, second.Bounds(), second, image.ZP, draw.Over) //склеиваем изображения
				}
				third, err := os.Create("result.png") // создаём изображение результ
				if err != nil {                       //если не получилось создать
					log.Fatalf("failed to create: %s", err) //выводим на экран ошибку и останавливаем прог
				}
				png.Encode(third, resultImage) //кодируем обратно изображение
				defer third.Close()            //отложенное закрытие
			} else { //**
				fmt.Println("не было необходимой картинки для склейки") //выводим текст
				return                                                  //****
			} //****
		} //закрываем цикл
	} //закрываем действия после успешного открытия json
} //закрываем функцию
