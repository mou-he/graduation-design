package image

import (
	"io"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/mou-he/graduation-design/common/image"
)

func RecognizeImage(file *multipart.FileHeader) (string, error) {

	modelPath := filepath.Join("common", "onnxmodel", "mobilenetv2-7.onnx")
	labelPath := filepath.Join("common", "onnxmodel", "imagenet_classes.txt")
	inputH, inputW := 224, 224

	recognizer, err := image.NewImageRecognizer(modelPath, labelPath, inputH, inputW)
	if err != nil {
		log.Println("NewImageRecognizer fail err is : ", err)
		return "", err
	}
	defer recognizer.Close()

	src, err := file.Open()
	if err != nil {
		log.Println("file open fail err is : ", err)
		return "", err
	}
	defer src.Close()

	buf, err := io.ReadAll(src)
	if err != nil {
		log.Println("io.ReadAll fail err is : ", err)
		return "", err
	}

	return recognizer.PredictFromBuffer(buf)
}
