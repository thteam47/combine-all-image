package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/thteam47/combine-all-image/models"
	"image"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Resize Image")
	w.Resize(fyne.NewSize(1100, 700))
	w.CenterOnScreen()
	//data := []models.Todo{}
	todos := binding.NewUntypedList()
	//for _, t := range data {
	//	todos.Append(t)
	//}

	newtodoDescTxtHeight := widget.NewEntry()
	newtodoDescTxtHeight.PlaceHolder = "Height"
	newtodoDescTxtHeight.SetText("1200")
	newtodoDescTxtWidth := widget.NewEntry()
	newtodoDescTxtWidth.PlaceHolder = "Width"
	newtodoDescTxtWidth.SetText("1200")
	newtodoDescTxtBorderSize := widget.NewEntry()
	newtodoDescTxtBorderSize.PlaceHolder = "Border Size"
	newtodoDescTxtBorderSize.SetText("90")
	newtodoDescTxtScaleLogo := widget.NewEntry()
	newtodoDescTxtScaleLogo.PlaceHolder = "Scale Logo"
	newtodoDescTxtScaleLogo.SetText("0.6")
	newtodoDescTxtPotisionLeftLogo := widget.NewEntry()
	newtodoDescTxtPotisionLeftLogo.PlaceHolder = "Position Left Logo"
	newtodoDescTxtPotisionLeftLogo.SetText("30")
	combinedSizeData := 4
	combinedSize := widget.NewRadioGroup([]string{"4", "9"}, func(selected string) {
		if selected == "9" {
			combinedSizeData = 9
		} else {
			combinedSizeData = 4
		}
	})
	combinedSize.Horizontal = true
	combinedSize.SetSelected("4")
	progressBar := widget.NewProgressBar()
	progressBar.SetValue(0)
	progressBar.Hide()
	logoTxt := widget.NewLabel("")
	w.SetContent(
		container.NewBorder(
			container.NewVBox(container.NewHBox(widget.NewLabel("Combine Images"),
				combinedSize, widget.NewButton("Logo Files", func() {
					dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
						if uri == nil {
							return
						}
						logoTxt.SetText(uri.URI().Path())
					}, w)
				}), logoTxt),
				container.NewGridWithColumns(5, widget.NewLabel("Height image"), widget.NewLabel("Width image"),
					widget.NewLabel("Scale logo"),
					widget.NewLabel("Border size logo"), widget.NewLabel("Position left logo")),
				container.NewGridWithColumns(5, newtodoDescTxtHeight, newtodoDescTxtWidth, newtodoDescTxtScaleLogo,
					newtodoDescTxtBorderSize, newtodoDescTxtPotisionLeftLogo),
				container.NewVBox(
					progressBar,
				),
				container.NewGridWithColumns(2,
					widget.NewButton("Open Files", func() {
						dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
							if uri == nil {
								return
							}
							imagesError := []string{}
							checkResolutionImage(&imagesError, uri.URI().Path())
							fmt.Println(imagesError)
							for _, path := range imagesError {
								todos.Append(models.NewTodo(path))
							}
							if len(imagesError) == 0 {
								dialog.ShowError(errors.New("No image invalid"), w)
							}
						}, w)
					}),
					widget.NewButton("Open Folder", func() {
						dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
							if err != nil {
								w.SetTitle("Error: " + err.Error())
								return
							}
							if uri == nil {
								return
							}
							if uri != nil {
								// Xử lý đường dẫn thư mục đã chọn ở đây
								imagesError := []string{}
								checkResolutionImage(&imagesError, uri.Path())
								fmt.Println(imagesError)
								for _, path := range imagesError {
									todos.Append(models.NewTodo(path))
								}
								if len(imagesError) == 0 {
									dialog.ShowError(errors.New("No image invalid"), w)
								}
							} else {
								w.SetTitle("No Folder Selected")
							}

						}, w)
					}),
				),
			),
			container.NewGridWithColumns(6,
				widget.NewButton("Resize", func() {
					dataTodos, _ := todos.Get()
					progressBar.SetValue(0)
					progressBar.Show()
					count := 0
					countDont := 0
					newTodos := []models.Todo{}
					for _, urlData := range dataTodos {
						todo := urlData.(models.Todo)
						if todo.Done {
							newTodos = append(newTodos, todo)
							countDont++
						}
					}
					for _, todo := range newTodos {
						err := fomattedImage(todo, newtodoDescTxtHeight, newtodoDescTxtWidth)
						if err != nil {
							dialog.ShowError(err, w)
							break
						}
						count++
						progressBar.SetValue(float64(count) / float64(len(newTodos)))
					}
					dialog.ShowInformation("Message", "Resize successful", w)
				}),
				widget.NewButton("Combine Image", func() {
					dataTodos, _ := todos.Get()
					progressBar.SetValue(0)
					progressBar.Show()
					images := []string{}
					for _, urlData := range dataTodos {
						todo := urlData.(models.Todo)
						if todo.Done {
							images = append(images, todo.Url)
						}
					}
					if len(images) != combinedSizeData {
						dialog.ShowError(errors.New(fmt.Sprintf("Number of photos must be %d", combinedSizeData)), w)
					}
					err := combineImage(images, newtodoDescTxtHeight, newtodoDescTxtWidth, combinedSizeData)
					if err != nil {
						dialog.ShowError(err, w)
					}
					progressBar.SetValue(100)
					dialog.ShowInformation("Message", "Combine successful", w)
				}),
				widget.NewButton("Insert Logo", func() {
					dataTodos, _ := todos.Get()
					progressBar.SetValue(0)
					progressBar.Show()
					count := 0
					countDont := 0
					newTodos := []models.Todo{}
					var err error
					borderSize := 90
					scaleLogo := 0.8
					positionLeftLogo := 50
					if newtodoDescTxtBorderSize.Text != "" {
						borderSize, err = strconv.Atoi(newtodoDescTxtBorderSize.Text)
						if err != nil {
							dialog.ShowError(errors.New("Border size invalid"), w)
							return
						}
					}
					if newtodoDescTxtScaleLogo.Text != "" {
						scaleLogo, err = strconv.ParseFloat(newtodoDescTxtScaleLogo.Text, 64)
						if err != nil {
							dialog.ShowError(errors.New("Scale logo invalid"), w)
							return
						}
					}
					if newtodoDescTxtPotisionLeftLogo.Text != "" {
						positionLeftLogo, err = strconv.Atoi(newtodoDescTxtPotisionLeftLogo.Text)
						if err != nil {
							dialog.ShowError(errors.New("Potision left logo invalid"), w)
							return
						}
					}
					for _, urlData := range dataTodos {
						todo := urlData.(models.Todo)
						if todo.Done {
							newTodos = append(newTodos, todo)
							countDont++
						}
					}
					for _, todo := range newTodos {
						err := combineImageWithLogo(todo.Url, logoTxt.Text, borderSize, scaleLogo, positionLeftLogo)
						if err != nil {
							dialog.ShowError(err, w)
							break
						}
						count++
						progressBar.SetValue(float64(count) / float64(len(newTodos)))
					}
					dialog.ShowInformation("Message", "Insert successful", w)
				}),
				widget.NewButton("Select All", func() {
					progressBar.Hide()
					dataTodos, _ := todos.Get()
					todos.Set(nil)
					for _, urlData := range dataTodos {
						todo := urlData.(models.Todo)
						todo.Done = true
						todos.Append(todo)
					}
				}),
				widget.NewButton("Reset Select", func() {
					progressBar.Hide()
					dataTodos, _ := todos.Get()
					todos.Set(nil)
					for _, urlData := range dataTodos {
						todo := urlData.(models.Todo)
						todo.Done = false
						todos.Append(todo)
					}
				}),
				widget.NewButton("Reset", func() {
					todos.Set(nil)
					progressBar.Hide()
				}),
			),
			nil, // Right
			nil,
			widget.NewListWithData(
				todos,
				func() fyne.CanvasObject {
					return container.NewBorder(
						nil, nil, nil,
						// left of the border
						widget.NewCheck("", func(b bool) {}),
						// takes the rest of the space
						widget.NewLabel(""),
					)
				},
				func(di binding.DataItem, o fyne.CanvasObject) {
					ctr, _ := o.(*fyne.Container)
					l := ctr.Objects[0].(*widget.Label)
					c := ctr.Objects[1].(*widget.Check)
					todo := models.NewTodoFromDataItem(di)
					l.SetText(todo.Url)
					c.SetChecked(todo.Done)

					c.OnChanged = func(checked bool) {
						// Cập nhật trạng thái Done của TODO dựa trên giá trị checked
						todo.Done = checked
						dataTodos, _ := todos.Get()
						todos.Set(nil)
						for _, urlData := range dataTodos {
							if urlData.(models.Todo).Url == todo.Url {
								todos.Append(todo)
							} else {
								todos.Append(urlData)
							}
						}

						// Gán danh sách mới cho biến todos
						//todos.Set(updatedTodos)
					}
				}),
		),
	)
	w.ShowAndRun()
}

func fomattedImage(todo models.Todo, newtodoDescTxtHeight *widget.Entry, newtodoDescTxtWidth *widget.Entry) error {
	srcImage, err := imaging.Open(todo.Url)
	if err != nil {
		panic(err)
	}
	newWidth := 1200
	newHeight := 1200
	if newtodoDescTxtHeight.Text != "" {
		newHeight, err = strconv.Atoi(newtodoDescTxtHeight.Text)
		if err != nil {
			return errors.New("Size height invalid")
		}
	}
	if newtodoDescTxtWidth.Text != "" {
		newWidth, err = strconv.Atoi(newtodoDescTxtWidth.Text)
		if err != nil {
			return errors.New("Size width invalid")
		}
	}
	dirName := filepath.Dir(todo.Url)
	dirNameFormatted := filepath.Join(dirName, "formatted")
	_, err = os.Stat(dirNameFormatted)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirNameFormatted, os.ModePerm)
		if err != nil {
			return errors.New(fmt.Sprintf("Lỗi khi tạo thư mục: %s", dirNameFormatted))
		}
	} else if err != nil {
		return errors.New(fmt.Sprintf("Lỗi khi kiểm tra thư mục: %s", dirNameFormatted))
	}
	formattedPath := filepath.Join(dirNameFormatted, filepath.Base(todo.Url))

	resizedImage := imaging.Resize(srcImage, newWidth, newHeight, imaging.Lanczos)
	contrastedImage := imaging.AdjustContrast(resizedImage, 10)
	sharpenedImage := imaging.Sharpen(contrastedImage, 1.0)

	err = imaging.Save(sharpenedImage, formattedPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Image %s format error", todo.Url))
	}
	return nil
}

func checkResolutionImage(imagesError *[]string, imageDir string) {
	fileInfo, err := os.Stat(imageDir)
	if err != nil {
		fmt.Printf("Lỗi khi kiểm tra đường dẫn: %v\n", err)
		return
	}
	if !fileInfo.IsDir() {
		fileName := filepath.Base(imageDir)

		if isImage(fileName) {
			// Đọc tệp tin ảnh
			imgFile, err := os.Open(imageDir)
			if err != nil {
				log.Println("Failed to open image:", imageDir)
				return
			}
			defer imgFile.Close()

			// Giải mã thông tin ảnh
			imgConfig, _, err := image.DecodeConfig(imgFile)
			if err != nil {
				log.Println("Failed to decode image:", imageDir)
			}

			if imgConfig.Width != imgConfig.Height {
				return
			} else if imgConfig.Width < 600 || imgConfig.Height < 600 {
				return
			}

			*imagesError = append(*imagesError, imageDir)
		}
		return
	}
	files, err := os.ReadDir(imageDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			if file.Name() == "formatted" {
				continue
			}
			filePath := filepath.Join(imageDir, file.Name())
			checkResolutionImage(imagesError, filePath)
		}
		// Kiểm tra xem tệp tin có phải là ảnh hay không
		if isImageFile(file) {
			// Đường dẫn đầy đủ đến tệp tin
			filePath := filepath.Join(imageDir, file.Name())

			// Đọc tệp tin ảnh
			imgFile, err := os.Open(filePath)
			if err != nil {
				log.Println("Failed to open image:", filePath)
				continue
			}
			defer imgFile.Close()

			// Giải mã thông tin ảnh
			imgConfig, _, err := image.DecodeConfig(imgFile)
			if err != nil {
				log.Println("Failed to decode image:", filePath)
				continue
			}

			if imgConfig.Width != imgConfig.Height {
				continue
			} else if imgConfig.Width < 600 || imgConfig.Height < 600 {
				continue
			}

			*imagesError = append(*imagesError, filePath)
		}
	}
}

// Kiểm tra xem tệp tin có phải là ảnh hay không
func isImageFile(fileInfo os.DirEntry) bool {
	extension := filepath.Ext(fileInfo.Name())
	switch extension {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func isImage(filename string) bool {
	// Chuyển đổi phần mở rộng của tệp thành chữ thường để so sánh dễ dàng hơn.
	ext := strings.ToLower(filepath.Ext(filename))

	// Danh sách các phần mở rộng thường được sử dụng cho hình ảnh.
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}

	for _, imageExt := range imageExtensions {
		if ext == imageExt {
			return true
		}
	}

	return false
}

func combineImage(childImages []string, newtodoDescTxtHeight *widget.Entry, newtodoDescTxtWidth *widget.Entry, size int) error {
	newWidth := 1200
	newHeight := 1200
	var err error
	if newtodoDescTxtHeight.Text != "" {
		newHeight, err = strconv.Atoi(newtodoDescTxtHeight.Text)
		if err != nil {
			return errors.New("Size height invalid")
		}
	}
	if newtodoDescTxtWidth.Text != "" {
		newWidth, err = strconv.Atoi(newtodoDescTxtWidth.Text)
		if err != nil {
			return errors.New("Size width invalid")
		}
	}
	childWidth := int(newWidth)
	childHeight := int(newHeight)

	// Kích thước ảnh cha
	parentWidth := childWidth * int(math.Sqrt(float64(size)))
	parentHeight := childHeight * int(math.Sqrt(float64(size)))

	// Tạo 1 ảnh mới với kích thước của ảnh cha
	parentContext := gg.NewContext(int(parentWidth), int(parentHeight))

	for i, childImagePath := range childImages {
		img, err := gg.LoadImage(childImagePath)
		if err != nil {
			return errors.New("gg.LoadImage")
		}

		x := (i % int(math.Sqrt(float64(size)))) * childWidth
		y := (i / int(math.Sqrt(float64(size)))) * childHeight
		scaleX := float64(childWidth) / float64(img.Bounds().Dx())
		scaleY := float64(childHeight) / float64(img.Bounds().Dy())

		// Scale ảnh con xuống kích thước mong muốn và vẽ lên ảnh cha
		parentContext.Push()
		parentContext.Translate(float64(x), float64(y))
		parentContext.Scale(scaleX, scaleY)
		parentContext.DrawImage(img, 0, 0)
		parentContext.Pop()
	}

	dirName := filepath.Dir(childImages[0])
	fileNameCombined := filepath.Join(dirName, "combined.png")
	if err := parentContext.SavePNG(fileNameCombined); err != nil {
		return errors.New("parentContext.SavePNG")
	}
	return nil
}
func getFileNameWithoutExtension(path string) string {
	// Lấy tên file từ đường dẫn
	_, file := filepath.Split(path)

	// Lấy tên file (không bao gồm phần mở rộng)
	fileName := file[:len(file)-len(filepath.Ext(file))]
	return fileName
}

func combineImageWithLogo(imagePath string, logoPath string, borderSizeLogo int, scaleLogo float64, positionLeftLogo int) error {
	image1, err := gg.LoadImage(imagePath)
	if err != nil {
		panic(err)
	}

	// Đọc ảnh logo
	logo, err := gg.LoadImage(logoPath)
	if err != nil {
		return errors.New("gg.LoadImage")
	}
	// Tạo ảnh mới với kích thước sau khi cắt bỏ viền logo
	newLogoWidth := logo.Bounds().Dx() - 2*borderSizeLogo
	newLogoHeight := logo.Bounds().Dy() - 2*borderSizeLogo
	newLogoBase := gg.NewContext(newLogoWidth, newLogoHeight)
	newLogo := newLogoBase.Image()
	// Sao chép phần logo đã được cắt bỏ vào ảnh mới
	ctx := gg.NewContextForImage(newLogo)
	ctx.Scale(scaleLogo, scaleLogo)
	ctx.DrawImage(logo, -int(borderSizeLogo-positionLeftLogo), -int(borderSizeLogo))
	newLogo = ctx.Image()

	// Vị trí để đặt logo vào góc trái của ảnh gốc (image1)
	logoX := 0
	logoY := 0

	// Tạo context cho ảnh gốc mới (sau khi thêm logo)
	context := gg.NewContext(image1.Bounds().Dx(), image1.Bounds().Dy())

	// Vị trí để vẽ ảnh gốc mới
	x := 0
	y := 0

	// Vẽ ảnh gốc mới (ảnh sau khi cắt bỏ viền logo và thêm logo)
	context.DrawImage(image1, x, y)

	// Vẽ logo đã được cắt bỏ vào ảnh gốc mới
	context.DrawImage(newLogo, logoX, logoY)
	dirName := filepath.Dir(imagePath)
	dirNameFormatted := filepath.Join(dirName, "formatted_with_logo")
	_, err = os.Stat(dirNameFormatted)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dirNameFormatted, os.ModePerm)
		if err != nil {
			return errors.New(fmt.Sprintf("Lỗi khi tạo thư mục: %s", dirNameFormatted))
		}
	} else if err != nil {
		return errors.New(fmt.Sprintf("Lỗi khi kiểm tra thư mục: %s", dirNameFormatted))
	}
	formattedPath := filepath.Join(dirNameFormatted, fmt.Sprintf("%s.png", getFileNameWithoutExtension(filepath.Base(imagePath))))
	// Lưu ảnh gốc mới có logo vào tệp
	if err := context.SavePNG(formattedPath); err != nil {
		return errors.New("context.SavePNG")
	}
	return nil
}
